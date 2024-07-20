package pixivfanbox

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils/threadsafe"
)

// Returns a defined request header needed to communicate with Pixiv Fanbox's API
func GetPixivFanboxHeaders() map[string]string {
	return map[string]string{
		"Origin":  constants.PIXIV_FANBOX_URL,
		"Referer": constants.PIXIV_FANBOX_URL,
	}
}

func getPostDetails(cacheKey, postId, url string, dlOptions *PixivFanboxDlOptions, useHttp3 bool) (*http.Response, string, error) {
	header := GetPixivFanboxHeaders()
	params := map[string]string{"postId": postId}
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Method:    "GET",
			Url:       url,
			Cookies:   dlOptions.Base.SessionCookies,
			Headers:   header,
			Params:    params,
			UserAgent: dlOptions.Base.Configs.UserAgent,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			Context:   dlOptions.ctx,
			CaptchaHandler: httpfuncs.CaptchaHandler{
				Check:                CaptchaChecker,
				Handler:              NewCaptchaHandler(dlOptions),
				InjectCaptchaCookies: GetCachedCfCookies,
			},
		},
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, "", err
		}
		return nil, "", fmt.Errorf(
			"pixiv fanbox error %d: failed to get post details for %s, more info => %w",
			cdlerrors.CONNECTION_ERROR,
			url,
			err,
		)
	}

	if res.Resp.StatusCode != 200 {
		return nil, "", fmt.Errorf(
			"pixiv fanbox error %d: failed to get post details for %s due to a %s response",
			cdlerrors.CONNECTION_ERROR,
			url,
			res.Resp.Status,
		)
	}

	if dlOptions.Base.UseCacheDb {
		cacheKey = database.ParsePostKey(cacheKey, constants.PIXIV_FANBOX)
		database.CachePostViaBatch(cacheKey)
	}
	return res.Resp, cacheKey, nil
}

type urlsChanVal struct {
	postUrls   []*httpfuncs.ToDownload
	gdriveUrls []*httpfuncs.ToDownload
}

// Query Pixiv Fanbox's API based on the slice of post IDs and
// returns a map of urls and a map of GDrive urls to download from.
func (pf *PixivFanboxDl) GetPostDetails(dlOptions *PixivFanboxDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, []error) {
	maxConcurrency := constants.PIXIV_FANBOX_MAX_CONCURRENCY
	postIdsLen := len(pf.PostIds)
	if postIdsLen < maxConcurrency {
		maxConcurrency = postIdsLen
	}
	var wg sync.WaitGroup
	queue := make(chan struct{}, maxConcurrency)
	urlsTsSlice := threadsafe.NewSliceWithCapacity[*urlsChanVal](postIdsLen)
	errTsSlice := threadsafe.NewSlice[error]()

	baseMsg := "Getting and processing post details from Pixiv Fanbox [%d/" + fmt.Sprintf("%d]...", postIdsLen)
	progress := dlOptions.Base.MainProgBar()
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting and processing %d post details from Pixiv Fanbox!",
			postIdsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting and processing %d post details from Pixiv Fanbox.\nPlease refer to the logs for more details.",
			postIdsLen,
		),
	)
	progress.SetToProgressBar()
	progress.UpdateMax(postIdsLen)
	progress.Start()
	defer progress.SnapshotTask()

	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV_FANBOX, true)
	url := fmt.Sprintf("%s/post.info", constants.PIXIV_FANBOX_API_URL)
	for _, postId := range pf.PostIds {
		var cacheKey string
		if dlOptions.Base.UseCacheDb {
			fullUrl := fmt.Sprintf("%s?postId=%s", url, postId)
			if database.PostCacheExists(fullUrl, constants.PIXIV_FANBOX) {
				progress.Increment()
				continue
			}
			cacheKey = database.ParsePostKey(fullUrl, constants.PIXIV_FANBOX)
		}

		wg.Add(1)
		go func() {
			defer func() {
				progress.Increment()
				wg.Done()
				<-queue
			}()

			queue <- struct{}{}
			res, parsedCacheKey, err := getPostDetails(cacheKey, postId, url, dlOptions, useHttp3)
			if err != nil {
				errTsSlice.Append(err)
				return
			}

			postUrls, postGdriveLinks, err := processFanboxPostJson(res, dlOptions)
			if err != nil {
				errTsSlice.Append(err)
			} else {
				if dlOptions.Base.UseCacheDb && parsedCacheKey != "" {
					for _, url := range postUrls {
						url.CacheKey = parsedCacheKey
						url.CacheFn = database.CachePost
					}
				}
				urlsTsSlice.Append(&urlsChanVal{
					postUrls:   postUrls,
					gdriveUrls: postGdriveLinks,
				})
			}
		}()
	}
	wg.Wait()
	close(queue)

	hasErr := errTsSlice.LenUnsafe() > 0
	hasCancelled := false
	var errSlice []error
	if hasErr {
		var errCtxCancelled bool
		if errCtxCancelled, errSlice = logger.LogSliceErrors(logger.ERROR, errTsSlice); errCtxCancelled {
			hasCancelled = true
		}
	}
	if hasCancelled {
		dlOptions.CancelCtx()
		progress.StopInterrupt("Stopped getting and processing post details from Pixiv Fanbox...")
		return nil, nil, errSlice
	}
	progress.Stop(hasErr)

	var urlsSlice, gdriveUrls []*httpfuncs.ToDownload
	urlsIter := urlsTsSlice.NewIter()
	for urlsIter.Next() {
		urls := urlsIter.Item()
		urlsSlice = append(urlsSlice, urls.postUrls...)
		gdriveUrls = append(gdriveUrls, urls.gdriveUrls...)
	}
	return urlsSlice, gdriveUrls, errSlice
}

