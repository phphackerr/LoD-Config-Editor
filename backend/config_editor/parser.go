package config_editor

import (
	"log"

	"gopkg.in/ini.v1"
)

type GameConfig struct {
	file *ini.File
	path string
}

// Загрузка INI с сохранением структуры и комментариев
func (c *GameConfig) Load(path string) error {
	cfg, err := ini.LoadSources(ini.LoadOptions{
		PreserveSurroundedQuote:  true, // не трогать кавычки, если появятся
		SpaceBeforeInlineComment: true, // сохранить inline-комментарии
		AllowBooleanKeys:         true,
		// Insensitive:              true,
	}, path)
	if err != nil {
		return err
	}
	c.file = cfg
	c.path = path

	ini.DefaultSection = ""

	return nil
}

// Получить значение
func (c *GameConfig) Get(section, key string) string {
	if c.file == nil {
		return ""
	}

	sec, err := c.file.GetSection(section)
	if err != nil {
		return "not found section"
	}

	k := sec.Key(key)
	if k == nil {
		return "not found key"
	}

	return k.String() // если значения нет → вернётся ""
}

// Обновить значение
func (c *GameConfig) Set(section, key, value string) {
	if c.file == nil {
		return
	}
	c.file.Section(section).Key(key).SetValue(value)
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
