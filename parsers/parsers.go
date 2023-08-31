package parsers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
)

// Returns the cookie info for the specified site
//
// Will panic if the site does not match any of the cases
func GetSessionCookieInfo(site string) *cookieInfo {
	switch site {
	case constants.FANTIA:
		return &cookieInfo{
			Domain:   "fantia.jp",
			Name:     "_session_id",
			SameSite: http.SameSiteLaxMode,
		}
	case constants.PIXIV_FANBOX:
		return &cookieInfo{
			Domain:   ".fanbox.cc",
			Name:     "FANBOXSESSID",
			SameSite: http.SameSiteNoneMode,
		}
	case constants.PIXIV:
		return &cookieInfo{
			Domain:   ".pixiv.net",
			Name:     "PHPSESSID",
			SameSite: http.SameSiteNoneMode,
		}
	case constants.KEMONO:
		return &cookieInfo{
			Domain:   constants.KEMONO_COOKIE_DOMAIN,
			Name:     constants.KEMONO_SESSION_COOKIE_NAME,
			SameSite: http.SameSiteNoneMode,
		}
	case constants.KEMONO_BACKUP:
		return &cookieInfo{
			Domain:   constants.KEMONO_COOKIE_BACKUP_DOMAIN,
			Name:     constants.KEMONO_SESSION_COOKIE_NAME,
			SameSite: http.SameSiteNoneMode,
		}
	default:
		panic(
			fmt.Errorf(
				"error %d, invalid site, %q in GetSessionCookieInfo",
				constants.DEV_ERROR,
				site,
			),
		)
	}
}

func ParseTxtCookieFile(f *os.File, filePath string, cookieArgs *cookieInfoArgs) ([]*http.Cookie, error) {
	var cookies []*http.Cookie
	reader := bufio.NewReader(f)
	for {
		lineBytes, err := iofuncs.ReadLine(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf(
				"error %d: reading cookie file at %s, more info => %v",
				constants.OS_ERROR,
				filePath,
				err,
			)
		}

		line := strings.TrimSpace(string(lineBytes))
		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip empty lines and comments
		}

		// split the line
		cookieInfos := strings.Split(line, "\t")
		if len(cookieInfos) < 7 {
			continue // too few values will be ignored
		}

		cookieName := cookieInfos[5]
		if cookieName != cookieArgs.name {
			continue // not the session cookie
		}

		// parse the values
		cookie := http.Cookie{
			Name:     cookieName,
			Value:    cookieInfos[6],
			Domain:   cookieInfos[0],
			Path:     cookieInfos[2],
			Secure:   cookieInfos[3] == "TRUE",
			HttpOnly: true,
			SameSite: cookieArgs.sameSite,
		}

		expiresUnixStr := cookieInfos[4]
		if expiresUnixStr != "" {
			expiresUnixInt, err := strconv.Atoi(expiresUnixStr)
			if err != nil {
				// should never happen but just in case
				return nil, fmt.Errorf(
					"error %d: parsing cookie expiration time, %q, more info => %v",
					constants.UNEXPECTED_ERROR,
					expiresUnixStr,
					err,
				)
			}
			if expiresUnixInt > 0 {
				cookie.Expires = time.Unix(int64(expiresUnixInt), 0)
			}
		}
		cookies = append(cookies, &cookie)
	}
	return cookies, nil
}

func ParseJsonCookieFile(f *os.File, filePath string, cookieArgs *cookieInfoArgs) ([]*http.Cookie, error) {
	var cookies []*http.Cookie
	var exportedCookies ExportedCookies
	if err := json.NewDecoder(f).Decode(&exportedCookies); err != nil {
		return nil, fmt.Errorf(
			"error %d: failed to decode cookie JSON file at %s, more info => %v",
			constants.JSON_ERROR,
			filePath,
			err,
		)
	}

	for _, cookie := range exportedCookies {
		if cookie.Name != cookieArgs.name {
			// not the session cookie
			continue
		}

		parsedCookie := &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HttpOnly,
			SameSite: cookieArgs.sameSite,
		}
		if !cookie.Session {
			parsedCookie.Expires = time.Unix(int64(cookie.Expire), 0)
		}

		cookies = append(cookies, parsedCookie)
	}
	return cookies, nil
}
