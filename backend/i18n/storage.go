package i18n

import (
	"encoding/json"
	"fmt"
	"lce/backend/fsutil"
	"lce/backend/runtimepaths"
	"os"
	"path/filepath"
	"strings"
)

func getLocalesPath() (string, error) {
	path, err := runtimepaths.LocalesDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve locales path: %w", err)
	}
	return path, nil
}

func validateLocaleCode(code string) error {
	return fsutil.ValidateSimpleName(strings.TrimSpace(code))
}

func readLocaleEntries(localesDir string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(localesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read locales directory: %w", err)
	}

	filtered := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		filtered = append(filtered, entry)
	}

	return filtered, nil
}

func readLocaleMetadata(localesDir string, entry os.DirEntry) (map[string]string, error) {
	filename := entry.Name()
	code := strings.TrimSuffix(filename, filepath.Ext(filename))
	if err := validateLocaleCode(code); err != nil {
		return nil, err
	}

	path := filepath.Join(localesDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	name := code
	if v, ok := raw["lang_name"].(string); ok && strings.TrimSpace(v) != "" {
		name = v
	}

	author := ""
	if v, ok := raw["author"].(string); ok {
		author = v
	}

	return map[string]string{
		"code":   code,
		"name":   name,
		"author": author,
	}, nil
}

func localePathForCode(localesDir, code string) (string, error) {
	if err := validateLocaleCode(code); err != nil {
		return "", err
	}
	return filepath.Join(localesDir, code+".json"), nil
}

func localeExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func readLocaleFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func readLocaleObject(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func writeLocaleJSON(path string, value interface{}) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	return fsutil.WriteFileAtomic(path, content, 0644)
}
