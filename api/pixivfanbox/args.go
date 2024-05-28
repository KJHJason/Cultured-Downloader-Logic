package pixivfanbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

// PixivFanboxDl is the struct that contains the IDs of the Pixiv Fanbox creators and posts to download.
type PixivFanboxDl struct {
	CreatorIds      []string
	CreatorPageNums []string

	PostIds []string
}

// ValidateArgs validates the IDs of the Pixiv Fanbox creators and posts to download.
//
// It also validates the page numbers of the creators to download.
//
// Should be called after initialising the struct.
func (pf *PixivFanboxDl) ValidateArgs() error {
	err := utils.ValidateIds(pf.PostIds)
	if err != nil {
		return err
	}

	pf.PostIds = utils.RemoveSliceDuplicates(pf.PostIds)

	for _, creatorId := range pf.CreatorIds {
		if !constants.PIXIV_FANBOX_CREATOR_ID_REGEX.MatchString(creatorId) {
			return fmt.Errorf(
				"error %d: invalid Pixiv Fanbox creator ID %q, must be alphanumeric with underscores, dashes, or periods",
				cdlerrors.INPUT_ERROR,
				creatorId,
			)
		}
	}

	if len(pf.CreatorPageNums) > 0 {
		err = utils.ValidatePageNumInput(
			len(pf.CreatorIds),
			pf.CreatorPageNums,
			[]string{
				"Number of Pixiv Fanbox Creator ID(s) and page numbers must be equal.",
			},
		)
		if err != nil {
			return err
		}
	} else {
		pf.CreatorPageNums = make([]string, len(pf.CreatorIds))
	}
	pf.CreatorIds, pf.CreatorPageNums = utils.RemoveDuplicateIdAndPageNum(
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
	UseCacheDb          bool
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

// CancelCtx releases the resources used and cancels the context of the PixivFanboxDlOptions struct.
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
			cdlerrors.DEV_ERROR,
		)
	}

	if pf.Configs == nil {
		return fmt.Errorf(
			"pixiv fanbox error %d, configs is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if pf.UseCacheDb && pf.Configs.OverwriteFiles {
		pf.UseCacheDb = false
	}

	if dlDirPath, err := utils.ValidateDlDirPath(pf.BaseDownloadDirPath, constants.PIXIV_FANBOX_TITLE); err != nil {
		return err
	} else {
		pf.BaseDownloadDirPath = dlDirPath
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
			cdlerrors.DEV_ERROR,
		)
	}

	if pf.DlGdrive && pf.GdriveClient == nil {
		pf.DlGdrive = false
	} else if !pf.DlGdrive && pf.GdriveClient != nil {
		pf.GdriveClient = nil
	}
	return nil
}
