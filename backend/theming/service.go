package theming

import (
	"fmt"
	"strings"
)

type ThemeService struct {
	storage ThemeStorage
}

func NewThemeService() (*ThemeService, error) {
	storage, err := NewFileSystemStorage()
	if err != nil {
		return nil, err
	}
	return &ThemeService{storage: storage}, nil
}

func (ts *ThemeService) GetThemes() ([]string, error) {
	return ts.storage.List()
}

func (ts *ThemeService) LoadRawTheme(name string) (Theme, error) {
	return ts.storage.Load(name)
}

func (ts *ThemeService) LoadResolvedTheme(name string) (ResolvedTheme, []ThemeError) {
	theme, err := ts.LoadRawTheme(name)
	if err != nil {
		return ResolvedTheme{}, []ThemeError{
			BaseThemeError{
				code: "io_error",
				path: "theme",
				msg:  err.Error(),
			},
		}
	}

	return ResolveTheme(theme)
}

func (ts *ThemeService) LoadThemeCSS(name string) (string, []ThemeError) {
	resolved, errs := ts.LoadResolvedTheme(name)
	if len(errs) > 0 {
		return "", errs
	}

	css, err := CompileCSS(resolved)
	if err != nil {
		return "", []ThemeError{
			BaseThemeError{
				code: "compile_error",
				path: "theme",
				msg:  err.Error(),
			},
		}
	}

	return css, nil
}

func (ts *ThemeService) SaveTheme(name string, theme Theme) error {
	if errs := ValidateTheme(theme); len(errs) > 0 {
		return fmt.Errorf("theme validation failed: %s", formatThemeErrors(errs))
	}

	return ts.storage.Save(name, theme)
}

func (ts *ThemeService) CreateTheme(name, base string) error {
	exists, err := ts.storage.Exists(name)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("theme '%s' already exists", name)
	}

	baseTheme, err := ts.storage.Load(base)
	if err != nil {
		return err
	}

	// Update meta for new theme
	baseTheme.Meta.ID = name
	baseTheme.Meta.Name = name

	return ts.storage.Save(name, baseTheme)
}

func ValidateTheme(theme Theme) []ThemeError {
	resolved, errs := ResolveTheme(theme)
	if len(errs) > 0 {
		return errs
	}

	if _, err := CompileCSS(resolved); err != nil {
		return []ThemeError{
			BaseThemeError{
				code: "compile_error",
				path: "theme",
				msg:  err.Error(),
			},
		}
	}

	return nil
}

func formatThemeErrors(errs []ThemeError) string {
	parts := make([]string, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}

		path := strings.TrimSpace(err.Path())
		msg := strings.TrimSpace(err.Error())
		switch {
		case path != "" && msg != "":
			parts = append(parts, path+": "+msg)
		case msg != "":
			parts = append(parts, msg)
		case path != "":
			parts = append(parts, path)
		default:
			parts = append(parts, "unknown error")
		}
	}

	if len(parts) == 0 {
		return "unknown validation error"
	}

	return strings.Join(parts, "; ")
}
