package kemono

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/PuerkitoBio/goquery"
)

type kemonoChanRes struct {
	urlsToDownload []*httpfuncs.ToDownload
	gdriveLinks    []*httpfuncs.ToDownload
	err            error
}

func getKemonoPartyHeaders(tld string) map[string]string {
	return map[string]string{
		"Host": getKemonoUrl(tld),
	}
}

func getKemonoUrl(tld string) string {
	if tld == constants.KEMONO_TLD || tld == "" {
		// if tld is empty, use the default url as the fallback
		return constants.KEMONO_URL
	}
	return constants.BACKUP_KEMONO_URL
}

func getKemonoApiUrl(tld string) string {
	if tld == constants.KEMONO_TLD || tld == "" {
		// if tld is empty, use the default url as the fallback
		return constants.KEMONO_API_URL
	}
	return constants.BACKUP_KEMONO_API_URL
}

func getKemonoUrlFromConditions(isBackup, isApi bool) string {
	if isApi {
		if isBackup {
			return constants.BACKUP_KEMONO_API_URL
		}
		return constants.KEMONO_API_URL
	}

	if isBackup {
		return constants.BACKUP_KEMONO_URL
	}
	return constants.KEMONO_URL
}

var errSessionCookieNotFound = errors.New("could not find session cookie")
func getKemonoUrlFromCookie(cookie []*http.Cookie, isApi bool) (string, string, error) {
	for _, c := range cookie {
		fmt.Println(c.Name)
		if c.Name == constants.KEMONO_SESSION_COOKIE_NAME {
			if c.Domain == constants.KEMONO_COOKIE_DOMAIN {
				return getKemonoUrlFromConditions(false, isApi), constants.KEMONO_TLD, nil
			} else {
				return getKemonoUrlFromConditions(true, isApi), constants.KEMONO_BACKUP_TLD, nil
			}
		}
	}
	return "", "", errSessionCookieNotFound
}

func parseCreatorHtml(res *http.Response, url string) (string, error) {
	// parse the response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	res.Body.Close()
	if err != nil {
		if err == context.Canceled {
			return "", err
		}

		err = fmt.Errorf(
			"kemono error %d, failed to parse response body when getting creator name from Kemono Party at %s\nmore info => %v",
			constants.HTML_ERROR,
			url,
			err,
		)
		return "", err
	}

	// <span itemprop="name">creator-name</span> => creator-name
	creatorName := doc.Find("span[itemprop=name]").Text()
	if creatorName == "" {
		return "", fmt.Errorf(
			"kemono error %d, failed to get creator name from Kemono Party at %s\nplease report this issue",
			constants.HTML_ERROR,
			url,
		)
	}

	return creatorName, nil
}

var creatorNameCacheLock sync.Mutex
var creatorNameCache = make(map[string]string)
func getCreatorName(service, userId string, dlOptions *KemonoDlOptions) (string, error) {
	creatorNameCacheLock.Lock()
	defer creatorNameCacheLock.Unlock()
	cacheKey := fmt.Sprintf("%s/%s", service, userId)
	if name, ok := creatorNameCache[cacheKey]; ok {
		return name, nil
	}

	kemonoUrl, tld, err := getKemonoUrlFromCookie(dlOptions.SessionCookies, false)
	if err != nil {
		return userId, err
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.KEMONO, true)
	url := fmt.Sprintf(
		"%s/%s/user/%s",
		kemonoUrl,
		service,
		userId,
	)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Url:         url,
			Method:      "GET",
			Headers:     getKemonoPartyHeaders(tld),
			UserAgent:   dlOptions.Configs.UserAgent,
			Cookies:     dlOptions.SessionCookies,
			Http2:       !useHttp3,
			Http3:       useHttp3,
			CheckStatus: true,
			Context:     dlOptions.Ctx,
		},
	)
	if err != nil {
		return userId, err
	}

	creatorName, err := parseCreatorHtml(res, url)
	if err != nil {
		return userId, err
	}

	creatorNameCache[cacheKey] = creatorName
	return creatorName, nil
}

