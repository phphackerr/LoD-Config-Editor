package updater

import (
	"net/http"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func NewUpdater(app *application.App) *Updater {
	return &Updater{
		app: app,
		client: &http.Client{
			Timeout: httpRequestTimeout,
		},
	}
}

func (u *Updater) emitProgress(status string, percent float64) {
	if u.app == nil {
		return
	}

	u.app.Event.Emit("update:progress", map[string]interface{}{
		"status":  status,
		"percent": percent,
		"ts":      time.Now().UnixMilli(),
	})
}
