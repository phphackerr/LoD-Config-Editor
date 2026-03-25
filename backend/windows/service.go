package windows

import (
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

type WindowService struct {
	app     *application.App
	windows map[WindowType]*application.WebviewWindow
	mu      sync.Mutex
}

func NewWindowService(app *application.App) *WindowService {
	return &WindowService{
		app:     app,
		windows: make(map[WindowType]*application.WebviewWindow),
	}
}

func (s *WindowService) createWindow(
	t WindowType,
	title string,
	width, height int,
) (*application.WebviewWindow, error) {

	opts := application.WebviewWindowOptions{
		Title:     title,
		Width:     width,
		Height:    height,
		Frameless: true,
		URL:       t.URL(),
	}

	win := s.app.Window.NewWithOptions(opts)
	win.OnWindowEvent(events.Common.WindowClosing, func(_ *application.WindowEvent) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if current, ok := s.windows[t]; ok && current == win {
			delete(s.windows, t)
		}
	})

	return win, nil
}

func (s *WindowService) OpenThemeEditor() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if win, ok := s.windows[WindowThemeEditor]; ok {
		win.Show()
		win.Focus()
		return nil
	}

	win, err := s.createWindow(
		WindowThemeEditor,
		"Theme Editor",
		1100,
		800,
	)
	if err != nil {
		return fmt.Errorf("failed to create theme editor window: %w", err)
	}

	s.windows[WindowThemeEditor] = win
	win.Center()
	win.Show()

	return nil
}

func (s *WindowService) IsThemeEditorOpen() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.windows[WindowThemeEditor]
	return ok
}

func (s *WindowService) OpenLanguageEditor() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if win, ok := s.windows[WindowLanguageEditor]; ok {
		win.Show()
		win.Focus()
		return nil
	}

	win, err := s.createWindow(
		WindowLanguageEditor,
		"Language Editor",
		1280,
		900,
	)
	if err != nil {
		return fmt.Errorf("failed to create language editor window: %w", err)
	}

	s.windows[WindowLanguageEditor] = win
	win.Center()
	win.Show()

	return nil
}

func (s *WindowService) IsLanguageEditorOpen() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.windows[WindowLanguageEditor]
	return ok
}