// To get all the paginated URL(s) from the api
func getCreatorPaginatedPosts(creatorId string, dlOptions *PixivFanboxDlOptions) ([]string, error) {
	params := map[string]string{"creatorId": creatorId}
	headers := GetPixivFanboxHeaders()
	url := fmt.Sprintf(
		"%s/post.paginateCreator",
		constants.PIXIV_FANBOX_API_URL,
	)
	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV_FANBOX, true)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Method:    "GET",
			Url:       url,
			Cookies:   dlOptions.Base.SessionCookies,
			Headers:   headers,
			Params:    params,
			UserAgent: dlOptions.Base.Configs.UserAgent,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			Context:   dlOptions.GetContext(),
			CaptchaHandler: httpfuncs.CaptchaHandler{
				Check:                CaptchaChecker,
				Handler:              NewCaptchaHandler(dlOptions),
				InjectCaptchaCookies: GetCachedCfCookies,
			},
		},
	)
	if err != nil || res.Resp.StatusCode != 200 {
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil, err
			}
			err = fmt.Errorf(
				"%s %d: failed to get creator's posts for %s due to %w",
				"pixiv fanbox error",
				cdlerrors.CONNECTION_ERROR,
				creatorId,
				err,
			)
		} else {
			res.Close()
			err = fmt.Errorf(
				"%s %d: failed to get creator's posts for %s due to %s response",
				"pixiv fanbox error",
				cdlerrors.RESPONSE_ERROR,
				creatorId,
				res.Resp.Status,
			)
		}
		return nil, err
	}

	var resJson CreatorPaginatedPostsJson
	if err := httpfuncs.LoadJsonFromResponse(res.Resp, &resJson); err != nil {
		return nil, err
	}
	return resJson.Body, nil
}

type resStruct struct {
	json *FanboxCreatorPostsJson
	err  error
}

func getFanboxPostsLogic(reqUrl string, headers map[string]string, dlOptions *PixivFanboxDlOptions, useHttp3 bool) *resStruct {
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Method:    "GET",
			Url:       reqUrl,
			Cookies:   dlOptions.Base.SessionCookies,
			Headers:   headers,
			UserAgent: dlOptions.Base.Configs.UserAgent,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			Context:   dlOptions.GetContext(),
			CaptchaHandler: httpfuncs.CaptchaHandler{
				Check:                CaptchaChecker,
				Handler:              NewCaptchaHandler(dlOptions),
				InjectCaptchaCookies: GetCachedCfCookies,
			},
		},
	)
	if err != nil || res.Resp.StatusCode != 200 {
		if err == nil {
			res.Close()
		}
		if !errors.Is(err, context.Canceled) {
			logger.LogError(
				fmt.Errorf(
					"failed to get post for %s\n%w",
					reqUrl,
					err,
				),
				logger.ERROR,
			)
		}
		return nil
	}

	var resJson *FanboxCreatorPostsJson
	if err := httpfuncs.LoadJsonFromResponse(res.Resp, &resJson); err != nil {
		return &resStruct{err: err}
	}
	return &resStruct{json: resJson}
}

