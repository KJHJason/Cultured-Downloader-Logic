package cf

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cdlsolvers/cdldocker"
	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
)

type cacheValues struct {
	cookies []*http.Cookie
	solved  time.Time
}

type CacheKey uint8

const (
	PixivCacheKey CacheKey = iota + 1
	PixivFanboxCacheKey
)

type CacheArgs struct {
	Key       CacheKey
	Url       string
	UserAgent string
	Timeout   time.Duration
	Notifier  notify.Notifier
}

var (
	cfCacheMu     sync.RWMutex
	cachedCookies = make(map[CacheKey]*cacheValues)
	failedKeys    = make(map[CacheKey]struct{})
)

func getFilteredCachedCookiesUnsafe(key CacheKey) []*http.Cookie {
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

func getCachedCfCookiesUnsafe(key CacheKey, timeout time.Duration) []*http.Cookie {
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

func addCacheCookiesToReq(req *http.Request, key CacheKey) {
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

func callMainLogicUnsafe(ctx context.Context, cacheArgs CacheArgs) error {
	if _, ok := failedKeys[cacheArgs.Key]; ok {
		return cdlerrors.ErrCaptchaPrevFailed
	}

	if cookies, err := sendReqAndGetCfCookies(cacheArgs.Url); err != nil {
		return err
	} else if len(cookies) > 0 {
		cachedCookies[cacheArgs.Key] = &cacheValues{
			cookies: cookies,
			solved:  time.Now(),
		}
		return nil
	}

	alert(cacheArgs.Notifier, "CF Captcha detected, solving it automatically...")

	var err error
	var cfCookies []*cdldocker.DevToolsCookie
	cfCookies, err = cdldocker.CallDockerImageForCf(ctx, cacheArgs.UserAgent, cacheArgs.Url)
	if err != nil {
		alert(cacheArgs.Notifier, "Failed to solve CF Captcha automatically...")
		return fmt.Errorf(
			"error %d: failed to solve CF Captcha automatically => %w",
			cdlerrors.CAPTCHA_ERROR,
			err,
		)
	}

	alert(cacheArgs.Notifier, "Successfully solved CF Captcha automatically!")
	cachedCookies[cacheArgs.Key] = &cacheValues{
		cookies: cdldocker.ConvertDevToolsCookies(cfCookies),
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
func Call(ctx context.Context, req *http.Request, cacheArgs CacheArgs) error {
	cfCacheMu.Lock()
	defer cfCacheMu.Unlock()

	if err := callMainLogicUnsafe(ctx, cacheArgs); err != nil {
		return err
	}
	addCacheCookiesToReq(req, cacheArgs.Key)
	return nil
}

// Similar to Call, but checks for cached cookies.
func CallIfReq(ctx context.Context, req *http.Request, cacheArgs CacheArgs) error {
	cfCacheMu.Lock()
	defer cfCacheMu.Unlock()

	if getCachedCfCookiesUnsafe(cacheArgs.Key, cacheArgs.Timeout) != nil {
		addCacheCookiesToReq(req, cacheArgs.Key)
		return nil
	}

	if err := callMainLogicUnsafe(ctx, cacheArgs); err != nil {
		return err
	}
	addCacheCookiesToReq(req, cacheArgs.Key)
	return nil
}
