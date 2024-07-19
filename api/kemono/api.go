package kemono

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils/threadsafe"
	"github.com/PuerkitoBio/goquery"
)

type kemonoChanRes struct {
	urlsToDownload []*httpfuncs.ToDownload
	gdriveLinks    []*httpfuncs.ToDownload
	err            error
}

func getKemonoPartyHeaders() map[string]string {
	return map[string]string{}
}

func parseCreatorHtmlAndGetName(res *http.Response, url string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(res.Body)
	res.Body.Close()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return "", err
		}

		err = fmt.Errorf(
			"kemono error %d, failed to parse response body when getting creator name from Kemono at %s\nmore info => %w",
			cdlerrors.HTML_ERROR,
			url,
			err,
		)
		return "", err
	}

	// <span itemprop="name">creator-name</span> => creator-name
	creatorName := doc.Find("span[itemprop=name]").Text()
	if creatorName == "" {
		return "", fmt.Errorf(
			"kemono error %d, failed to get creator name from Kemono at %s\nplease report this issue",
			cdlerrors.HTML_ERROR,
			url,
		)
	}

	return creatorName, nil
}

var creatorNameCacheLock sync.Mutex
var creatorNameCache = make(map[string]string)

func getCreatorName(service, userId string, dlOptions *KemonoDlOptions) (string, error) {
	url := fmt.Sprintf(
		"%s/%s/user/%s",
		constants.KEMONO_URL,
		service,
		userId,
	)

	var cacheKey string
	if dlOptions.Base.UseCacheDb {
		if name := database.GetKemonoCreatorCache(url); name != "" {
			return name, nil
		}
	} else {
		creatorNameCacheLock.Lock()
		defer creatorNameCacheLock.Unlock()
		cacheKey = fmt.Sprintf("%s/%s", service, userId)
		if name, ok := creatorNameCache[cacheKey]; ok {
			return name, nil
		}
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.KEMONO, true)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Url:         url,
			Method:      "GET",
			Headers:     getKemonoPartyHeaders(),
			UserAgent:   dlOptions.Base.Configs.UserAgent,
			Cookies:     dlOptions.Base.SessionCookies,
			Http2:       !useHttp3,
			Http3:       useHttp3,
			CheckStatus: true,
			Context:     dlOptions.GetContext(),
		},
	)
	if err != nil {
		return userId, err
	}

	creatorName, err := parseCreatorHtmlAndGetName(res.Resp, url)
	if err != nil {
		return userId, err
	}

	if dlOptions.Base.UseCacheDb {
		database.CacheKemonoCreatorName(url, creatorName)
	} else {
		creatorNameCache[cacheKey] = creatorName
	}
	return creatorName, nil
}

func getPostDetails(post *KemonoPostToDl, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	url := fmt.Sprintf(
		"%s/%s/user/%s/post/%s",
		constants.KEMONO_API_URL,
		post.Service,
		post.CreatorId,
		post.PostId,
	)
	var cacheKey string
	if dlOptions.Base.UseCacheDb {
		if database.PostCacheExists(url, constants.KEMONO) {
			return nil, nil, nil
		}
		cacheKey = database.ParsePostKey(url, constants.KEMONO)
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.KEMONO, true)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Url:         url,
			Method:      "GET",
			Headers:     getKemonoPartyHeaders(),
			UserAgent:   dlOptions.Base.Configs.UserAgent,
			Cookies:     dlOptions.Base.SessionCookies,
			Http2:       !useHttp3,
			Http3:       useHttp3,
			CheckStatus: true,
			Context:     dlOptions.GetContext(),
		},
	)
	if err != nil {
		return nil, nil, err
	}

	// https://github.com/KJHJason/Cultured-Downloader-CLI/commit/e8d05e4a8e1db05d721964a93d933ca2504d0e1f
	var resJson MainKemonoJson
	if err := httpfuncs.LoadJsonFromResponse(res.Resp, &resJson); err != nil {
		return nil, nil, err
	}

	postsToDl, gdriveLinks := processMultipleJson(KemonoJson{&resJson}, dlOptions)
	if dlOptions.Base.UseCacheDb {
		for _, post := range postsToDl {
			post.CacheKey = cacheKey
			post.CacheFn = database.CachePost
		}
	}
	return postsToDl, gdriveLinks, nil
}

