package map_downloader

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// GetChangelogCommand — Wails-команда для получения списка изменений
func (md *MapDownloader) GetChangelog(version string) (string, error) {
	re := regexp.MustCompile(`[^\d]`)
	versionNumber := re.ReplaceAllString(version, "")

	url := fmt.Sprintf("https://d1stats.ru/lod-%s-changelog/?lang=en", versionNumber)
	resp, err := md.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("ошибка HTTP-запроса к changelog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("получен не-200 статус код для changelog: %d", resp.StatusCode)
	}

	htmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать тело ответа changelog: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return "", fmt.Errorf("не удалось создать goquery документ для changelog: %w", err)
	}

	selection := doc.Find("div.text").First()
	if selection.Length() == 0 {
		return "", fmt.Errorf("не найден div с классом text в changelog")
	}

	changelogHTML, err := selection.Html()
	if err != nil {
		return "", fmt.Errorf("не удалось получить HTML changelog: %w", err)
	}

	return changelogHTML, nil
}
