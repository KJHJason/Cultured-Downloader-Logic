package pixivfanbox

import (
	"context"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cdlsolvers/cf"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
)

var (
	CaptchaChecker = cf.CaptchaChecker
)

func NewHttpCaptchaHandler(ctx context.Context, userAgent string, notifier notify.Notifier) httpfuncs.CaptchaHandler {
	handler := CaptchaHandler{
		ctx:       ctx,
		notifier:  notifier,
		userAgent: userAgent,
	}
	return httpfuncs.CaptchaHandler{
		Check:         CaptchaChecker,
		Handler:       handler,
		CallBeforeReq: true,
		ReqModifier:   handler.CallIfReq,
	}
}

type CaptchaHandler struct {
	ctx       context.Context
	notifier  notify.Notifier
	userAgent string
}

func (ch CaptchaHandler) getCacheArgs() cf.CacheArgs {
	return cf.CacheArgs{
		Key:       cf.PixivFanboxCacheKey,
		Url:       constants.PIXIV_FANBOX_URL,
		UserAgent: ch.userAgent,
		Timeout:   constants.CF_BOT_COOKIE_TIMEOUT,
		Notifier:  ch.notifier,
	}
}

func (ch CaptchaHandler) Call(req *http.Request) error {
	return cf.Call(
		ch.ctx,
		req,
		ch.getCacheArgs(),
	)
}

func (ch CaptchaHandler) CallIfReq(req *http.Request) error {
	return cf.CallIfReq(
		ch.ctx,
		req,
		ch.getCacheArgs(),
	)
}