// GetFanboxCreatorPosts returns a slice of post IDs for a given creator
func getFanboxPosts(creatorId, pageNum string, dlOptions *PixivFanboxDlOptions) (postIds []string, errSlice []error, hasCancelled bool) {
	paginatedUrls, err := getCreatorPaginatedPosts(creatorId, dlOptions)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, nil, true
		}
		return nil, []error{err}, false
	}

	minPage, maxPage, hasMax, err := utils.GetMinMaxFromStr(pageNum)
	if err != nil {
		return nil, []error{err}, false
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV_FANBOX, true)
	headers := GetPixivFanboxHeaders()
	var wg sync.WaitGroup
	maxConcurrency := constants.PIXIV_FANBOX_MAX_CONCURRENCY
	if len(paginatedUrls) < maxConcurrency {
		maxConcurrency = len(paginatedUrls)
	}
	queue := make(chan struct{}, maxConcurrency)
	resTsSlice := threadsafe.NewSliceWithCapacity[*resStruct](len(paginatedUrls))
	for idx, paginatedUrl := range paginatedUrls {
		curPage := idx + 1
		if curPage < minPage {
			continue
		}
		if hasMax && curPage > maxPage {
			break
		}

		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				<-queue
			}()
			queue <- struct{}{}
			resTsSlice.Append(
				getFanboxPostsLogic(paginatedUrl, headers, dlOptions, useHttp3),
			)
		}()
	}
	wg.Wait()
	close(queue)

	// parse the JSON response
	resIter := resTsSlice.NewIter()
	for resIter.Next() {
		res := resIter.Item()
		if res.err != nil {
			errSlice = append(errSlice, res.err)
			continue
		}

		for _, postInfoMap := range res.json.Body.Items {
			if dlOptions.Base.Filters.IsPostDateValid(postInfoMap.PublishedDatetime) {
				postIds = append(postIds, postInfoMap.ID)
			}
		}
	}

	hasCancelled = false
	if len(errSlice) > 0 {
		hasCancelled = logger.LogErrors(logger.ERROR, errSlice...)
		if hasCancelled {
			dlOptions.CancelCtx()
		}
	}
	return postIds, errSlice, hasCancelled
}

// Retrieves all the posts based on the slice of creator IDs and updates its slice of post IDs accordingly
func (pf *PixivFanboxDl) GetCreatorsPosts(dlOptions *PixivFanboxDlOptions) []error {
	creatorIdsLen := len(pf.CreatorIds)
	if creatorIdsLen != len(pf.CreatorPageNums) {
		return []error{
			fmt.Errorf(
				"pixiv fanbox error %d: length of creator IDs and page numbers are not equal",
				cdlerrors.DEV_ERROR,
			),
		}
	}

	var errSlice []error
	progress := dlOptions.Base.MainProgBar()
	if len(pf.CreatorIds) > 1 {
		baseMsg := "Getting post ID(s) from creator(s) on Pixiv Fanbox [%d/" + fmt.Sprintf("%d]...", creatorIdsLen)
		progress.UpdateBaseMsg(baseMsg)
		progress.UpdateSuccessMsg(
			fmt.Sprintf(
				"Finished getting post ID(s) from %d creator(s) on Pixiv Fanbox!",
				creatorIdsLen,
			),
		)
		progress.UpdateErrorMsg(
			fmt.Sprintf(
				"Something went wrong while getting post IDs from %d creator(s) on Pixiv Fanbox!\nPlease refer to logs for more details.",
				creatorIdsLen,
			),
		)
		progress.SetToProgressBar()
		progress.UpdateMax(creatorIdsLen)
	} else {
		progress.SetToSpinner()
		creatorId := pf.CreatorIds[0]
		progress.UpdateBaseMsg("Getting post ID(s) from creator, " + creatorId + ", on Pixiv Fanbox...")
		progress.UpdateSuccessMsg("Finished getting post ID(s) from creator, " + creatorId + ", on Pixiv Fanbox!")
		progress.UpdateErrorMsg("Something went wrong while getting post ID(s) from creator, " + creatorId + ", on Pixiv Fanbox.\nPlease refer to the logs for more details.")
	}

	progress.Start()
	defer progress.SnapshotTask()
	for idx, creatorId := range pf.CreatorIds {
		retrievedPostIds, err, hasCancelled := getFanboxPosts(
			creatorId,
			pf.CreatorPageNums[idx],
			dlOptions,
		)
		if hasCancelled {
			dlOptions.CancelCtx()
			progress.StopInterrupt("Stopped getting post IDs from creator(s) on Pixiv Fanbox...")
			return nil
		}

		if err != nil {
			errSlice = append(errSlice, err...)
		} else {
			pf.PostIds = append(pf.PostIds, retrievedPostIds...)
		}
		progress.Increment()
	}

	hasErr := false
	if len(errSlice) > 0 {
		hasErr = true
		logger.LogErrors(logger.ERROR, errSlice...)
	}
	progress.Stop(hasErr)
	pf.PostIds = utils.RemoveSliceDuplicates(pf.PostIds)
	return errSlice
}
