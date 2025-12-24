package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"lce/backend/app_settings"
)

// I18N handles internationalization
type I18N struct {
	cache map[string]map[string]string
	mu    sync.RWMutex
}

// NewI18N creates a new I18N instance
func NewI18N() *I18N {
	return &I18N{
		cache: make(map[string]map[string]string),
	}
}

// getLocalesPath returns the absolute path to the locales directory
func getLocalesPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	
	path := filepath.Join(filepath.Dir(ex), "locales")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	// Fallback to CWD for dev environment if needed, or just return error
	// For portable app, we prefer exe dir.
	
	return path, nil // Return path even if not exists, caller might handle or it might be created later
}

// GetLanguages returns available languages
func (i *I18N) GetLanguages() ([]map[string]string, error) {
	var langs []map[string]string
	localesDir, err := getLocalesPath()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(localesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if filepath.Ext(filename) != ".json" {
			continue
		}

		code := strings.TrimSuffix(filename, filepath.Ext(filename))
		path := filepath.Join(localesDir, filename)

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var jsonMap map[string]interface{}
		if err := json.Unmarshal(data, &jsonMap); err != nil {
			continue
		}

		name := code
		if langName, ok := jsonMap["lang_name"].(string); ok {
			name = langName
		}

		author := ""
		if auth, ok := jsonMap["author"].(string); ok {
			author = auth
		}

		langs = append(langs, map[string]string{
			"code":   code,
			"name":   name,
			"author": author,
		})
	}

	return langs, nil
}

// GetCurrentLanguage returns the current language from settings
func (i *I18N) GetCurrentLanguage() (string, error) {
	settings, err := app_settings.LoadSettings()
	if err != nil {
		return "en", nil // Return default if settings fail
	}
	return settings.Language, nil
}

// SwitchLanguage changes the language in settings
func (i *I18N) SwitchLanguage(newLang string) error {
	settings, err := app_settings.LoadSettings()
	if err != nil {
		return err
	}

	settings.Language = newLang
	return app_settings.SaveSettings(&settings)
}

// GetTranslationsCurrent returns translations for the current language
func (i *I18N) GetTranslationsCurrent() (map[string]string, error) {
	lang, err := i.GetCurrentLanguage()
	if err != nil {
		return nil, err
	}
	return i.GetTranslations(lang)
}

// GetTranslations returns translations for a specific language
func (i *I18N) GetTranslations(langCode string) (map[string]string, error) {
	// Check cache first
	i.mu.RLock()
	if val, ok := i.cache[langCode]; ok {
		i.mu.RUnlock()
		return val, nil
	}
	i.mu.RUnlock()

	localesDir, err := getLocalesPath()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(localesDir, langCode+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return make(map[string]string), nil // Return empty if not found
	}

	var jsonValue interface{}
	if err := json.Unmarshal(data, &jsonValue); err != nil {
		return nil, err
	}

	translations := make(map[string]string)
	flattenJSON("", jsonValue, translations)

	// Update cache
	i.mu.Lock()
	i.cache[langCode] = translations
	i.mu.Unlock()

	return translations, nil
}

// flattenJSON recursively flattens JSON with dot-notation
func flattenJSON(prefix string, value interface{}, result map[string]string) {
	if obj, ok := value.(map[string]interface{}); ok {
		for k, v := range obj {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			// Lowercase key for consistency if desired, or keep as is.
			// The original code lowercased keys, so we'll keep that behavior for the last segment?
			// Or better, lowercase the whole key to be case-insensitive?
			// Let's stick to the original logic: lowercasing keys.
			// newKey = strings.ToLower(newKey) // Removed to preserve case for nested keys like TITLE.window_title
			
			switch val := v.(type) {
			case map[string]interface{}:
				flattenJSON(newKey, val, result)
			case string:
				result[newKey] = val
			}
		}
	}
}
