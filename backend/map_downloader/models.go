package map_downloader

import (
	"net/http"
	"sync"

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
	app        *application.App // Для отправки событий Wails
	client     *http.Client     // Для HTTP-запросов
	mapsDir    string           // Путь к папке для карт
	settingsMu sync.RWMutex     // Для защиты доступа к настройкам\
}

// NewMapDownloader создает новый экземпляр MapDownloader
func NewMapDownloader(app *application.App) *MapDownloader {
	return &MapDownloader{
		app:     app,
		client:  &http.Client{},
		mapsDir: "",
	}
}
