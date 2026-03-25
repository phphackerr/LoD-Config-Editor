package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateSimpleName(t *testing.T) {
	t.Parallel()

	valid := []string{"default", "dark-theme", "pt-BR", "theme.v2"}
	for _, name := range valid {
		if err := ValidateSimpleName(name); err != nil {
			t.Fatalf("expected name %q to be valid, got error: %v", name, err)
		}
	}

	invalid := []string{"", ".", "..", "../x", "x/y", `x\y`, "white space", "name*"}
	for _, name := range invalid {
		if err := ValidateSimpleName(name); err == nil {
			t.Fatalf("expected name %q to be invalid", name)
		}
	}
}

func TestWriteFileAtomic(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	content := []byte(`{"ok":true}`)

	if err := WriteFileAtomic(path, content, 0644); err != nil {
		t.Fatalf("WriteFileAtomic failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("unexpected file content: got %q, want %q", string(got), string(content))
	}
}
