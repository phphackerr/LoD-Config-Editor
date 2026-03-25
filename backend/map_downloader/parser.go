package map_downloader

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	epicwarBaseURL    = "https://www.epicwar.com"
	epicwarSearchURL  = epicwarBaseURL + "/maps/?sort=time&order=desc&a=Vordik"
	defaultMapVersion = "v1.0"
)

var (
	dateRegex            = regexp.MustCompile(`\d{1,2}\s+[A-Za-z]{3}\s+\d{4}`)
	versionRegex         = regexp.MustCompile(`(?i)\bv(\d+\.\d+[a-z]?)\b`)
	sizeRegex            = regexp.MustCompile(`\((\d+(?:\.\d+)?)\s*(B|KB|MB|GB)\)`)
	mapDownloadPathRegex = regexp.MustCompile(`/maps/(\d+)/download/?`)
	errNoMapsFound       = errors.New("не найдено ни одной карты")
)

type parsedMapInfo struct {
	Name         string
	Version      string
	DownloadLink string
	Date         string
	Size         int64
}

// parseMapInfos parses map metadata from EpicWar HTML.
func (md *MapDownloader) parseMapInfos(htmlContent string) ([]parsedMapInfo, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать goquery документ: %w", err)
	}

	results := make([]parsedMapInfo, 0, 8)
	seenMaps := make(map[string]struct{}, 8)

	doc.Find("a[href*='/download/']").Each(func(_ int, s *goquery.Selection) {
		downloadLink, ok := s.Attr("href")
		if !ok {
			return
		}

		normalizedLink, err := normalizeDownloadURL(downloadLink)
		if err != nil {
			return
		}

		mapID := extractMapID(normalizedLink)
		if mapID == "" {
			return
		}

		if _, exists := seenMaps[mapID]; exists {
			return
		}
		seenMaps[mapID] = struct{}{}

		record := findMapRecordContainer(s, mapID)
		if record.Length() == 0 {
			return
		}

		name := strings.TrimSpace(findMapNameLink(record, mapID).Text())
		if name == "" {
			name = strings.TrimSpace(s.Text())
		}
		if name == "" {
			return
		}

		date := extractMapDate(record)

		size := parseMapSize(record.Text())
		if size == 0 {
			size = parseMapSize(s.Parent().Text())
		}

		results = append(results, parsedMapInfo{
			Name:         name,
			Version:      extractVersion(name),
			DownloadLink: normalizedLink,
			Date:         date,
			Size:         size,
		})
	})

	if len(results) == 0 {
		return nil, fmt.Errorf("%w", errNoMapsFound)
	}

	return results, nil
}