func GetMultiplePosts(posts []*KemonoPostToDl, dlOptions *KemonoDlOptions) (urlsToDownload []*httpfuncs.ToDownload, gdriveLinks []*httpfuncs.ToDownload, errSlice []error) {
	var maxConcurrency int
	postLen := len(posts)
	if postLen > constants.KEMONO_MAX_CONCURRENCY {
		maxConcurrency = constants.KEMONO_MAX_CONCURRENCY
	} else {
		maxConcurrency = postLen
	}
	wg := sync.WaitGroup{}
	queue := make(chan struct{}, maxConcurrency)
	resTsSlice := threadsafe.NewSliceWithCapacity[*kemonoChanRes](postLen)

	baseMsg := "Getting post details from Kemono [%d/" + fmt.Sprintf("%d]...", postLen)
	progress := dlOptions.Base.MainProgBar()
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting %d post details from Kemono!",
			postLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting %d post details from Kemono.\nPlease refer to the logs for more details.",
			postLen,
		),
	)
	progress.SetToProgressBar()
	progress.UpdateMax(postLen)
	progress.Start()
	defer progress.SnapshotTask()
	for _, post := range posts {
		wg.Add(1)
		go func() {
			defer func() {
				progress.Increment()
				wg.Done()
				<-queue
			}()

			queue <- struct{}{}
			toDownload, foundGdriveLinks, err := getPostDetails(post, dlOptions)
			if err != nil {
				resTsSlice.Append(&kemonoChanRes{
					err: err,
				})
				return
			}
			resTsSlice.Append(&kemonoChanRes{
				urlsToDownload: toDownload,
				gdriveLinks:    foundGdriveLinks,
			})
		}()
	}
	wg.Wait()
	close(queue)

	hasError, hasCancelled := false, false
	resIter := resTsSlice.NewIter()
	for resIter.Next() {
		res := resIter.Item()
		if res.err == nil {
			urlsToDownload = append(urlsToDownload, res.urlsToDownload...)
			gdriveLinks = append(gdriveLinks, res.gdriveLinks...)
			continue
		}

		if errors.Is(res.err, context.Canceled) {
			hasCancelled = true
			continue
		}
		if !hasError {
			hasError = true
		}
		logger.LogError(res.err, logger.ERROR)
		errSlice = append(errSlice, res.err)
	}

	if hasCancelled {
		dlOptions.CancelCtx()
		progress.StopInterrupt("Stopped getting post details from Kemono...")
		return nil, nil, errSlice
	}
	progress.Stop(hasError)
	return urlsToDownload, gdriveLinks, errSlice
}

func getCreatorPosts(creator *KemonoCreatorToDl, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	useHttp3 := httpfuncs.IsHttp3Supported(constants.KEMONO, true)
	minPage, maxPage, hasMax, err := utils.GetMinMaxFromStr(creator.PageNum)
	if err != nil {
		return nil, nil, err
	}
	minOffset, maxOffset := utils.ConvertPageNumToOffset(minPage, maxPage, constants.KEMONO_PER_PAGE)

	var postsToDl, gdriveLinksToDl []*httpfuncs.ToDownload
	params := make(map[string]string)
	curOffset := minOffset
	for {
		params["o"] = strconv.Itoa(curOffset)
		res, err := httpfuncs.CallRequest(
			&httpfuncs.RequestArgs{
				Url: fmt.Sprintf(
					"%s/%s/user/%s",
					constants.KEMONO_API_URL,
					creator.Service,
					creator.CreatorId,
				),
				Method:      "GET",
				UserAgent:   dlOptions.Base.Configs.UserAgent,
				Headers:     getKemonoPartyHeaders(),
				Cookies:     dlOptions.Base.SessionCookies,
				Params:      params,
				Http2:       !useHttp3,
				Http3:       useHttp3,
				CheckStatus: true,
				Context:     dlOptions.GetContext(),
			},
		)
		if err != nil {
			return nil, nil, err
		}

		var resJson KemonoJson
		if err := httpfuncs.LoadJsonFromResponse(res.Resp, &resJson); err != nil {
			return nil, nil, err
		}

		if len(resJson) == 0 {
			break
		}

		posts, gdriveLinks := processMultipleJson(resJson, dlOptions)
		postsToDl = append(postsToDl, posts...)
		gdriveLinksToDl = append(gdriveLinksToDl, gdriveLinks...)

		if hasMax && curOffset >= maxOffset {
			break
		}
		curOffset += constants.KEMONO_PER_PAGE
	}
	return postsToDl, gdriveLinksToDl, nil
}

