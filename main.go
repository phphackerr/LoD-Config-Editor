package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"

	"lce/backend/app_settings"
	"lce/backend/config_editor"
	"lce/backend/config_watcher"
	"lce/backend/i18n"
	"lce/backend/map_downloader"
	"lce/backend/paths_scanner"
	"lce/backend/taskbar"
	"lce/backend/theming"
	"lce/backend/updater"
	"lce/backend/utils"
	"lce/backend/version"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed manifest.json
var manifestData []byte

//go:embed themes/*
var themesFS embed.FS

//go:embed locales/*
var localesFS embed.FS

func main() {
	// Get user config directory for manifest
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("failed to get user config dir: %v", err)
	}
	appDataDir := filepath.Join(configDir, "LCE")

	// Get executable directory for assets
	ex, err := os.Executable()
	if err != nil {
		log.Fatalf("failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(ex)

	// Initialize version info: manifest in AppData, assets in ExeDir
	version.Init(manifestData, themesFS, localesFS, appDataDir, exeDir)

	app := application.New(application.Options{
		Name:        "LoD Config Editor",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(i18n.NewI18N()),
			application.NewService(theming.NewThemeService()),
			application.NewService(paths_scanner.NewScanner()),
			application.NewService(config_editor.NewConfigEditor()),
			application.NewService(taskbar.NewTaskbarUtils()),
			application.NewService(utils.NewUtils()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	_ = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "LoD Config Editor",
		Width:     1300,
		Height:    800,
		Frameless: true,
		URL:       "/",
	})

	appSettings := app_settings.NewAppSettings(app)
	app.RegisterService(application.NewService(appSettings))

	configWatcher := config_watcher.New(app)
	app.RegisterService(application.NewService(configWatcher))

	mapDownloader := map_downloader.NewMapDownloader(app)
	app.RegisterService(application.NewService(mapDownloader))

	updaterService := updater.NewUpdater(app)
	app.RegisterService(application.NewService(updaterService))

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}


}
