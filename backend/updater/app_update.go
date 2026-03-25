package updater

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"lce/backend/version"
	"lce/backend/versioncmp"
)

const (
	minExecutableBytes = 512 * 1024 // 512 KiB sanity floor for PE executable
)

// CheckForUpdates checks if a new app version is available on GitHub.
func (u *Updater) CheckForUpdates() UpdateCheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), httpRequestTimeout)
	defer cancel()

	release, err := u.fetchLatestRelease(ctx)
	if err != nil {
		return UpdateCheckResult{Error: fmt.Sprintf("failed to check for updates: %v", err)}
	}

	if versioncmp.Compare(release.TagName, version.App.Version) > 0 {
		return UpdateCheckResult{
			Available: true,
			Version:   release.TagName,
			Body:      release.Body,
		}
	}

	return UpdateCheckResult{
		Available: false,
		Version:   version.App.Version,
	}
}

// DoUpdate downloads a new executable side-by-side and marks update as ready.
func (u *Updater) DoUpdate(targetVersion string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("in-app update is supported only on Windows")
	}

	targetVersion = strings.TrimSpace(targetVersion)
	if targetVersion == "" {
		return fmt.Errorf("target version is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*httpRequestTimeout)
	defer cancel()

	release, err := u.fetchReleaseByTag(ctx, targetVersion)
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}

	asset, err := selectExecutableAsset(release.Assets)
	if err != nil {
		return err
	}

	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeDir := filepath.Dir(ex)
	newExePath := filepath.Join(exeDir, "lce_new.exe")

	checksumMap, _ := u.fetchChecksumMap(ctx, release.Assets)

	u.emitProgress("downloading", 0)
	_, err = u.downloadToFile(ctx, asset.BrowserDownloadURL, newExePath, updateAssetMaxBytes, func(percent float64) {
		u.emitProgress("downloading", percent)
	})
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	sum, err := verifyDownloadedExecutable(newExePath, minExecutableBytes)
	if err != nil {
		_ = os.Remove(newExePath)
		return fmt.Errorf("downloaded executable validation failed: %w", err)
	}

	if expected, ok := checksumMap[asset.Name]; ok && expected != "" {
		if !strings.EqualFold(sum, expected) {
			_ = os.Remove(newExePath)
			return fmt.Errorf("checksum mismatch for %s", asset.Name)
		}
	}

	u.mu.Lock()
	u.newExePath = newExePath
	u.mu.Unlock()

	u.emitProgress("ready", 100)
	return nil
}

// RestartApp launches the new executable and quits current process.
func (u *Updater) RestartApp() error {
	u.mu.Lock()
	newExePath := u.newExePath
	u.mu.Unlock()

	if newExePath == "" {
		return fmt.Errorf("no update ready to install")
	}

	if _, err := verifyDownloadedExecutable(newExePath, minExecutableBytes); err != nil {
		return fmt.Errorf("prepared update is invalid: %w", err)
	}

	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	exeDir := filepath.Dir(ex)
	oldExePath := filepath.Join(exeDir, "lce_old.exe")
	targetExePath := filepath.Join(exeDir, "lce.exe")

	_ = os.Remove(oldExePath)

	if err := os.Rename(ex, oldExePath); err != nil {
		return fmt.Errorf("failed to rename current executable: %w", err)
	}

	if err := os.Rename(newExePath, targetExePath); err != nil {
		_ = os.Rename(oldExePath, ex)
		return fmt.Errorf("failed to activate new executable: %w", err)
	}

	cmd := exec.Command(targetExePath)
	if err := cmd.Start(); err != nil {
		// Best effort rollback.
		_ = os.Rename(targetExePath, newExePath)
		_ = os.Rename(oldExePath, ex)
		return fmt.Errorf("failed to launch new version: %w", err)
	}

	u.mu.Lock()
	u.newExePath = ""
	u.mu.Unlock()

	if u.app != nil {
		u.app.Quit()
	}
	return nil
}

