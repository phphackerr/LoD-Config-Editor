package config_editor

import (
	"errors"
	"log"
	"strings"
	"sync"
)

type ConfigEditor struct {
	mu     sync.Mutex
	config *GameConfig
}

func NewConfigEditor() *ConfigEditor {
	return &ConfigEditor{
		config: &GameConfig{},
	}
}

func (c *GameConfig) ToMap() map[string]map[string]string {
	result := make(map[string]map[string]string)
	for _, section := range c.file.Sections() {
		secMap := make(map[string]string)
		for _, key := range section.Keys() {
			secMap[key.Name()] = key.Value()
		}
		result[section.Name()] = secMap
	}
	return result
}

// Загрузить конфиг
func (e *ConfigEditor) LoadConfig(configPath string) (map[string]map[string]string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if err := e.config.Load(configPath); err != nil {
		e.config = &GameConfig{}
		return nil, err
	}
	return e.config.ToMap(), nil
}

// Проверить наличие
func (e *ConfigEditor) IsConfigAvailable() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.config != nil && e.config.file != nil
}

// GetConfigValueStrict returns config value with strict error semantics.
func (e *ConfigEditor) GetConfigValueStrict(section, option string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.config == nil {
		return "", ErrConfigNotLoaded
	}

	return e.config.Get(section, option)
}

// Получить значение
func (e *ConfigEditor) GetConfigValue(section, option string) string {
	value, err := e.GetConfigValueStrict(section, option)
	if err != nil {
		if !errors.Is(err, ErrSectionNotFound) && !errors.Is(err, ErrKeyNotFound) {
			log.Printf("GetConfigValue error for [%s] %s: %v", section, option, err)
		}
		return ""
	}
	return value
}

// Установить значение
func (e *ConfigEditor) SetConfigValue(section, option, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.config.Set(section, option, value); err != nil {
		return err
	}
	return e.config.Save()
}

// Получить значение как Hotkey
func (e *ConfigEditor) GetHotkeyValue(section, option string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config == nil || e.config.file == nil {
		return "", ErrConfigNotLoaded
	}

	rawValue, err := e.config.Get(section, option)
	if err != nil {
		if errors.Is(err, ErrSectionNotFound) || errors.Is(err, ErrKeyNotFound) {
			return "", nil
		}
		return "", err
	}

	if rawValue == "" {
		return "", nil
	}

	// Если ctrl / shift / alt
	if rawValue == "ctrl" || rawValue == "shift" || rawValue == "alt" {
		return string(rawValue[0]-32) + rawValue[1:], nil // Ctrl / Shift / Alt
	}

	// Если начинается с 0x
	if len(rawValue) > 2 && rawValue[:2] == "0x" {
		keyName := Lookup(rawValue)
		if keyName == "" {
			return rawValue, nil
		}
		// F-клавиши
		if keyName[0] == 'f' && len(keyName) > 1 {
			return "F" + keyName[1:], nil
		}
		// Одиночные буквы
		if len(keyName) == 1 {
			return strings.ToUpper(keyName), nil
		}
		return keyName, nil
	}

	return rawValue, nil
}
