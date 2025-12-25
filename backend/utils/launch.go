package utils

import (
	"encoding/json"
	"fmt"
	"lce/backend/version"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type Utils struct {
	window *application.WebviewWindow
}

func NewUtils(window *application.WebviewWindow) *Utils {
	return &Utils{
		window: window,
	}
}

func (u *Utils) OpenDevTools() {
	u.window.OpenDevTools()
}

// Открыть папку в проводнике
func (g *Utils) OpenFolderInExplorer(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("ошибка проверки пути: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("путь '%s' не является папкой", path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("ошибка получения абсолютного пути: %w", err)
	}

	cmd := exec.Command("explorer", absPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ошибка запуска explorer: %w", err)
	}
	return nil
}

// Запуск игры с аргументами
func (g *Utils) LaunchGameExe(path string, args ...string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("ошибка проверки пути: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("указанный путь '%s' не является папкой", path)
	}

	var war3Path, frozenPath string

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("ошибка чтения директории: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if name == "war3.exe" {
			war3Path = filepath.Join(path, entry.Name())
			break
		} else if name == "frozen throne.exe" {
			frozenPath = filepath.Join(path, entry.Name())
		}
	}

	var exeToLaunch string
	if war3Path != "" {
		exeToLaunch = war3Path
	} else if frozenPath != "" {
		exeToLaunch = frozenPath
	} else {
		return fmt.Errorf("файлы 'war3.exe' или 'Frozen Throne.exe' не найдены в '%s'", path)
	}

	cmd := exec.Command(exeToLaunch, args...)
	cmd.Dir = path

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ошибка запуска '%s': %w", exeToLaunch, err)
	}

	return nil
}

func (g *Utils) OpenFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("ошибка проверки пути: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("путь '%s' является папкой, а не файлом", path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("ошибка получения абсолютного пути: %w", err)
	}

	cmd := exec.Command("cmd", "/C", "start", "", absPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ошибка запуска команды: %w", err)
	}
	return nil
}

func (g *Utils) GetAppVersion() string {
	return version.App.Version
}

func (g *Utils) OpenURL(url string) error {
	cmd := exec.Command("cmd", "/C", "start", "", url)
	return cmd.Start()
}

type DiscordInvite struct {
	ApproximatePresenceCount int `json:"approximate_presence_count"`
	ApproximateMemberCount   int `json:"approximate_member_count"`
}

func (g *Utils) GetDiscordStats(inviteCode string) (*DiscordInvite, error) {
	url := fmt.Sprintf("https://discord.com/api/v9/invites/%s?with_counts=true", inviteCode)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord api returned status: %d", resp.StatusCode)
	}

	var invite DiscordInvite
	if err := json.NewDecoder(resp.Body).Decode(&invite); err != nil {
		return nil, err
	}

	return &invite, nil
}
