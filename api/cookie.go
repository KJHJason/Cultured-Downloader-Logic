package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/parsers"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

// Returns a cookie with given value and website to be used in requests
func GetCookie(sessionID, website string) *http.Cookie {
	if sessionID == "" {
		return &http.Cookie{}
	}

	sessionCookieInfo := parsers.GetSessionCookieInfo(website)
	domain := sessionCookieInfo.Domain
	cookieName := sessionCookieInfo.Name
	sameSite := sessionCookieInfo.SameSite

	cookie := http.Cookie{
		Name:     cookieName,
		Value:    sessionID,
		Domain:   domain,
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		Path:     "/",
		SameSite: sameSite,
		Secure:   true,
		HttpOnly: true,
	}
	return &cookie
}

func getHeaders(website, userAgent string) map[string]string {
	headers := map[string]string{
		"User-Agent": userAgent,
	}

	var referer, origin string
	switch website {
	case constants.PIXIV :
		referer = constants.PIXIV_URL
		origin = constants.PIXIV_URL
	case constants.PIXIV_FANBOX :
		referer = constants.PIXIV_FANBOX_URL
		origin = constants.PIXIV_FANBOX_URL
	case constants.FANTIA :
		referer = constants.FANTIA_URL
		origin = constants.FANTIA_URL
	case constants.KEMONO :
		referer = constants.KEMONO_URL
		origin = constants.KEMONO_URL
	case constants.KEMONO_BACKUP :
		referer = constants.BACKUP_KEMONO_URL
		origin = constants.BACKUP_KEMONO_URL
	default :
		// Shouldn't happen but could happen during development
		panic(
			fmt.Errorf(
				"error %d, invalid website, %q, in getHeaders",
				constants.DEV_ERROR,
				website,
			),
		)
	}

	headers["Referer"] = referer
	headers["Origin"] = origin
	return headers
}

// Verifies the given cookie by making a request to the website
// and returns true if the cookie is valid
func VerifyCookie(cookie *http.Cookie, website, userAgent string) (bool, error) {
	// sends a request to the website to verify the cookie
	var websiteUrl string
	switch website {
	case constants.FANTIA:
		websiteUrl = constants.FANTIA_URL + "/mypage/users/plans"
	case constants.PIXIV_FANBOX:
		websiteUrl = constants.PIXIV_FANBOX_URL + "/creators/supporting"
	case constants.PIXIV:
		websiteUrl = constants.PIXIV_URL + "/dashboard"
	case constants.KEMONO:
		websiteUrl = constants.KEMONO_URL + "/favorites"
	case constants.KEMONO_BACKUP:
		websiteUrl = constants.BACKUP_KEMONO_URL + "/favorites"
	default:
		// Shouldn't happen but could happen during development
		panic(
			fmt.Errorf(
				"error %d, invalid website, %q, in VerifyCookie",
				constants.DEV_ERROR,
				website,
			),
		)
	}

	if cookie.Value == "" {
		return false, nil
	}

	useHttp3 := httpfuncs.IsHttp3Supported(website, false)
	cookies := []*http.Cookie{cookie}
	resp, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Method:      "HEAD",
			Url:         websiteUrl,
			Cookies:     cookies,
			CheckStatus: true,
			Http3:       useHttp3,
			Http2:       !useHttp3,
			Headers:     getHeaders(website, userAgent),
		},
	)
	if err != nil {
		return false, fmt.Errorf("error occurred when trying to verify cookie...\n%v", err)
	}
	resp.Body.Close()

	// check if the cookie is valid
	resUrl := resp.Request.URL.String()
	if website == constants.FANTIA && strings.HasPrefix(resUrl, constants.FANTIA_RECAPTCHA_URL) {
		// This would still mean that the cookie is still valid.
		return true, nil
	}
	return resUrl == websiteUrl, nil
}

// Prints out the error message and exits the program if the cookie verification fails
func processCookieVerification(website string, err error) error {
	if err != nil {
		logger.LogError(
			fmt.Errorf("error occurred when trying to verify %s cookie...\n%v", GetReadableSiteStr(website), err),
			false,
			logger.ERROR,
		)
		return fmt.Errorf(
			"error %d: could not verify %s cookie.\nPlease refer to the log file for more details",
			constants.INPUT_ERROR,
			GetReadableSiteStr(website),
		)
	}
	return nil
}

// Verifies the given cookie by making a request to the backup domain and checks if the cookie is valid
func backupVerifyCookie(website, cookieValue, userAgent string) (*http.Cookie, error) {
	var backupWebsite string
	switch website {
	case constants.KEMONO:
		backupWebsite = constants.KEMONO_BACKUP
	default:
		// Shouldn't happen but could happen during development
		panic(
			fmt.Sprintf(
				"error %d: %s is not supported for cookie verification on a backup domain.",
				constants.DEV_ERROR,
				GetReadableSiteStr(website),
			),
		)
	}

	cookie := GetCookie(cookieValue, backupWebsite)
	cookieIsValid, err := VerifyCookie(cookie, backupWebsite, userAgent)
	processCookieVerification(backupWebsite, err)
	if !cookieIsValid {
		return nil, fmt.Errorf(
			"error %d: %s cookie is invalid",
			constants.INPUT_ERROR,
			GetReadableSiteStr(backupWebsite),
		)
	}
	return cookie, nil
}

// Verifies the given cookie by making a request to the website and checks if the cookie is valid
// If the cookie is valid, the cookie will be returned
//
// However, if the cookie is invalid, an error message will be printed out and the program will shutdown
func VerifyAndGetCookie(website, cookieValue, userAgent string) (*http.Cookie, error) {
	cookie := GetCookie(cookieValue, website)
	cookieIsValid, err := VerifyCookie(cookie, website, userAgent)
	processCookieVerification(website, err)

	if !cookieIsValid {
		if website != constants.KEMONO {
			return nil, fmt.Errorf(
				"error %d: %s cookie is invalid",
				constants.INPUT_ERROR,
				GetReadableSiteStr(website),
			)
		} else {
			// try to verify the cookie on the backup domain
			cookie, err = backupVerifyCookie(website, cookieValue, userAgent)
			if err != nil {
				return nil, err
			}
		}
	}
	return cookie, nil
}
