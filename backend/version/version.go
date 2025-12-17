package version

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Manifest struct {
	AppVersion string            `json:"app_version"`
	Themes     map[string]string `json:"themes"`
	Locales    map[string]string `json:"locales"`
}

var (
	// AppVersion - текущая версия приложения
	AppVersion string

	// Themes - версии тем
	Themes map[string]string

	// Locales - версии языков
	Locales map[string]string
)

// Init initializes the version package.
// It tries to read manifest.json from appDataDir.
// If not found, it writes the embedded manifestData to that location.
// It then loads the versions into memory.
func Init(manifestData []byte, appDataDir string) {
	manifestPath := filepath.Join(appDataDir, "manifest.json")

	// Ensure directory exists
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		log.Printf("failed to create app data dir: %v", err)
	}

	var m Manifest

	// Try to read local file
	data, err := os.ReadFile(manifestPath)
	if err == nil {
		// Found local file, parse it
		if err := json.Unmarshal(data, &m); err != nil {
			log.Printf("failed to parse local manifest: %v, falling back to embedded", err)
			parseEmbedded(manifestData, &m)
		}
	} else {
		// No local file (or error), use embedded and write it to disk
		parseEmbedded(manifestData, &m)
		if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
			log.Printf("failed to write bootstrap manifest: %v", err)
		}
	}

	AppVersion = m.AppVersion
	Themes = m.Themes
	Locales = m.Locales
}

func parseEmbedded(data []byte, m *Manifest) {
	if err := json.Unmarshal(data, m); err != nil {
		log.Fatalf("failed to parse embedded manifest: %v", err)
	}
}
