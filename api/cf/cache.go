package cf

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
)

type cacheValues struct {
	cookies []*http.Cookie
	solved  time.Time
}

var (
	cfCacheMu     sync.RWMutex
	cachedCookies = make(map[string]*cacheValues)
)

func getFilteredCachedCookiesUnsafe(key string) []*http.Cookie {
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

func GetCachedCfCookies(key string, timeout time.Duration) []*http.Cookie {
	cfCacheMu.RLock()
	defer cfCacheMu.RUnlock()
	return getCachedCfCookiesUnsafe(key, timeout)
}

func getCachedCfCookiesUnsafe(key string, timeout time.Duration) []*http.Cookie {
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

func addCacheCookiesToReq(req *http.Request, key string) {
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

func callMainLogicUnsafe(ctx context.Context, key string, url string, notifier notify.Notifier) error {
	var err error
	var cfCookies Cookies

	alert(notifier, "CF Captcha detected, solving it automatically...")
	cfCookies, err = CallDockerImage(ctx, url)
	if err != nil {
		alert(notifier, "Failed to solve CF Captcha automatically...")
		return err
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

func Call(ctx context.Context, req *http.Request, key, url string, timeout time.Duration, notifier notify.Notifier) error {
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

func GetCfCookies(ctx context.Context, key, url string, timeout time.Duration, notifier notify.Notifier) ([]*http.Cookie, error) {
	cfCacheMu.Lock()
	defer cfCacheMu.Unlock()

	if cfCookies := getCachedCfCookiesUnsafe(key, timeout); cfCookies != nil {
		return cfCookies, nil
	}

	if err := callMainLogicUnsafe(ctx, key, url, notifier); err != nil {
		return nil, err
	}
	return getFilteredCachedCookiesUnsafe(key), nil
}
