package config_editor

import (
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

// Получить значение
func (e *ConfigEditor) GetConfigValue(section, option string) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.config == nil || e.config.file == nil {
		return "config not loaded"
	}
	return e.config.Get(section, option)
}

// Установить значение
func (e *ConfigEditor) SetConfigValue(section, option, value string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config.Set(section, option, value)
	return e.config.Save()
}

// Получить значение как Hotkey
func (e *ConfigEditor) GetHotkeyValue(section, option string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	rawValue := e.config.Get(section, option)
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
