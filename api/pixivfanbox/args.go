package pixivfanbox

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
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
				constants.INPUT_ERROR,
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
	ctx           context.Context
	cancel        context.CancelFunc
	DlThumbnails  bool
	DlImages      bool
	DlAttachments bool
	DlGdrive      bool

	Configs *configs.Config

	// GdriveClient is the Google Drive client to be
	// used in the download process for Pixiv Fanbox posts
	GdriveClient *gdrive.GDrive

	SessionCookieId string
	SessionCookies  []*http.Cookie

	Notifier notify.Notifier

	// Prog bar
	PostProgBar             progress.Progress
	CreatorPostsProgBar     progress.Progress
	ProcessJsonProgBar      progress.Progress
	GdriveApiProgBar        progress.Progress
	GdriveDlProgBar         progress.Progress
}

func (pf *PixivFanboxDlOptions) GetContext() context.Context {
	return pf.ctx
}

func (pf *PixivFanboxDlOptions) GetCancel() context.CancelFunc {
	return pf.cancel
}

func (pf *PixivFanboxDlOptions) SetContext(ctx context.Context) {
	pf.ctx, pf.cancel = context.WithCancel(ctx)
}

// Cancel cancels the context of the PixivFanboxDlOptions struct.
func (pf *PixivFanboxDlOptions) Cancel() {
	pf.cancel()
}

// ValidateArgs validates the session cookie ID of the Pixiv Fanbox account to download from.
//
// Should be called after initialising the struct.
func (pf *PixivFanboxDlOptions) ValidateArgs(userAgent string) error {
	if pf.GetContext() == nil {
		pf.SetContext(context.Background())
	}

	captchaMap := map[string]progress.Progress{
		"post":              pf.PostProgBar,
		"get creator posts": pf.CreatorPostsProgBar,
		"process json":      pf.ProcessJsonProgBar,
		"gdrive api":        pf.GdriveApiProgBar,
		"gdrive download":   pf.GdriveDlProgBar,
	}
	for captchaName, captchaProgBar := range captchaMap {
		if captchaProgBar == nil {
			return fmt.Errorf(
				"kemono error %d, %s progress bar is nil",
				constants.DEV_ERROR,
				captchaName,
			)
		}
	}

	if pf.SessionCookieId != "" {
		if cookie, err := api.VerifyAndGetCookie(constants.PIXIV_FANBOX, pf.SessionCookieId, userAgent); err != nil {
			return err
		} else {
			pf.SessionCookies = []*http.Cookie{
				cookie,
			}
		}
	}

	if pf.DlGdrive && pf.GdriveClient == nil {
		pf.DlGdrive = false
	} else if !pf.DlGdrive && pf.GdriveClient != nil {
		pf.GdriveClient = nil
	}
	return nil
}
