package config_editor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGameConfigLoadGetSetSave(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.lod.ini")
	initial := "[General]\nGameSpeed = Fast\n\n[Hotkeys]\ninventory1 = 0x41\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatalf("failed to write test ini: %v", err)
	}

	cfg := &GameConfig{}
	if err := cfg.Load(path); err != nil {
		t.Fatalf("failed to load ini: %v", err)
	}

	got, err := cfg.Get("general", "gamespeed")
	if err != nil {
		t.Fatalf("failed to get key case-insensitively: %v", err)
	}
	if got != "Fast" {
		t.Fatalf("unexpected value: got %q, want %q", got, "Fast")
	}

	if err := cfg.Set("GENERAL", "GAMESPEED", "Slow"); err != nil {
		t.Fatalf("failed to set existing key: %v", err)
	}
	if err := cfg.Set("Custom", "MyKey", "123"); err != nil {
		t.Fatalf("failed to set new key: %v", err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to save ini: %v", err)
	}

	reloaded := &GameConfig{}
	if err := reloaded.Load(path); err != nil {
		t.Fatalf("failed to reload ini: %v", err)
	}

	got, err = reloaded.Get("General", "GameSpeed")
	if err != nil {
		t.Fatalf("failed to get updated key: %v", err)
	}
	if got != "Slow" {
		t.Fatalf("unexpected updated value: got %q, want %q", got, "Slow")
	}

	got, err = reloaded.Get("custom", "mykey")
	if err != nil {
		t.Fatalf("failed to get newly added key: %v", err)
	}
	if got != "123" {
		t.Fatalf("unexpected new key value: got %q, want %q", got, "123")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read saved ini: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "GameSpeed = Slow") {
		t.Fatalf("saved ini does not contain updated key, got:\n%s", text)
	}
}

func TestGameConfigErrorsAndStrictEditorAPI(t *testing.T) {
	t.Parallel()

	cfg := &GameConfig{}
	if _, err := cfg.Get("General", "GameSpeed"); !errors.Is(err, ErrConfigNotLoaded) {
		t.Fatalf("Get should return ErrConfigNotLoaded, got: %v", err)
	}
	if err := cfg.Set("General", "GameSpeed", "Fast"); !errors.Is(err, ErrConfigNotLoaded) {
		t.Fatalf("Set should return ErrConfigNotLoaded, got: %v", err)
	}
	if err := cfg.Save(); !errors.Is(err, ErrConfigNotLoaded) {
		t.Fatalf("Save should return ErrConfigNotLoaded, got: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.lod.ini")
	if err := os.WriteFile(path, []byte("[General]\nGameSpeed = Fast\n"), 0644); err != nil {
		t.Fatalf("failed to write test ini: %v", err)
	}

	editor := NewConfigEditor()
	if _, err := editor.GetConfigValueStrict("General", "GameSpeed"); !errors.Is(err, ErrConfigNotLoaded) {
		t.Fatalf("strict getter should fail before load, got: %v", err)
	}
	if got := editor.GetConfigValue("General", "GameSpeed"); got != "" {
		t.Fatalf("non-strict getter should return empty string before load, got: %q", got)
	}

	if _, err := editor.LoadConfig(path); err != nil {
		t.Fatalf("failed to load config through editor: %v", err)
	}

	if _, err := editor.GetConfigValueStrict("Missing", "GameSpeed"); !errors.Is(err, ErrSectionNotFound) {
		t.Fatalf("missing section should return ErrSectionNotFound, got: %v", err)
	}
	if _, err := editor.GetConfigValueStrict("General", "Missing"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("missing key should return ErrKeyNotFound, got: %v", err)
	}
}
