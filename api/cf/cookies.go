package cf

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
)

const (
	// https://developers.cloudflare.com/fundamentals/reference/policies-compliances/cloudflare-cookies/
	BOT_COOKIE       = "__cf_bm"
	CLEARANCE_COOKIE = "cf_clearance"
)

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
type Cookie struct {
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

type Cookies []*Cookie

func parseCookies(filePath string) (Cookies, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf(
			"error %d: failed to read cookies file %s => %w",
			cdlerrors.OS_ERROR,
			filePath,
			err,
		)
	}

	var cookies Cookies
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

func convertExpiresToTime(unix float64) time.Time {
	seconds := int64(unix)
	nanoSeconds := int64((unix - float64(seconds)) * 1e9)
	return time.Unix(seconds, nanoSeconds)
}

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

func ConvertCookies(cookies []*Cookie) []*http.Cookie {
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