// CleanupOldExecutables removes temporary update files.
func CleanupOldExecutables() {
	ex, err := os.Executable()
	if err != nil {
		return
	}
	exeDir := filepath.Dir(ex)

	newExePath := filepath.Join(exeDir, "lce_new.exe")
	if _, err := os.Stat(newExePath); err == nil {
		_ = os.Remove(newExePath)
	}

	oldExePath := filepath.Join(exeDir, "lce_old.exe")
	if _, err := os.Stat(oldExePath); err == nil {
		if err := os.Remove(oldExePath); err != nil {
			time.Sleep(oldExeCleanupBackoff)
			_ = os.Remove(oldExePath)
		}
	}
}

func (u *Updater) fetchLatestRelease(ctx context.Context) (ReleaseInfo, error) {
	var release ReleaseInfo
	err := u.fetchJSON(ctx, githubAPIBaseURL+"/releases/latest", &release)
	if err != nil {
		return ReleaseInfo{}, err
	}
	return release, nil
}

func (u *Updater) fetchReleaseByTag(ctx context.Context, tag string) (ReleaseInfo, error) {
	var release ReleaseInfo
	endpoint := githubAPIBaseURL + "/releases/tags/" + url.PathEscape(tag)
	err := u.fetchJSON(ctx, endpoint, &release)
	if err != nil {
		return ReleaseInfo{}, err
	}
	return release, nil
}

func selectExecutableAsset(assets []ReleaseAsset) (ReleaseAsset, error) {
	if len(assets) == 0 {
		return ReleaseAsset{}, fmt.Errorf("release has no assets")
	}

	// Prefer portable executable, avoid installers/setups if possible.
	candidates := make([]ReleaseAsset, 0, len(assets))
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.HasSuffix(name, ".exe") {
			candidates = append(candidates, asset)
		}
	}
	if len(candidates) == 0 {
		return ReleaseAsset{}, fmt.Errorf("no executable asset found in release")
	}

	slices.SortStableFunc(candidates, func(a, b ReleaseAsset) int {
		return compareAssetPriority(a.Name, b.Name)
	})

	return candidates[0], nil
}

func compareAssetPriority(a, b string) int {
	score := func(name string) int {
		n := strings.ToLower(name)
		s := 0
		if strings.Contains(n, "setup") || strings.Contains(n, "installer") {
			s += 10
		}
		if strings.Contains(n, "portable") {
			s -= 5
		}
		if strings.Contains(n, "lce") {
			s -= 1
		}
		return s
	}

	sa := score(a)
	sb := score(b)
	if sa < sb {
		return -1
	}
	if sa > sb {
		return 1
	}
	if strings.ToLower(a) < strings.ToLower(b) {
		return -1
	}
	if strings.ToLower(a) > strings.ToLower(b) {
		return 1
	}
	return 0
}

func (u *Updater) fetchChecksumMap(ctx context.Context, assets []ReleaseAsset) (map[string]string, error) {
	checksumAsset, ok := findChecksumAsset(assets)
	if !ok {
		return map[string]string{}, nil
	}

	data, err := u.fetchBytes(ctx, checksumAsset.BrowserDownloadURL, checksumMaxBytes)
	if err != nil {
		return nil, err
	}

	return parseChecksumFile(string(data)), nil
}

func findChecksumAsset(assets []ReleaseAsset) (ReleaseAsset, bool) {
	for _, asset := range assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, "sha256") || strings.Contains(name, "checksum") {
			if strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".sha256") || strings.HasSuffix(name, ".sha256sum") {
				return asset, true
			}
		}
	}
	return ReleaseAsset{}, false
}

func parseChecksumFile(content string) map[string]string {
	result := make(map[string]string)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: "<sha256>  <file>"
		fields := strings.Fields(line)
		if len(fields) >= 2 && len(fields[0]) == 64 {
			result[filepath.Base(strings.TrimPrefix(fields[1], "*"))] = strings.ToLower(fields[0])
			continue
		}

		// Format: "<file>: <sha256>"
		if idx := strings.Index(line, ":"); idx > 0 {
			name := strings.TrimSpace(line[:idx])
			hash := strings.TrimSpace(line[idx+1:])
			hashFields := strings.Fields(hash)
			if len(hashFields) == 0 {
				continue
			}
			hash = hashFields[0]
			if len(hash) == 64 {
				result[filepath.Base(name)] = strings.ToLower(hash)
			}
		}
	}

	return result
}
