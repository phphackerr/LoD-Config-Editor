package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

var simpleNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// ValidateSimpleName validates an identifier-like file stem (theme/locale/component name).
// It rejects path separators and any characters outside [A-Za-z0-9._-].
func ValidateSimpleName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is empty")
	}

	if name == "." || name == ".." {
		return fmt.Errorf("invalid name: %q", name)
	}

	if filepath.Base(name) != name || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("path separators are not allowed in name: %q", name)
	}

	if !simpleNamePattern.MatchString(name) {
		return fmt.Errorf("name contains unsupported characters: %q", name)
	}

	return nil
}

// WriteFileAtomic writes file contents atomically in the same directory.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}

	tmpPath := tmp.Name()
	cleanup := func() {
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}

	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}

	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cleanup()
		return err
	}

	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return err
	}

	return nil
}

// SanitizeFilename removes filesystem-invalid characters and reserved names.
func SanitizeFilename(name, fallback string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = fallback
	}

	var b strings.Builder
	for _, r := range name {
		if r < 32 || strings.ContainsRune(`<>:"/\|?*`, r) {
			b.WriteRune('_')
			continue
		}
		if r == 0 || !utf8.ValidRune(r) {
			b.WriteRune('_')
			continue
		}
		b.WriteRune(r)
	}

	clean := strings.Trim(b.String(), " .")
	if clean == "" {
		clean = fallback
	}

	upper := strings.ToUpper(clean)
	switch upper {
	case "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		clean = "_" + clean
	}

	runes := []rune(clean)
	if len(runes) > 180 {
		clean = string(runes[:180])
	}

	return clean
}
