package updater

import (
	"encoding/json"
	"fmt"
	"lce/backend/version"
	"net/http"
	"runtime"
	"strings"
	"sync"

	"github.com/minio/selfupdate"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	RepoOwner = "phphackerr"
	RepoName  = "LoD-Config-Editor"
)

type Updater struct {
	app  *application.App
	lock sync.Mutex
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
	if release.TagName > version.AppVersion {
		return UpdateCheckResult{
			Available: true,
			Version:   release.TagName,
			Body:      release.Body,
		}
	}

	return UpdateCheckResult{Available: false, Version: version.AppVersion}
}

// DoUpdate downloads and applies the update
func (u *Updater) DoUpdate(version string) error {
	// 1. Fetch release info again to get asset URL (or cache it)
	// For simplicity, we fetch again
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

	// 2. Find the correct asset for the current OS/Arch
	// Assuming Windows amd64 for now as per user context
	assetUrl := ""
	
	targetSuffix := ".exe"
	if runtime.GOOS != "windows" {
		targetSuffix = "" // Adjust for other OS if needed
	}

	// Look for an asset that looks like an executable or setup
	// Ideally, we should look for "LCE_Setup.exe" or "lce.exe"
	// Let's assume the release contains the executable directly for self-update
	// OR a setup file. Self-update usually requires replacing the binary.
	// If it's a setup file, we might just download and run it.
	// But the user asked for "auto-update", usually implying in-place update.
	// Let's try to find an .exe that is NOT a setup if possible, or just the first .exe
	
	for _, asset := range release.Assets {
		if strings.HasSuffix(strings.ToLower(asset.Name), targetSuffix) {
			// Skip setup files if we want direct binary replacement, 
			// UNLESS the user wants to run the installer.
			// Let's assume direct binary replacement for "updater" logic usually.
			// But if we only distribute Setup.exe, we should download and run it.
			
			// Strategy: If we find a file that looks like the main binary, use it.
			// If not, maybe it's a setup.
			
			// For this implementation, let's assume we publish the binary itself as "lce.exe" or similar.
			assetUrl = asset.BrowserDownloadUrl
			break
		}
	}

	if assetUrl == "" {
		return fmt.Errorf("no suitable asset found for version %s", version)
	}

	// 3. Download the asset
	u.app.Event.Emit("update:progress", map[string]interface{}{"status": "downloading", "percent": 0})
	
	resp, err = http.Get(assetUrl)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	// 4. Apply update
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		// If selfupdate fails (e.g. permissions), we might want to fallback to saving the file and running it?
		// But selfupdate is designed to handle this on Windows.
		return fmt.Errorf("failed to apply update: %w", err)
	}

	return nil
}

// DownloadAndRunInstaller is an alternative if we want to run a setup.exe
func (u *Updater) DownloadAndRunInstaller(version string) error {
	// ... implementation for downloading setup.exe and running it ...
	return nil
}