func getPostDetails(post *KemonoPostToDl, downloadPath string, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	useHttp3 := httpfuncs.IsHttp3Supported(constants.KEMONO, true)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Url: fmt.Sprintf(
				"%s/%s/user/%s/post/%s",
				getKemonoApiUrl(post.Tld),
				post.Service,
				post.CreatorId,
				post.PostId,
			),
			Method:      "GET",
			Headers:     getKemonoPartyHeaders(post.Tld),
			UserAgent:   dlOptions.Configs.UserAgent,
			Cookies:     dlOptions.SessionCookies,
			Http2:       !useHttp3,
			Http3:       useHttp3,
			CheckStatus: true,
			Context:     dlOptions.Ctx,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	var resJson KemonoJson
	if err := httpfuncs.LoadJsonFromResponse(res, &resJson); err != nil {
		return nil, nil, err
	}

	postsToDl, gdriveLinks := processMultipleJson(resJson, post.Tld, downloadPath, dlOptions)
	return postsToDl, gdriveLinks, nil
}

func GetMultiplePosts(posts []*KemonoPostToDl, downloadPath string, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload) {
	var maxConcurrency int
	postLen := len(posts)
	if postLen > API_MAX_CONCURRENT {
		maxConcurrency = API_MAX_CONCURRENT
	} else {
		maxConcurrency = postLen
	}
	wg := sync.WaitGroup{}
	queue := make(chan struct{}, maxConcurrency)
	resChan := make(chan *kemonoChanRes, postLen)

	baseMsg := "Getting post details from Kemono [%d/" + fmt.Sprintf("%d]...", postLen)
	progress := dlOptions.PostProgBar
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
	progress.UpdateMax(postLen)
	progress.Start()
	for _, post := range posts {
		wg.Add(1)
		go func(post *KemonoPostToDl) {
			defer func() {
				progress.Increment()
				wg.Done()
				<-queue
			}()

			queue <- struct{}{}
			toDownload, foundGdriveLinks, err := getPostDetails(post, downloadPath, dlOptions)
			if err != nil {
				resChan <- &kemonoChanRes{
					err: err,
				}
				return
			}
			resChan <- &kemonoChanRes{
				urlsToDownload: toDownload,
				gdriveLinks:    foundGdriveLinks,
			}
		}(post)
	}
	wg.Wait()
	close(queue)
	close(resChan)

	hasError, hasCancelled := false, false
	var urlsToDownload, gdriveLinks []*httpfuncs.ToDownload
	for res := range resChan {
		if res.err != nil {
			if res.err == context.Canceled {
				hasCancelled = true
				continue
			}
			if !hasError {
				hasError = true
			}
			logger.LogError(res.err, false, logger.ERROR)
			continue
		}
		urlsToDownload = append(urlsToDownload, res.urlsToDownload...)
		gdriveLinks = append(gdriveLinks, res.gdriveLinks...)
	}

	if hasCancelled {
		progress.StopInterrupt("Stopped getting post details from Kemono...")
		return nil, nil
	}
	progress.Stop(hasError)
	return urlsToDownload, gdriveLinks
}

func getCreatorPosts(creator *KemonoCreatorToDl, downloadPath string, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	useHttp3 := httpfuncs.IsHttp3Supported(constants.KEMONO, true)
	minPage, maxPage, hasMax, err := api.GetMinMaxFromStr(creator.PageNum)
	if err != nil {
		return nil, nil, err
	}
	minOffset, maxOffset := api.ConvertPageNumToOffset(minPage, maxPage, constants.KEMONO_PER_PAGE)

	var postsToDl, gdriveLinksToDl []*httpfuncs.ToDownload
	params := make(map[string]string)
	curOffset := minOffset
	for {
		params["o"] = strconv.Itoa(curOffset)
		res, err := httpfuncs.CallRequest(
			&httpfuncs.RequestArgs{
				Url: fmt.Sprintf(
					"%s/%s/user/%s",
					getKemonoApiUrl(creator.Tld),
					creator.Service,
					creator.CreatorId,
				),
				Method:      "GET",
				UserAgent:   dlOptions.Configs.UserAgent,
				Headers:     getKemonoPartyHeaders(creator.Tld),
				Cookies:     dlOptions.SessionCookies,
				Params:      params,
				Http2:       !useHttp3,
				Http3:       useHttp3,
				CheckStatus: true,
				Context:     dlOptions.Ctx,
			},
		)
		if err != nil {
			return nil, nil, err
		}

		var resJson KemonoJson
		if err := httpfuncs.LoadJsonFromResponse(res, &resJson); err != nil {
			return nil, nil, err
		}

		if len(resJson) == 0 {
			break
		}

		posts, gdriveLinks := processMultipleJson(resJson, creator.Tld, downloadPath, dlOptions)
		postsToDl = append(postsToDl, posts...)
		gdriveLinksToDl = append(gdriveLinksToDl, gdriveLinks...)

		if (hasMax && curOffset >= maxOffset) {
			break
		}
		curOffset += 25
	}
	return postsToDl, gdriveLinksToDl, nil
}

