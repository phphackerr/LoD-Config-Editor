package map_downloader

import (
	"context"
	"errors"
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

	downloadCtx, err := md.beginDownloadControl()
	if err != nil {
		return MapMetadata{}, err
	}
	defer md.endDownloadControl()

	currentInfo := mapInfo
	var lastErr error

	for attempt := 1; attempt <= downloadMapAttempts; attempt++ {
		filePath, partPath := md.resolveDownloadTarget(currentInfo)
		_ = os.Remove(partPath)

		result, err := md.downloadMapOnce(downloadCtx, currentInfo, filePath, partPath)
		if err == nil {
			return result, nil
		}
		if isStoppedError(err) {
			return MapMetadata{}, err
		}
		lastErr = err

		_ = os.Remove(partPath)

		if attempt == downloadMapAttempts || !isRetryableRequestError(err) {
			return MapMetadata{}, err
		}

		refreshedInfo, refreshErr := md.tryRefreshDownloadInfo(currentInfo)
		if refreshErr == nil {
			if refreshedInfo.DownloadURL != "" && refreshedInfo.DownloadURL != currentInfo.DownloadURL {
				currentInfo = refreshedInfo
				log.Printf("DownloadMap: получена свежая ссылка, повтор без задержки")
				continue
			}
			currentInfo = refreshedInfo
		} else {
			log.Printf("DownloadMap: не удалось обновить ссылку на карту перед повтором: %v", refreshErr)
		}

		delay := retryDelay(attempt)
		log.Printf("DownloadMap: ошибка попытки %d/%d: %v. Повтор через %s", attempt, downloadMapAttempts, err, delay)
		select {
		case <-downloadCtx.Done():
			if md.isStopRequested() {
				return MapMetadata{}, errDownloadStopped
			}
		case <-time.After(delay):
		}
	}

	return MapMetadata{}, lastErr
}

func (md *MapDownloader) resolveDownloadTarget(mapInfo MapInfo) (filePath string, partPath string) {
	fileName := mapFileName(mapInfo.Name)
	filePath = filepath.Join(md.mapsDir, fileName)
	return filePath, filePath + ".part"
}

func (md *MapDownloader) tryRefreshDownloadInfo(currentInfo MapInfo) (MapInfo, error) {
	freshInfo, err := md.fetchMapInfoOnce()
	if err != nil {
		return currentInfo, err
	}

	if freshInfo.DownloadURL == "" {
		return currentInfo, fmt.Errorf("обновление ссылки на карту вернуло пустой URL")
	}

	if freshInfo.Name == "" {
		freshInfo.Name = currentInfo.Name
	}
	if freshInfo.Version == "" {
		freshInfo.Version = currentInfo.Version
	}
	if freshInfo.Size == 0 {
		freshInfo.Size = currentInfo.Size
	}

	return freshInfo, nil
}

func readResponseSnippet(body io.Reader, maxBytes int64) string {
	if body == nil || maxBytes <= 0 {
		return ""
	}

	data, err := io.ReadAll(io.LimitReader(body, maxBytes))
	if err != nil || len(data) == 0 {
		return ""
	}

	flattened := strings.Join(strings.Fields(string(data)), " ")
	if len(flattened) > 180 {
		return flattened[:180] + "..."
	}
	return flattened
}

func (md *MapDownloader) downloadMapOnce(downloadCtx context.Context, mapInfo MapInfo, filePath string, partPath string) (MapMetadata, error) {
	resp, err := md.getWithContextTimeout(downloadCtx, mapInfo.DownloadURL, downloadMapTimeout)
	if err != nil {
		return MapMetadata{}, stopAwareWrap("ошибка HTTP-запроса при загрузке", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet := readResponseSnippet(resp.Body, 512)
		contextMessage := "получен не-200 статус код при загрузке"
		if snippet != "" {
			contextMessage = fmt.Sprintf("%s (%s)", contextMessage, snippet)
		}
		return MapMetadata{}, &httpStatusError{
			StatusCode: resp.StatusCode,
			Context:    contextMessage,
		}
	}

	file, err := os.Create(partPath)
	if err != nil {
		return MapMetadata{}, fmt.Errorf("не удалось создать файл карты: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	totalSize := mapInfo.Size
	if totalSize == 0 {
		// Попытка получить Content-Length, если размер не был в mapInfo
		if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
			if s, parseErr := strconv.ParseInt(contentLength, 10, 64); parseErr == nil {
				totalSize = s
			}
		}
	}

	var downloadedBytes int64
	var lastUpdate time.Time = time.Now()
	var lastBytes int64

	buffer := make([]byte, 32*1024) // Буфер 32KB

	for {
		if err := md.waitWhilePaused(); err != nil {
			_ = os.Remove(partPath)
			return MapMetadata{}, err
		}

		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, writeErr := file.Write(buffer[:n])
			if writeErr != nil {
				_ = os.Remove(partPath)
				return MapMetadata{}, fmt.Errorf("ошибка записи в файл: %w", writeErr)
			}
			downloadedBytes += int64(n)
		}

		if err == io.EOF {
			break // Конец файла
		}
		if err != nil {
			_ = os.Remove(partPath)
			if errors.Is(err, context.Canceled) || md.isStopRequested() {
				return MapMetadata{}, errDownloadStopped
			}
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
				md.emitDownloadProgress(DownloadProgressEvent{
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

	if err := file.Sync(); err != nil {
		_ = os.Remove(partPath)
		return MapMetadata{}, fmt.Errorf("не удалось синхронизировать файл карты: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(partPath)
		return MapMetadata{}, fmt.Errorf("не удалось закрыть временный файл карты: %w", err)
	}

	if totalSize > 0 {
		md.emitDownloadProgress(DownloadProgressEvent{
			Progress:   100.0,
			Downloaded: downloadedBytes,
			Total:      totalSize,
			Speed:      0,
		})
	}

	if err := os.Rename(partPath, filePath); err != nil {
		_ = os.Remove(partPath)
		return MapMetadata{}, fmt.Errorf("не удалось переименовать временный файл карты: %w", err)
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
