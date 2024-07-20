package cf

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
)

type cacheValues struct {
	cookies []*http.Cookie
	solved  time.Time
}

const (
	PixivCacheKey       = 1
	PixivFanboxCacheKey = 2
)

var (
	cfCacheMu     sync.RWMutex
	cachedCookies = make(map[int]*cacheValues)
	failedKeys    = make(map[int]struct{})
)

func getFilteredCachedCookiesUnsafe(key int) []*http.Cookie {
	var ok bool
	var cachedValues *cacheValues
	if cachedValues, ok = cachedCookies[key]; !ok {
		return nil
	}

	var cfCookies []*http.Cookie
	for _, cookie := range cachedValues.cookies {
		if IsCfCookies(cookie.Name) {
			cfCookies = append(cfCookies, cookie)
		}
	}
	return cfCookies
}

func getCachedCfCookiesUnsafe(key int, timeout time.Duration) []*http.Cookie {
	var ok bool
	var cachedValues *cacheValues
	if cachedValues, ok = cachedCookies[key]; !ok {
		return nil
	}

	solvedTime := cachedValues.solved
	if !solvedTime.IsZero() && time.Since(solvedTime) < timeout {
		return getFilteredCachedCookiesUnsafe(key)
	}
	return nil
}

func checkHasCfCookies(req *http.Request) bool {
	for _, cookie := range req.Cookies() {
		if IsCfCookies(cookie.Name) {
			return true
		}
	}
	return false
}

func addCacheCookiesToReq(req *http.Request, key int) {
	if checkHasCfCookies(req) {
		cookiesCopy := make([]*http.Cookie, len(req.Cookies()))
		copy(cookiesCopy, req.Cookies())
		req.Header.Del("Cookie")
		for _, cookie := range cookiesCopy {
			if !IsCfCookies(cookie.Name) {
				req.AddCookie(cookie)
			}
		}
	}

	for _, cookie := range getFilteredCachedCookiesUnsafe(key) {
		req.AddCookie(cookie)
	}
}

func alert(notifier notify.Notifier, msg string) {
	if notifier != nil {
		notifier.Alert(msg)
	}
}

func callMainLogicUnsafe(ctx context.Context, key int, url string, notifier notify.Notifier) error {
	if _, ok := failedKeys[key]; ok {
		return cdlerrors.ErrCaptchaPrevFailed
	}

	if cookies, err := sendReqAndGetCfCookies(url); err != nil {
		return err
	} else if len(cookies) > 0 {
		cachedCookies[key] = &cacheValues{
			cookies: cookies,
			solved:  time.Now(),
		}
		return nil
	}

	alert(notifier, "CF Captcha detected, solving it automatically...")

	var err error
	var cfCookies Cookies
	cfCookies, err = CallDockerImage(ctx, url)
	if err != nil {
		alert(notifier, "Failed to solve CF Captcha automatically...")
		return fmt.Errorf(
			"error %d: failed to solve CF Captcha automatically => %w",
			cdlerrors.CAPTCHA_ERROR,
			err,
		)
	}

	alert(notifier, "Successfully solved CF Captcha automatically!")
	cachedCookies[key] = &cacheValues{
		cookies: ConvertCookies(cfCookies),
		solved:  time.Now(),
	}
	return nil
}

func CaptchaChecker(res *httpfuncs.ResponseWrapper) (bool, error) {
	if res.Resp.StatusCode != http.StatusForbidden {
		return false, nil
	}
	return true, nil
}

// Note: This function does not check for cached cookies.
func Call(ctx context.Context, req *http.Request, key int, url string, timeout time.Duration, notifier notify.Notifier) error {
	cfCacheMu.Lock()
	defer cfCacheMu.Unlock()

	if err := callMainLogicUnsafe(ctx, key, url, notifier); err != nil {
		return err
	}
	addCacheCookiesToReq(req, key)
	return nil
}

// Similar to Call, but checks for cached cookies.
func CallIfReq(ctx context.Context, req *http.Request, key int, url string, timeout time.Duration, notifier notify.Notifier) error {
	cfCacheMu.Lock()
	defer cfCacheMu.Unlock()

	if getCachedCfCookiesUnsafe(key, timeout) != nil {
		addCacheCookiesToReq(req, key)
		return nil
	}

	if err := callMainLogicUnsafe(ctx, key, url, notifier); err != nil {
		return err
	}
	addCacheCookiesToReq(req, key)
	return nil
}
