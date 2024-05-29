package api

import (
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/filters"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

type BaseDl struct {
	Notifier notify.Notifier

	DlThumbnails       bool // Fantia, PixivFanbox
	DlImages           bool // Fantia, PixivFanbox
	OrganiseImages     bool // Fantia
	DlAttachments      bool // Fantia, PixivFanbox, Kemono
	DlGdrive           bool // Fantia, PixivFanbox, Kemono
	DetectOtherDlLinks bool // Fantia
	UseCacheDb         bool
	DownloadDirPath    string

	Filters      *filters.Filters
	Configs      *configs.Config
	GdriveClient *gdrive.GDrive // Fantia, PixivFanbox, Kemono

	SessionCookieId string
	SessionCookies  []*http.Cookie

	// Progress indicators
	ProgressBarInfo *progress.ProgressBarInfo
}

func (b *BaseDl) MainProgBar() progress.ProgressBar {
	if b.ProgressBarInfo == nil {
		return nil
	}
	return b.ProgressBarInfo.MainProgressBar
}

func (b *BaseDl) DownloadProgressBars() *[]*progress.DownloadProgressBar {
	if b.ProgressBarInfo == nil {
		return nil
	}
	return b.ProgressBarInfo.DownloadProgressBars
}
