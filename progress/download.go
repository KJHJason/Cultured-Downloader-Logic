package progress

import (
	"context"
)

type DlProgressBars []*DlProgress

type NewDlProgressBar func(context.Context, Messages) *DlProgress

type DlProgress interface {
	// Update the filename of the download.
	UpdateFilename(string)

	// Update the download speed.
	UpdateDownloadSpeed(float64)

	// Update the download eta
	UpdateDownloadETA(float64)

	// Update the download percentage.
	UpdateDownloadPercentage(float64)

	// Stop the download progress with a bool to indicate if there is an error or not.
	Stop(bool)

	// Update the error message of the download progress.
	UpdateErrMsg(string)
}
