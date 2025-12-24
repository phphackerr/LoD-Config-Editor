package version

import (
	"embed"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type MultiLineString string

func (m *MultiLineString) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*m = MultiLineString(single)
		return nil
	}

	var multi []string
	if err := json.Unmarshal(data, &multi); err == nil {
		*m = MultiLineString(strings.Join(multi, "\n"))
		return nil
	}

	return nil
}

type ComponentInfo struct {
	Version   string          `json:"version"`
	Changelog MultiLineString `json:"changelog,omitempty"`
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

	var m Manifest
	var embeddedManifest Manifest
	forceUpdate := false

	// Always parse embedded manifest first to get the binary's version
	parseEmbedded(manifestData, &embeddedManifest)

	// Try to read local file
	data, err := os.ReadFile(manifestPath)
	if err == nil {
		// Found local file, parse it
		if err := json.Unmarshal(data, &m); err != nil {
			log.Printf("failed to parse local manifest: %v, falling back to embedded", err)
			m = embeddedManifest
			forceUpdate = true
		} else {
			// Check if embedded version is newer than local version
			embeddedVer := strings.TrimPrefix(embeddedManifest.AppVersion.Version, "v")
			localVer := strings.TrimPrefix(m.AppVersion.Version, "v")

			if embeddedVer > localVer {
				log.Printf("Embedded version (%s) is newer than local (%s). Updating local manifest and assets.", embeddedVer, localVer)
				m = embeddedManifest
				forceUpdate = true
			}
		}
	} else {
		// No local file (or error), use embedded
		m = embeddedManifest
		forceUpdate = true
	}

	// Bootstrap Themes (in ExeDir)
	themesDir := filepath.Join(exeDir, "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		log.Printf("failed to create themes dir: %v", err)
	}

	// Unpack default themes
	entries, _ := themesFS.ReadDir("themes")
	for _, entry := range entries {
		targetPath := filepath.Join(themesDir, entry.Name())
		_, err := os.Stat(targetPath)
		if os.IsNotExist(err) || forceUpdate {
			// File doesn't exist or forced update, unpack it
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

	// Unpack default locales
	entries, _ = localesFS.ReadDir("locales")
	for _, entry := range entries {
		targetPath := filepath.Join(localesDir, entry.Name())
		_, err := os.Stat(targetPath)
		if os.IsNotExist(err) || forceUpdate {
			// File doesn't exist or forced update, unpack it
			data, _ := localesFS.ReadFile("locales/" + entry.Name())
			if err := os.WriteFile(targetPath, data, 0644); err != nil {
				log.Printf("failed to unpack locale %s: %v", entry.Name(), err)
			}
		}
	}

	// Save the updated manifest to disk if needed
	if forceUpdate {
		if newManifestData, err := json.MarshalIndent(m, "", "  "); err == nil {
			if err := os.WriteFile(manifestPath, newManifestData, 0644); err != nil {
				log.Printf("failed to update local manifest: %v", err)
			}
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
