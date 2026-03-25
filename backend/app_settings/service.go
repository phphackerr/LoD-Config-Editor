package app_settings

import (
	"log"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppSettings is a Wails service exposing settings APIs to frontend.
type AppSettings struct {
	app      *application.App
	settings Settings
	lock     sync.RWMutex
}

// NewAppSettings creates an AppSettings service with loaded settings.
func NewAppSettings(app *application.App) *AppSettings {
	s, err := LoadSettings()
	if err != nil {
		log.Printf("Ошибка при загрузке настроек: %v. Используются настройки по умолчанию.", err)
		s = DefaultSettings()
	}

	return &AppSettings{
		app:      app,
		settings: s,
	}
}

// GetDefaultSettings returns the default settings.
func (a *AppSettings) GetDefaultSettings() Settings {
	return DefaultSettings()
}

// GetSettings returns current settings.
func (a *AppSettings) GetSettings() Settings {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return cloneSettings(a.settings)
}

// UpdateSettings validates and persists settings.
func (a *AppSettings) UpdateSettings(newSettings Settings) (Settings, error) {
	normalized, _ := normalizeSettings(cloneSettings(newSettings))

	if err := SaveSettings(&normalized); err != nil {
		log.Printf("Ошибка сохранения настроек: %v", err)
		a.lock.RLock()
		defer a.lock.RUnlock()
		return cloneSettings(a.settings), err
	}

	a.lock.Lock()
	a.settings = normalized
	current := cloneSettings(a.settings)
	a.lock.Unlock()

	if a.app != nil {
		a.app.Event.Emit("app-settings:updated", current)
	}

	return current, nil
}

// GetOption returns option value by key.
func (a *AppSettings) GetOption(key string) interface{} {
	a.lock.RLock()
	defer a.lock.RUnlock()

	switch key {
	case "width":
		return a.settings.Width
	case "height":
		return a.settings.Height
	case "language":
		return a.settings.Language
	case "game_path":
		return a.settings.GamePath
	case "first_run":
		return a.settings.FirstRun
	case "all_paths":
		return append([]string(nil), a.settings.AllPaths...)
	case "theme":
		return a.settings.Theme
	case "windowed_mode":
		return a.settings.WindowedMode
	default:
		return nil
	}
}
