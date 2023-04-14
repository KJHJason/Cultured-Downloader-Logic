package parsers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
			Domain:   "kemono.party",
			Name:     "session",
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

func parseTxtCookieFile(f *os.File, filePath string, cookieArgs *cookieInfoArgs) ([]*http.Cookie, error) {
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

func parseJsonCookieFile(f *os.File, filePath string, cookieArgs *cookieInfoArgs) ([]*http.Cookie, error) {
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

// parse the Netscape cookie file generated by extensions like Get cookies.txt LOCALLY
func ParseNetscapeCookieFile(filePath, sessionId, website string) ([]*http.Cookie, error) {
	if filePath != "" && sessionId != "" {
		return nil, fmt.Errorf(
			"error %d: cannot use both cookie file and session id flags",
			constants.INPUT_ERROR,
		)
	}

	sessionCookieInfo := GetSessionCookieInfo(website)
	sessionCookieName := sessionCookieInfo.Name
	sessionCookieSameSite := sessionCookieInfo.SameSite

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf(
			"error %d: opening cookie file at %s, more info => %v",
			constants.OS_ERROR,
			filePath,
			err,
		)
	}
	defer f.Close()

	cookieArgs := &cookieInfoArgs{
		name:     sessionCookieName,
		sameSite: sessionCookieSameSite,
	}
	var cookies []*http.Cookie
	if ext := filepath.Ext(filePath); ext == ".txt" {
		cookies, err = parseTxtCookieFile(f, filePath, cookieArgs)
	} else if ext == ".json" {
		cookies, err = parseJsonCookieFile(f, filePath, cookieArgs)
	} else {
		err = fmt.Errorf(
			"error %d: invalid cookie file extension, %q, at %s...\nOnly .txt and .json files are supported",
			constants.INPUT_ERROR,
			ext,
			filePath,
		)
	}

	if err != nil {
		return nil, err
	}

	if len(cookies) == 0 {
		return nil, fmt.Errorf(
			"error %d: no session cookie found in cookie file at %s for website %q",
			constants.INPUT_ERROR,
			filePath,
			website,
		)
	}
	return cookies, nil
}
