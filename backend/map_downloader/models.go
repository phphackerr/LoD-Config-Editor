package map_downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// MapInfo соответствует MapInfo из Rust
type MapInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	DownloadURL  string `json:"download_url"`
	Date         string `json:"date"`
	Size         int64  `json:"size"` // int64 для размера
	SavePath     string `json:"save_path"`
	IsDownloaded bool   `json:"is_downloaded"`
}

// MapMetadata соответствует MapMetadata из Rust
type MapMetadata struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Size         int64  `json:"size"`
	IsDownloaded bool   `json:"is_downloaded"`
}

// DownloadProgressEvent используется для отправки прогресса загрузки во фронтенд
type DownloadProgressEvent struct {
	Progress   float64 `json:"progress"`
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Speed      float64 `json:"speed"`
}

// MapDownloader соответствует MapDownloader из Rust
type MapDownloader struct {
	app          *application.App // Для отправки событий Wails
	client       *http.Client     // Для HTTP-запросов
	mapsDir      string           // Путь к папке для карт
	settingsMu   sync.RWMutex     // Для защиты доступа к настройкам
	downloadMu   sync.Mutex
	downloadCond *sync.Cond
	downloadCtrl downloadControlState
}

const (
	fetchMapInfoTimeout = 30 * time.Second
	changelogTimeout    = 30 * time.Second
	downloadMapTimeout  = 4 * time.Hour

	fetchMapInfoAttempts = 3
	changelogAttempts    = 2
	downloadMapAttempts  = 3

	retryBaseDelay = 500 * time.Millisecond
	retryMaxDelay  = 5 * time.Second
)

const (
	defaultRequestUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36"
	defaultRequestAccept    = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
	defaultRequestLanguage  = "en-US,en;q=0.9"
)

// NewMapDownloader создает новый экземпляр MapDownloader
func NewMapDownloader(app *application.App) *MapDownloader {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = 45 * time.Second
	transport.TLSHandshakeTimeout = 15 * time.Second
	transport.ExpectContinueTimeout = 1 * time.Second
	transport.IdleConnTimeout = 90 * time.Second
	transport.MaxIdleConns = 20
	transport.MaxIdleConnsPerHost = 10

	md := &MapDownloader{
		app:     app,
		client:  &http.Client{Transport: transport},
		mapsDir: "",
	}
	md.downloadCond = sync.NewCond(&md.downloadMu)
	return md
}

func (md *MapDownloader) getWithTimeout(rawURL string, timeout time.Duration) (*http.Response, error) {
	return md.getWithContextTimeout(context.Background(), rawURL, timeout)
}

func (md *MapDownloader) getWithContextTimeout(parent context.Context, rawURL string, timeout time.Duration) (*http.Response, error) {
	if parent == nil {
		parent = context.Background()
	}

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, timeout)
	} else {
		ctx, cancel = context.WithCancel(parent)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	setDefaultRequestHeaders(req)

	resp, err := md.client.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}

	resp.Body = &cancelOnCloseReadCloser{
		ReadCloser: resp.Body,
		cancel:     cancel,
	}

	return resp, nil
}

type cancelOnCloseReadCloser struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnCloseReadCloser) Close() error {
	err := c.ReadCloser.Close()
	c.cancel()
	return err
}

type httpStatusError struct {
	StatusCode int
	Context    string
}

func (e *httpStatusError) Error() string {
	if e == nil {
		return "unexpected http status"
	}

	if e.Context == "" {
		return fmt.Sprintf("unexpected http status: %d", e.StatusCode)
	}

	return fmt.Sprintf("%s: %d", e.Context, e.StatusCode)
}

func setDefaultRequestHeaders(req *http.Request) {
	if req == nil {
		return
	}

	req.Header.Set("User-Agent", defaultRequestUserAgent)
	req.Header.Set("Accept", defaultRequestAccept)
	req.Header.Set("Accept-Language", defaultRequestLanguage)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	if strings.Contains(strings.ToLower(req.URL.Hostname()), "epicwar.com") {
		req.Header.Set("Referer", epicwarBaseURL+"/")
	}
}

func retryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return retryBaseDelay
	}

	delay := retryBaseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay >= retryMaxDelay {
			return retryMaxDelay
		}
	}
	return delay
}

func isRetryableRequestError(err error) bool {
	if err == nil {
		return false
	}

	var statusErr *httpStatusError
	if errors.As(err, &statusErr) {
		switch statusErr.StatusCode {
		case http.StatusRequestTimeout,
			http.StatusTooManyRequests,
			http.StatusForbidden,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout:
			return true
		default:
			return statusErr.StatusCode >= 500
		}
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "timeout") ||
		strings.Contains(message, "connection reset") ||
		strings.Contains(message, "connection aborted") ||
		strings.Contains(message, "broken pipe") ||
		strings.Contains(message, "unexpected eof")
}

func (md *MapDownloader) emitDownloadProgress(progress DownloadProgressEvent) {
	if md == nil || md.app == nil {
		return
	}
	md.app.Event.Emit("download-progress", progress)
}
