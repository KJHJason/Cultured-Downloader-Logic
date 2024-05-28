package api

import (
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
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
}
