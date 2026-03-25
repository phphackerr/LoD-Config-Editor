package theming

import (
	"strings"
	"testing"
)

type stubThemeStorage struct {
	saved bool
}

func (s *stubThemeStorage) List() ([]string, error) {
	return nil, nil
}

func (s *stubThemeStorage) Load(name string) (Theme, error) {
	return Theme{}, nil
}

func (s *stubThemeStorage) Save(name string, theme Theme) error {
	s.saved = true
	return nil
}

func (s *stubThemeStorage) Exists(name string) (bool, error) {
	return false, nil
}

func validThemeFixture() Theme {
	return Theme{
		Meta: ThemeMeta{
			ID:      "test",
			Name:    "Test",
			Author:  "test",
			Version: 1,
		},
		Tokens: []Token{
			{
				ID:   "app.bg",
				Type: TokenColor,
				Value: TokenValue{
					Literal: &Literal{
						Color: &ColorLiteral{
							Space: "srgb",
							R:     40,
							G:     50,
							B:     60,
							A:     1,
						},
					},
				},
			},
			{
				ID:   "app.hover.bg",
				Type: TokenColor,
				Value: TokenValue{
					Derived: &Derived{
						Op: "lighten",
						From: TokenValue{
							Ref: &Ref{ID: "app.bg"},
						},
						Amount: 5,
					},
				},
			},
		},
	}
}

func invalidThemeFixtureUnknownRef() Theme {
	theme := validThemeFixture()
	theme.Tokens[1].Value = TokenValue{
		Ref: &Ref{ID: "missing.token"},
	}
	return theme
}

func TestValidateTheme_OK(t *testing.T) {
	t.Parallel()

	errs := ValidateTheme(validThemeFixture())
	if len(errs) != 0 {
		t.Fatalf("ValidateTheme returned unexpected errors: %v", errs)
	}
}

func TestValidateTheme_UnknownRef(t *testing.T) {
	t.Parallel()

	errs := ValidateTheme(invalidThemeFixtureUnknownRef())
	if len(errs) == 0 {
		t.Fatalf("ValidateTheme should return errors for unknown ref")
	}
}

func TestThemeServiceSaveTheme_ValidationPreventsWrite(t *testing.T) {
	t.Parallel()

	storage := &stubThemeStorage{}
	service := &ThemeService{storage: storage}

	err := service.SaveTheme("test", invalidThemeFixtureUnknownRef())
	if err == nil {
		t.Fatalf("SaveTheme should fail on invalid theme")
	}
	if storage.saved {
		t.Fatalf("storage.Save should not be called for invalid theme")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "validation") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestThemeServiceSaveTheme_WritesValidTheme(t *testing.T) {
	t.Parallel()

	storage := &stubThemeStorage{}
	service := &ThemeService{storage: storage}

	if err := service.SaveTheme("test", validThemeFixture()); err != nil {
		t.Fatalf("SaveTheme should save valid theme, got error: %v", err)
	}
	if !storage.saved {
		t.Fatalf("storage.Save should be called for valid theme")
	}
}
