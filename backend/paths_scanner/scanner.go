//go:build windows

package paths_scanner

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows/registry"
)

const (
	defaultMaxDepth       = 5
	defaultScanTimeout    = 45 * time.Second
	defaultScanWorkers    = 4
	defaultMaxMatchesRoot = 10
)

var (
	errMaxMatchesReached = errors.New("max matches per root reached")
)

type Scanner struct {
	maxDepth          int
	timeout           time.Duration
	workerCount       int
	maxMatchesPerRoot int
	excludedFolders   map[string]struct{}
}

func NewScanner() *Scanner {
	return &Scanner{
		maxDepth:          defaultMaxDepth,
		timeout:           defaultScanTimeout,
		workerCount:       defaultScanWorkers,
		maxMatchesPerRoot: defaultMaxMatchesRoot,
		excludedFolders: map[string]struct{}{
			"windows":                   {},
			"users":                     {},
			"programdata":               {},
			"system volume information": {},
		},
	}
}

func isTargetFile(filename string) bool {
	lower := strings.ToLower(filename)
	return lower == "config.lod.ini" || lower == "war3.exe"
}

func (s *Scanner) FindConfigOrExeParallel() []string {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.findConfigOrExeWithContext(ctx)
}

func (s *Scanner) findConfigOrExeWithContext(ctx context.Context) []string {
	drives := getLogicalDrives()
	if len(drives) == 0 {
		return []string{}
	}

	workers := s.workerCount
	if workers <= 0 {
		workers = defaultScanWorkers
	}
	if workers > len(drives) {
		workers = len(drives)
	}

	jobs := make(chan string)
	results := make(chan []string, len(drives))

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for drive := range jobs {
			matches, err := s.findInstallPathsInRoot(ctx, drive)
			if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
				log.Printf("paths_scanner: failed scanning %s: %v", drive, err)
			}
			if len(matches) > 0 {
				results <- matches
			}
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

	go func() {
		defer close(jobs)
		for _, drive := range drives {
			select {
			case <-ctx.Done():
				return
			case jobs <- drive:
			}
		}
	}()

	wg.Wait()
	close(results)

	collected := make([]string, 0, len(drives))
	for matched := range results {
		collected = append(collected, matched...)
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		log.Printf("paths_scanner: scan timed out after %s, returning partial results", s.timeout)
	}

	return uniqueSortedPaths(collected)
}

func (s *Scanner) findInstallPathsInRoot(ctx context.Context, root string) ([]string, error) {
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return nil, nil
	}

	maxDepth := s.maxDepth
	if maxDepth <= 0 {
		maxDepth = defaultMaxDepth
	}

	maxMatches := s.maxMatchesPerRoot
	if maxMatches <= 0 {
		maxMatches = defaultMaxMatchesRoot
	}

	var found []string
	foundSet := make(map[string]struct{}, 4)

	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if walkErr != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		depth, err := depthFromRoot(root, path)
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if depth > maxDepth {
				return filepath.SkipDir
			}
			if depth > 0 {
				if _, skip := s.excludedFolders[strings.ToLower(d.Name())]; skip {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if depth > maxDepth {
			return nil
		}
		if !d.Type().IsRegular() || !isTargetFile(d.Name()) {
			return nil
		}

		dir := filepath.Dir(path)
		key := normalizePathKey(dir)
		if _, exists := foundSet[key]; exists {
			return nil
		}

		foundSet[key] = struct{}{}
		found = append(found, dir)

		if len(found) >= maxMatches {
			return errMaxMatchesReached
		}
		return nil
	})

	switch {
	case walkErr == nil:
	case errors.Is(walkErr, errMaxMatchesReached):
	case errors.Is(walkErr, context.Canceled), errors.Is(walkErr, context.DeadlineExceeded):
		return found, walkErr
	default:
		return found, walkErr
	}

	return found, nil
}

func getLogicalDrives() []string {
	drives := make([]string, 0, 8)
	for char := 'A'; char <= 'Z'; char++ {
		drive := fmt.Sprintf("%c:\\", char)
		if _, err := os.Stat(drive); err == nil {
			drives = append(drives, drive)
		}
	}
	return drives
}

func findPathInRegistry() (string, bool) {
	const regPath = `SOFTWARE\Blizzard Entertainment\Warcraft III`

	hives := []registry.Key{
		registry.LOCAL_MACHINE,
		registry.CURRENT_USER,
	}

	for _, hive := range hives {
		k, err := registry.OpenKey(hive, regPath, registry.READ)
		if err != nil {
			continue
		}

		installPath, _, err := k.GetStringValue("InstallPath")
		_ = k.Close()
		if err != nil {
			continue
		}

		installPath = strings.TrimSpace(installPath)
		if installPath == "" {
			continue
		}

		clean := filepath.Clean(installPath)
		if hasInstallFiles(clean) {
			return clean, true
		}
	}

	return "", false
}

func hasInstallFiles(path string) bool {
	_, errConfig := os.Stat(filepath.Join(path, "config.lod.ini"))
	_, errExe := os.Stat(filepath.Join(path, "war3.exe"))
	return errConfig == nil || errExe == nil
}

// CheckAndFindPaths scans registry + disks and returns unique sorted install paths.
func (s *Scanner) CheckAndFindPaths() ([]string, error) {
	collected := make([]string, 0, 8)

	if registryPath, ok := findPathInRegistry(); ok {
		collected = append(collected, registryPath)
	}

	collected = append(collected, s.FindConfigOrExeParallel()...)
	return uniqueSortedPaths(collected), nil
}
