package theming

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ThemeService отвечает за загрузку/сохранение тем
type ThemeService struct{}

// NewThemeService создаёт сервис
func NewThemeService() *ThemeService {
	return &ThemeService{}
}

// путь к директории с темами
func (ts *ThemeService) getThemesDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("не удалось получить путь к исполняемому файлу: %w", err)
	}
	themesDir := filepath.Join(filepath.Dir(ex), "themes")

	_, err = os.ReadDir(themesDir)
	if err != nil {
		if os.IsNotExist(err) {
			// если папки нет — создаём
			if err := os.MkdirAll(themesDir, 0755); err != nil {
				return "", fmt.Errorf("не удалось создать директорию 'themes': %w", err)
			}
		} else {
			return "", fmt.Errorf("не удалось прочитать директорию 'themes': %w", err)
		}
	}

	return themesDir, nil
}

// LoadTheme загружает и разрешает все переменные
func (ts *ThemeService) LoadTheme(name string) (map[string]string, error) {
	themesDir, err := ts.getThemesDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(themesDir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать тему: %w", err)
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("не удалось разобрать JSON: %w", err)
	}

	resolved := make(map[string]string)
	for key := range raw {
		val, err := ts.resolveKey(raw, key, map[string]bool{})
		if err != nil {
			return nil, fmt.Errorf("ошибка при разрешении ключа %s: %w", key, err)
		}
		resolved[key] = val
	}

	// Проверка всех цветов
	for key, val := range resolved {
		if !ValidateColor(val) {
			return nil, fmt.Errorf("ключ '%s' содержит некорректный цвет: %s", key, val)
		}
	}

	return resolved, nil
}

// resolveKey рекурсивно подставляет значения ссылок (foreground: "background")
func (ts *ThemeService) resolveKey(theme map[string]string, key string, seen map[string]bool) (string, error) {
	if seen[key] {
		return "", fmt.Errorf("обнаружена циклическая ссылка в ключе '%s'", key)
	}
	seen[key] = true

	value := theme[key]
	if _, ok := theme[value]; ok {
		return ts.resolveKey(theme, value, seen)
	}
	return value, nil
}
