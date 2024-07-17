package pixivfanbox

import (
	"net/http"
	"sync"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cf"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

var (
	cfCacheMu     sync.RWMutex
	cachedCookies []*http.Cookie
	solvedTime    *time.Time
)

type CaptchaHandler struct {
	dlOptions *PixivFanboxDlOptions
}

func newCaptchaHandler(dlOptions *PixivFanboxDlOptions) CaptchaHandler {
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

	for _, cookie := range cachedCookies {
		if isCfCookies(cookie.Name) {
			req.AddCookie(cookie)
		}
	}
}

func (ch CaptchaHandler) Call(req *http.Request) error {
	cfCacheMu.Lock()
	defer cfCacheMu.Unlock()

	if solvedTime != nil && time.Since(solvedTime) < constants.PIXIV_FANBOX_CAPTCHA_TIMEOUT*time.Second {
		return nil
	}

	if cachedCookies != nil {
		addCacheCookiesToReq(req)
		return nil
	}

	var err error
	var cfCookies cf.Cookies
	cfArgs := cf.NewCfArgs(constants.PIXIV_FANBOX_URL)
	if utils.UseDockerForCf {
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

	solvedTime = time.Now().UnixNano()
	cachedCookies = cf.ConvertCookies(cfCookies)
	addCacheCookiesToReq(req)
	return nil
}

func CaptchaChecker(res *httpfuncs.ResponseWrapper) (bool, error) {
	if res.Resp.StatusCode != http.StatusForbidden {
		return false, nil
	}
	return true, nil
}
