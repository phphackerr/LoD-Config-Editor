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
	app        *application.App
	filePath   string
	watcher    *fsnotify.Watcher
	stop       chan struct{}
	stopOnce   sync.Once
	lastEvent  time.Time
	debounceMs int
}

// New создаёт пустой вотчер
func New(app *application.App) *ConfigWatcher {
	return &ConfigWatcher{
		app:  app,
		stop: make(chan struct{}),
	}
}

// StartWatching — вызывается из фронтенда
func (cw *ConfigWatcher) StartWatching(path string, debounceMs int) error {
	// если вотчер уже активен — останавливаем
	if cw.watcher != nil {
		cw.StopWatching()
	}

	if debounceMs <= 0 {
		debounceMs = 200 // дефолт
	}

	cw.filePath = path
	cw.debounceMs = debounceMs
	cw.stopOnce = sync.Once{}
	cw.stop = make(chan struct{})

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	cw.watcher = w

	dir := filepath.Dir(cw.filePath)
	if err := w.Add(dir); err != nil {
		return err
	}

	go cw.run()
	return nil
}

func (cw *ConfigWatcher) run() {
	for {
		select {
		case event := <-cw.watcher.Events:
			if filepath.Clean(event.Name) == filepath.Clean(cw.filePath) &&
				(event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Rename == fsnotify.Rename ||
					event.Op&fsnotify.Chmod == fsnotify.Chmod) {

				now := time.Now()
				if now.Sub(cw.lastEvent) < time.Duration(cw.debounceMs)*time.Millisecond {
					continue
				}
				cw.lastEvent = now

				log.Println("⚡ Config file changed:", event)
				cw.app.Event.Emit("config-changed", cw.filePath)
			}

		case err := <-cw.watcher.Errors:
			log.Println("Watcher error:", err)

		case <-cw.stop:
			_ = cw.watcher.Close()
			return
		}
	}
}

// StopWatching — останавливает наблюдение
func (cw *ConfigWatcher) StopWatching() {
	cw.stopOnce.Do(func() {
		close(cw.stop)
		if cw.watcher != nil {
			_ = cw.watcher.Close()
			cw.watcher = nil
		}
	})
}
