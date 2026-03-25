package app_settings

import (
	"encoding/json"
	"fmt"
	"lce/backend/fsutil"
	"lce/backend/runtimepaths"
	"os"
	"path/filepath"
	"sync"
)

var settingsFileMu sync.Mutex

func getSettingsPath() (string, error) {
	appConfigDir, err := runtimepaths.AppDataDir()
	if err != nil {
		return "", fmt.Errorf("не удалось получить директорию конфигурации приложения: %w", err)
	}

	return filepath.Join(appConfigDir, "settings.json"), nil
}

// LoadSettings loads settings from disk or creates defaults on first run.
func LoadSettings() (Settings, error) {
	settingsFileMu.Lock()
	defer settingsFileMu.Unlock()
	return loadSettingsLocked()
}

// SaveSettings validates and saves settings to disk atomically.
func SaveSettings(newSettings *Settings) error {
	if newSettings == nil {
		return fmt.Errorf("settings must not be nil")
	}

	settingsFileMu.Lock()
	defer settingsFileMu.Unlock()

	normalized, _ := normalizeSettings(cloneSettings(*newSettings))
	return saveSettingsLocked(normalized)
}

func loadSettingsLocked() (Settings, error) {
	settingsPath, err := getSettingsPath()
	if err != nil {
		return DefaultSettings(), err
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			defaults := DefaultSettings()
			if err := saveSettingsLocked(defaults); err != nil {
				return defaults, fmt.Errorf("не удалось сохранить настройки по умолчанию: %w", err)
			}
			return defaults, nil
		}
		return DefaultSettings(), fmt.Errorf("не удалось прочитать файл настроек: %w", err)
	}

	settings := DefaultSettings()
	if err := json.Unmarshal(data, &settings); err != nil {
		defaults := DefaultSettings()
		if saveErr := saveSettingsLocked(defaults); saveErr != nil {
			return defaults, fmt.Errorf("ошибка десериализации: %w (доп. ошибка сохранения defaults: %v)", err, saveErr)
		}
		return defaults, fmt.Errorf("ошибка десериализации: %w", err)
	}

	normalized, changed := normalizeSettings(settings)
	if changed {
		if err := saveSettingsLocked(normalized); err != nil {
			return normalized, fmt.Errorf("не удалось сохранить нормализованные настройки: %w", err)
		}
	}

	return cloneSettings(normalized), nil
}

func saveSettingsLocked(settings Settings) error {
	settingsPath, err := getSettingsPath()
	if err != nil {
		return err
	}

	jsonBytes, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		return fmt.Errorf("не удалось сериализовать настройки в JSON: %w", err)
	}

	if err := fsutil.WriteFileAtomic(settingsPath, jsonBytes, 0644); err != nil {
		return fmt.Errorf("не удалось записать файл настроек: %w", err)
	}
	return nil
}
