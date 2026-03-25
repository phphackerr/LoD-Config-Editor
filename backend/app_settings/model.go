package app_settings

import (
	"lce/backend/fsutil"
	"path/filepath"
	"slices"
	"strings"
)

const (
	minWindowWidth  uint32 = 650
	minWindowHeight uint32 = 350
)

// Settings stores application settings.
type Settings struct {
	Width        uint32   `json:"width"`
	Height       uint32   `json:"height"`
	Language     string   `json:"language"`
	GamePath     string   `json:"game_path"`
	FirstRun     bool     `json:"first_run"`
	AllPaths     []string `json:"all_paths"`
	Theme        string   `json:"theme"`
	WindowedMode bool     `json:"windowed_mode"`
}

// DefaultSettings returns default settings.
func DefaultSettings() Settings {
	return Settings{
		Width:        1600,
		Height:       900,
		Language:     "en",
		GamePath:     "",
		FirstRun:     true,
		AllPaths:     []string{},
		Theme:        "default",
		WindowedMode: false,
	}
}

func normalizeSettings(input Settings) (Settings, bool) {
	normalized := input
	changed := false

	if normalized.Width < minWindowWidth {
		normalized.Width = minWindowWidth
		changed = true
	}
	if normalized.Height < minWindowHeight {
		normalized.Height = minWindowHeight
		changed = true
	}

	lang := strings.TrimSpace(normalized.Language)
	if lang == "" || fsutil.ValidateSimpleName(lang) != nil {
		normalized.Language = DefaultSettings().Language
		changed = true
	}

	theme := strings.TrimSpace(normalized.Theme)
	if theme == "" || fsutil.ValidateSimpleName(theme) != nil {
		normalized.Theme = DefaultSettings().Theme
		changed = true
	} else if theme != normalized.Theme {
		normalized.Theme = theme
		changed = true
	}

	gamePath := strings.TrimSpace(normalized.GamePath)
	if gamePath != normalized.GamePath {
		normalized.GamePath = gamePath
		changed = true
	}

	sanitizedPaths := sanitizePaths(normalized.AllPaths)
	if !slices.Equal(normalized.AllPaths, sanitizedPaths) {
		normalized.AllPaths = sanitizedPaths
		changed = true
	}

	if normalized.GamePath != "" && !containsPath(normalized.AllPaths, normalized.GamePath) {
		normalized.AllPaths = append([]string{normalized.GamePath}, normalized.AllPaths...)
		changed = true
	}

	return normalized, changed
}

func sanitizePaths(paths []string) []string {
	if len(paths) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))

	for _, raw := range paths {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}

		clean := filepath.Clean(path)
		key := normalizePathKey(clean)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, clean)
	}

	return result
}

func containsPath(paths []string, target string) bool {
	key := normalizePathKey(filepath.Clean(target))
	for _, path := range paths {
		if normalizePathKey(filepath.Clean(path)) == key {
			return true
		}
	}
	return false
}

func normalizePathKey(path string) string {
	return strings.ToLower(strings.ReplaceAll(path, "/", "\\"))
}

func cloneSettings(s Settings) Settings {
	out := s
	if s.AllPaths != nil {
		out.AllPaths = append([]string(nil), s.AllPaths...)
	} else {
		out.AllPaths = []string{}
	}
	return out
}
