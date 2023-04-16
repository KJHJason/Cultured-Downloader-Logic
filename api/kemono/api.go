package kemono

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/spinner"
	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/PuerkitoBio/goquery"
)

type kemonoChanRes struct {
	urlsToDownload []*httpfuncs.ToDownload
	gdriveLinks    []*httpfuncs.ToDownload
	err            error
}

func getKemonoPartyHeaders() map[string]string {
	return map[string]string{
		"Host": constants.KEMONO_URL,
	}
}

func parseCreatorHtml(res *http.Response, url string) (string, error) {
	// parse the response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	res.Body.Close()
	if err != nil {
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
			"kemono error %d, failed to get creator name from Kemono Party at %s\nplease report this issue!",
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

	useHttp3 := httpfuncs.IsHttp3Supported(constants.KEMONO, true)
	url := fmt.Sprintf(
		"%s/%s/user/%s",
		constants.KEMONO_URL,
		service,
		userId,
	)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Url:         url,
			Method:      "GET",
			Headers:     getKemonoPartyHeaders(),
			UserAgent:   dlOptions.Configs.UserAgent,
			Cookies:     dlOptions.SessionCookies,
			Http2:       !useHttp3,
			Http3:       useHttp3,
			CheckStatus: true,
		},
	)
	if err != nil {
		return "", err
	}

	creatorName, err := parseCreatorHtml(res, url)
	if err != nil {
		return "", err
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
				constants.KEMONO_API_URL,
				post.Service,
				post.CreatorId,
				post.PostId,
			),
			Method:      "GET",
			Headers:     getKemonoPartyHeaders(),
			UserAgent:   dlOptions.Configs.UserAgent,
			Cookies:     dlOptions.SessionCookies,
			Http2:       !useHttp3,
			Http3:       useHttp3,
			CheckStatus: true,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	var resJson KemonoJson
	if err := httpfuncs.LoadJsonFromResponse(res, &resJson); err != nil {
		return nil, nil, err
	}

	postsToDl, gdriveLinks := processMultipleJson(resJson, downloadPath, dlOptions)
	return postsToDl, gdriveLinks, nil
}

func getMultiplePosts(posts []*KemonoPostToDl, downloadPath string, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload) {
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

	baseMsg := "Getting post details from Kemono Party [%d/" + fmt.Sprintf("%d]...", postLen)
	progress := spinner.New(
		spinner.REQ_SPINNER,
		"fgHiYellow",
		fmt.Sprintf(
			baseMsg,
			0,
		),
		fmt.Sprintf(
			"Finished getting %d post details from Kemono Party!",
			postLen,
		),
		fmt.Sprintf(
			"Something went wrong while getting %d post details from Kemono Party.\nPlease refer to the logs for more details.",
			postLen,
		),
		postLen,
	)
	progress.Start()
	for _, post := range posts {
		wg.Add(1)
		go func(post *KemonoPostToDl) {
			defer func() {
				progress.MsgIncrement(baseMsg)
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

	hasError := false
	var urlsToDownload, gdriveLinks []*httpfuncs.ToDownload
	for res := range resChan {
		if res.err != nil {
			if !hasError {
				hasError = true
			}
			logger.LogError(res.err, false, logger.ERROR)
			continue
		}
		urlsToDownload = append(urlsToDownload, res.urlsToDownload...)
		gdriveLinks = append(gdriveLinks, res.gdriveLinks...)
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
					constants.KEMONO_API_URL,
					creator.Service,
					creator.CreatorId,
				),
				Method:      "GET",
				UserAgent:   dlOptions.Configs.UserAgent,
				Headers:     getKemonoPartyHeaders(),
				Cookies:     dlOptions.SessionCookies,
				Params:      params,
				Http2:       !useHttp3,
				Http3:       useHttp3,
				CheckStatus: true,
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

		posts, gdriveLinks := processMultipleJson(resJson, downloadPath, dlOptions)
		postsToDl = append(postsToDl, posts...)
		gdriveLinksToDl = append(gdriveLinksToDl, gdriveLinks...)

		if (hasMax && curOffset >= maxOffset) {
			break
		}
		curOffset += 25
	}
	return postsToDl, gdriveLinksToDl, nil
}

func getMultipleCreators(creators []*KemonoCreatorToDl, downloadPath string, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload) {
	var errSlice []error
	var urlsToDownload, gdriveLinks []*httpfuncs.ToDownload
	creatorLen := len(creators)
	baseMsg := "Getting creator's posts from Kemono Party [%d/" + fmt.Sprintf("%d]...", creatorLen)
	progress := spinner.New(
		spinner.REQ_SPINNER,
		"fgHiYellow",
		fmt.Sprintf(
			baseMsg,
			0,
		),
		fmt.Sprintf(
			"Finished getting %d creator's posts from Kemono Party!",
			creatorLen,
		),
		fmt.Sprintf(
			"Something went wrong while getting %d creator's posts from Kemono Party.\nPlease refer to the logs for more details.",
			creatorLen,
		),
		creatorLen,
	)
	progress.Start()
	for _, creator := range creators {
		postsToDl, gdriveLinksToDl, err := getCreatorPosts(creator, downloadPath, dlOptions)
		if err != nil {
			errSlice = append(errSlice, err)
			progress.MsgIncrement(baseMsg)
			continue
		}
		urlsToDownload = append(urlsToDownload, postsToDl...)
		gdriveLinks = append(gdriveLinks, gdriveLinksToDl...)
		progress.MsgIncrement(baseMsg)
	}

	hasError := false
	if len(errSlice) > 0 {
		hasError = true
		logger.LogErrors(false, logger.ERROR, errSlice...)
	}
	progress.Stop(hasError)
	return urlsToDownload, gdriveLinks
}

func processFavCreator(resJson KemonoFavCreatorJson) []*KemonoCreatorToDl {
	var creators []*KemonoCreatorToDl
	for _, creator := range resJson {
		creators = append(creators, &KemonoCreatorToDl{
			CreatorId: creator.Id,
			Service:   creator.Service,
			PageNum:   "", // download all pages
		})
	}
	return creators
}

func getFavourites(downloadPath string, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	useHttp3 := httpfuncs.IsHttp3Supported(constants.KEMONO, true)
	params := map[string]string{
		"type": "artist",
	}
	reqArgs := &httpfuncs.RequestArgs{
		Url:         fmt.Sprintf("%s/v1/account/favorites", constants.KEMONO_API_URL),
		Method:      "GET",
		Cookies:     dlOptions.SessionCookies,
		Params:      params,
		Headers:     getKemonoPartyHeaders(),
		UserAgent:   dlOptions.Configs.UserAgent,
		Http2:       !useHttp3,
		Http3:       useHttp3,
		CheckStatus: true,
	}
	res, err := httpfuncs.CallRequest(reqArgs)
	if err != nil {
		return nil, nil, err
	}

	var creatorResJson KemonoFavCreatorJson
	if err := httpfuncs.LoadJsonFromResponse(res, &creatorResJson); err != nil {
		return nil, nil, err
	}
	artistToDl := processFavCreator(creatorResJson)

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
	urlsToDownload, gdriveLinks := processMultipleJson(postResJson, downloadPath, dlOptions)

	creatorsPost, creatorsGdrive := getMultipleCreators(artistToDl, downloadPath, dlOptions)
	urlsToDownload = append(urlsToDownload, creatorsPost...)
	gdriveLinks = append(gdriveLinks, creatorsGdrive...)

	return urlsToDownload, gdriveLinks, nil
}
