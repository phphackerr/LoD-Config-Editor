package version

import (
	"embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type ComponentInfo struct {
	Version   string `json:"version"`
	Changelog string `json:"changelog,omitempty"`
}

type Manifest struct {
	AppVersion ComponentInfo            `json:"app_version"`
	Themes     map[string]ComponentInfo `json:"themes"`
	Locales    map[string]ComponentInfo `json:"locales"`
}

var (
	// App - текущая версия приложения и список изменений
	App ComponentInfo

	// Themes - версии тем
	Themes map[string]ComponentInfo

	// Locales - версии языков
	Locales map[string]ComponentInfo
)

// Init initializes the version package.
func Init(manifestData []byte, themesFS, localesFS embed.FS, appDataDir, exeDir string) {
	manifestPath := filepath.Join(appDataDir, "manifest.json")

	// Ensure AppData directory exists
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		log.Printf("failed to create app data dir: %v", err)
	}

	// Bootstrap Themes (in ExeDir)
	themesDir := filepath.Join(exeDir, "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		log.Printf("failed to create themes dir: %v", err)
	}
	
	// Always check and unpack default themes if missing
	entries, _ := themesFS.ReadDir("themes")
	for _, entry := range entries {
		targetPath := filepath.Join(themesDir, entry.Name())
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			// File doesn't exist, unpack it
			data, _ := themesFS.ReadFile("themes/" + entry.Name())
			if err := os.WriteFile(targetPath, data, 0644); err != nil {
				log.Printf("failed to unpack theme %s: %v", entry.Name(), err)
			}
		}
	}

	// Bootstrap Locales (in ExeDir)
	localesDir := filepath.Join(exeDir, "locales")
	if err := os.MkdirAll(localesDir, 0755); err != nil {
		log.Printf("failed to create locales dir: %v", err)
	}

	// Always check and unpack default locales if missing
	entries, _ = localesFS.ReadDir("locales")
	for _, entry := range entries {
		targetPath := filepath.Join(localesDir, entry.Name())
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			// File doesn't exist, unpack it
			data, _ := localesFS.ReadFile("locales/" + entry.Name())
			if err := os.WriteFile(targetPath, data, 0644); err != nil {
				log.Printf("failed to unpack locale %s: %v", entry.Name(), err)
			}
		}
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

	App = m.AppVersion
	Themes = m.Themes
	Locales = m.Locales
}

func parseEmbedded(data []byte, m *Manifest) {
	if err := json.Unmarshal(data, m); err != nil {
		log.Fatalf("failed to parse embedded manifest: %v", err)
	}
}
