package pixivfanbox

import (
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cf"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

var (
	CaptchaChecker = cf.CaptchaChecker
)

func GetCachedCfCookies() []*http.Cookie {
	return cf.GetCachedCfCookies(cf.PixivFanboxCacheKey, constants.CF_BOT_COOKIE_TIMEOUT)
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
		cf.PixivFanboxCacheKey,
		constants.PIXIV_FANBOX_URL,
		constants.CF_BOT_COOKIE_TIMEOUT,
		ch.dlOptions.Base.Notifier,
	)
}

func (ch CaptchaHandler) GetCfCookies() ([]*http.Cookie, error) {
	return cf.GetCfCookies(
		ch.dlOptions.GetContext(),
		cf.PixivFanboxCacheKey,
		constants.PIXIV_FANBOX_URL,
		constants.CF_BOT_COOKIE_TIMEOUT,
		ch.dlOptions.Base.Notifier,
	)
}
