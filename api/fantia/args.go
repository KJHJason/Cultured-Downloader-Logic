package fantia

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/PuerkitoBio/goquery"
)

// FantiaDl is the struct that contains the
// IDs of the Fantia fanclubs and posts to download.
type FantiaDl struct {
	FanclubIds      []string
	FanclubPageNums []string
	PostIds         []string
}

// ValidateArgs validates the IDs of the Fantia fanclubs and posts to download.
//
// It also validates the page numbers of the fanclubs to download.
//
// Should be called after initialising the struct.
func (f *FantiaDl) ValidateArgs() {
	api.ValidateIds(f.PostIds)
	api.ValidateIds(f.FanclubIds)
	f.PostIds = api.RemoveSliceDuplicates(f.PostIds)

	if len(f.FanclubPageNums) > 0 {
		api.ValidatePageNumInput(
			len(f.FanclubIds),
			f.FanclubPageNums,
			[]string{
				"Number of Fantia Fanclub ID(s) and page numbers must be equal.",
			},
		)
	} else {
		f.FanclubPageNums = make([]string, len(f.FanclubIds))
	}

	f.FanclubIds, f.FanclubPageNums = api.RemoveDuplicateIdAndPageNum(
		f.FanclubIds,
		f.FanclubPageNums,
	)
}

// FantiaDlOptions is the struct that contains the options for downloading from Fantia.
type FantiaDlOptions struct {
	DlThumbnails     bool
	DlImages         bool
	DlAttachments    bool
	DlGdrive         bool
	AutoSolveCaptcha bool // whether to use chromedp to solve reCAPTCHA automatically

	GdriveClient *gdrive.GDrive

	Configs *configs.Config

	SessionCookieId string
	SessionCookies  []*http.Cookie

	csrfMu    sync.Mutex
	CsrfToken string

	Notifier       notify.Notifier
	captchaHandler constants.CAPTCHA_FN

	// Progress indicators
	CaptchaSolverProgBar   progress.Progress
	PostProgBar            progress.Progress
	GetFanclubPostsProgBar progress.Progress
	ProcessJsonProgBar     progress.Progress
}

func (f *FantiaDlOptions) GetConfigs() *configs.Config {
	return f.Configs
}

func (f *FantiaDlOptions) GetSessionCookies() []*http.Cookie {
	return f.SessionCookies
}

func (f *FantiaDlOptions) GetAutoSolveCaptcha() bool {
	return f.AutoSolveCaptcha
}

func (f *FantiaDlOptions) SetAutoSolveCaptcha(autoSolveCaptcha bool) {
	f.AutoSolveCaptcha = autoSolveCaptcha
}

func (f *FantiaDlOptions) GetNotifier() notify.Notifier {
	return f.Notifier
}

func (f *FantiaDlOptions) GetCaptchaSolverProgBar() progress.Progress {
	return f.CaptchaSolverProgBar
}

func (f *FantiaDlOptions) GetCaptchaHandler() constants.CAPTCHA_FN {
	return f.captchaHandler
}

// GetCsrfToken gets the CSRF token from Fantia's index HTML
// which is required to communicate with their API.
func (f *FantiaDlOptions) GetCsrfToken(userAgent string) error {
	f.csrfMu.Lock()
	defer f.csrfMu.Unlock()

	useHttp3 := httpfuncs.IsHttp3Supported(constants.FANTIA, false)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Method:      "GET",
			Url:         "https://fantia.jp/",
			Cookies:     f.SessionCookies,
			Http2:       !useHttp3,
			Http3:       useHttp3,
			CheckStatus: true,
			UserAgent:   userAgent,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"fantia error %d, failed to get CSRF token from Fantia: %w",
			constants.CONNECTION_ERROR,
			err,
		)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf(
			"fantia error %d, failed to get CSRF token from Fantia: %w",
			constants.RESPONSE_ERROR,
			err,
		)
	}

	// parse the response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return fmt.Errorf(
			"fantia error %d, failed to parse response body when getting CSRF token from Fantia: %w",
			constants.HTML_ERROR,
			err,
		)
	}

	if csrfToken, ok := doc.Find("meta[name=csrf-token]").Attr("content"); !ok {
		// shouldn't happen but just in case if Fantia's csrf token changes
		docHtml, err := doc.Html()
		if err != nil {
			docHtml = "failed to get HTML"
		}
		return fmt.Errorf(
			"fantia error %d, failed to get CSRF Token from Fantia, please report this issue!\nHTML: %s",
			constants.HTML_ERROR,
			docHtml,
		)
	} else {
		f.CsrfToken = csrfToken
	}
	return nil
}

// ValidateArgs validates the options for downloading from Fantia.
//
// Should be called after initialising the struct.
func (f *FantiaDlOptions) ValidateArgs(userAgent string) error {
	if f.SessionCookieId != "" {
		if cookie, err := api.VerifyAndGetCookie(constants.FANTIA, f.SessionCookieId, userAgent); err != nil {
			return err
		} else {
			f.SessionCookies = []*http.Cookie{cookie}
		}
	}

	if f.DlGdrive && f.GdriveClient == nil {
		f.DlGdrive = false
	} else if !f.DlGdrive && f.GdriveClient != nil {
		f.GdriveClient = nil
	}

	return f.GetCsrfToken(userAgent)
}
