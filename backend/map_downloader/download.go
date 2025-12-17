package map_downloader

import (
	"fmt"
	"io"
	"lce/backend/app_settings"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// initMapsDir инициализирует mapsDir на основе настроек приложения
func (md *MapDownloader) initMapsDir() error {
	md.settingsMu.Lock()
	defer md.settingsMu.Unlock()

	// md.mapsDir кэшировался, но это мешает смене пути.
	// Убираем проверку, чтобы всегда брать актуальный путь из настроек.
	// if md.mapsDir != "" {
	// 	return nil
	// }

	settings, err := app_settings.LoadSettings()
	if err != nil {
		return fmt.Errorf("не удалось загрузить настройки: %w", err)
	}

	if settings.GamePath == "" {
		return fmt.Errorf("укажите путь к Warcraft III в настройках")
	}

	mapsPath := filepath.Join(settings.GamePath, "maps", "download")
	if err := os.MkdirAll(mapsPath, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию для карт: %w", err)
	}

	md.mapsDir = mapsPath
	return nil
}

// DownloadMapCommand — Wails-команда для загрузки карты
func (md *MapDownloader) DownloadMap(mapInfo MapInfo) (MapMetadata, error) {
	log.Printf("Загрузка карты %s...", mapInfo.Name)

	if err := md.initMapsDir(); err != nil {
		return MapMetadata{}, fmt.Errorf("MapDownloader не инициализирован: %w", err)
	}

	fileName := mapInfo.Name
	if !strings.HasSuffix(strings.ToLower(fileName), ".w3x") && !strings.HasSuffix(strings.ToLower(fileName), ".w3m") {
		fileName += ".w3x"
	}
	filePath := filepath.Join(md.mapsDir, fileName)

	resp, err := md.client.Get(mapInfo.DownloadURL)
	if err != nil {
		return MapMetadata{}, fmt.Errorf("ошибка HTTP-запроса при загрузке: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MapMetadata{}, fmt.Errorf("получен не-200 статус код при загрузке: %d", resp.StatusCode)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return MapMetadata{}, fmt.Errorf("не удалось создать файл карты: %w", err)
	}
	defer file.Close()

	totalSize := mapInfo.Size
	if totalSize == 0 {
		// Попытка получить Content-Length, если размер не был в mapInfo
		if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
			if s, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
				totalSize = s
			}
		}
	}

	var downloadedBytes int64
	var lastUpdate time.Time = time.Now()
	var lastBytes int64

	buffer := make([]byte, 32*1024) // Буфер 32KB

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := file.Write(buffer[:n])
			if writeErr != nil {
				return MapMetadata{}, fmt.Errorf("ошибка записи в файл: %w", writeErr)
			}
			downloadedBytes += int64(n)
		}

		if err == io.EOF {
			break // Конец файла
		}
		if err != nil {
			return MapMetadata{}, fmt.Errorf("ошибка чтения из ответа: %w", err)
		}

		if totalSize > 0 {
			now := time.Now()
			elapsed := now.Sub(lastUpdate).Seconds()

			// Обновляем скорость каждые 0.5 секунды
			if elapsed >= 0.5 {
				speed := float64(downloadedBytes-lastBytes) / elapsed
				progress := (float64(downloadedBytes) / float64(totalSize)) * 100.0

				// Отправляем прогресс во фронтенд
				md.app.Event.Emit("download-progress", DownloadProgressEvent{
					Progress:   progress,
					Downloaded: downloadedBytes,
					Total:      totalSize,
					Speed:      speed,
				})

				lastUpdate = now
				lastBytes = downloadedBytes
			}
		}
	}

	log.Println("Карта успешно загружена")

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return MapMetadata{}, fmt.Errorf("не удалось получить информацию о файле: %w", err)
	}

	return MapMetadata{
		Name:         mapInfo.Name,
		Version:      mapInfo.Version,
		Size:         fileInfo.Size(),
		IsDownloaded: true,
	}, nil
}
