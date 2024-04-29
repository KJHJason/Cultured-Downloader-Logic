package fantia

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/chromedp/chromedp"
)

const captchaBtnSelector = `//input[@name='commit']`

type CaptchaOptions interface {
	GetConfigs()        *configs.Config
	GetSessionCookies() []*http.Cookie
	GetNotifier()       notify.Notifier
	GetContext()        context.Context
}

// Automatically try to solve the reCAPTCHA for Fantia.
func autoSolveCaptcha(captchaOptions CaptchaOptions) error {
	readableSite := api.GetReadableSiteStr(constants.FANTIA)
	notifier := captchaOptions.GetNotifier()
	notifier.Alert(
		fmt.Sprintf("reCAPTCHA detected for the current %s session! Trying to solve it automatically...", readableSite),
	)

	configs := captchaOptions.GetConfigs()
	allocCtx, cancel := api.GetDefaultChromedpAlloc(configs.UserAgent, captchaOptions.GetContext())
	defer cancel()

	actions := []chromedp.Action{
		api.SetChromedpAllocCookies(captchaOptions.GetSessionCookies()),
		chromedp.Navigate(constants.FANTIA_RECAPTCHA_URL),
		chromedp.WaitVisible(captchaBtnSelector, chromedp.BySearch),
		chromedp.Click(captchaBtnSelector, chromedp.BySearch),
		chromedp.WaitVisible(`//h3[@class='mb-15'][contains(text(), 'ファンティアでクリエイターを応援しよう！')]`, chromedp.BySearch),
	}

	allocCtx, cancel = context.WithTimeout(allocCtx, 45 * time.Second)
	if err := api.ExecuteChromedpActions(allocCtx, cancel, actions...); err != nil {
		var fmtErr error
		if errors.Is(err, context.DeadlineExceeded) {
			fmtErr = fmt.Errorf(
				"fantia error %d: failed to solve reCAPTCHA for %s due to timeout, please visit %s with the SAME session cookies to solve it manually and try again", 
				errs.CAPTCHA_ERROR,
				readableSite,
				constants.FANTIA_RECAPTCHA_URL,
			)
		} else {
			fmtErr = fmt.Errorf(
				"fantia error %d: failed to solve reCAPTCHA for %s, more info => %w",
				errs.CAPTCHA_ERROR, 
				readableSite,
				err,
			)
			logger.LogError(fmtErr, false, logger.ERROR)
		}
		notifier.Alert("Failed to solve reCAPTCHA automatically...")
		return fmtErr
	}
	notifier.Alert("Successfully solved reCAPTCHA automatically!")
	return nil
}

func SolveCaptcha(captchaOptions CaptchaOptions) error {
	if len(captchaOptions.GetSessionCookies()) == 0 {
		// Since reCAPTCHA is per session for Fantia, the program shall avoid 
		// trying to solve it and alert the user to login or create a Fantia account.
		// It is possible that the reCAPTCHA is per IP address for guests, but I'm not sure.
		return fmt.Errorf(
			"fantia error %d: reCAPTCHA detected but you are not logged in. Please login to Fantia and try again",
			errs.CAPTCHA_ERROR,
		)
	}
	return autoSolveCaptcha(captchaOptions)
}