func GetMultipleCreators(creators []*KemonoCreatorToDl, dlOptions *KemonoDlOptions) (urlsToDownload []*httpfuncs.ToDownload, gdriveLinks []*httpfuncs.ToDownload, errSlice []error) {
	creatorLen := len(creators)

	progress := dlOptions.Base.MainProgBar()
	if creatorLen > 1 {
		baseMsg := "Getting creator's posts from Kemono [%d/" + fmt.Sprintf("%d]...", creatorLen)
		progress.UpdateBaseMsg(baseMsg)
		progress.UpdateSuccessMsg(
			fmt.Sprintf(
				"Finished getting %d creator's posts from Kemono!",
				creatorLen,
			),
		)
		progress.UpdateErrorMsg(
			fmt.Sprintf(
				"Something went wrong while getting %d creator's posts from Kemono.\nPlease refer to the logs for more details.",
				creatorLen,
			),
		)
		progress.SetToProgressBar()
		progress.UpdateMax(creatorLen)
	} else {
		progress.SetToSpinner()
		creatorId := creators[0].CreatorId
		progress.UpdateBaseMsg("Getting posts from creator, " + creatorId + ", on Kemono...")
		progress.UpdateSuccessMsg("Finished getting posts from creator, " + creatorId + ", on Kemono!")
		progress.UpdateErrorMsg("Something went wrong while getting posts from creator, " + creatorId + ", on Kemono.\nPlease refer to the logs for more details.")
	}
	progress.Start()
	defer progress.SnapshotTask()

	hasCancelled := false
	for _, creator := range creators {
		postsToDl, gdriveLinksToDl, err := getCreatorPosts(creator, dlOptions)
		if err == nil {
			urlsToDownload = append(urlsToDownload, postsToDl...)
			gdriveLinks = append(gdriveLinks, gdriveLinksToDl...)
			progress.Increment()
			continue
		}

		if errors.Is(err, context.Canceled) {
			hasCancelled = true
			progress.StopInterrupt("Stopped getting creator's posts from Kemono...")
			break
		}
		errSlice = append(errSlice, err)
		progress.Increment()
	}

	hasErr := false
	if len(errSlice) > 0 {
		hasErr = true
		logger.LogErrors(logger.ERROR, errSlice...)
	}
	if hasCancelled {
		dlOptions.CancelCtx()
		return nil, nil, errSlice
	}
	progress.Stop(hasErr)
	return urlsToDownload, gdriveLinks, errSlice
}

func processFavCreator(resJson KemonoFavCreatorJson) []*KemonoCreatorToDl {
	var creators []*KemonoCreatorToDl
	for _, creator := range resJson {
		creators = append(creators, &KemonoCreatorToDl{
			CreatorId: creator.Id,
			Service:   creator.Service,
			PageNum:   "", // download all pages,
		})
	}
	return creators
}

func GetFavourites(dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, []error) {
	useHttp3 := httpfuncs.IsHttp3Supported(constants.KEMONO, true)
	params := map[string]string{
		"type": "artist",
	}
	reqArgs := &httpfuncs.RequestArgs{
		Url:         constants.KEMONO_API_URL + "/account/favorites",
		Method:      "GET",
		Cookies:     dlOptions.Base.SessionCookies,
		Params:      params,
		Headers:     getKemonoPartyHeaders(),
		UserAgent:   dlOptions.Base.Configs.UserAgent,
		Http2:       !useHttp3,
		Http3:       useHttp3,
		CheckStatus: true,
		Context:     dlOptions.GetContext(),
	}
	res, err := httpfuncs.CallRequest(reqArgs)
	if err != nil {
		return nil, nil, []error{err}
	}

	var creatorResJson KemonoFavCreatorJson
	if err := httpfuncs.LoadJsonFromResponse(res.Resp, &creatorResJson); err != nil {
		return nil, nil, []error{err}
	}
	artistToDl := processFavCreator(creatorResJson)

	reqArgs.Params = map[string]string{
		"type": "post",
	}
	res, err = httpfuncs.CallRequest(reqArgs)
	if err != nil {
		return nil, nil, []error{err}
	}

	var postResJson KemonoJson
	if err := httpfuncs.LoadJsonFromResponse(res.Resp, &postResJson); err != nil {
		return nil, nil, []error{err}
	}
	urlsToDownload, gdriveLinks := processMultipleJson(postResJson, dlOptions)

	creatorsPost, creatorsGdrive, errSlice := GetMultipleCreators(artistToDl, dlOptions)
	urlsToDownload = append(urlsToDownload, creatorsPost...)
	gdriveLinks = append(gdriveLinks, creatorsGdrive...)

	return urlsToDownload, gdriveLinks, errSlice
}
