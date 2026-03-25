package windows

type WindowType string

const (
	WindowMain           WindowType = "main"
	WindowThemeEditor    WindowType = "themeEditor"
	WindowLanguageEditor WindowType = "languageEditor"
)

func (wt WindowType) URL() string {
	switch wt {
	case WindowThemeEditor:
		return "/?window=themeEditor"
	case WindowLanguageEditor:
		return "/?window=languageEditor"
	case WindowMain:
	default:
		return "/"
	}
	return "/"
}
