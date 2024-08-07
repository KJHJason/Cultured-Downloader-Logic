package pixivcommon

import (
	"context"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cf"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
)

var (
	CaptchaChecker = cf.CaptchaChecker
)

func NewHttpCaptchaHandler(ctx context.Context, url string, notifier notify.Notifier) httpfuncs.CaptchaHandler {
	handler := CaptchaHandler{
		url:      url,
		ctx:      ctx,
		notifier: notifier,
	}
	return httpfuncs.CaptchaHandler{
		Check:         CaptchaChecker,
		Handler:       handler,
		CallBeforeReq: true,
		ReqModifier:   handler.CallIfReq,
	}
}

type CaptchaHandler struct {
	url      string
	ctx      context.Context
	notifier notify.Notifier
}

func (ch CaptchaHandler) Call(req *http.Request) error {
	return cf.Call(
		ch.ctx,
		req,
		cf.PixivCacheKey,
		ch.url,
		constants.CF_BOT_COOKIE_TIMEOUT,
		ch.notifier,
	)
}

func (ch CaptchaHandler) CallIfReq(req *http.Request) error {
	return cf.CallIfReq(
		ch.ctx,
		req,
		cf.PixivCacheKey,
		ch.url,
		constants.CF_BOT_COOKIE_TIMEOUT,
		ch.notifier,
	)
}
