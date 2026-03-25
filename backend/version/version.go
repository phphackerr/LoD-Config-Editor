package version

import (
	"embed"
	"encoding/json"
	"lce/backend/versioncmp"
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

// Init initializes version metadata and writable theme/locale assets in AppData.
func Init(manifestData []byte, themesFS, localesFS embed.FS, appDataDir, exeDir string) {
	manifestPath := filepath.Join(appDataDir, "manifest.json")

	// Ensure AppData directory exists
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		log.Printf("failed to create app data dir: %v", err)
	}

	var m Manifest
	var embeddedManifest Manifest
	forceUpdate := false
	forceThemesUpdate := false
	forceLocalesUpdate := false

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
			embeddedVer := embeddedManifest.AppVersion.Version
			localVer := m.AppVersion.Version

			if versioncmp.Compare(embeddedVer, localVer) > 0 {
				log.Printf("Embedded version (%s) is newer than local (%s). Updating local manifest and assets.", embeddedVer, localVer)
				m = embeddedManifest
				forceUpdate = true
			} else {
				// App version is unchanged, but component versions may still be newer.
				forceThemesUpdate = hasNewerComponentVersions(embeddedManifest.Themes, m.Themes)
				forceLocalesUpdate = hasNewerComponentVersions(embeddedManifest.Locales, m.Locales)
				if forceThemesUpdate {
					log.Printf("Embedded themes are newer than local manifest. Updating bundled themes.")
					m.Themes = embeddedManifest.Themes
				}
				if forceLocalesUpdate {
					log.Printf("Embedded locales are newer than local manifest. Updating bundled locales.")
					m.Locales = embeddedManifest.Locales
				}
			}
		}
	} else {
		// No local file (or error), use embedded
		m = embeddedManifest
		forceUpdate = true
	}

	themesDir := filepath.Join(appDataDir, "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		log.Printf("failed to create themes dir: %v", err)
	}

	copyMissingFromLegacyDir(filepath.Join(exeDir, "themes"), themesDir)

	// Unpack default themes
	entries, _ := themesFS.ReadDir("themes")
	for _, entry := range entries {
		targetPath := filepath.Join(themesDir, entry.Name())
		_, err := os.Stat(targetPath)
		if os.IsNotExist(err) || forceUpdate || forceThemesUpdate {
			// File doesn't exist or forced update, unpack it
			data, _ := themesFS.ReadFile("themes/" + entry.Name())
			if err := os.WriteFile(targetPath, data, 0644); err != nil {
				log.Printf("failed to unpack theme %s: %v", entry.Name(), err)
			}
		}
	}

	localesDir := filepath.Join(appDataDir, "locales")
	if err := os.MkdirAll(localesDir, 0755); err != nil {
		log.Printf("failed to create locales dir: %v", err)
	}

	copyMissingFromLegacyDir(filepath.Join(exeDir, "locales"), localesDir)

	// Unpack default locales
	entries, _ = localesFS.ReadDir("locales")
	for _, entry := range entries {
		targetPath := filepath.Join(localesDir, entry.Name())
		_, err := os.Stat(targetPath)
		if os.IsNotExist(err) || forceUpdate || forceLocalesUpdate {
			// File doesn't exist or forced update, unpack it
			data, _ := localesFS.ReadFile("locales/" + entry.Name())
			if err := os.WriteFile(targetPath, data, 0644); err != nil {
				log.Printf("failed to unpack locale %s: %v", entry.Name(), err)
			}
		}
	}

	// Save the updated manifest to disk if needed
	if forceUpdate || forceThemesUpdate || forceLocalesUpdate {
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

func copyMissingFromLegacyDir(srcDir, dstDir string) {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if _, err := os.Stat(dstPath); err == nil {
			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			continue
		}

		if err := os.WriteFile(dstPath, data, 0644); err != nil {
			log.Printf("failed to migrate %s: %v", entry.Name(), err)
		}
	}
}

func hasNewerComponentVersions(embedded, local map[string]ComponentInfo) bool {
	if len(embedded) == 0 {
		return false
	}

	for key, emb := range embedded {
		loc, ok := local[key]
		if !ok {
			return true
		}

		if versioncmp.Compare(emb.Version, loc.Version) > 0 {
			return true
		}
	}

	return false
}
