package fantia

import (
	"errors"
	"fmt"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

type fantiaPostArgs struct {
	msgSuffix string
	postId    string
	url       string
}

func getFantiaPostDetails(postArg *fantiaPostArgs, dlOptions *FantiaDlOptions) (*httpfuncs.ResponseWrapper, error) {
	// Now that we have the post ID, we can query Fantia's API
	// to get the post's contents from the JSON response.
	progress := dlOptions.Base.MainProgBar()
	progress.SetToSpinner()
	progress.UpdateBaseMsg(
		fmt.Sprintf(
			"Getting post %s's contents from Fantia %s...",
			postArg.postId,
			postArg.msgSuffix,
		),
	)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting post %s's contents from Fantia %s!",
			postArg.postId,
			postArg.msgSuffix,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting post %s's cotents from Fantia %s.\nPlease refer to the logs for more details.",
			postArg.postId,
			postArg.msgSuffix,
		),
	)
	progress.Start()
	defer progress.SnapshotTask()

	postApiUrl := postArg.url + postArg.postId
	header := map[string]string{
		"Referer":          fmt.Sprintf("%s/posts/%s", constants.FANTIA_URL, postArg.postId),
		"X-Csrf-Token":     dlOptions.CsrfToken,
		"X-Requested-With": "XMLHttpRequest",
	}
	useHttp3 := httpfuncs.IsHttp3Supported(constants.FANTIA, true)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Method:         "GET",
			Url:            postApiUrl,
			Cookies:        dlOptions.Base.SessionCookies,
			Headers:        header,
			Http2:          !useHttp3,
			Http3:          useHttp3,
			UserAgent:      dlOptions.Base.Configs.UserAgent,
			Context:        dlOptions.GetContext(),
			CaptchaCheck:   CaptchaChecker,
			CaptchaHandler: newCaptchaHandler(dlOptions),
		},
	)
	if err != nil || res.Resp.StatusCode != 200 {
		errCode := cdlerrors.CONNECTION_ERROR
		if err == nil {
			errCode = res.Resp.StatusCode
		}

		errMsg := fmt.Sprintf(
			"fantia error %d: failed to get post details for %s",
			errCode,
			postApiUrl,
		)
		if err != nil {
			err = fmt.Errorf(
				"%s, more info => %w",
				errMsg,
				err,
			)
		} else {
			err = errors.New(errMsg)
		}

		progress.Stop(true)
		return nil, err
	}

	progress.Stop(false)
	return res, nil
}

func DlFantiaPost(count, maxCount int, postId string, dlOptions *FantiaDlOptions) (cancelled bool, gdriveUrls []*httpfuncs.ToDownload, errSlice []error) {
	msgSuffix := fmt.Sprintf(
		"[%d/%d]",
		count,
		maxCount,
	)

	var cacheKey string
	if dlOptions.Base.UseCacheDb {
		url := constants.FANTIA_POST_API_URL + postId
		if database.PostCacheExists(url, constants.FANTIA) {
			return false, nil, nil
		}
		cacheKey = database.ParsePostKey(url, constants.FANTIA)
	}

	respWrapper, err := getFantiaPostDetails(
		&fantiaPostArgs{
			msgSuffix: msgSuffix,
			postId:    postId,
			url:       constants.FANTIA_POST_API_URL,
		},
		dlOptions,
	)
	if err != nil {
		return false, nil, []error{err}
	}

	urlsToDownload, postGdriveUrls, err := processIllustDetailApiRes(
		&processIllustArgs{
			respWrapper: respWrapper,
			postId:      postId,
			postIdsLen:  maxCount,
			msgSuffix:   msgSuffix,
		},
		dlOptions,
	)
	if err != nil {
		return false, nil, []error{err}
	}

	// Download the urls
	useHttp3 := httpfuncs.IsHttp3Supported(constants.FANTIA, false)
	cancelled, errorSlice := httpfuncs.DownloadUrls(
		urlsToDownload,
		&httpfuncs.DlOptions{
			Context:         dlOptions.GetContext(),
			MaxConcurrency:  constants.FANTIA_MAX_CONCURRENCY,
			Headers:         nil,
			Cookies:         dlOptions.Base.SessionCookies,
			UseHttp3:        useHttp3,
			HeadReqTimeout:  constants.DEFAULT_HEAD_REQ_TIMEOUT,
			SupportRange:    constants.FANTIA_RANGE_SUPPORTED,
			SetMetadata:     dlOptions.Base.SetMetadata,
			Filters:         dlOptions.Base.Filters,
			ProgressBarInfo: dlOptions.Base.ProgressBarInfo,
		},
		dlOptions.Base.Configs,
	)

	if cancelled {
		dlOptions.CancelCtx()
		return true, nil, errorSlice
	}
	if dlOptions.Base.UseCacheDb {
		// No need to use batch since posts are downloaded sequentially
		database.CachePost(cacheKey)
	}
	return false, postGdriveUrls, nil
}

// Query Fantia's API based on the slice of post IDs and get a map of urls to download from.
//
// Note that only the downloading of the URL(s) is/are executed concurrently
// to reduce the chance of the signed AWS S3 URL(s) from expiring before the download is
// executed or completed due to a download queue to avoid resource exhaustion of the user's system.
func (f *FantiaDl) DlFantiaPosts(dlOptions *FantiaDlOptions) ([]*httpfuncs.ToDownload, []error) {
	var errSlice []error
	var gdriveLinks []*httpfuncs.ToDownload
	postIdsLen := len(f.PostIds)
	for i, postId := range f.PostIds {
		cancelled, postGdriveLinks, err := DlFantiaPost(i+1, postIdsLen, postId, dlOptions)
		if len(err) > 0 {
			if cancelled {
				return nil, nil
			}

			errSlice = append(errSlice, err...)
			continue
		}

		if len(postGdriveLinks) > 0 {
			gdriveLinks = append(gdriveLinks, postGdriveLinks...)
		}
	}

	if len(errSlice) > 0 {
		logger.LogErrors(logger.ERROR, errSlice...)
	}
	return gdriveLinks, errSlice
}
