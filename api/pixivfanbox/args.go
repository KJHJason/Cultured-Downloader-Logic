package pixivfanbox

import (
	"context"
	"fmt"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
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
	ctx    context.Context
	cancel context.CancelFunc
	Base   *api.BaseDl

	CfCookies []*http.Cookie
}

func (pf *PixivFanboxDlOptions) GetCaptchaHandler() httpfuncs.CaptchaHandler {
	handler := newCaptchaHandlerWithDlOptions(pf)
	return httpfuncs.CaptchaHandler{
		Check:         CaptchaChecker,
		Handler:       handler,
		CallBeforeReq: true,
		ReqModifier:   handler.CallIfReq,
	}
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

	if pf.Base == nil {
		return fmt.Errorf(
			"pixiv fanbox error %d: Base cannot be nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if pf.Base.Notifier == nil {
		return fmt.Errorf(
			"pixiv fanbox error %d: Notifier cannot be nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if pf.Base.MainProgBar() == nil {
		return fmt.Errorf(
			"pixiv fanbox error %d, main progress bar is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if pf.Base.Configs == nil {
		return fmt.Errorf(
			"pixiv fanbox error %d, configs is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if pf.Base.UseCacheDb && pf.Base.Configs.OverwriteFiles {
		pf.Base.UseCacheDb = false
	}

	if dlDirPath, err := utils.ValidateDlDirPath(pf.Base.DownloadDirPath, constants.PIXIV_FANBOX_TITLE); err != nil {
		return err
	} else {
		pf.Base.DownloadDirPath = dlDirPath
	}

	if pf.Base.SessionCookieId != "" {
		pf.Base.SessionCookies = []*http.Cookie{
			api.GetCookie(pf.Base.SessionCookieId, constants.PIXIV_FANBOX),
		}
		pf.Base.SessionCookieId = ""
	}

	if len(pf.Base.SessionCookies) > 0 {
		captchaHandler := pf.GetCaptchaHandler()
		if err := api.VerifyCookies(constants.PIXIV_FANBOX, userAgent, pf.Base.SessionCookies, captchaHandler); err != nil {
			return err
		}
	} else {
		return fmt.Errorf(
			"pixiv fanbox error %d: Session cookies cannot be empty",
			cdlerrors.INPUT_ERROR,
		)
	}

	if pf.Base.DlGdrive && pf.Base.GdriveClient == nil {
		pf.Base.DlGdrive = false
	} else if !pf.Base.DlGdrive && pf.Base.GdriveClient != nil {
		pf.Base.GdriveClient = nil
	}
	return nil
}
