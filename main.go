package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"lce/backend/app_settings"
	"lce/backend/config_editor"
	"lce/backend/config_watcher"
	"lce/backend/i18n"
	"lce/backend/map_downloader"
	"lce/backend/paths_scanner"
	"lce/backend/taskbar"
	"lce/backend/theming"
	"lce/backend/utils"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {

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

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
