package map_downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// parseMapInfos парсит информацию о картах из HTML-документа.
// Возвращает список кортежей: (name, version, download_link, date, size)
func (md *MapDownloader) parseMapInfos(htmlContent string) ([]struct {
	Header       string
	Name         string
	Version      string
	DownloadLink string
	Date         string
	Size         int64 // Используем int64
}, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать goquery документ: %w", err)
	}

	var results []struct {
		Header       string
		Name         string
		Version      string
		DownloadLink string
		Date         string
		Size         int64
	}

	// Стратегия: Ищем ссылки на скачивание.
	// URL имеет вид: /maps/12345/download/?token=...
	// Селектор [href*='/maps/download/'] ищет точное совпадение подстроки, которого нет.
	// Поэтому ищем просто по наличию "/download/"
	doc.Find("a[href*='/download/']").Each(func(i int, s *goquery.Selection) {
		downloadLink, _ := s.Attr("href")
		
		// Проверяем, что это действительно ссылка на карту (содержит /maps/)
		if !strings.Contains(downloadLink, "/maps/") {
			return
		}
		
		// Ищем контейнер. Поднимаемся вверх, пока не найдем контейнер, в котором есть ДРУГАЯ ссылка на /maps/ (имя карты)
		var container *goquery.Selection
		var nameLink *goquery.Selection

		// Пробуем 3 уровня вверх
		curr := s.Parent()
		for k := 0; k < 3; k++ {
			// Ищем ссылку на карту (не скачивание, не картинка)
			candidate := curr.Find("a[href^='/maps/']").FilterFunction(func(_ int, sel *goquery.Selection) bool {
				h, _ := sel.Attr("href")
				return !strings.Contains(h, "/download/") && strings.TrimSpace(sel.Text()) != "" && sel.Find("img").Length() == 0
			}).First()

			if candidate.Length() > 0 {
				container = curr
				nameLink = candidate
				break
			}
			curr = curr.Parent()
		}

		if container == nil || nameLink == nil {
			return // Не нашли контейнер или имя
		}

		name := strings.TrimSpace(nameLink.Text())

		// Дата обычно в соседней ячейке (следующей или через одну)
		// Если мы в td, то дата в td:nth-child(4) (по старой логике) или просто в следующей ячейке?
		// В новой верстке:
		// td[2] -> Name + Download
		// td[3] -> Category
		// td[4] -> Date
		// Попробуем найти строку (tr) и в ней дату
		row := container.Parent()
		date := ""
		if row.Is("tr") {
			date = strings.TrimSpace(row.Find("td:nth-child(4)").Text())
		}
		// Если дата пустая, пробуем найти просто текст похожий на дату в контейнере или рядом
		if date == "" {
			// Fallback: ищем текст даты (DD Mon YYYY)
			dateRegex := regexp.MustCompile(`\d{1,2}\s+[A-Za-z]{3}\s+\d{4}`)
			date = dateRegex.FindString(container.Parent().Text())
		}

		// Определяем версию
		version := "v1.0"
		versionRegex := regexp.MustCompile(`v(\d+\.\d+[a-z]?)`)
		if matches := versionRegex.FindStringSubmatch(name); len(matches) > 0 {
			version = matches[0]
		}

		// Размер. Он находится в текстовом узле рядом со ссылкой скачивания в том же контейнере.
		// Текст контейнера: "DotA ... (145.54 MB)"
		containerText := container.Text()
		sizeRegex := regexp.MustCompile(`\((\d+\.?\d*)\s*(B|KB|MB|GB)\)`)
		size := int64(0)
		sizeMatches := sizeRegex.FindStringSubmatch(containerText)
		if len(sizeMatches) > 2 {
			valueStr := sizeMatches[1]
			unit := sizeMatches[2]
			value, err := strconv.ParseFloat(valueStr, 64)
			if err == nil {
				var bytes float64
				switch unit {
				case "B":
					bytes = value
				case "KB":
					bytes = value * 1024.0
				case "MB":
					bytes = value * 1024.0 * 1024.0
				case "GB":
					bytes = value * 1024.0 * 1024.0 * 1024.0
				}
				size = int64(bytes)
			}
		}

		results = append(results, struct {
			Header       string
			Name         string
			Version      string
			DownloadLink string
			Date         string
			Size         int64
		}{
			Header:       name,
			Name:         name,
			Version:      version,
			DownloadLink: downloadLink,
			Date:         date,
			Size:         size,
		})
	})

	if len(results) == 0 {
		return nil, fmt.Errorf("не найдено ни одной карты")
	}

	return results, nil
}

// GetMapInfoCommand — Wails-команда для получения информации о карте
func (md *MapDownloader) GetMapInfo() (MapInfo, error) {
	log.Println("Получение информации о карте...")

	if err := md.initMapsDir(); err != nil {
		return MapInfo{}, fmt.Errorf("MapDownloader не инициализирован: %w", err)
	}

	// Новый URL поиска
	searchURL := "https://www.epicwar.com/maps/?sort=time&order=desc&a=Vordik"
	resp, err := md.client.Get(searchURL)
	if err != nil {
		return MapInfo{}, fmt.Errorf("ошибка HTTP-запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MapInfo{}, fmt.Errorf("получен не-200 статус код: %d", resp.StatusCode)
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return MapInfo{}, fmt.Errorf("не удалось прочитать тело ответа: %w", err)
	}

	mapInfos, err := md.parseMapInfos(string(htmlBytes))
	if err != nil {
		return MapInfo{}, fmt.Errorf("ошибка парсинга информации о картах: %w", err)
	}

	// Берем первую карту, так как сортировка по времени уже есть в URL
	foundMapInfo := mapInfos[0]

	savePath := md.mapsDir
	// Используем то же имя файла, что и в DownloadMap
	fileName := foundMapInfo.Name
	if !strings.HasSuffix(strings.ToLower(fileName), ".w3x") && !strings.HasSuffix(strings.ToLower(fileName), ".w3m") {
		fileName += ".w3x"
	}
	filePath := filepath.Join(savePath, fileName)

	isDownloaded := false
	if _, err := os.Stat(filePath); err == nil {
		isDownloaded = true
	}

	// Ссылка теперь абсолютная, не нужно добавлять домен
	downloadURL := foundMapInfo.DownloadLink
	if !strings.HasPrefix(downloadURL, "http") {
		downloadURL = "https://www.epicwar.com" + downloadURL
	}

	return MapInfo{
		Name:         foundMapInfo.Name,
		Version:      foundMapInfo.Version,
		DownloadURL:  downloadURL,
		Date:         foundMapInfo.Date,
		Size:         foundMapInfo.Size,
		SavePath:     savePath,
		IsDownloaded: isDownloaded,
	}, nil
}
