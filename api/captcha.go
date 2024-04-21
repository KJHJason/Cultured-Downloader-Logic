package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/chromedp/chromedp"
)

const captchaBtnSelector = `//input[@name='commit']`

type DlOptions interface {
	GetConfigs()                     *configs.Config
	GetSessionCookies()              []*http.Cookie
	GetAutoSolveCaptcha()            bool
	GetNotifier()                    notify.Notifier
	GetCaptchaHandler()              constants.CAPTCHA_FN
	GetContext()					 context.Context

	// Prog bars
	GetCaptchaSolverProgBar()        progress.Progress

	SetAutoSolveCaptcha(bool)
}

func getChromedpActions(website string, cookies []*http.Cookie) []chromedp.Action {
	switch website {
	case constants.FANTIA:
		return []chromedp.Action{
			SetChromedpAllocCookies(cookies),
			chromedp.Navigate(constants.FANTIA_RECAPTCHA_URL),
			chromedp.WaitVisible(captchaBtnSelector, chromedp.BySearch),
			chromedp.Click(captchaBtnSelector, chromedp.BySearch),
			chromedp.WaitVisible(`//h3[@class='mb-15'][contains(text(), 'ファンティアでクリエイターを応援しよう！')]`, chromedp.BySearch),
		}
	default:
		panic(fmt.Sprintf("unsupported website %q in getChromedpActions()", website))
	}
}

// Automatically try to solve the reCAPTCHA for Fantia.
func autoSolveCaptcha(dlOptions DlOptions, website string) error {
	readableSite := GetReadableSiteStr(website)
	notifier := dlOptions.GetNotifier()
	notifier.Alert(
		fmt.Sprintf("reCAPTCHA detected for the current %s session! Trying to solve it automatically...", readableSite),
	)

	prog := dlOptions.GetCaptchaSolverProgBar()
	prog.UpdateBaseMsg(
		fmt.Sprintf("Solving reCAPTCHA for %s...", readableSite),
	)
	prog.UpdateSuccessMsg(
		fmt.Sprintf("Successfully solved reCAPTCHA for %s!", readableSite),
	)
	prog.Start()

	actions := getChromedpActions(website, dlOptions.GetSessionCookies())

	configs := dlOptions.GetConfigs()
	allocCtx, cancel := GetDefaultChromedpAlloc(configs.UserAgent, dlOptions.GetContext())
	defer cancel()

	allocCtx, cancel = context.WithTimeout(allocCtx, 45 * time.Second)
	if err := ExecuteChromedpActions(allocCtx, cancel, actions...); err != nil {
		var fmtErr error
		if errors.Is(err, context.DeadlineExceeded) {
			fmtErr = fmt.Errorf(
				"error %d: failed to solve reCAPTCHA for %s due to timeout, please visit %s to solve it manually and try again", 
				constants.CAPTCHA_ERROR,
				readableSite,
				constants.FANTIA_RECAPTCHA_URL,
			)
		} else {
			fmtErr = fmt.Errorf(
				"error %d: failed to solve reCAPTCHA for %s, more info => %v",
				constants.CAPTCHA_ERROR, 
				readableSite,
				err,
			)
			logger.LogError(fmtErr, false, logger.ERROR)
		}

		prog.UpdateErrorMsg(fmtErr.Error() + "\n")
		prog.Stop(true)
		notifier.Alert("Failed to solve reCAPTCHA automatically...")
		return fmtErr
	}
	prog.Stop(false)
	notifier.Alert("Successfully solved reCAPTCHA automatically!")
	return nil
}

func getCaptchaAndVerificationUrls(website string) (string, string) {
	switch website {
	case constants.FANTIA:
		return constants.FANTIA_RECAPTCHA_URL, constants.FANTIA_URL + "/mypage/users/plans"
	default:
		panic(fmt.Sprintf("unsupported website %q in getCaptchaUrl()", website))
	}
}

func SolveCaptcha(dlOptions DlOptions, website string) error {
	if len(dlOptions.GetSessionCookies()) == 0 && website == constants.FANTIA {
		// Since reCAPTCHA is per session for Fantia, the program shall avoid 
		// trying to solve it and alert the user to login or create a Fantia account.
		// It is possible that the reCAPTCHA is per IP address for guests, but I'm not sure.
		return fmt.Errorf(
			"fantia error %d: reCAPTCHA detected but you are not logged in. Please login to Fantia and try again",
			constants.CAPTCHA_ERROR,
		)
	}

	if dlOptions.GetAutoSolveCaptcha() {
		return autoSolveCaptcha(dlOptions, website)
	}
	panic("missing captcha auto solver")
}

// try the alternative method if the first one fails.
//
// E.g. User preferred to solve the reCAPTCHA automatically, but the program failed to do so,
//      The program will then ask the user to solve the reCAPTCHA manually on their browser with the SAME session.
func HandleCaptchaErr(err error, dlOptions DlOptions, website string) error {
	if err == nil {
		return nil
	}

	dlOptions.SetAutoSolveCaptcha(!dlOptions.GetAutoSolveCaptcha())
	return SolveCaptcha(dlOptions, website)
}
