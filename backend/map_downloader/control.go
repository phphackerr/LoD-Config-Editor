package map_downloader

import (
	"context"
	"errors"
	"fmt"
)

var (
	errDownloadInProgress = errors.New("загрузка карты уже выполняется")
	errNoActiveDownload   = errors.New("активная загрузка карты не найдена")
	errDownloadStopped    = errors.New("загрузка карты остановлена пользователем")
)

type downloadControlState struct {
	active bool
	paused bool
	stop   bool
	cancel context.CancelFunc
}

func (md *MapDownloader) beginDownloadControl() (context.Context, error) {
	md.downloadMu.Lock()
	defer md.downloadMu.Unlock()

	if md.downloadCtrl.active {
		return nil, errDownloadInProgress
	}

	ctx, cancel := context.WithCancel(context.Background())
	md.downloadCtrl = downloadControlState{
		active: true,
		cancel: cancel,
	}

	return ctx, nil
}

func (md *MapDownloader) endDownloadControl() {
	md.downloadMu.Lock()
	ctrl := md.downloadCtrl
	md.downloadCtrl = downloadControlState{}
	if md.downloadCond != nil {
		md.downloadCond.Broadcast()
	}
	md.downloadMu.Unlock()

	if ctrl.cancel != nil {
		ctrl.cancel()
	}
}

func (md *MapDownloader) waitWhilePaused() error {
	md.downloadMu.Lock()
	defer md.downloadMu.Unlock()

	for md.downloadCtrl.active && md.downloadCtrl.paused && !md.downloadCtrl.stop {
		md.downloadCond.Wait()
	}

	if md.downloadCtrl.stop {
		return errDownloadStopped
	}

	if !md.downloadCtrl.active {
		return errNoActiveDownload
	}

	return nil
}

func (md *MapDownloader) isStopRequested() bool {
	md.downloadMu.Lock()
	defer md.downloadMu.Unlock()
	return md.downloadCtrl.stop
}

// PauseDownload приостанавливает текущую загрузку карты.
func (md *MapDownloader) PauseDownload() error {
	md.downloadMu.Lock()
	defer md.downloadMu.Unlock()

	if !md.downloadCtrl.active {
		return errNoActiveDownload
	}

	if md.downloadCtrl.stop {
		return errDownloadStopped
	}

	md.downloadCtrl.paused = true
	return nil
}

// ResumeDownload возобновляет ранее приостановленную загрузку карты.
func (md *MapDownloader) ResumeDownload() error {
	md.downloadMu.Lock()
	defer md.downloadMu.Unlock()

	if !md.downloadCtrl.active {
		return errNoActiveDownload
	}

	if md.downloadCtrl.stop {
		return errDownloadStopped
	}

	md.downloadCtrl.paused = false
	md.downloadCond.Broadcast()
	return nil
}

// StopDownload полностью останавливает текущую загрузку карты.
func (md *MapDownloader) StopDownload() error {
	md.downloadMu.Lock()
	if !md.downloadCtrl.active {
		md.downloadMu.Unlock()
		return nil
	}

	if md.downloadCtrl.stop {
		md.downloadMu.Unlock()
		return nil
	}

	md.downloadCtrl.stop = true
	md.downloadCtrl.paused = false
	cancel := md.downloadCtrl.cancel
	md.downloadCond.Broadcast()
	md.downloadMu.Unlock()

	if cancel != nil {
		cancel()
	}

	return nil
}

func isStoppedError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, errDownloadStopped)
}

func asStopAwareError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return errDownloadStopped
	}
	return err
}

func stopAwareWrap(prefix string, err error) error {
	if err == nil {
		return nil
	}

	err = asStopAwareError(err)
	if isStoppedError(err) {
		return err
	}

	return fmt.Errorf("%s: %w", prefix, err)
}
