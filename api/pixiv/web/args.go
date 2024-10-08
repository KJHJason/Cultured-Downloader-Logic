package pixivweb

import (
	"context"
	"fmt"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

// PixivToDl is the struct that contains the arguments of Pixiv download options.
type PixivWebDlOptions struct {
	ctx    context.Context
	cancel context.CancelFunc

	Base *api.BaseDl

	pFilters *pixivcommon.PixivFilters
}

func (p *PixivWebDlOptions) GetCaptchaHandler() httpfuncs.CaptchaHandler {
	return pixivcommon.NewHttpCaptchaHandler(
		p.ctx,
		constants.PIXIV_URL,
		p.Base.Configs.UserAgent,
		p.Base.Notifier,
	)
}

func (p *PixivWebDlOptions) GetContext() context.Context {
	return p.ctx
}

// CancelCtx releases the resources used and cancels the context of the PixivWebDlOptions struct.
func (p *PixivWebDlOptions) CancelCtx() {
	p.cancel()
}

func (p *PixivWebDlOptions) SetContext(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
}

func (p *PixivWebDlOptions) CtxIsActive() bool {
	return p.ctx.Err() == nil
}

func (p *PixivWebDlOptions) SetPixivFilters(filters pixivcommon.PixivFilters) error {
	if err := filters.ValidateForWebApi(); err != nil {
		return err
	}
	p.pFilters = &filters
	return nil
}

// ValidateArgs validates the arguments of the Pixiv download options.
//
// Should be called after initialising the struct.
func (p *PixivWebDlOptions) ValidateArgs(userAgent string) error {
	if p.GetContext() == nil {
		p.SetContext(context.Background())
	}

	if p.Base.Configs == nil {
		return fmt.Errorf(
			"pixiv web error %d, configs is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if p.Base.MainProgBar() == nil {
		return fmt.Errorf(
			"pixiv web error %d, main progress bar is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if p.Base.Notifier == nil {
		return fmt.Errorf(
			"pixiv web error %d: Notifier cannot be nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if p.Base.UseCacheDb && p.Base.Configs.OverwriteFiles {
		p.Base.UseCacheDb = false
	}

	if p.Base.SessionCookieId != "" {
		p.Base.SessionCookies = []*http.Cookie{
			api.GetCookie(p.Base.SessionCookieId, constants.PIXIV),
		}
		p.Base.SessionCookieId = ""
	}

	if len(p.Base.SessionCookies) > 0 {
		ch := p.GetCaptchaHandler()
		if err := api.VerifyCookies(constants.PIXIV, userAgent, p.Base.SessionCookies, ch); err != nil {
			return err
		}
	} else {
		return fmt.Errorf(
			"pixiv web error %d: Session cookies cannot be empty",
			cdlerrors.INPUT_ERROR,
		)
	}

	if dlDirPath, err := utils.ValidateDlDirPath(p.Base.DownloadDirPath, constants.PIXIV_TITLE); err != nil {
		return err
	} else {
		p.Base.DownloadDirPath = dlDirPath
	}
	return nil
}
