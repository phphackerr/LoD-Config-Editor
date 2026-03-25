package config_editor

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"gopkg.in/ini.v1"
)

type GameConfig struct {
	file   *ini.File
	path   string
	keyMap map[string]map[string]string // section -> key -> originalKey
}

var iniGlobalsMu sync.Mutex

// Загрузка INI с сохранением структуры и комментариев
func (c *GameConfig) Load(path string) error {
	cfg, err := loadINI(path)
	if err != nil {
		c.file = nil
		c.path = ""
		c.keyMap = nil
		return err
	}
	c.file = cfg
	c.path = path
	c.rebuildKeyMap()

	return nil
}

// Получить значение
func (c *GameConfig) Get(section, key string) (string, error) {
	if c.file == nil {
		return "", ErrConfigNotLoaded
	}

	sec, secName, err := c.resolveSection(section)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrSectionNotFound, section)
	}

	// 1. Пытаемся найти точное совпадение
	if sec.HasKey(key) {
		return sec.Key(key).String(), nil
	}

	// 2. Если не нашли, ищем через карту (case-insensitive)
	secNameLower := strings.ToLower(secName)
	keyLower := strings.ToLower(key)

	if mapping, ok := c.keyMap[secNameLower]; ok {
		if realKey, exists := mapping[keyLower]; exists && sec.HasKey(realKey) {
			return sec.Key(realKey).String(), nil
		}
	}

	// Если в мапе нет, но вдруг ключ есть в секции с другим регистром.
	for _, k := range sec.Keys() {
		if strings.EqualFold(k.Name(), key) {
			return k.String(), nil
		}
	}

	return "", fmt.Errorf("%w: [%s] %s", ErrKeyNotFound, secName, key)
}

// Обновить значение
func (c *GameConfig) Set(section, key, value string) error {
	if c.file == nil {
		return ErrConfigNotLoaded
	}

	secNameLower := strings.ToLower(section)
	keyLower := strings.ToLower(key)

	// 1. Определяем реальное имя секции
	realSection, _, _ := c.resolveSection(section)

	// Если секции нет - создаем (с тем именем, которое передали)
	if realSection == nil {
		var err error
		realSection, err = c.file.NewSection(section)
		if err != nil {
			return fmt.Errorf("failed to create section %s: %w", section, err)
		}
		// Обновляем мапу
		if c.keyMap == nil {
			c.keyMap = make(map[string]map[string]string)
		}
		c.keyMap[secNameLower] = make(map[string]string)
	}

	// 2. Определяем реальное имя ключа
	realKeyName := key // По умолчанию - как передали

	if mapping, ok := c.keyMap[secNameLower]; ok {
		if existingKey, ok := mapping[keyLower]; ok {
			realKeyName = existingKey
		}
	}

	// 3. Устанавливаем значение
	realSection.Key(realKeyName).SetValue(value)

	// 4. Обновляем мапу (на случай если это новый ключ)
	if c.keyMap[secNameLower] == nil {
		c.keyMap[secNameLower] = make(map[string]string)
	}
	c.keyMap[secNameLower][keyLower] = realKeyName

	return nil
}

// Сохранить обратно в файл
func (c *GameConfig) Save() error {
	if c.file == nil || c.path == "" {
		return ErrConfigNotLoaded
	}
	log.Println("💾 Saving INI to:", c.path)

	return saveINI(c.file, c.path)
}

func (c *GameConfig) resolveSection(section string) (*ini.Section, string, error) {
	if s, err := c.file.GetSection(section); err == nil {
		return s, s.Name(), nil
	}

	for _, s := range c.file.Sections() {
		if strings.EqualFold(s.Name(), section) {
			return s, s.Name(), nil
		}
	}

	return nil, "", fmt.Errorf("%w: %s", ErrSectionNotFound, section)
}

func loadINI(path string) (*ini.File, error) {
	iniGlobalsMu.Lock()
	defer iniGlobalsMu.Unlock()

	origDefaultSection := ini.DefaultSection
	ini.DefaultSection = ""
	defer func() {
		ini.DefaultSection = origDefaultSection
	}()

	return ini.LoadSources(ini.LoadOptions{
		PreserveSurroundedQuote:  true,
		SpaceBeforeInlineComment: true,
		AllowBooleanKeys:         true,
	}, path)
}

func saveINI(file *ini.File, path string) error {
	iniGlobalsMu.Lock()
	defer iniGlobalsMu.Unlock()

	origDefaultSection := ini.DefaultSection
	origPrettyFormat := ini.PrettyFormat
	origPrettyEqual := ini.PrettyEqual

	ini.DefaultSection = ""
	ini.PrettyFormat = true
	ini.PrettyEqual = false

	defer func() {
		ini.DefaultSection = origDefaultSection
		ini.PrettyFormat = origPrettyFormat
		ini.PrettyEqual = origPrettyEqual
	}()

	return file.SaveTo(path)
}

func (c *GameConfig) Path() string {
	return c.path
}

func (c *GameConfig) rebuildKeyMap() {
	c.keyMap = make(map[string]map[string]string)
	if c.file == nil {
		return
	}

	for _, section := range c.file.Sections() {
		secNameLower := strings.ToLower(section.Name())
		c.keyMap[secNameLower] = make(map[string]string)
		for _, key := range section.Keys() {
			c.keyMap[secNameLower][strings.ToLower(key.Name())] = key.Name()
		}
	}
}
