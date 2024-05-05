package fantia

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"path/filepath"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
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
func (f *FantiaDl) ValidateArgs() error {
	err := api.ValidateIds(f.PostIds)
	if err != nil {
		return err
	}

	err = api.ValidateIds(f.FanclubIds)
	if err != nil {
		return err
	}

	f.PostIds = api.RemoveSliceDuplicates(f.PostIds)
	if len(f.FanclubPageNums) > 0 {
		err := api.ValidatePageNumInput(
			len(f.FanclubIds),
			f.FanclubPageNums,
			[]string{
				"Number of Fantia Fanclub ID(s) and page numbers must be equal.",
			},
		)
		if err != nil {
			return err
		}
	} else {
		f.FanclubPageNums = make([]string, len(f.FanclubIds))
	}

	f.FanclubIds, f.FanclubPageNums = api.RemoveDuplicateIdAndPageNum(
		f.FanclubIds,
		f.FanclubPageNums,
	)
	return err
}

// FantiaDlOptions is the struct that contains the options for downloading from Fantia.
type FantiaDlOptions struct {
	ctx                 context.Context
	cancel              context.CancelFunc
	DlThumbnails        bool
	DlImages            bool
	DlAttachments       bool
	DlGdrive            bool
	DetectOtherDlLinks  bool
	BaseDownloadDirPath string

	GdriveClient *gdrive.GDrive

	Configs *configs.Config

	SessionCookieId string
	SessionCookies  []*http.Cookie

	csrfMu    sync.Mutex
	CsrfToken string

	Notifier       notify.Notifier

	// Progress indicators
	MainProgBar          progress.ProgressBar
	DownloadProgressBars *[]*progress.DownloadProgressBar
}

func (f *FantiaDlOptions) GetConfigs() *configs.Config {
	return f.Configs
}

func (f *FantiaDlOptions) GetSessionCookies() []*http.Cookie {
	return f.SessionCookies
}

func (f *FantiaDlOptions) GetNotifier() notify.Notifier {
	return f.Notifier
}

func (f *FantiaDlOptions) GetContext() context.Context {
	return f.ctx
}

func (f *FantiaDlOptions) GetCancel() context.CancelFunc {
	return f.cancel
}

func (f *FantiaDlOptions) SetContext(ctx context.Context) {
	f.ctx, f.cancel = context.WithCancel(ctx)
}

// Cancel cancels the context of the FantiaDlOptions struct.
func (f *FantiaDlOptions) Cancel() {
	f.cancel()
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
			errs.CONNECTION_ERROR,
			err,
		)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf(
			"fantia error %d, failed to get CSRF token from Fantia: %w",
			errs.RESPONSE_ERROR,
			err,
		)
	}

	// parse the response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return fmt.Errorf(
			"fantia error %d, failed to parse response body when getting CSRF token from Fantia: %w",
			errs.HTML_ERROR,
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
			errs.HTML_ERROR,
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
	if f.GetContext() == nil {
		f.SetContext(context.Background())
	}

	if f.Notifier == nil {
		return fmt.Errorf(
			"fantia error %d, notifier is nil",
			errs.DEV_ERROR,
		)
	}

	if f.BaseDownloadDirPath == "" {
		f.BaseDownloadDirPath = filepath.Join(iofuncs.DOWNLOAD_PATH, constants.FANTIA_TITLE)
	} else {
		if !iofuncs.DirPathExists(f.BaseDownloadDirPath) {
			return fmt.Errorf(
				"fantia error %d, download path does not exist or is not a directory, please create the directory and try again",
				errs.INPUT_ERROR,
			)
		}
	}

	if f.MainProgBar == nil {
		return fmt.Errorf(
			"fantia error %d, main progress bar is nil",
			errs.DEV_ERROR,
		)
	}

	if f.SessionCookieId != "" {
		if cookie, err := api.VerifyAndGetCookie(constants.FANTIA, f.SessionCookieId, userAgent); err != nil {
			return err
		} else {
			f.SessionCookies = []*http.Cookie{cookie}
		}
	} else if len(f.SessionCookies) > 0 {
		if err := api.VerifyCookies(constants.FANTIA, userAgent, f.SessionCookies); err != nil {
			return err
		}
	} 

	if f.DlGdrive && f.GdriveClient == nil {
		f.DlGdrive = false
	} else if !f.DlGdrive && f.GdriveClient != nil {
		f.GdriveClient = nil
	}

	return f.GetCsrfToken(userAgent)
}