func GetMultipleCreators(creators []*KemonoCreatorToDl, downloadPath string, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload) {
	var errSlice []error
	var urlsToDownload, gdriveLinks []*httpfuncs.ToDownload
	creatorLen := len(creators)
	baseMsg := "Getting creator's posts from Kemono [%d/" + fmt.Sprintf("%d]...", creatorLen)
	progress := dlOptions.GetCreatorPostProgBar
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
	progress.UpdateMax(creatorLen)
	progress.Start()
	hasCancelled := false
	for _, creator := range creators {
		postsToDl, gdriveLinksToDl, err := getCreatorPosts(creator, downloadPath, dlOptions)
		if err != nil {
			errSlice = append(errSlice, err)
			if err == context.Canceled {
				hasCancelled = true
				progress.StopInterrupt("Stopped getting creator's posts from Kemono...")
				break
			}
			progress.Increment()
			continue
		}
		urlsToDownload = append(urlsToDownload, postsToDl...)
		gdriveLinks = append(gdriveLinks, gdriveLinksToDl...)
		progress.Increment()
	}

	hasErr := false
	if len(errSlice) > 0 {
		hasErr = true
		logger.LogErrors(false, logger.ERROR, errSlice...)
	}
	if hasCancelled {
		return nil, nil
	}
	progress.Stop(hasErr)
	return urlsToDownload, gdriveLinks
}

func processFavCreator(resJson KemonoFavCreatorJson, tld string) []*KemonoCreatorToDl {
	var creators []*KemonoCreatorToDl
	for _, creator := range resJson {
		creators = append(creators, &KemonoCreatorToDl{
			CreatorId: creator.Id,
			Service:   creator.Service,
			PageNum:   "", // download all pages
			Tld:       tld,
		})
	}
	return creators
}

func GetFavourites(downloadPath string, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	apiUrl, tld, err := getKemonoUrlFromCookie(dlOptions.SessionCookies, true)
	if err != nil {
		return nil, nil, err
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.KEMONO, true)
	params := map[string]string{
		"type": "artist",
	}
	reqArgs := &httpfuncs.RequestArgs{
		Url:         fmt.Sprintf("%s/v1/account/favorites", apiUrl),
		Method:      "GET",
		Cookies:     dlOptions.SessionCookies,
		Params:      params,
		Headers:     getKemonoPartyHeaders(tld),
		UserAgent:   dlOptions.Configs.UserAgent,
		Http2:       !useHttp3,
		Http3:       useHttp3,
		CheckStatus: true,
		Context:     dlOptions.Ctx,
	}
	res, err := httpfuncs.CallRequest(reqArgs)
	if err != nil {
		return nil, nil, err
	}

	var creatorResJson KemonoFavCreatorJson
	if err := httpfuncs.LoadJsonFromResponse(res, &creatorResJson); err != nil {
		return nil, nil, err
	}
	artistToDl := processFavCreator(creatorResJson, tld)

	reqArgs.Params = map[string]string{
		"type": "post",
	}
	res, err = httpfuncs.CallRequest(reqArgs)
	if err != nil {
		return nil, nil, err
	}

	var postResJson KemonoJson
	if err := httpfuncs.LoadJsonFromResponse(res, &postResJson); err != nil {
		return nil, nil, err
	}
	urlsToDownload, gdriveLinks := processMultipleJson(postResJson, tld, downloadPath, dlOptions)

	creatorsPost, creatorsGdrive := GetMultipleCreators(artistToDl, downloadPath, dlOptions)
	urlsToDownload = append(urlsToDownload, creatorsPost...)
	gdriveLinks = append(gdriveLinks, creatorsGdrive...)

	return urlsToDownload, gdriveLinks, nil
}
