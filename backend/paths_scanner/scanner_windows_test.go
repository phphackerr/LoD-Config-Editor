//go:build windows

package paths_scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindInstallPathsInRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	validDir := filepath.Join(root, "Games", "Warcraft III")
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatalf("failed to create valid directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(validDir, "war3.exe"), []byte("x"), 0644); err != nil {
		t.Fatalf("failed to create war3.exe: %v", err)
	}

	excludedDir := filepath.Join(root, "Windows", "War3")
	if err := os.MkdirAll(excludedDir, 0755); err != nil {
		t.Fatalf("failed to create excluded directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(excludedDir, "config.lod.ini"), []byte("x"), 0644); err != nil {
		t.Fatalf("failed to create config in excluded dir: %v", err)
	}

	deepDir := filepath.Join(root, "A", "B", "C", "War3")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatalf("failed to create deep directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(deepDir, "config.lod.ini"), []byte("x"), 0644); err != nil {
		t.Fatalf("failed to create deep config: %v", err)
	}

	s := NewScanner()
	s.maxDepth = 3
	s.maxMatchesPerRoot = 10
	s.timeout = 2 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	found, err := s.findInstallPathsInRoot(ctx, root)
	if err != nil {
		t.Fatalf("findInstallPathsInRoot returned error: %v", err)
	}

	if len(found) != 1 {
		t.Fatalf("expected exactly one found path, got %d (%v)", len(found), found)
	}
	if filepath.Clean(found[0]) != filepath.Clean(validDir) {
		t.Fatalf("unexpected found path: got %q, want %q", found[0], validDir)
	}
}
