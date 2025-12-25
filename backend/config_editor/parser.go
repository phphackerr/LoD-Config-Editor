package config_editor

import (
	"log"
	"strings"

	"gopkg.in/ini.v1"
)

type GameConfig struct {
	file   *ini.File
	path   string
	keyMap map[string]map[string]string // section -> key -> originalKey
}

// Загрузка INI с сохранением структуры и комментариев
func (c *GameConfig) Load(path string) error {
	// Устанавливаем имя дефолтной секции в пустую строку ПЕРЕД загрузкой,
	// чтобы ключи без секции попадали в секцию "" (а не "DEFAULT").
	// При сохранении, если имя секции совпадает с DefaultSection, заголовок не пишется.
	ini.DefaultSection = ""

	cfg, err := ini.LoadSources(ini.LoadOptions{
		PreserveSurroundedQuote:  true, // не трогать кавычки, если появятся
		SpaceBeforeInlineComment: true, // сохранить inline-комментарии
		AllowBooleanKeys:         true,
		// Insensitive:              true, // Мы реализуем свою нечувствительность
	}, path)
	if err != nil {
		return err
	}
	c.file = cfg
	c.path = path
	c.keyMap = make(map[string]map[string]string)

	// Строим карту ключей
	for _, section := range cfg.Sections() {
		secNameLower := strings.ToLower(section.Name())
		c.keyMap[secNameLower] = make(map[string]string)
		for _, key := range section.Keys() {
			c.keyMap[secNameLower][strings.ToLower(key.Name())] = key.Name()
		}
	}

	return nil
}

// Получить значение
func (c *GameConfig) Get(section, key string) string {
	if c.file == nil {
		return ""
	}

	// 1. Пытаемся найти точное совпадение
	sec, err := c.file.GetSection(section)
	if err == nil && sec.HasKey(key) {
		return sec.Key(key).String()
	}

	// 2. Если не нашли, ищем через карту (case-insensitive)
	secNameLower := strings.ToLower(section)
	keyLower := strings.ToLower(key)

	// Ищем реальное имя секции
	var realSectionName string
	
	// Сначала проверяем, может секция с таким именем есть (но ключ не нашли выше)
	if s, err := c.file.GetSection(section); err == nil {
		realSectionName = s.Name()
	} else {
		// Если нет, ищем перебором (так как keyMap хранит только lower case ключи секций)
		for _, s := range c.file.Sections() {
			if strings.EqualFold(s.Name(), section) {
				realSectionName = s.Name()
				break
			}
		}
	}

	if realSectionName == "" {
		return "not found section"
	}

	// Теперь ищем ключ в этой секции
	sec, _ = c.file.GetSection(realSectionName)
	
	// Проверяем маппинг ключа
	if mapping, ok := c.keyMap[secNameLower]; ok {
		if realKey, ok := mapping[keyLower]; ok {
			return sec.Key(realKey).String()
		}
	}
	
	// Если в мапе нет, но вдруг он есть в файле (добавили динамически?)
	if sec.HasKey(key) {
		return sec.Key(key).String()
	}

	return "not found key"
}

// Обновить значение
func (c *GameConfig) Set(section, key, value string) {
	if c.file == nil {
		return
	}

	secNameLower := strings.ToLower(section)
	keyLower := strings.ToLower(key)

	// 1. Определяем реальное имя секции
	var realSection *ini.Section
	if s, err := c.file.GetSection(section); err == nil {
		realSection = s
	} else {
		for _, s := range c.file.Sections() {
			if strings.EqualFold(s.Name(), section) {
				realSection = s
				break
			}
		}
	}

	// Если секции нет - создаем (с тем именем, которое передали)
	if realSection == nil {
		realSection, _ = c.file.NewSection(section)
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
}

// Сохранить обратно в файл
func (c *GameConfig) Save() error {
	if c.file == nil || c.path == "" {
		log.Println("⚠ Save skipped: file or path is nil")
		return nil
	}
	log.Println("💾 Saving INI to:", c.path)

	ini.PrettyFormat = true // Выравнивание знака '='
	ini.PrettyEqual = false // Пробелы вокруг '='

	return c.file.SaveTo(c.path)
}

func (c *GameConfig) Path() string {
	return c.path
}
