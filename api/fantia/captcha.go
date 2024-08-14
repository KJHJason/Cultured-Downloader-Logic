package fantia

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cdlsolvers/cdldocker"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
)

var (
	captchaMu  sync.Mutex
	solvedTime time.Time
)

type CaptchaOptions struct {
	Ctx            context.Context
	UserAgent      string
	SessionCookies []*http.Cookie
	Notifier       notify.Notifier
}

// Automatically try to solve the reCAPTCHA for Fantia.
func autoSolveCaptcha(captchaOptions CaptchaOptions) error {
	readableSite := database.GetReadableSiteStr(constants.FANTIA)
	notifier := captchaOptions.Notifier
	notifier.Alert(
		fmt.Sprintf("reCAPTCHA detected for the current %s session! Trying to solve it automatically...", readableSite),
	)

	err := cdldocker.CallDockerImageForFantia(
		captchaOptions.Ctx,
		captchaOptions.UserAgent,
		captchaOptions.SessionCookies,
	)
	if err != nil {
		logger.MainLogger.Errorf(
			"fantia error %d: failed to solve reCAPTCHA for %s, more info => %v",
			cdlerrors.CAPTCHA_ERROR,
			readableSite,
			err,
		)
		notifier.Alert("Failed to solve reCAPTCHA automatically...")
	}

	notifier.Alert("Successfully solved reCAPTCHA automatically!")
	solvedTime = time.Now()
	return nil
}

func newHttpCaptchaHandler(dlOptions *FantiaDlOptions) httpfuncs.CaptchaHandler {
	handler := CaptchaHandler{
		options: CaptchaOptions{
			Ctx:            dlOptions.GetContext(),
			UserAgent:      dlOptions.Base.Configs.UserAgent,
			SessionCookies: dlOptions.Base.SessionCookies,
			Notifier:       dlOptions.Base.Notifier,
		},
	}
	return httpfuncs.CaptchaHandler{
		Check:         CaptchaChecker,
		Handler:       handler,
		CallBeforeReq: false,
		ReqModifier:   nil,
	}
}

func NewHttpCaptchaHandler(options CaptchaOptions) httpfuncs.CaptchaHandler {
	handler := CaptchaHandler{
		options: CaptchaOptions{
			Ctx:            options.Ctx,
			UserAgent:      options.UserAgent,
			SessionCookies: options.SessionCookies,
			Notifier:       options.Notifier,
		},
	}
	return httpfuncs.CaptchaHandler{
		Check:         CaptchaChecker,
		Handler:       handler,
		CallBeforeReq: false,
		ReqModifier:   nil,
	}
}

type CaptchaHandler struct {
	options CaptchaOptions
}

func NewCaptchaHandlerWithOptions(options CaptchaOptions) CaptchaHandler {
	return CaptchaHandler{
		options: options,
	}
}

func (ch CaptchaHandler) Call(*http.Request) error {
	captchaMu.Lock()
	defer captchaMu.Unlock()

	if !solvedTime.IsZero() && time.Since(solvedTime) < constants.FANTIA_CAPTCHA_CACHE_TIMEOUT {
		// if the reCAPTCHA was solved within the last few seconds,
		// then skip solving it to avoid solving it multiple times
		return nil
	}

	if len(ch.options.SessionCookies) == 0 {
		// Since reCAPTCHA is per session for Fantia, the program shall avoid
		// trying to solve it and alert the user to login or create a Fantia account.
		// It is possible that the reCAPTCHA is per IP address for guests, but I'm not sure.
		return fmt.Errorf(
			"fantia error %d: reCAPTCHA detected but you are not logged in. Please login to Fantia and try again",
			cdlerrors.CAPTCHA_ERROR,
		)
	}
	return autoSolveCaptcha(ch.options)
}

func CaptchaChecker(res *httpfuncs.ResponseWrapper) (bool, error) {
	finalUrl := res.Url()
	if finalUrl == constants.FANTIA_RECAPTCHA_URL {
		return true, nil
	}

	// check if response is json
	contentType := res.Resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return false, nil
	}

	body, err := res.GetBody()
	if err != nil {
		return false, err
	}

	var captchaResp CaptchaResponse
	if err := httpfuncs.LoadJsonFromBytes(res.Url(), body, &captchaResp); err != nil {
		return false, err
	}
	return captchaResp.Redirect != "", nil
}
