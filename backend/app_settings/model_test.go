package app_settings

import "testing"

func TestNormalizeSettings(t *testing.T) {
	t.Parallel()

	input := Settings{
		Width:    100,
		Height:   200,
		Language: "../bad",
		GamePath: ` C:\Games\Warcraft III\ `,
		AllPaths: []string{
			`C:\Games\Warcraft III`,
			`C:/Games/Warcraft III`,
			``,
		},
		Theme: "",
	}

	normalized, changed := normalizeSettings(input)
	if !changed {
		t.Fatalf("expected settings to be changed by normalization")
	}

	if normalized.Width != minWindowWidth {
		t.Fatalf("expected width=%d, got=%d", minWindowWidth, normalized.Width)
	}
	if normalized.Height != minWindowHeight {
		t.Fatalf("expected height=%d, got=%d", minWindowHeight, normalized.Height)
	}
	if normalized.Language != DefaultSettings().Language {
		t.Fatalf("expected default language, got=%q", normalized.Language)
	}
	if normalized.Theme != DefaultSettings().Theme {
		t.Fatalf("expected default theme, got=%q", normalized.Theme)
	}
	if normalized.GamePath != `C:\Games\Warcraft III\` {
		t.Fatalf("unexpected normalized game path: %q", normalized.GamePath)
	}
	if len(normalized.AllPaths) != 1 {
		t.Fatalf("expected deduplicated all_paths length 1, got %d (%v)", len(normalized.AllPaths), normalized.AllPaths)
	}
	if !containsPath(normalized.AllPaths, normalized.GamePath) {
		t.Fatalf("expected game_path to be present in all_paths")
	}
}

func TestCloneSettings(t *testing.T) {
	t.Parallel()

	original := Settings{
		AllPaths: []string{"a", "b"},
	}
	cloned := cloneSettings(original)
	cloned.AllPaths[0] = "x"

	if original.AllPaths[0] != "a" {
		t.Fatalf("clone must not mutate original slice")
	}
}
