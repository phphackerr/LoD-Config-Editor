//go:build !windows

package taskbar

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Taskbar states
const (
	TBPF_NOPROGRESS    = 0
	TBPF_INDETERMINATE = 0x1
	TBPF_NORMAL        = 0x2
	TBPF_ERROR         = 0x4
	TBPF_PAUSED        = 0x8
)

type TaskbarUtils struct{}

func NewTaskbarUtils() *TaskbarUtils {
	return &TaskbarUtils{}
}

func (t *TaskbarUtils) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	return nil
}

func (t *TaskbarUtils) SetTaskbarProgress(completed, total uint64) {}

func (t *TaskbarUtils) SetTaskbarError() {}

func (t *TaskbarUtils) SetTaskbarPaused(completed, total uint64) {}

func (t *TaskbarUtils) SetTaskbarCompleteAndFlash() {}
