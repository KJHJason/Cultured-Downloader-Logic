package progress

import (
	"context"
	"sync"
)

type DownloadProgressBar struct {
	msg        string
	successMsg string
	errMsg     string
	hasError   bool
	finished   bool

	totalBytes int64
	percentage int // -1 if unknown, 0-100 otherwise if there's a known ETA

	filename      string
	downloadSpeed float64
	downloadETA   float64
	mu            sync.RWMutex
}

func NewDlProgressBar(ctx context.Context, messages Messages) *DownloadProgressBar {
	return &DownloadProgressBar{
		msg:        messages.Msg,
		successMsg: messages.SuccessMsg,
		errMsg:     messages.ErrMsg,
		hasError:   false,
		finished:   false,

		totalBytes: 0,
		percentage: 0,

		filename:      "",
		downloadSpeed: 0,
		downloadETA:   -1,
		mu:            sync.RWMutex{},
	}
}

func (dlP *DownloadProgressBar) UpdatePercentage(percentage int) {
	dlP.mu.Lock()
	defer dlP.mu.Unlock()

	dlP.percentage = percentage
}

func (dlP *DownloadProgressBar) GetPercentage() int {
	dlP.mu.RLock()
	defer dlP.mu.RUnlock()

	return dlP.percentage
}

func (dlP *DownloadProgressBar) UpdateFilename(filename string) {
	dlP.mu.Lock()
	defer dlP.mu.Unlock()

	dlP.filename = filename
}

func (dlP *DownloadProgressBar) GetFilename() string {
	dlP.mu.RLock()
	defer dlP.mu.RUnlock()

	return dlP.filename
}

func (dlP *DownloadProgressBar) UpdateDownloadSpeed(speed float64) {
	dlP.mu.Lock()
	defer dlP.mu.Unlock()

	dlP.downloadSpeed = speed
}

func (dlP *DownloadProgressBar) GetDownloadSpeed() float64 {
	dlP.mu.RLock()
	defer dlP.mu.RUnlock()

	return dlP.downloadSpeed
}

func (dlP *DownloadProgressBar) UpdateDownloadETA(eta float64) {
	dlP.mu.Lock()
	defer dlP.mu.Unlock()

	dlP.downloadETA = eta
}

func (dlP *DownloadProgressBar) GetDownloadETA() float64 {
	dlP.mu.RLock()
	defer dlP.mu.RUnlock()

	return dlP.downloadETA
}

func (dlP *DownloadProgressBar) UpdateTotalBytes(bytes int64) {
	dlP.mu.Lock()
	defer dlP.mu.Unlock()

	dlP.totalBytes = bytes
}

func (dlP *DownloadProgressBar) GetTotalBytes() int64 {
	dlP.mu.RLock()
	defer dlP.mu.RUnlock()

	return dlP.totalBytes
}

func (dlP *DownloadProgressBar) Stop(err bool) {
	dlP.mu.Lock()
	defer dlP.mu.Unlock()

	dlP.finished = true
	dlP.hasError = err
}

func (dlP *DownloadProgressBar) IsFinished() bool {
	dlP.mu.RLock()
	defer dlP.mu.RUnlock()

	return dlP.finished
}

func (dlP *DownloadProgressBar) HasError() bool {
	dlP.mu.RLock()
	defer dlP.mu.RUnlock()

	return dlP.hasError
}

func (dlP *DownloadProgressBar) UpdateErrMsg(errMsg string) {
	dlP.mu.Lock()
	defer dlP.mu.Unlock()

	dlP.errMsg = errMsg
}

func (dlP *DownloadProgressBar) GetErrMsg() string {
	dlP.mu.RLock()
	defer dlP.mu.RUnlock()

	return dlP.errMsg
}

func (dlP *DownloadProgressBar) UpdateSuccessMsg(successMsg string) {
	dlP.mu.Lock()
	defer dlP.mu.Unlock()

	dlP.successMsg = successMsg
}

func (dlP *DownloadProgressBar) GetSuccessMsg() string {
	dlP.mu.RLock()
	defer dlP.mu.RUnlock()

	return dlP.successMsg
}

func (dlP *DownloadProgressBar) UpdateMsg(msg string) {
	dlP.mu.Lock()
	defer dlP.mu.Unlock()

	dlP.msg = msg
}

func (dlP *DownloadProgressBar) GetMsg() string {
	dlP.mu.RLock()
	defer dlP.mu.RUnlock()

	return dlP.msg
}
