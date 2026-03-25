package i18n

import (
	"encoding/json"
	"fmt"
	"lce/backend/app_settings"
	"sort"
	"sync"
)

// I18N handles internationalization.
type I18N struct {
	cache map[string]map[string]string
	mu    sync.RWMutex
}

// NewI18N creates a new I18N instance.
func NewI18N() *I18N {
	return &I18N{
		cache: make(map[string]map[string]string),
	}
}

// GetLanguages returns available languages.
func (i *I18N) GetLanguages() ([]map[string]string, error) {
	localesDir, err := getLocalesPath()
	if err != nil {
		return nil, err
	}

	entries, err := readLocaleEntries(localesDir)
	if err != nil {
		return nil, err
	}

	langs := make([]map[string]string, 0, len(entries))
	for _, entry := range entries {
		metadata, err := readLocaleMetadata(localesDir, entry)
		if err != nil {
			continue
		}
		langs = append(langs, metadata)
	}

	sort.Slice(langs, func(a, b int) bool {
		return langs[a]["code"] < langs[b]["code"]
	})

	return langs, nil
}

// GetCurrentLanguage returns the current language from settings.
func (i *I18N) GetCurrentLanguage() (string, error) {
	settings, err := app_settings.LoadSettings()
	if err != nil {
		return app_settings.DefaultSettings().Language, nil
	}
	return settings.Language, nil
}

// SwitchLanguage changes language in settings.
func (i *I18N) SwitchLanguage(newLang string) error {
	if err := validateLocaleCode(newLang); err != nil {
		return err
	}

	settings, err := app_settings.LoadSettings()
	if err != nil {
		return err
	}

	if settings.Language == newLang {
		return nil
	}

	settings.Language = newLang
	return app_settings.SaveSettings(&settings)
}

// GetTranslationsCurrent returns translations for current language.
func (i *I18N) GetTranslationsCurrent() (map[string]string, error) {
	lang, err := i.GetCurrentLanguage()
	if err != nil {
		return nil, err
	}
	return i.GetTranslations(lang)
}

// GetTranslations returns translations for a specific language.
func (i *I18N) GetTranslations(langCode string) (map[string]string, error) {
	if err := validateLocaleCode(langCode); err != nil {
		return nil, err
	}

	i.mu.RLock()
	if cached, ok := i.cache[langCode]; ok {
		copy := cloneStringMap(cached)
		i.mu.RUnlock()
		return copy, nil
	}
	i.mu.RUnlock()

	localesDir, err := getLocalesPath()
	if err != nil {
		return nil, err
	}

	path, err := localePathForCode(localesDir, langCode)
	if err != nil {
		return nil, err
	}

	data, err := readLocaleFile(path)
	if err != nil {
		return make(map[string]string), nil
	}

	var jsonValue interface{}
	if err := json.Unmarshal(data, &jsonValue); err != nil {
		return nil, err
	}

	translations := make(map[string]string)
	flattenJSON("", jsonValue, translations)

	i.mu.Lock()
	i.cache[langCode] = cloneStringMap(translations)
	i.mu.Unlock()

	return translations, nil
}

// CreateLanguage creates a new language file.
func (i *I18N) CreateLanguage(code, name, author string) error {
	if err := validateLocaleCode(code); err != nil {
		return err
	}

	localesDir, err := getLocalesPath()
	if err != nil {
		return err
	}

	path, err := localePathForCode(localesDir, code)
	if err != nil {
		return err
	}

	exists, err := localeExists(path)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("language %s already exists", code)
	}

	data := map[string]string{
		"lang_name": name,
		"author":    author,
	}

	return writeLocaleJSON(path, data)
}

// SaveTranslation saves translations for a language.
func (i *I18N) SaveTranslation(langCode string, translations map[string]string) error {
	if err := validateLocaleCode(langCode); err != nil {
		return err
	}

	localesDir, err := getLocalesPath()
	if err != nil {
		return err
	}

	path, err := localePathForCode(localesDir, langCode)
	if err != nil {
		return err
	}

	existingData, _ := readLocaleObject(path)
	nested := unflattenJSON(translations)

	if val, ok := existingData["lang_name"]; ok {
		if _, exists := nested["lang_name"]; !exists {
			nested["lang_name"] = val
		}
	}
	if val, ok := existingData["author"]; ok {
		if _, exists := nested["author"]; !exists {
			nested["author"] = val
		}
	}

	if err := writeLocaleJSON(path, nested); err != nil {
		return err
	}

	i.mu.Lock()
	i.cache[langCode] = cloneStringMap(translations)
	i.mu.Unlock()

	return nil
}

func cloneStringMap(input map[string]string) map[string]string {
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}
