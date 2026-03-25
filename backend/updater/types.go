package updater

import (
	"net/http"
	"sync"
	"time"

	"lce/backend/version"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	RepoOwner = "phphackerr"
	RepoName  = "LoD-Config-Editor"
)

const (
	githubAPIBaseURL     = "https://api.github.com/repos/" + RepoOwner + "/" + RepoName
	manifestMainURL      = "https://raw.githubusercontent.com/phphackerr/LoD-Config-Editor/main/manifest.json"
	manifestMasterURL    = "https://raw.githubusercontent.com/phphackerr/LoD-Config-Editor/master/manifest.json"
	componentMaxBytes    = 4 * 1024 * 1024   // 4 MiB
	checksumMaxBytes     = 1 * 1024 * 1024   // 1 MiB
	updateAssetMaxBytes  = 500 * 1024 * 1024 // 500 MiB
	httpRequestTimeout   = 30 * time.Second
	oldExeCleanupBackoff = 2 * time.Second
)

type Updater struct {
	app    *application.App
	client *http.Client

	mu         sync.Mutex
	newExePath string
}

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type ReleaseInfo struct {
	TagName string         `json:"tag_name"`
	Body    string         `json:"body"`
	Assets  []ReleaseAsset `json:"assets"`
}

type UpdateCheckResult struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Body      string `json:"body"`
	Error     string `json:"error,omitempty"`
}

// ComponentVersions matches the structure of manifest.json.
type ComponentVersions struct {
	AppVersion version.ComponentInfo            `json:"app_version"`
	Themes     map[string]version.ComponentInfo `json:"themes"`
	Locales    map[string]version.ComponentInfo `json:"locales"`
}

type ComponentUpdate struct {
	Type      string `json:"type"` // "theme" or "locale"
	Name      string `json:"name"`
	Version   string `json:"version"`
	Changelog string `json:"changelog"`
}
