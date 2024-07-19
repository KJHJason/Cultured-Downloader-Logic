package pixivfanbox

import (
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cf"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

const (
	cacheKey = "fanbox"
)

var (
	CaptchaChecker = cf.CaptchaChecker
)

func GetCachedCfCookies() []*http.Cookie {
	return cf.GetCachedCfCookies(cacheKey, constants.CF_BOT_COOKIE_TIMEOUT)
}

type CaptchaHandler struct {
	dlOptions *PixivFanboxDlOptions
}

func NewCaptchaHandler(dlOptions *PixivFanboxDlOptions) CaptchaHandler {
	return CaptchaHandler{dlOptions: dlOptions}
}

func (ch CaptchaHandler) Call(req *http.Request) error {
	return cf.Call(
		ch.dlOptions.GetContext(),
		req,
		cacheKey,
		constants.PIXIV_FANBOX_URL,
		constants.CF_BOT_COOKIE_TIMEOUT,
		ch.dlOptions.Base.Notifier,
	)
}

func (ch CaptchaHandler) GetCfCookies() ([]*http.Cookie, error) {
	return cf.GetCfCookies(
		ch.dlOptions.GetContext(),
		cacheKey,
		constants.PIXIV_FANBOX_URL,
		constants.CF_BOT_COOKIE_TIMEOUT,
		ch.dlOptions.Base.Notifier,
	)
}
