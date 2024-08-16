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

// E.g.
// [
//	{
//	  "name": "_session_id",
//	  "value": "cookie-value",
//	  "domain": "fantia.jp",
//	  "path": "/",
//	  "secure": true,
//	  "expires": 1721925436
//	}
// ]
type DevToolsCookieParam struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Secure   *bool  `json:"secure,omitempty"`
	Expires  int64  `json:"expires,omitempty"`
}

func convertCookiesToDevToolsCookiesParam(cookies []*http.Cookie) []*DevToolsCookieParam {
	devToolsCookies := make([]*DevToolsCookieParam, len(cookies))
	for idx, cookie := range cookies {
		devToolsCookies[idx] = &DevToolsCookieParam{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Secure:   &cookie.Secure,
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
// {
// 	"name": "cf_clearance",
// 	"value": "cookie-value",
// 	"domain": ".nopecha.com",
// 	"path": "/",
// 	"expires": 1755272665.550445,
// 	"size": 161,
// 	"httpOnly": true,
// 	"secure": true,
// 	"session": false,
// 	"sameSite": "None",
// 	"priority": "Medium",
// 	"sameParty": false,
// 	"sourceScheme": "Secure",
// 	"sourcePort": 443,
// 	"partitionKey": {
// 		"topLevelSite": "https://nopecha.com",
// 		"hasCrossSiteAncestor": false
// 	}
// }
type DevToolsCookie struct {
	Name         string  `json:"name"`
	Value        string  `json:"value"`
	Domain       string  `json:"domain"`
	Path         string  `json:"path"`
	Expires      float64 `json:"expires"`
	Size         int     `json:"size"`
	HTTPOnly     bool    `json:"httpOnly"`
	Secure       bool    `json:"secure"`
	Session      bool    `json:"session"`
	SameSite     string  `json:"sameSite"`
	Priority     string  `json:"priority"`
	SameParty    bool    `json:"sameParty"`
	SourceScheme string  `json:"sourceScheme"`
	SourcePort   int     `json:"sourcePort"`
	PartitionKey struct {
		TopLevelSite         string `json:"topLevelSite"`
		HasCrossSiteAncestor bool   `json:"hasCrossSiteAncestor"`
	} `json:"partitionKey"`
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

	if len(data) == 0 {
		return nil, fmt.Errorf(
			"error %d: empty cookies file %s",
			cdlerrors.CAPTCHA_ERROR,
			filePath,
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
