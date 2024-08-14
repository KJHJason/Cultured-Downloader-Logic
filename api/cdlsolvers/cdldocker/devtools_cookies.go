package cdldocker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
)

func convertSameSite(sameSite string) http.SameSite {
	switch sameSite {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

func convertSameSiteToString(sameSite http.SameSite) string {
	switch sameSite {
	case http.SameSiteLaxMode:
		return "Lax"
	case http.SameSiteStrictMode:
		return "Strict"
	case http.SameSiteNoneMode:
		return "None"
	default:
		return ""
	}
}

// https://chromedevtools.github.io/devtools-protocol/tot/Network/#type-CookieParam
// E.g.
// [
//
//	{
//	  "name": "_session_id",
//	  "value": "cookie-value",
//	  "url": null,
//	  "domain": "fantia.jp",
//	  "path": "/",
//	  "secure": true,
//	  "httpOnly": true,
//	  "sameSite": "Lax",
//	  "expires": 1721925436
//	}
//
// ]
type DevToolsCookieParam struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Url      string `json:"url,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Secure   *bool  `json:"secure,omitempty"`
	HttpOnly *bool  `json:"httpOnly,omitempty"`
	SameSite string `json:"sameSite,omitempty"`
	Expires  int64  `json:"expires,omitempty"`
}

func convertCookiesToDevToolsCookiesParam(cookies []*http.Cookie) []*DevToolsCookieParam {
	devToolsCookies := make([]*DevToolsCookieParam, len(cookies))
	for idx, cookie := range cookies {
		devToolsCookies[idx] = &DevToolsCookieParam{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Url:      "",
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Secure:   &cookie.Secure,
			HttpOnly: &cookie.HttpOnly,
			SameSite: convertSameSiteToString(cookie.SameSite),
			Expires:  cookie.Expires.Unix(),
		}
	}
	return devToolsCookies
}

func makeTempCookieParamFile(cdlTempDir string, cookies []*DevToolsCookieParam) (string, error) {
	data, err := json.Marshal(cookies)
	if err != nil {
		return "", fmt.Errorf(
			"error %d: failed to marshal cookies data => %w",
			cdlerrors.JSON_ERROR,
			err,
		)
	}

	filePath := filepath.Join(cdlTempDir, "cookie.json")
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return "", fmt.Errorf(
			"error %d: failed to write cookies data to file => %w",
			cdlerrors.OS_ERROR,
			err,
		)
	}
	return filePath, nil
}

// E.g.
//
//	{
//		"name": "cf_clearance",
//		"value": "Y.y1g4TLzz8EImSH0MQ09QB7hKr6azxd4hoKE8tkpWQ-1720623556-1.0.1.1-v7ULqmIQEgT0xCuYh8R4WIh45_m_jC_2eXoPWGea2GXyuYdkJ216zo9dTnSwK3Wiaat1Xjhg8d.zGJLqk0X19g",
//		"domain": ".nopecha.com",
//		"path": "/",
//		"expires": 1752159568.841206,
//		"size": 161,
//		"httpOnly": true,
//		"secure": true,
//		"session": false,
//		"sameSite": "None",
//		"priority": "Medium",
//		"sameParty": false,
//		"sourceScheme": "Secure",
//		"sourcePort": 443,
//		"partitionKey": "https://nopecha.com"
//	}
//
// https://chromedevtools.github.io/devtools-protocol/tot/Network/#type-Cookie
type DevToolsCookie struct {
	Name         string  `json:"name"`
	Value        string  `json:"value"`
	Domain       string  `json:"domain"`
	Path         string  `json:"path"`
	Expires      float64 `json:"expires"`
	Size         int     `json:"size"`
	HttpOnly     bool    `json:"httpOnly"`
	Secure       bool    `json:"secure"`
	Session      bool    `json:"session"`
	SameSite     string  `json:"sameSite"`
	Priority     string  `json:"priority"`
	SameParty    bool    `json:"sameParty"`
	SourceScheme string  `json:"sourceScheme"`
	SourcePort   int     `json:"sourcePort"`
}

func convertExpiresToTime(unix float64) time.Time {
	seconds := int64(unix)
	nanoSeconds := int64((unix - float64(seconds)) * 1e9)
	return time.Unix(seconds, nanoSeconds)
}

func ConvertDevToolsCookies(cookies []*DevToolsCookie) []*http.Cookie {
	httpCookies := make([]*http.Cookie, len(cookies))
	for i, cookie := range cookies {
		httpCookies[i] = &http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,

			Path:    cookie.Path,
			Domain:  cookie.Domain,
			Expires: convertExpiresToTime(cookie.Expires),

			Secure:   cookie.Secure,
			HttpOnly: cookie.HttpOnly,
			SameSite: convertSameSite(cookie.SameSite),
		}
	}
	return httpCookies
}

func parseCookiesFromFile(filePath string) ([]*DevToolsCookie, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf(
			"error %d: failed to read cookies file %s => %w",
			cdlerrors.OS_ERROR,
			filePath,
			err,
		)
	}

	var cookies []*DevToolsCookie
	err = json.Unmarshal(data, &cookies)
	if err != nil {
		return nil, fmt.Errorf(
			"error %d: failed to unmarshal cookies data => %w",
			cdlerrors.JSON_ERROR,
			err,
		)
	}
	return cookies, nil
}
