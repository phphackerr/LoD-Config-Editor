package map_downloader

import (
	"lce/backend/fsutil"
	"path/filepath"
	"strings"
)

func mapFileName(raw string) string {
	name := strings.TrimSpace(raw)
	ext := strings.ToLower(filepath.Ext(name))

	base := name
	if ext == ".w3x" || ext == ".w3m" {
		base = strings.TrimSuffix(name, filepath.Ext(name))
	} else {
		ext = ".w3x"
	}

	safeBase := fsutil.SanitizeFilename(base, "map")
	return safeBase + ext
}
