package pixivfanbox

import (
	"net/http"
	"sync"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cf"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/startup"
)

var (
	cfCacheMu      sync.RWMutex
	cachedCookies  []*http.Cookie
	solvedUnixTime int64
)

func getFilteredCachedCookiesUnsafe() []*http.Cookie {
	var cfCookies []*http.Cookie
	for _, cookie := range cachedCookies {
		if isCfCookies(cookie.Name) {
			cfCookies = append(cfCookies, cookie)
		}
	}
	return cfCookies
}

func GetCachedCfCookies() []*http.Cookie {
	cfCacheMu.RLock()
	defer cfCacheMu.RUnlock()
	return getCachedCfCookiesUnsafe()
}

func getCachedCfCookiesUnsafe() []*http.Cookie {
	if solvedUnixTime != 0 && (time.Now().UnixMilli()-solvedUnixTime) < constants.PIXIV_FANBOX_CAPTCHA_CACHE_TIMEOUT {
		return getFilteredCachedCookiesUnsafe()
	}
	return nil
}

type CaptchaHandler struct {
	dlOptions *PixivFanboxDlOptions
}

func NewCaptchaHandler(dlOptions *PixivFanboxDlOptions) CaptchaHandler {
	return CaptchaHandler{dlOptions: dlOptions}
}

func isCfCookies(name string) bool {
	return name == cf.BOT_COOKIE || name == cf.CLEARANCE_COOKIE
}

func checkHasCfCookies(req *http.Request) bool {
	for _, cookie := range req.Cookies() {
		if isCfCookies(cookie.Name) {
			return true
		}
	}
	return false
}

func addCacheCookiesToReq(req *http.Request) {
	if checkHasCfCookies(req) {
		cookiesCopy := make([]*http.Cookie, len(req.Cookies()))
		copy(cookiesCopy, req.Cookies())
		req.Header.Del("Cookie")
		for _, cookie := range cookiesCopy {
			if !isCfCookies(cookie.Name) {
				req.AddCookie(cookie)
			}
		}
	}

	for _, cookie := range getFilteredCachedCookiesUnsafe() {
		req.AddCookie(cookie)
	}
}

func (ch CaptchaHandler) callMainLogicUnsafe() error {
	var err error
	var cfCookies cf.Cookies
	cfArgs := cf.NewCfArgs(constants.PIXIV_FANBOX_URL)
	if startup.UseDockerForCf {
		cfCookies, err = cf.CallDockerImage(ch.dlOptions.GetContext(), cfArgs)
		if err != nil {
			return err
		}
	} else {
		cfCookies, err = cf.CallPyScript(cfArgs)
		if err != nil {
			return err
		}
	}

	solvedUnixTime = time.Now().UnixMilli()
	cachedCookies = cf.ConvertCookies(cfCookies)
	return nil
}

func (ch CaptchaHandler) Call(req *http.Request) error {
	cfCacheMu.Lock()
	defer cfCacheMu.Unlock()

	if getCachedCfCookiesUnsafe() != nil {
		addCacheCookiesToReq(req)
		return nil
	}

	if err := ch.callMainLogicUnsafe(); err != nil {
		return err
	}
	addCacheCookiesToReq(req)
	return nil
}

func (ch CaptchaHandler) GetCfCookies() ([]*http.Cookie, error) {
	cfCacheMu.Lock()
	defer cfCacheMu.Unlock()

	if cfCookies := getCachedCfCookiesUnsafe(); cfCookies != nil {
		return cfCookies, nil
	}

	if err := ch.callMainLogicUnsafe(); err != nil {
		return nil, err
	}
	return getFilteredCachedCookiesUnsafe(), nil
}

func CaptchaChecker(res *httpfuncs.ResponseWrapper) (bool, error) {
	if res.Resp.StatusCode != http.StatusForbidden {
		return false, nil
	}
	return true, nil
}
