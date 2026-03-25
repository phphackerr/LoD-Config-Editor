package runtimepaths

import (
	"fmt"
	"os"
	"path/filepath"
)

const appFolderName = "LCE"

func AppDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	dir := filepath.Join(configDir, appFolderName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app data dir: %w", err)
	}

	return dir, nil
}

func ThemesDir() (string, error) {
	base, err := AppDataDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "themes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create themes dir: %w", err)
	}
	return dir, nil
}

func LocalesDir() (string, error) {
	base, err := AppDataDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "locales")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create locales dir: %w", err)
	}
	return dir, nil
}

func ManifestPath() (string, error) {
	base, err := AppDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "manifest.json"), nil
}
