package pixivcommon

import (
	"context"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cf"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
)

const (
	cacheKey = "pixiv"
)

var (
	CaptchaChecker = cf.CaptchaChecker
)

func GetCachedCfCookies() []*http.Cookie {
	return cf.GetCachedCfCookies(cacheKey, constants.CF_BOT_COOKIE_TIMEOUT)
}

type CaptchaHandler struct {
	url      string
	ctx      context.Context
	notifier notify.Notifier
}

func NewCaptchaHandler(ctx context.Context, url string, notifier notify.Notifier) CaptchaHandler {
	return CaptchaHandler{
		url:      url,
		ctx:      ctx,
		notifier: notifier,
	}
}

func (ch CaptchaHandler) Call(req *http.Request) error {
	return cf.Call(
		ch.ctx,
		req,
		cacheKey,
		ch.url,
		constants.CF_BOT_COOKIE_TIMEOUT,
		ch.notifier,
	)
}

func (ch CaptchaHandler) GetCfCookies() ([]*http.Cookie, error) {
	return cf.GetCfCookies(
		ch.ctx,
		cacheKey,
		ch.url,
		constants.CF_BOT_COOKIE_TIMEOUT,
		ch.notifier,
	)
}
