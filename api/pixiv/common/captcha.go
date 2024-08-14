package pixivcommon

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

func NewHttpCaptchaHandler(ctx context.Context, url, userAgent string, notifier notify.Notifier) httpfuncs.CaptchaHandler {
	handler := CaptchaHandler{
		url:       url,
		userAgent: userAgent,
		ctx:       ctx,
		notifier:  notifier,
	}
	return httpfuncs.CaptchaHandler{
		Check:         CaptchaChecker,
		Handler:       handler,
		CallBeforeReq: true,
		ReqModifier:   handler.CallIfReq,
	}
}

type CaptchaHandler struct {
	url       string
	userAgent string
	ctx       context.Context
	notifier  notify.Notifier
}

func (ch CaptchaHandler) getCacheArgs() cf.CacheArgs {
	return cf.CacheArgs{
		Key:       cf.PixivCacheKey,
		Url:       ch.url,
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
