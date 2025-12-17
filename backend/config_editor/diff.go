package config_editor

import (
	"fmt"
)

// CheckConfigDiff сравнивает текущий конфиг в памяти с тем, что на диске,
// возвращает map[section]map[key]value только с изменёнными значениями
func (e *ConfigEditor) CheckConfigDiff() (map[string]map[string]map[string]string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config == nil || e.config.Path() == "" {
		return nil, fmt.Errorf("config not loaded")
	}

	// Загружаем версию с диска
	diskCfg := &GameConfig{}
	if err := diskCfg.Load(e.config.Path()); err != nil {
		return nil, fmt.Errorf("failed to load config from disk: %w", err)
	}

	diff := make(map[string]map[string]map[string]string)

	// --- Проверяем новые и изменённые секции/ключи ---
	for _, section := range diskCfg.file.Sections() {
		secName := section.Name()
		memSection := e.config.file.Section(secName)

		if memSection == nil || memSection.KeysHash() == nil {
			// Секция полностью новая
			diff[secName] = make(map[string]map[string]string)
			diff[secName]["<section>"] = map[string]string{"old": "", "new": "<added>", "status": "added"}
			for _, key := range section.Keys() {
				diff[secName][key.Name()] = map[string]string{
					"old":    "",
					"new":    key.Value(),
					"status": "added",
				}
			}
			continue
		}

		// Проверяем ключи
		for _, key := range section.Keys() {
			diskVal := key.Value()
			if memSection.HasKey(key.Name()) {
				memVal := memSection.Key(key.Name()).String()
				if diskVal != memVal {
					if diff[secName] == nil {
						diff[secName] = make(map[string]map[string]string)
					}
					diff[secName][key.Name()] = map[string]string{
						"old":    memVal,
						"new":    diskVal,
						"status": "modified",
					}
				}
			} else {
				if diff[secName] == nil {
					diff[secName] = make(map[string]map[string]string)
				}
				diff[secName][key.Name()] = map[string]string{
					"old":    "",
					"new":    diskVal,
					"status": "added",
				}
			}
		}
	}

	// --- Проверяем удалённые секции/ключи ---
	for _, section := range e.config.file.Sections() {
		secName := section.Name()
		diskSection := diskCfg.file.Section(secName)
		if diskSection == nil || diskSection.KeysHash() == nil {
			// Секция полностью удалена
			diff[secName] = make(map[string]map[string]string)
			diff[secName]["<section>"] = map[string]string{
				"old":    "<deleted>",
				"new":    "",
				"status": "deleted",
			}
			continue
		}

		for _, key := range section.Keys() {
			if !diskSection.HasKey(key.Name()) {
				if diff[secName] == nil {
					diff[secName] = make(map[string]map[string]string)
				}
				diff[secName][key.Name()] = map[string]string{
					"old":    key.Value(),
					"new":    "",
					"status": "deleted",
				}
			}
		}
	}

	return diff, nil
}

// ApplyChangesToConfig — применяет изменения из diff к текущему конфигу в памяти и сохраняет его.
func (e *ConfigEditor) ApplyChangesToConfig(diff map[string]map[string]map[string]string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config == nil || e.config.Path() == "" {
		return fmt.Errorf("config not loaded")
	}

	for secName, keys := range diff {
		section, err := e.config.file.GetSection(secName)
		if err != nil && err.Error() == "section not found" && (keys["<section>"] == nil || keys["<section>"]["status"] != "deleted") {
			// Если секция не найдена, но не помечена как удаленная, создаем ее
			section, _ = e.config.file.NewSection(secName)
		} else if err != nil && err.Error() != "section not found" {
			return fmt.Errorf("failed to get section %s: %w", secName, err)
		}

		for keyName, info := range keys {
			if keyName == "<section>" { // Пропускаем метаинформацию о секции
				if info["status"] == "deleted" {
					e.config.file.DeleteSection(secName)
				}
				continue
			}

			switch info["status"] {
			case "added", "modified":
				if section != nil {
					section.Key(keyName).SetValue(info["new"])
				}
			case "deleted":
				if section != nil {
					section.DeleteKey(keyName)
				}
			}
		}
	}

	return e.config.Save()
}

// DiscardExternalChanges — перезаписывает конфиг на диске текущим состоянием из памяти.
// Это отменяет любые внешние изменения, применяя нашу in-memory версию.
func (e *ConfigEditor) DiscardExternalChanges() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.config == nil || e.config.Path() == "" {
		return fmt.Errorf("config not loaded")
	}

	// Просто сохраняем текущее состояние из памяти на диск
	err := e.config.Save()
	if err != nil {
		return fmt.Errorf("не удалось перезаписать конфиг на диске: %w", err)
	}
	return nil
}
