package paths_scanner

import (
	"path/filepath"
	"sort"
	"strings"
)

func normalizePathKey(path string) string {
	clean := filepath.Clean(path)
	return strings.ToLower(strings.ReplaceAll(clean, "/", `\`))
}

func uniqueSortedPaths(paths []string) []string {
	seen := make(map[string]string, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		clean := filepath.Clean(path)
		key := normalizePathKey(clean)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = clean
	}

	result := make([]string, 0, len(seen))
	for _, path := range seen {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func depthFromRoot(root, path string) (int, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return 0, err
	}

	if rel == "." {
		return 0, nil
	}

	return strings.Count(rel, string(filepath.Separator)) + 1, nil
}
