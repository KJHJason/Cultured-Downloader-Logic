package pixivfanbox

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

// PixivFanboxDl is the struct that contains the IDs of the Pixiv Fanbox creators and posts to download.
type PixivFanboxDl struct {
	CreatorIds      []string
	CreatorPageNums []string

	PostIds []string
}

var creatorIdRegex = regexp.MustCompile(`^[\w.-]+$`)

// ValidateArgs validates the IDs of the Pixiv Fanbox creators and posts to download.
//
// It also validates the page numbers of the creators to download.
//
// Should be called after initialising the struct.
func (pf *PixivFanboxDl) ValidateArgs() error {
	api.ValidateIds(pf.PostIds)
	pf.PostIds = api.RemoveSliceDuplicates(pf.PostIds)

	for _, creatorId := range pf.CreatorIds {
		if !creatorIdRegex.MatchString(creatorId) {
			return fmt.Errorf(
				"error %d: invalid Pixiv Fanbox creator ID %q, must be alphanumeric with underscores, dashes, or periods",
				errs.INPUT_ERROR,
				creatorId,
			)
		}
	}

	if len(pf.CreatorPageNums) > 0 {
		api.ValidatePageNumInput(
			len(pf.CreatorIds),
			pf.CreatorPageNums,
			[]string{
				"Number of Pixiv Fanbox Creator ID(s) and page numbers must be equal.",
			},
		)
	} else {
		pf.CreatorPageNums = make([]string, len(pf.CreatorIds))
	}
	pf.CreatorIds, pf.CreatorPageNums = api.RemoveDuplicateIdAndPageNum(
		pf.CreatorIds,
		pf.CreatorPageNums,
	)
	return nil
}

// PixivFanboxDlOptions is the struct that contains the options for downloading from Pixiv Fanbox.
type PixivFanboxDlOptions struct {
	ctx                 context.Context
	cancel              context.CancelFunc
	DlThumbnails        bool
	DlImages            bool
	DlAttachments       bool
	DlGdrive            bool
	BaseDownloadDirPath string

	Configs *configs.Config

	// GdriveClient is the Google Drive client to be
	// used in the download process for Pixiv Fanbox posts
	GdriveClient *gdrive.GDrive

	SessionCookieId string
	SessionCookies  []*http.Cookie

	Notifier notify.Notifier

	// Progress indicators
	MainProgBar          progress.ProgressBar
	DownloadProgressBars *[]*progress.DownloadProgressBar
}

func (pf *PixivFanboxDlOptions) GetContext() context.Context {
	return pf.ctx
}

func (pf *PixivFanboxDlOptions) SetContext(ctx context.Context) {
	pf.ctx, pf.cancel = context.WithCancel(ctx)
}

// Cancel releases the resources used and cancels the context of the PixivFanboxDlOptions struct.
func (pf *PixivFanboxDlOptions) CancelCtx() {
	pf.cancel()
}

func (pf *PixivFanboxDlOptions) CtxIsActive() bool {
	return pf.ctx.Err() == nil
}

// ValidateArgs validates the session cookie ID of the Pixiv Fanbox account to download from.
//
// Should be called after initialising the struct.
func (pf *PixivFanboxDlOptions) ValidateArgs(userAgent string) error {
	if pf.GetContext() == nil {
		pf.SetContext(context.Background())
	}

	if pf.Notifier == nil {
		return fmt.Errorf(
			"pixiv fanbox error %d: Notifier cannot be nil",
			errs.DEV_ERROR,
		)
	}

	if pf.BaseDownloadDirPath == "" {
		pf.BaseDownloadDirPath = filepath.Join(iofuncs.DOWNLOAD_PATH, constants.PIXIV_FANBOX_TITLE)
	} else {
		if !iofuncs.DirPathExists(pf.BaseDownloadDirPath) {
			return fmt.Errorf(
				"pixiv fanbox error %d, download path does not exist or is not a directory, please create the directory and try again",
				errs.INPUT_ERROR,
			)
		}
		pf.BaseDownloadDirPath = filepath.Join(pf.BaseDownloadDirPath, constants.PIXIV_FANBOX_TITLE)
	}

	if len(pf.SessionCookies) > 0 {
		if err := api.VerifyCookies(constants.PIXIV_FANBOX, userAgent, pf.SessionCookies); err != nil {
			return err
		}
		pf.SessionCookieId = ""
	} else if pf.SessionCookieId != "" {
		if cookie, err := api.VerifyAndGetCookie(constants.PIXIV_FANBOX, pf.SessionCookieId, userAgent); err != nil {
			return err
		} else {
			pf.SessionCookies = []*http.Cookie{cookie}
		}
	}

	if pf.MainProgBar == nil {
		return fmt.Errorf(
			"pixiv fanbox error %d, main progress bar is nil",
			errs.DEV_ERROR,
		)
	}

	if pf.DlGdrive && pf.GdriveClient == nil {
		pf.DlGdrive = false
	} else if !pf.DlGdrive && pf.GdriveClient != nil {
		pf.GdriveClient = nil
	}
	return nil
}
