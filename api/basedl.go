package api

import (
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

type BaseDl struct {
	Notifier notify.Notifier

	DlThumbnails        bool
	DlImages            bool
	OrganiseImages      bool
	DlAttachments       bool
	DlGdrive            bool
	DetectOtherDlLinks  bool
	UseCacheDb          bool
	BaseDownloadDirPath string

	Filters *Filters
	Configs *configs.Config

	SessionCookieId string
	SessionCookies  []*http.Cookie

	// Progress indicators
	MainProgBar          progress.ProgressBar
	DownloadProgressBars *[]*progress.DownloadProgressBar
}
