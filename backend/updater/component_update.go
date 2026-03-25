package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"lce/backend/fsutil"
	"lce/backend/runtimepaths"
	"lce/backend/theming"
	"lce/backend/version"
	"lce/backend/versioncmp"
)

func getVersionsPath() (string, error) {
	return runtimepaths.ManifestPath()
}

// LoadLocalVersions loads versions from manifest.json in AppData.
func (u *Updater) LoadLocalVersions() (ComponentVersions, error) {
	path, err := getVersionsPath()
	if err != nil {
		return ComponentVersions{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			defaults := ComponentVersions{
				AppVersion: version.App,
				Themes:     cloneComponentMap(version.Themes),
				Locales:    cloneComponentMap(version.Locales),
			}
			if err := u.SaveLocalVersions(defaults); err != nil {
				return ComponentVersions{}, err
			}
			return defaults, nil
		}
		return ComponentVersions{}, err
	}

	var versions ComponentVersions
	if err := json.Unmarshal(data, &versions); err != nil {
		return ComponentVersions{}, err
	}

	ensureComponentVersionsDefaults(&versions)
	return versions, nil
}

// SaveLocalVersions saves versions to manifest.json in AppData.
func (u *Updater) SaveLocalVersions(versions ComponentVersions) error {
	ensureComponentVersionsDefaults(&versions)

	path, err := getVersionsPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return err
	}

	return fsutil.WriteFileAtomic(path, data, 0644)
}

// CheckForComponentUpdates fetches manifest and compares with local versions.
func (u *Updater) CheckForComponentUpdates() ([]ComponentUpdate, error) {
	local, err := u.LoadLocalVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to load local versions: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpRequestTimeout)
	defer cancel()

	remote, _, err := u.fetchRemoteManifest(ctx)
	if err != nil {
		return nil, err
	}

	updates := make([]ComponentUpdate, 0, 8)

	for name, remoteInfo := range remote.Themes {
		localInfo, exists := local.Themes[name]
		if !exists || versioncmp.Compare(remoteInfo.Version, localInfo.Version) > 0 {
			updates = append(updates, ComponentUpdate{
				Type:      "theme",
				Name:      name,
				Version:   remoteInfo.Version,
				Changelog: string(remoteInfo.Changelog),
			})
		}
	}

	for name, remoteInfo := range remote.Locales {
		localInfo, exists := local.Locales[name]
		if !exists || versioncmp.Compare(remoteInfo.Version, localInfo.Version) > 0 {
			updates = append(updates, ComponentUpdate{
				Type:      "locale",
				Name:      name,
				Version:   remoteInfo.Version,
				Changelog: string(remoteInfo.Changelog),
			})
		}
	}

	sort.Slice(updates, func(i, j int) bool {
		if updates[i].Type != updates[j].Type {
			return updates[i].Type < updates[j].Type
		}
		return updates[i].Name < updates[j].Name
	})

	return updates, nil
}

// UpdateComponent downloads component file and updates local component versions.
func (u *Updater) UpdateComponent(update ComponentUpdate) error {
	update.Name = strings.TrimSpace(update.Name)
	update.Type = strings.TrimSpace(update.Type)

	if err := fsutil.ValidateSimpleName(update.Name); err != nil {
		return fmt.Errorf("invalid component name %q: %w", update.Name, err)
	}
	if update.Type != "theme" && update.Type != "locale" {
		return fmt.Errorf("unknown component type: %s", update.Type)
	}

	ctx, cancel := context.WithTimeout(context.Background(), httpRequestTimeout)
	defer cancel()

	_, baseURL, err := u.fetchRemoteManifest(ctx)
	if err != nil {
		return err
	}

	componentURL, destination, err := resolveComponentURLAndPath(baseURL, update.Type, update.Name)
	if err != nil {
		return err
	}

	content, err := u.fetchBytes(ctx, componentURL, componentMaxBytes)
	if err != nil {
		return fmt.Errorf("failed to download component: %w", err)
	}

	if err := validateComponentContent(update.Type, content); err != nil {
		return err
	}

	if err := fsutil.WriteFileAtomic(destination, content, 0644); err != nil {
		return fmt.Errorf("failed to save component: %w", err)
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	local, err := u.LoadLocalVersions()
	if err != nil {
		return err
	}

	info := version.ComponentInfo{
		Version:   update.Version,
		Changelog: version.MultiLineString(update.Changelog),
	}
	if update.Type == "theme" {
		local.Themes[update.Name] = info
	} else {
		local.Locales[update.Name] = info
	}

	return u.SaveLocalVersions(local)
}

func (u *Updater) fetchRemoteManifest(ctx context.Context) (ComponentVersions, string, error) {
	candidates := []struct {
		url     string
		baseURL string
	}{
		{url: manifestMainURL, baseURL: strings.TrimSuffix(manifestMainURL, "/manifest.json")},
		{url: manifestMasterURL, baseURL: strings.TrimSuffix(manifestMasterURL, "/manifest.json")},
	}

	var lastErr error
	for _, candidate := range candidates {
		data, err := u.fetchBytes(ctx, candidate.url, componentMaxBytes)
		if err != nil {
			lastErr = err
			continue
		}

		var manifest ComponentVersions
		if err := json.Unmarshal(data, &manifest); err != nil {
			lastErr = err
			continue
		}
		ensureComponentVersionsDefaults(&manifest)
		return manifest, candidate.baseURL, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("failed to fetch manifest from known branches")
	}
	return ComponentVersions{}, "", lastErr
}

func resolveComponentURLAndPath(baseURL, componentType, name string) (string, string, error) {
	switch componentType {
	case "theme":
		themesDir, err := runtimepaths.ThemesDir()
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("%s/themes/%s.json", baseURL, name), filepath.Join(themesDir, name+".json"), nil
	case "locale":
		localesDir, err := runtimepaths.LocalesDir()
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("%s/locales/%s.json", baseURL, name), filepath.Join(localesDir, name+".json"), nil
	default:
		return "", "", fmt.Errorf("unsupported component type: %s", componentType)
	}
}

func validateComponentContent(componentType string, content []byte) error {
	if !json.Valid(content) {
		return fmt.Errorf("downloaded component payload is not valid JSON")
	}

	switch componentType {
	case "theme":
		var theme theming.Theme
		if err := json.Unmarshal(content, &theme); err != nil {
			return fmt.Errorf("invalid theme JSON: %w", err)
		}
		if strings.TrimSpace(theme.Meta.ID) == "" {
			return fmt.Errorf("invalid theme JSON: missing meta.id")
		}
	case "locale":
		var locale map[string]interface{}
		if err := json.Unmarshal(content, &locale); err != nil {
			return fmt.Errorf("invalid locale JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported component type: %s", componentType)
	}

	return nil
}

func ensureComponentVersionsDefaults(versions *ComponentVersions) {
	if versions.Themes == nil {
		versions.Themes = make(map[string]version.ComponentInfo)
	}
	if versions.Locales == nil {
		versions.Locales = make(map[string]version.ComponentInfo)
	}
}

func cloneComponentMap(input map[string]version.ComponentInfo) map[string]version.ComponentInfo {
	out := make(map[string]version.ComponentInfo, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}
