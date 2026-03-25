package theming

import (
	"encoding/json"
	"fmt"
	"lce/backend/fsutil"
	"lce/backend/runtimepaths"
	"os"
	"path/filepath"
	"sort"
)

type ThemeStorage interface {
	List() ([]string, error)
	Load(name string) (Theme, error)
	Save(name string, theme Theme) error
	Exists(name string) (bool, error)
}

type FileSystemStorage struct {
	BaseDir string
}

func NewFileSystemStorage() (*FileSystemStorage, error) {
	themesDir, err := runtimepaths.ThemesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve themes directory: %w", err)
	}

	return &FileSystemStorage{BaseDir: themesDir}, nil
}

func (s *FileSystemStorage) List() ([]string, error) {
	entries, err := os.ReadDir(s.BaseDir)
	if err != nil {
		return nil, err
	}

	var themes []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			themes = append(themes, entry.Name()[:len(entry.Name())-5])
		}
	}

	sort.Strings(themes)
	return themes, nil
}

func (s *FileSystemStorage) Load(name string) (Theme, error) {
	path, err := s.themePath(name)
	if err != nil {
		return Theme{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Theme{}, err
	}

	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return Theme{}, err
	}

	return theme, nil
}

func (s *FileSystemStorage) Save(name string, theme Theme) error {
	path, err := s.themePath(name)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return err
	}

	return fsutil.WriteFileAtomic(path, data, 0644)
}

func (s *FileSystemStorage) Exists(name string) (bool, error) {
	path, err := s.themePath(name)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *FileSystemStorage) themePath(name string) (string, error) {
	if err := fsutil.ValidateSimpleName(name); err != nil {
		return "", err
	}

	return filepath.Join(s.BaseDir, name+".json"), nil
}
