package map_downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// GetChangelogCommand — Wails-команда для получения списка изменений
func (md *MapDownloader) GetChangelog(version string) (string, error) {
	re := regexp.MustCompile(`[^\d]`)
	versionNumber := re.ReplaceAllString(version, "")

	url := fmt.Sprintf("https://d1stats.ru/lod-%s-changelog/?lang=en", versionNumber)

	var lastErr error
	for attempt := 1; attempt <= changelogAttempts; attempt++ {
		result, err := md.getChangelogOnce(url)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if attempt == changelogAttempts || !isRetryableRequestError(err) {
			break
		}

		delay := retryDelay(attempt)
		log.Printf("GetChangelog: ошибка попытки %d/%d: %v. Повтор через %s", attempt, changelogAttempts, err, delay)
		time.Sleep(delay)
	}

	return "", lastErr
}

func (md *MapDownloader) getChangelogOnce(url string) (string, error) {
	resp, err := md.getWithTimeout(url, changelogTimeout)
	if err != nil {
		return "", fmt.Errorf("ошибка HTTP-запроса к changelog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", &httpStatusError{
			StatusCode: resp.StatusCode,
			Context:    "получен не-200 статус код для changelog",
		}
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
