package pixivfanbox

import (
	"context"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cf"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
)

var (
	CaptchaChecker = cf.CaptchaChecker
)

type CaptchaHandler struct {
	ctx      context.Context
	notifier notify.Notifier
}

func NewCaptchaHandler(ctx context.Context, notifier notify.Notifier) CaptchaHandler {
	return CaptchaHandler{
		ctx:      ctx,
		notifier: notifier,
	}
}

func newCaptchaHandlerWithDlOptions(dlOptions *PixivFanboxDlOptions) CaptchaHandler {
	return CaptchaHandler{
		ctx:      dlOptions.GetContext(),
		notifier: dlOptions.Base.Notifier,
	}
}

func (ch CaptchaHandler) Call(req *http.Request) error {
	return cf.Call(
		ch.ctx,
		req,
		cf.PixivFanboxCacheKey,
		constants.PIXIV_FANBOX_URL,
		constants.CF_BOT_COOKIE_TIMEOUT,
		ch.notifier,
	)
}

func (ch CaptchaHandler) CallIfReq(req *http.Request) error {
	return cf.CallIfReq(
		ch.ctx,
		req,
		cf.PixivFanboxCacheKey,
		constants.PIXIV_FANBOX_URL,
		constants.CF_BOT_COOKIE_TIMEOUT,
		ch.notifier,
	)
}
