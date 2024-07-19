package cf

import (
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

func IsCfCookies(name string) bool {
	return name == BOT_COOKIE || name == CLEARANCE_COOKIE
}

func SendReqAndGetCfCookies(url string, http3 bool) ([]*http.Cookie, error) {
	reqArgs := httpfuncs.RequestArgs{
		Method:      "GET",
		Url:         url,
		CheckStatus: false,
	}
	res, err := httpfuncs.CallRequest(&reqArgs)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var cfCookies []*http.Cookie
	cookies := res.Resp.Cookies()
	for _, cookie := range cookies {
		if IsCfCookies(cookie.Name) {
			cfCookies = append(cfCookies, cookie)
		}
	}
	return cfCookies, nil
}
