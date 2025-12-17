package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"lce/backend/version"
	"net/http"
	"os"
	"path/filepath"
)

const (
	ManifestURL = "https://raw.githubusercontent.com/phphackerr/LoD-Config-Editor/master/manifest.json"
	BaseRepoURL = "https://raw.githubusercontent.com/phphackerr/LoD-Config-Editor/master"
)

// ComponentVersions matches the structure of manifest.json
type ComponentVersions struct {
	AppVersion version.ComponentInfo        `json:"app_version"`
	Themes     map[string]version.ComponentInfo `json:"themes"`
	Locales    map[string]version.ComponentInfo `json:"locales"`
}

type ComponentUpdate struct {
	Type      string `json:"type"` // "theme" or "locale"
	Name      string `json:"name"`
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
}

// getVersionsPath returns the path to manifest.json in AppData/LCE
func getVersionsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	appConfigDir := filepath.Join(configDir, "LCE")
	if _, err := os.Stat(appConfigDir); os.IsNotExist(err) {
		if err := os.MkdirAll(appConfigDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create app config dir: %w", err)
		}
	}

	return filepath.Join(appConfigDir, "manifest.json"), nil
}

// LoadLocalVersions loads versions from versions.json or initializes with defaults
func (u *Updater) LoadLocalVersions() (ComponentVersions, error) {
	path, err := getVersionsPath()
	if err != nil {
		return ComponentVersions{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize with defaults from compiled binary (embedded manifest)
			defaults := ComponentVersions{
				AppVersion: version.App,
				Themes:     version.Themes,
				Locales:    version.Locales,
			}
			u.SaveLocalVersions(defaults)
			return defaults, nil
		}
		return ComponentVersions{}, err
	}

	var versions ComponentVersions
	if err := json.Unmarshal(data, &versions); err != nil {
		return ComponentVersions{}, err
	}

	// Ensure maps are not nil
	if versions.Themes == nil {
		versions.Themes = make(map[string]version.ComponentInfo)
	}
	if versions.Locales == nil {
		versions.Locales = make(map[string]version.ComponentInfo)
	}

	return versions, nil
}

// SaveLocalVersions saves versions to versions.json
func (u *Updater) SaveLocalVersions(versions ComponentVersions) error {
	path, err := getVersionsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// CheckForComponentUpdates fetches manifest and compares with local versions
func (u *Updater) CheckForComponentUpdates() ([]ComponentUpdate, error) {
	// 1. Load local versions
	local, err := u.LoadLocalVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to load local versions: %w", err)
	}

	// 2. Fetch remote manifest
	resp, err := http.Get(ManifestURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest fetch returned status: %s", resp.Status)
	}

	var remote ComponentVersions
	if err := json.NewDecoder(resp.Body).Decode(&remote); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	var updates []ComponentUpdate

	// 3. Compare Themes
	for name, remoteInfo := range remote.Themes {
		localInfo, exists := local.Themes[name]
		// If new theme or newer version
		if !exists || remoteInfo.Version > localInfo.Version {
			updates = append(updates, ComponentUpdate{
				Type:      "theme",
				Name:      name,
				Version:   remoteInfo.Version,
				Changelog: remoteInfo.Changelog,
			})
		}
	}

	// 4. Compare Locales
	for name, remoteInfo := range remote.Locales {
		localInfo, exists := local.Locales[name]
		if !exists || remoteInfo.Version > localInfo.Version {
			updates = append(updates, ComponentUpdate{
				Type:      "locale",
				Name:      name,
				Version:   remoteInfo.Version,
				Changelog: remoteInfo.Changelog,
			})
		}
	}

	return updates, nil
}

// UpdateComponent downloads the component file and updates versions.json
func (u *Updater) UpdateComponent(update ComponentUpdate) error {
	// 1. Determine URL and target path
	var url, filename string
	
	// We need to know where to save files. 
	// Themes are in "themes/" relative to executable? Or AppData?
	// Locales are in "locales/"?
	// Usually in development they are in "frontend/src/..." or "backend/..." but in production
	// they might be embedded or in a specific folder.
	// Wails apps often embed assets. If we want to support dynamic updates, 
	// the app needs to look for assets in a local directory FIRST, then embedded.
	
	// Let's assume we save them to AppData/LCE/themes and AppData/LCE/locales
	// and the app logic (theming/i18n) needs to support loading from there.
	
	// Determine target path relative to executable
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(ex)

	if update.Type == "theme" {
		url = fmt.Sprintf("%s/themes/%s.css", BaseRepoURL, update.Name)
		filename = filepath.Join(exeDir, "themes", update.Name+".json") // Note: User changed extension to .json in previous step, assuming themes are JSON now?
		// Wait, themes are JSON in this project (theming.go loads .json).
		// But in component_updater.go line 178 it says .css in URL.
		// Let's check theming.go again. It loads .json.
		// So URL should probably be .json too?
		// The user manually changed .css to .json in step 2423.
		// So I should respect that.
		url = fmt.Sprintf("%s/themes/%s.json", BaseRepoURL, update.Name)
	} else if update.Type == "locale" {
		url = fmt.Sprintf("%s/locales/%s.json", BaseRepoURL, update.Name)
		filename = filepath.Join(exeDir, "locales", update.Name+".json")
	} else {
		return fmt.Errorf("unknown component type: %s", update.Type)
	}

	// 2. Download file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download component: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status: %s", resp.Status)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// 3. Update versions.json
	// We need a mutex here if we were doing concurrent updates, but for now it's fine
	// or we can add a lock to Updater struct.
	
	// Re-load to get latest state
	local, err := u.LoadLocalVersions()
	if err != nil {
		return err // Should not happen if we just wrote it
	}

	if update.Type == "theme" {
		local.Themes[update.Name] = version.ComponentInfo{
			Version:   update.Version,
			Changelog: update.Changelog,
		}
	} else {
		local.Locales[update.Name] = version.ComponentInfo{
			Version:   update.Version,
			Changelog: update.Changelog,
		}
	}

	return u.SaveLocalVersions(local)
}
