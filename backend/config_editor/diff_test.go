package config_editor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckConfigDiffAndApplyChanges(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.lod.ini")

	initial := "[General]\na = 1\nb = 2\n\n[Old]\nx = 9\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	editor := NewConfigEditor()
	if _, err := editor.LoadConfig(path); err != nil {
		t.Fatalf("failed to load initial config: %v", err)
	}

	external := "[General]\na = 10\nc = 3\n\n[New]\nn = 5\n"
	if err := os.WriteFile(path, []byte(external), 0644); err != nil {
		t.Fatalf("failed to write external config: %v", err)
	}

	diff, err := editor.CheckConfigDiff()
	if err != nil {
		t.Fatalf("CheckConfigDiff failed: %v", err)
	}

	if len(diff) != 3 {
		t.Fatalf("unexpected diff sections count: got %d, want %d", len(diff), 3)
	}

	bySection := make(map[string]SectionDiff, len(diff))
	for _, section := range diff {
		bySection[section.Section] = section
	}

	general, ok := bySection["General"]
	if !ok {
		t.Fatalf("missing General section diff")
	}
	if general.Status != DiffStatusModified {
		t.Fatalf("General status: got %q, want %q", general.Status, DiffStatusModified)
	}

	generalKeys := make(map[string]KeyDiff, len(general.Keys))
	for _, key := range general.Keys {
		generalKeys[key.Key] = key
	}
	if generalKeys["a"].Status != DiffStatusModified || generalKeys["a"].Old != "1" || generalKeys["a"].New != "10" {
		t.Fatalf("unexpected diff for key a: %+v", generalKeys["a"])
	}
	if generalKeys["b"].Status != DiffStatusDeleted || generalKeys["b"].Old != "2" || generalKeys["b"].New != "" {
		t.Fatalf("unexpected diff for key b: %+v", generalKeys["b"])
	}
	if generalKeys["c"].Status != DiffStatusAdded || generalKeys["c"].Old != "" || generalKeys["c"].New != "3" {
		t.Fatalf("unexpected diff for key c: %+v", generalKeys["c"])
	}

	newSection, ok := bySection["New"]
	if !ok {
		t.Fatalf("missing New section diff")
	}
	if newSection.Status != DiffStatusAdded {
		t.Fatalf("New status: got %q, want %q", newSection.Status, DiffStatusAdded)
	}

	oldSection, ok := bySection["Old"]
	if !ok {
		t.Fatalf("missing Old section diff")
	}
	if oldSection.Status != DiffStatusDeleted {
		t.Fatalf("Old status: got %q, want %q", oldSection.Status, DiffStatusDeleted)
	}

	if err := editor.ApplyChangesToConfig(diff); err != nil {
		t.Fatalf("ApplyChangesToConfig failed: %v", err)
	}

	reloaded := &GameConfig{}
	if err := reloaded.Load(path); err != nil {
		t.Fatalf("failed to reload config after apply: %v", err)
	}

	if v, err := reloaded.Get("General", "a"); err != nil || v != "10" {
		t.Fatalf("unexpected General.a after apply: value=%q err=%v", v, err)
	}
	if v, err := reloaded.Get("General", "c"); err != nil || v != "3" {
		t.Fatalf("unexpected General.c after apply: value=%q err=%v", v, err)
	}
	if _, err := reloaded.Get("General", "b"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected General.b to be deleted, got err=%v", err)
	}
	if _, err := reloaded.Get("Old", "x"); !errors.Is(err, ErrSectionNotFound) {
		t.Fatalf("expected Old section to be deleted, got err=%v", err)
	}
	if v, err := reloaded.Get("New", "n"); err != nil || v != "5" {
		t.Fatalf("unexpected New.n after apply: value=%q err=%v", v, err)
	}
}
