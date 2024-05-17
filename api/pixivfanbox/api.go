package pixivfanbox

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/cache"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

// Returns a defined request header needed to communicate with Pixiv Fanbox's API
func GetPixivFanboxHeaders() map[string]string {
	return map[string]string{
		"Origin":  constants.PIXIV_FANBOX_URL,
		"Referer": constants.PIXIV_FANBOX_URL,
	}
}

type resChanVal struct {
	cacheKey string
	response *http.Response
}

// Query Pixiv Fanbox's API based on the slice of post IDs and
// returns a map of urls and a map of GDrive urls to download from.
func (pf *PixivFanboxDl) GetPostDetails(dlOptions *PixivFanboxDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, []error) {
	maxConcurrency := constants.PIXIV_FANBOX_MAX_CONCURRENT
	postIdsLen := len(pf.PostIds)
	if postIdsLen < maxConcurrency {
		maxConcurrency = postIdsLen
	}
	var wg sync.WaitGroup
	queue := make(chan struct{}, maxConcurrency)
	resChan := make(chan *resChanVal, postIdsLen)
	errChan := make(chan error, postIdsLen)

	baseMsg := "Getting post details from Pixiv Fanbox [%d/" + fmt.Sprintf("%d]...", postIdsLen)
	progress := dlOptions.MainProgBar
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting %d post details from Pixiv Fanbox!",
			postIdsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting %d post details from Pixiv Fanbox.\nPlease refer to the logs for more details.",
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
		if dlOptions.UseCacheDb {
			cacheKey = fmt.Sprintf("%s?postId=%s", url, postId)
			if cache.PostCacheExists(cacheKey, constants.PIXIV_FANBOX) {
				progress.Increment()
				continue
			}
		}

		wg.Add(1)
		go func(postId, cacheKey string) {
			defer func() {
				wg.Done()
				<-queue
			}()

			queue <- struct{}{}
			header := GetPixivFanboxHeaders()
			params := map[string]string{"postId": postId}
			res, err := httpfuncs.CallRequest(
				&httpfuncs.RequestArgs{
					Method:    "GET",
					Url:       url,
					Cookies:   dlOptions.SessionCookies,
					Headers:   header,
					Params:    params,
					UserAgent: dlOptions.Configs.UserAgent,
					Http2:     !useHttp3,
					Http3:     useHttp3,
					Context:   dlOptions.ctx,
				},
			)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					errChan <- err
				} else {
					errChan <- fmt.Errorf(
						"pixiv fanbox error %d: failed to get post details for %s, more info => %w",
						cdlerrors.CONNECTION_ERROR,
						url,
						err,
					)
				}
			} else if res.StatusCode != 200 {
				errChan <- fmt.Errorf(
					"pixiv fanbox error %d: failed to get post details for %s due to a %s response",
					cdlerrors.CONNECTION_ERROR,
					url,
					res.Status,
				)
			} else {
				if dlOptions.UseCacheDb {
					cache.CachePost(cacheKey)
				}
				resChan <- &resChanVal{cacheKey: cacheKey, response: res}
			}
			progress.Increment()
		}(postId, cacheKey)
	}
	wg.Wait()
	close(queue)
	close(resChan)
	close(errChan)

	hasErr := false
	hasCancelled := false
	var errSlice []error
	if len(errChan) > 0 {
		hasErr = true
		var errCtxCancelled bool
		if errCtxCancelled, errSlice = logger.LogChanErrors(logger.ERROR, errChan); errCtxCancelled {
			hasCancelled = true
		}
	}
	if hasCancelled {
		dlOptions.CancelCtx()
		progress.StopInterrupt("Stopped getting post details from Pixiv Fanbox...")
		return nil, nil, errSlice
	}
	progress.Stop(hasErr)

	urlsSlice, gdriveUrls := processMultiplePostJson(resChan, dlOptions)
	return urlsSlice, gdriveUrls, errSlice
}

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
			Cookies:   dlOptions.SessionCookies,
			Headers:   headers,
			Params:    params,
			UserAgent: dlOptions.Configs.UserAgent,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			Context:   dlOptions.GetContext(),
		},
	)
	if err != nil || res.StatusCode != 200 {
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
			res.Body.Close()
			err = fmt.Errorf(
				"%s %d: failed to get creator's posts for %s due to %s response",
				"pixiv fanbox error",
				cdlerrors.RESPONSE_ERROR,
				creatorId,
				res.Status,
			)
		}
		return nil, err
	}

	var resJson CreatorPaginatedPostsJson
	if err := httpfuncs.LoadJsonFromResponse(res, &resJson); err != nil {
		return nil, err
	}
	return resJson.Body, nil
}

type resStruct struct {
	json *FanboxCreatorPostsJson
	err  error
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

	minPage, maxPage, hasMax, err := api.GetMinMaxFromStr(pageNum)
	if err != nil {
		return nil, []error{err}, false
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV_FANBOX, true)
	headers := GetPixivFanboxHeaders()
	var wg sync.WaitGroup
	maxConcurrency := constants.PIXIV_FANBOX_MAX_CONCURRENT
	if len(paginatedUrls) < maxConcurrency {
		maxConcurrency = len(paginatedUrls)
	}
	queue := make(chan struct{}, maxConcurrency)
	resChan := make(chan *resStruct, len(paginatedUrls))
	for idx, paginatedUrl := range paginatedUrls {
		curPage := idx + 1
		if curPage < minPage {
			continue
		}
		if hasMax && curPage > maxPage {
			break
		}

		wg.Add(1)
		go func(reqUrl string) {
			defer func() {
				wg.Done()
				<-queue
			}()
			queue <- struct{}{}
			res, err := httpfuncs.CallRequest(
				&httpfuncs.RequestArgs{
					Method:    "GET",
					Url:       reqUrl,
					Cookies:   dlOptions.SessionCookies,
					Headers:   headers,
					UserAgent: dlOptions.Configs.UserAgent,
					Http2:     !useHttp3,
					Http3:     useHttp3,
					Context:   dlOptions.GetContext(),
				},
			)
			if err != nil || res.StatusCode != 200 {
				if err == nil {
					res.Body.Close()
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
				return
			}

			var resJson *FanboxCreatorPostsJson
			if err := httpfuncs.LoadJsonFromResponse(res, &resJson); err != nil {
				resChan <- &resStruct{err: err}
			} else {
				resChan <- &resStruct{json: resJson}
			}
		}(paginatedUrl)
	}
	wg.Wait()
	close(queue)
	close(resChan)

	// parse the JSON response
	for res := range resChan {
		if res.err != nil {
			errSlice = append(errSlice, res.err)
			continue
		}

		for _, postInfoMap := range res.json.Body.Items {
			postIds = append(postIds, postInfoMap.Id)
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
	progress := dlOptions.MainProgBar
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
	pf.PostIds = api.RemoveSliceDuplicates(pf.PostIds)
	return errSlice
}
