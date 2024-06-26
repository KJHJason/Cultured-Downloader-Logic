package progress

import (
	"sync"
)

type ProgressBar interface {
	// Add adds the given number to the progress.
	Add(int)

	// Start starts the progress.
	Start()

	// Stop stops the progress.
	// Requires a bool to indicate if there is an error or not after the progress is stopped.
	Stop(bool)

	// Change to spinner, meaning there is no clear indicator of the progress.
	SetToSpinner()
	GetIsSpinner() bool

	// Change to progress bar, meaning there is a clear indicator of the progress.
	SetToProgressBar()
	GetIsProgBar() bool

	// Stops the progress due to an interrupt signal.
	StopInterrupt(string)

	// UpdateBaseMsg updates the base message of the progress.
	UpdateBaseMsg(string)

	// UpdateMax updates the maximum value of the progress.
	UpdateMax(int)

	// Increment increments the progress by 1.
	Increment()

	// Updates the success message of the progress.
	UpdateSuccessMsg(string)

	// Updates the message of the progress in the event of an error.
	UpdateErrorMsg(string)

	// Mainly for frontend usage.
	SnapshotTask()           // Saves the current progress state for nested progress bars.
	MakeLatestSnapshotMain() // Makes the latest snapshot the main progress bar.
}

type ProgressBarInfo struct {
	MainProgressBar      ProgressBar
	DownloadProgressBars *[]*DownloadProgressBar

	mu sync.Mutex
}

func (pgi *ProgressBarInfo) AppendDlProgBar(progBar *DownloadProgressBar) {
	pgi.mu.Lock()
	defer pgi.mu.Unlock()

	if pgi.DownloadProgressBars == nil {
		return
	}
	*pgi.DownloadProgressBars = append(*pgi.DownloadProgressBars, progBar)
}