func extractMapID(downloadURL string) string {
	matches := mapDownloadPathRegex.FindStringSubmatch(downloadURL)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func findMapRecordContainer(downloadLink *goquery.Selection, mapID string) *goquery.Selection {
	candidates := make([]*goquery.Selection, 0, 8)
	current := downloadLink
	for depth := 0; depth < 10; depth++ {
		current = current.Parent()
		if current.Length() == 0 {
			break
		}

		switch goquery.NodeName(current) {
		case "tr", "li", "article", "section", "div":
			candidates = append(candidates, current)
		}
	}

	if len(candidates) == 0 {
		return downloadLink
	}

	for _, candidate := range candidates {
		if findMapNameLink(candidate, mapID).Length() > 0 && dateRegex.FindString(candidate.Text()) != "" {
			return candidate
		}
	}
	for _, candidate := range candidates {
		if findMapNameLink(candidate, mapID).Length() > 0 {
			return candidate
		}
	}
	for _, candidate := range candidates {
		if dateRegex.FindString(candidate.Text()) != "" {
			return candidate
		}
	}

	return candidates[0]
}

func findMapNameLink(container *goquery.Selection, mapID string) *goquery.Selection {
	selector := "a[href^='/maps/']"
	if mapID != "" {
		selector = fmt.Sprintf("a[href*='/maps/%s/']", mapID)
	}

	return container.Find(selector).FilterFunction(func(_ int, sel *goquery.Selection) bool {
		href, _ := sel.Attr("href")
		if strings.Contains(href, "/download/") {
			return false
		}
		if strings.TrimSpace(sel.Text()) == "" {
			return false
		}
		return sel.Find("img").Length() == 0
	}).First()
}

func extractMapDate(record *goquery.Selection) string {
	date := strings.TrimSpace(record.Find("td:nth-child(4)").First().Text())
	if date != "" {
		return date
	}
	return dateRegex.FindString(record.Text())
}

func extractVersion(name string) string {
	matches := versionRegex.FindStringSubmatch(name)
	if len(matches) == 0 {
		return defaultMapVersion
	}
	return strings.ToLower(matches[0])
}

func parseMapSize(text string) int64 {
	matches := sizeRegex.FindStringSubmatch(text)
	if len(matches) < 3 {
		return 0
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	switch strings.ToUpper(matches[2]) {
	case "B":
		return int64(value)
	case "KB":
		return int64(value * 1024)
	case "MB":
		return int64(value * 1024 * 1024)
	case "GB":
		return int64(value * 1024 * 1024 * 1024)
	default:
		return 0
	}
}

func normalizeDownloadURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty download url")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	base, _ := url.Parse(epicwarBaseURL)
	if !u.IsAbs() {
		u = base.ResolveReference(u)
	}

	if scheme := strings.ToLower(u.Scheme); scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("unsupported download scheme: %s", u.Scheme)
	}

	host := strings.ToLower(u.Hostname())
	if host != "epicwar.com" && host != "www.epicwar.com" {
		return "", fmt.Errorf("unexpected download host: %s", u.Host)
	}

	return u.String(), nil
}

// FetchMapInfo — Wails-команда для получения информации о карте
func (md *MapDownloader) FetchMapInfo() (MapInfo, error) {
	log.Println("Получение информации о карте...")

	if err := md.initMapsDir(); err != nil {
		return MapInfo{}, fmt.Errorf("MapDownloader не инициализирован: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= fetchMapInfoAttempts; attempt++ {
		mapInfo, err := md.fetchMapInfoOnce()
		if err == nil {
			return mapInfo, nil
		}
		lastErr = err

		if attempt == fetchMapInfoAttempts || !shouldRetryFetchMapInfo(err) {
			break
		}

		delay := retryDelay(attempt)
		log.Printf("FetchMapInfo: ошибка попытки %d/%d: %v. Повтор через %s", attempt, fetchMapInfoAttempts, err, delay)
		time.Sleep(delay)
	}

	return MapInfo{}, lastErr
}

func shouldRetryFetchMapInfo(err error) bool {
	return errors.Is(err, errNoMapsFound) || isRetryableRequestError(err)
}

func (md *MapDownloader) fetchMapInfoOnce() (MapInfo, error) {
	resp, err := md.getWithTimeout(epicwarSearchURL, fetchMapInfoTimeout)
	if err != nil {
		return MapInfo{}, fmt.Errorf("ошибка HTTP-запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MapInfo{}, &httpStatusError{
			StatusCode: resp.StatusCode,
			Context:    "получен не-200 статус код",
		}
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return MapInfo{}, fmt.Errorf("не удалось прочитать тело ответа: %w", err)
	}

	mapInfos, err := md.parseMapInfos(string(htmlBytes))
	if err != nil {
		return MapInfo{}, fmt.Errorf("ошибка парсинга информации о картах: %w", err)
	}

	// Search endpoint is sorted by latest map, use first entry.
	foundMapInfo := mapInfos[0]

	savePath := md.mapsDir
	fileName := mapFileName(foundMapInfo.Name)
	filePath := filepath.Join(savePath, fileName)

	isDownloaded := false
	if _, err := os.Stat(filePath); err == nil {
		isDownloaded = true
	}

	return MapInfo{
		Name:         foundMapInfo.Name,
		Version:      foundMapInfo.Version,
		DownloadURL:  foundMapInfo.DownloadLink,
		Date:         foundMapInfo.Date,
		Size:         foundMapInfo.Size,
		SavePath:     savePath,
		IsDownloaded: isDownloaded,
	}, nil
}
