package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var allowedUpdateHosts = map[string]struct{}{
	"api.github.com":                {},
	"github.com":                    {},
	"www.github.com":                {},
	"objects.githubusercontent.com": {},
	"raw.githubusercontent.com":     {},
}

func (u *Updater) fetchJSON(ctx context.Context, rawURL string, out interface{}) error {
	if err := validateRemoteURL(rawURL); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "lce-updater")

	resp, err := u.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(out); err != nil {
		return err
	}

	return nil
}

func (u *Updater) fetchBytes(ctx context.Context, rawURL string, maxBytes int64) ([]byte, error) {
	if err := validateRemoteURL(rawURL); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "lce-updater")

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	if maxBytes > 0 && resp.ContentLength > maxBytes {
		return nil, fmt.Errorf("response too large: %d bytes (max %d)", resp.ContentLength, maxBytes)
	}

	reader := io.LimitReader(resp.Body, maxBytes+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("response exceeded max size %d", maxBytes)
	}
	return data, nil
}

func (u *Updater) downloadToFile(ctx context.Context, rawURL, finalPath string, maxBytes int64, onProgress func(float64)) (int64, error) {
	if err := validateRemoteURL(rawURL); err != nil {
		return 0, err
	}

	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return 0, err
	}

	tempPath := finalPath + ".part"
	_ = os.Remove(tempPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "lce-updater")

	resp, err := u.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("download failed with status: %s", resp.Status)
	}
	if maxBytes > 0 && resp.ContentLength > maxBytes {
		return 0, fmt.Errorf("asset too large: %d bytes (max %d)", resp.ContentLength, maxBytes)
	}

	out, err := os.Create(tempPath)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = out.Close()
	}()

	pr := &ProgressReader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
		OnProgress: func(percent float64) {
			if onProgress != nil {
				onProgress(percent)
			}
		},
		MaxBytes: maxBytes,
	}

	written, err := io.Copy(out, pr)
	if err != nil {
		_ = os.Remove(tempPath)
		return 0, err
	}

	if err := out.Sync(); err != nil {
		_ = os.Remove(tempPath)
		return 0, err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tempPath)
		return 0, err
	}

	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return 0, err
	}

	return written, nil
}

func validateRemoteURL(rawURL string) error {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}

	if u.Scheme != "https" {
		return fmt.Errorf("only https URLs are allowed: %s", rawURL)
	}

	host := strings.ToLower(u.Hostname())
	if _, ok := allowedUpdateHosts[host]; !ok {
		return fmt.Errorf("remote host is not allowed: %s", host)
	}

	return nil
}

// ProgressReader tracks reading progress and optionally enforces max bytes.
type ProgressReader struct {
	io.Reader
	Total      int64
	Current    int64
	OnProgress func(float64)
	MaxBytes   int64
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)

	if pr.MaxBytes > 0 && pr.Current > pr.MaxBytes {
		return n, fmt.Errorf("download exceeded max bytes: %d", pr.MaxBytes)
	}

	if pr.Total > 0 && pr.OnProgress != nil {
		percent := float64(pr.Current) / float64(pr.Total) * 100
		if percent > 100 {
			percent = 100
		}
		pr.OnProgress(percent)
	}

	return n, err
}
