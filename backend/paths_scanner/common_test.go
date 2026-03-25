package paths_scanner

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestUniqueSortedPaths(t *testing.T) {
	t.Parallel()

	paths := []string{
		`C:\Games\Warcraft III`,
		`C:/Games/Warcraft III`,
		`D:\War3`,
		` `,
	}

	got := uniqueSortedPaths(paths)
	want := []string{`C:\Games\Warcraft III`, `D:\War3`}

	if !slices.Equal(got, want) {
		t.Fatalf("uniqueSortedPaths mismatch: got=%v want=%v", got, want)
	}
}

func TestDepthFromRoot(t *testing.T) {
	t.Parallel()

	root := filepath.Clean(filepath.Join(`C:\`, "Games"))

	cases := []struct {
		path string
		want int
	}{
		{path: root, want: 0},
		{path: filepath.Join(root, "Warcraft III"), want: 1},
		{path: filepath.Join(root, "A", "B"), want: 2},
	}

	for _, tc := range cases {
		got, err := depthFromRoot(root, tc.path)
		if err != nil {
			t.Fatalf("depthFromRoot(%q) error: %v", tc.path, err)
		}
		if got != tc.want {
			t.Fatalf("depthFromRoot(%q)=%d want=%d", tc.path, got, tc.want)
		}
	}
}
