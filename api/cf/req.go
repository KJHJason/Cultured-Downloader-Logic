package cf

import (
	"net/http"

	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

func IsCfCookies(name string) bool {
	return name == BOT_COOKIE || name == CLEARANCE_COOKIE
}

func sendReqAndGetCfCookies(url string) ([]*http.Cookie, error) {
	logger.MainLogger.Info("Sending request to get cf cookies")
	reqArgs := httpfuncs.RequestArgs{
		Method:      "GET",
		Url:         url,
		CheckStatus: false,
	}
	res, err := httpfuncs.CallRequest(&reqArgs)
	if err != nil {
		logger.MainLogger.Errorf(
			"error %d: failed to send request to get cf cookies => %v",
			cdlerrors.CONNECTION_ERROR, err,
		)
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

	if len(cfCookies) == 0 {
		logger.MainLogger.Errorf("failed to get cf cookies from %s", url)
	}
	return cfCookies, nil
}
