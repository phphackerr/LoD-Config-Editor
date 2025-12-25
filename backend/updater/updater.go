package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"lce/backend/version"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	RepoOwner = "phphackerr"
	RepoName  = "LoD-Config-Editor"
)

type Updater struct {
	app        *application.App
	newExePath string
}

type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	} `json:"assets"`
}

type UpdateCheckResult struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Body      string `json:"body"`
	Error     string `json:"error,omitempty"`
}

func NewUpdater(app *application.App) *Updater {
	return &Updater{
		app: app,
	}
}

// CheckForUpdates checks if a new version is available on GitHub
func (u *Updater) CheckForUpdates() UpdateCheckResult {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
	resp, err := http.Get(url)
	if err != nil {
		return UpdateCheckResult{Error: fmt.Sprintf("Failed to check for updates: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UpdateCheckResult{Error: fmt.Sprintf("GitHub API returned status: %s", resp.Status)}
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return UpdateCheckResult{Error: fmt.Sprintf("Failed to decode release info: %v", err)}
	}

	// Simple version comparison (lexicographical for now, assuming vX.Y.Z format)
	// In a real app, use semver library
	remoteVer := strings.TrimPrefix(release.TagName, "v")
	localVer := strings.TrimPrefix(version.App.Version, "v")

	if remoteVer > localVer {
		return UpdateCheckResult{
			Available: true,
			Version:   release.TagName,
			Body:      release.Body,
		}
	}

	return UpdateCheckResult{Available: false, Version: version.App.Version}
}

// DoUpdate downloads the new version side-by-side and launches it
func (u *Updater) DoUpdate(version string) error {
	// 1. Fetch release info
	// url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", RepoOwner, RepoName, version)
	// Actually we need to fetch the release info to get the assets.
	// The previous code was:
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", RepoOwner, RepoName, version)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode release info: %w", err)
	}

	// 2. Find the correct asset
	var assetUrl string
	targetSuffix := ".exe"
	if runtime.GOOS != "windows" {
		targetSuffix = ""
	}

	for _, asset := range release.Assets {
		if strings.HasSuffix(strings.ToLower(asset.Name), targetSuffix) {
			assetUrl = asset.BrowserDownloadUrl
			break
		}
	}

	if assetUrl == "" {
		return fmt.Errorf("no suitable asset found for version %s", version)
	}

	// 3. Determine target path
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(ex)
	newExePath := filepath.Join(exeDir, "lce_new.exe")

	// Check if we are already running this version (unlikely if we got here, but good sanity check)
	if newExePath == ex {
		return fmt.Errorf("already running version %s", version)
	}

	// 4. Download the asset
	u.app.Event.Emit("update:progress", map[string]interface{}{"status": "downloading", "percent": 0})
	
	resp, err = http.Get(assetUrl)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	// Create new file
	out, err := os.Create(newExePath)
	if err != nil {
		return fmt.Errorf("failed to create new executable: %w", err)
	}
	defer out.Close()

	// Wrap body in ProgressReader
	reader := &ProgressReader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
		OnProgress: func(p float64) {
			u.app.Event.Emit("update:progress", map[string]interface{}{"status": "downloading", "percent": p})
		},
	}

	// Copy content
	if _, err := io.Copy(out, reader); err != nil {
		return fmt.Errorf("failed to write new executable: %w", err)
	}

	// 5. Launch new executable and quit
	u.app.Event.Emit("update:progress", map[string]interface{}{"status": "ready", "percent": 100})
	
	// Store path for restart
	u.newExePath = newExePath
	return nil
}

// RestartApp launches the new executable and quits the current one
func (u *Updater) RestartApp() error {
	if u.newExePath == "" {
		return fmt.Errorf("no update ready to install")
	}

	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(ex)
	oldExePath := filepath.Join(exeDir, "lce_old.exe")
	targetExePath := filepath.Join(exeDir, "lce.exe")

	// 1. Remove old backup if exists
	// We ignore error because it might not exist
	_ = os.Remove(oldExePath)

	// 2. Rename current to old
	// Windows allows renaming the running executable
	if err := os.Rename(ex, oldExePath); err != nil {
		return fmt.Errorf("failed to rename current executable to lce_old.exe: %w", err)
	}

	// 3. Rename new to target (lce.exe)
	if err := os.Rename(u.newExePath, targetExePath); err != nil {
		// Try to rollback: rename old back to current
		_ = os.Rename(oldExePath, ex)
		return fmt.Errorf("failed to rename new executable to lce.exe: %w", err)
	}

	// 4. Launch target
	cmd := exec.Command(targetExePath)
	if err := cmd.Start(); err != nil {
		// Try to rollback is hard here because we already renamed things.
		// But at least we have lce.exe on disk.
		return fmt.Errorf("failed to start new version: %w", err)
	}

	u.app.Quit()
	return nil
}

// ProgressReader tracks reading progress
type ProgressReader struct {
	io.Reader
	Total      int64
	Current    int64
	OnProgress func(float64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.Current += int64(n)
	if pr.Total > 0 && pr.OnProgress != nil {
		percent := float64(pr.Current) / float64(pr.Total) * 100
		pr.OnProgress(percent)
	}
	return n, err
}

// CleanupOldExecutables removes temporary update files
func CleanupOldExecutables() {
	ex, err := os.Executable()
	if err != nil {
		return
	}
	exeDir := filepath.Dir(ex)
	
	// Clean up lce_new.exe if it exists (failed update leftover)
	newExePath := filepath.Join(exeDir, "lce_new.exe")
	if _, err := os.Stat(newExePath); err == nil {
		_ = os.Remove(newExePath)
	}

	// Clean up lce_old.exe (previous version backup)
	oldExePath := filepath.Join(exeDir, "lce_old.exe")
	if _, err := os.Stat(oldExePath); err == nil {
		err := os.Remove(oldExePath)
		if err != nil {
			// One retry attempt
			time.Sleep(2 * time.Second)
			_ = os.Remove(oldExePath)
		}
	}
}
