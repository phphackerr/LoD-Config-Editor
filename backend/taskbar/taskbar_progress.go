package taskbar

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	ole32                = syscall.NewLazyDLL("ole32.dll")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")

	user32            = syscall.NewLazyDLL("user32.dll")
	procFlashWindowEx = user32.NewProc("FlashWindowEx")
)

// Taskbar states
const (
	TBPF_NOPROGRESS    = 0
	TBPF_INDETERMINATE = 0x1
	TBPF_NORMAL        = 0x2
	TBPF_ERROR         = 0x4
	TBPF_PAUSED        = 0x8
)

// Flash flags
const (
	FLASHW_STOP      = 0
	FLASHW_CAPTION   = 0x00000001
	FLASHW_TRAY      = 0x00000002
	FLASHW_ALL       = FLASHW_CAPTION | FLASHW_TRAY
	FLASHW_TIMER     = 0x00000004
	FLASHW_TIMERNOFG = 0x0000000C
)

type FLASHWINFO struct {
	cbSize    uint32
	hwnd      syscall.Handle
	dwFlags   uint32
	uCount    uint32
	dwTimeout uint32
}

// vtable для ITaskbarList3
type iTaskbarList3Vtbl struct {
	QueryInterface       uintptr
	AddRef               uintptr
	Release              uintptr
	HrInit               uintptr
	AddTab               uintptr
	DeleteTab            uintptr
	ActivateTab          uintptr
	SetActiveAlt         uintptr
	MarkFullscreenWindow uintptr
	SetProgressValue     uintptr
	SetProgressState     uintptr
}

type ITaskbarList3 struct {
	lpVtbl *iTaskbarList3Vtbl
}

type TaskbarUtils struct {
	taskbar *ITaskbarList3
}

func NewTaskbarUtils() *TaskbarUtils {
	return &TaskbarUtils{}
}

func (t *TaskbarUtils) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	obj, err := createTaskbarList()
	if err != nil {
		return fmt.Errorf("не удалось инициализировать TaskbarList: %w", err)
	}
	t.taskbar = obj
	return nil
}

func (t *TaskbarUtils) SetTaskbarProgress(completed, total uint64) {
	if t.taskbar == nil {
		return
	}

	win := application.Get().Window.Current().NativeWindow()
	hwnd := (syscall.Handle)(win)

	if total == 0 {
		t.callSetProgressState(hwnd, TBPF_NOPROGRESS)
	} else {
		t.callSetProgressState(hwnd, TBPF_NORMAL)
		t.callSetProgressValue(hwnd, completed, total)
	}
}

func (t *TaskbarUtils) SetTaskbarError() {
	if t.taskbar == nil {
		return
	}
	win := application.Get().Window.Current().NativeWindow()
	hwnd := (syscall.Handle)(win)
	t.callSetProgressState(hwnd, TBPF_ERROR)
}

func (t *TaskbarUtils) SetTaskbarPaused(completed, total uint64) {
	if t.taskbar == nil {
		return
	}
	win := application.Get().Window.Current().NativeWindow()
	hwnd := (syscall.Handle)(win)
	t.callSetProgressState(hwnd, TBPF_PAUSED)
	t.callSetProgressValue(hwnd, completed, total)
}

// --- новый метод ---
func (t *TaskbarUtils) SetTaskbarCompleteAndFlash() {
	if t.taskbar == nil {
		return
	}
	win := application.Get().Window.Current().NativeWindow()
	hwnd := (syscall.Handle)(win)

	// зелёный прогресс 100%
	t.callSetProgressState(hwnd, TBPF_NORMAL)
	t.callSetProgressValue(hwnd, 100, 100)

	// мигание в панели задач
	f := FLASHWINFO{
		cbSize:    uint32(unsafe.Sizeof(FLASHWINFO{})),
		hwnd:      hwnd,
		dwFlags:   FLASHW_TRAY | FLASHW_TIMERNOFG,
		uCount:    5, // сколько раз мигнёт
		dwTimeout: 0,
	}
	procFlashWindowEx.Call(uintptr(unsafe.Pointer(&f)))
}

// --- low-level COM calls ---

func createTaskbarList() (*ITaskbarList3, error) {
	clsid, _ := ole.CLSIDFromString("{56FDF344-FD6D-11d0-958A-006097C9A090}") // CLSID_TaskbarList
	iid, _ := ole.IIDFromString("{EA1AFB91-9E28-4B86-90E9-9E9F8A5EEFAF}")     // IID_ITaskbarList3

	unknown, err := ole.CreateInstance(clsid, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateInstance failed: %w", err)
	}

	taskbar, err := unknown.QueryInterface(iid)
	if err != nil {
		return nil, fmt.Errorf("QueryInterface failed: %w", err)
	}

	return (*ITaskbarList3)(unsafe.Pointer(taskbar)), nil
}

func (t *TaskbarUtils) callSetProgressValue(hwnd syscall.Handle, completed, total uint64) {
	syscall.SyscallN(
		t.taskbar.lpVtbl.SetProgressValue,
		uintptr(unsafe.Pointer(t.taskbar)),
		uintptr(hwnd),
		uintptr(completed),
		uintptr(total),
		0, 0,
	)
}

func (t *TaskbarUtils) callSetProgressState(hwnd syscall.Handle, state uint32) {
	syscall.SyscallN(
		t.taskbar.lpVtbl.SetProgressState,
		uintptr(unsafe.Pointer(t.taskbar)),
		uintptr(hwnd),
		uintptr(state),
	)
}
