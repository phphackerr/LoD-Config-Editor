package config_watcher

import (
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type ConfigWatcher struct {
	app       *application.App
	filePath  string
	watcher   *fsnotify.Watcher
	stop      chan struct{}
	done      chan struct{}
	mu        sync.Mutex
	lastEvent time.Time

	debounceMs int
}

// New создаёт пустой вотчер
func New(app *application.App) *ConfigWatcher {
	return &ConfigWatcher{
		app: app,
	}
}

// StartWatching — вызывается из фронтенда
func (cw *ConfigWatcher) StartWatching(path string, debounceMs int) error {
	cw.StopWatching()

	if debounceMs <= 0 {
		debounceMs = 200 // дефолт
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := w.Add(dir); err != nil {
		_ = w.Close()
		return err
	}

	cw.mu.Lock()
	cw.filePath = path
	cw.debounceMs = debounceMs
	cw.watcher = w
	cw.stop = make(chan struct{})
	cw.done = make(chan struct{})
	cw.lastEvent = time.Time{}
	stop := cw.stop
	done := cw.done
	cw.mu.Unlock()

	go cw.run(w, stop, done)
	return nil
}

func (cw *ConfigWatcher) run(w *fsnotify.Watcher, stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	defer func() {
		_ = w.Close()
	}()

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}

			cw.mu.Lock()
			filePath := cw.filePath
			debounce := cw.debounceMs
			lastEvent := cw.lastEvent
			cw.mu.Unlock()

			if filepath.Clean(event.Name) == filepath.Clean(filePath) &&
				(event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Rename == fsnotify.Rename ||
					event.Op&fsnotify.Chmod == fsnotify.Chmod) {

				now := time.Now()
				if now.Sub(lastEvent) < time.Duration(debounce)*time.Millisecond {
					continue
				}

				cw.mu.Lock()
				cw.lastEvent = now
				cw.mu.Unlock()

				log.Println("⚡ Config file changed:", event)
				cw.app.Event.Emit("config-changed", filePath)
			}

		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			if err != nil {
				log.Println("Watcher error:", err)
			}

		case <-stop:
			return
		}
	}
}

// StopWatching — останавливает наблюдение
func (cw *ConfigWatcher) StopWatching() {
	cw.mu.Lock()
	stop := cw.stop
	done := cw.done
	cw.stop = nil
	cw.done = nil
	cw.watcher = nil
	cw.mu.Unlock()

	if stop != nil {
		close(stop)
		if done != nil {
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				log.Println("Watcher stop timeout")
			}
		}
	}
}
