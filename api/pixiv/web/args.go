package pixivweb

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

// PixivToDl is the struct that contains the arguments of Pixiv download options.
type PixivWebDlOptions struct {
	ctx    context.Context
	cancel context.CancelFunc

	Base *api.BaseDl

	// Sort order of the results. Can be "date_desc" or "date_asc".
	SortOrder    string
	SearchMode   string
	SearchAiMode int // 1: filter AI works, 0: Display AI works
	RatingMode   string
	ArtworkType  string
}

func (p *PixivWebDlOptions) GetCaptchaHandler() httpfuncs.CaptchaHandler {
	return httpfuncs.CaptchaHandler{
		Check: pixivcommon.CaptchaChecker,
		Handler: pixivcommon.NewCaptchaHandler(
			p.ctx,
			constants.PIXIV_MOBILE_URL,
			p.Base.Notifier,
		),
		InjectCaptchaCookies: pixivcommon.GetCachedCfCookies,
	}
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

	if p.Base.UseCacheDb && p.Base.Configs.OverwriteFiles {
		p.Base.UseCacheDb = false
	}

	if len(p.Base.SessionCookies) > 0 {
		if err := api.VerifyCookies(constants.PIXIV, userAgent, p.Base.SessionCookies, httpfuncs.CaptchaHandler{}); err != nil {
			return err
		}
		p.Base.SessionCookieId = ""
	} else if p.Base.SessionCookieId != "" {
		if cookie, err := api.VerifyAndGetCookie(constants.PIXIV, p.Base.SessionCookieId, userAgent, httpfuncs.CaptchaHandler{}); err != nil {
			return err
		} else {
			p.Base.SessionCookies = []*http.Cookie{cookie}
		}
	}

	if dlDirPath, err := utils.ValidateDlDirPath(p.Base.DownloadDirPath, constants.PIXIV_TITLE); err != nil {
		return err
	} else {
		p.Base.DownloadDirPath = dlDirPath
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

	// Web API:
	// - 0: Display AI works
	// - 1: Filter AI works
	if p.SearchAiMode != 0 && p.SearchAiMode != 1 {
		p.SearchAiMode = 1 // Default to filter AI works
	}

	p.SortOrder = strings.ToLower(p.SortOrder)
	_, err := utils.ValidateStrArgs(
		p.SortOrder,
		constants.ACCEPTED_SORT_ORDER,
		[]string{
			fmt.Sprintf(
				"pixiv web error %d: Sort order %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.SortOrder,
			),
		},
	)
	if err != nil {
		return err
	}

	p.SearchMode = strings.ToLower(p.SearchMode)
	_, err = utils.ValidateStrArgs(
		p.SearchMode,
		constants.ACCEPTED_SEARCH_MODE,
		[]string{
			fmt.Sprintf(
				"pixiv web error %d: Search order %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.SearchMode,
			),
		},
	)
	if err != nil {
		return err
	}

	p.RatingMode = strings.ToLower(p.RatingMode)
	_, err = utils.ValidateStrArgs(
		p.RatingMode,
		constants.ACCEPTED_RATING_MODE,
		[]string{
			fmt.Sprintf(
				"pixiv web error %d: Rating order %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.RatingMode,
			),
		},
	)
	if err != nil {
		return err
	}

	p.ArtworkType = strings.ToLower(p.ArtworkType)
	_, err = utils.ValidateStrArgs(
		p.ArtworkType,
		constants.ACCEPTED_ARTWORK_TYPE,
		[]string{
			fmt.Sprintf(
				"pixiv web error %d: Artwork type %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.ArtworkType,
			),
		},
	)
	if err != nil {
		return err
	}
	return nil
}
