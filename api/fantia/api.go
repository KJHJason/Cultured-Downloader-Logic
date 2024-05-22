package fantia

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/PuerkitoBio/goquery"
)

type fantiaPostArgs struct {
	msgSuffix string
	postId    string
	url       string
}

func getFantiaPostDetails(postArg *fantiaPostArgs, dlOptions *FantiaDlOptions) (*http.Response, error) {
	// Now that we have the post ID, we can query Fantia's API
	// to get the post's contents from the JSON response.
	progress := dlOptions.MainProgBar
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
			Method:    "GET",
			Url:       postApiUrl,
			Cookies:   dlOptions.SessionCookies,
			Headers:   header,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			UserAgent: dlOptions.Configs.UserAgent,
			Context:   dlOptions.GetContext(),
		},
	)
	if err != nil || res.StatusCode != 200 {
		errCode := cdlerrors.CONNECTION_ERROR
		if err == nil {
			errCode = res.StatusCode
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
	if dlOptions.UseCacheDb {
		url := constants.FANTIA_POST_API_URL + postId
		if database.PostCacheExists(url, constants.FANTIA) {
			return false, nil, nil
		}
		cacheKey = database.ParsePostKey(url, constants.FANTIA)
	}

	res, err := getFantiaPostDetails(
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
			res:        res,
			postId:     postId,
			postIdsLen: maxCount,
			msgSuffix:  msgSuffix,
		},
		dlOptions,
	)
	if errors.Is(err, cdlerrors.ErrRecaptcha) {
		err = SolveCaptcha(dlOptions)
		if err != nil {
			// stop the download if the captcha auto-solving fails
			dlOptions.CancelCtx()
			return false, nil, []error{err}
		}

		return DlFantiaPost(count, maxCount, postId, dlOptions)
	} else if err != nil {
		return false, nil, []error{err}
	}

	// Download the urls
	useHttp3 := httpfuncs.IsHttp3Supported(constants.FANTIA, false)
	cancelled, errorSlice := httpfuncs.DownloadUrls(
		urlsToDownload,
		&httpfuncs.DlOptions{
			Context:        dlOptions.GetContext(),
			MaxConcurrency: constants.FANTIA_MAX_CONCURRENCY,
			Headers:        nil,
			Cookies:        dlOptions.SessionCookies,
			UseHttp3:       useHttp3,
			HeadReqTimeout: constants.DEFAULT_HEAD_REQ_TIMEOUT,
			SupportRange:   constants.FANTIA_RANGE_SUPPORTED,
			ProgressBarInfo: &progress.ProgressBarInfo{
				MainProgressBar:      dlOptions.MainProgBar,
				DownloadProgressBars: dlOptions.DownloadProgressBars,
			},
		},
		dlOptions.Configs,
	)

	if cancelled {
		dlOptions.CancelCtx()
		return true, nil, errorSlice
	}
	if dlOptions.UseCacheDb {
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

func getFantiaProductPaidContent(purchaseRelativeUrl, productId string, dlOptions *FantiaDlOptions) (*http.Response, error) {
	useHttp3 := httpfuncs.IsHttp3Supported(constants.FANTIA, false)
	purchaseUrl := constants.FANTIA_URL + purchaseRelativeUrl // https://fantia.jp/mypage/users/purchases/123456
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Method:    "GET",
			Url:       purchaseUrl,
			Cookies:   dlOptions.SessionCookies,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			UserAgent: dlOptions.Configs.UserAgent,
			Context:   dlOptions.GetContext(),
		},
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, fmt.Errorf(
			"fantia error %d: failed to get purchase details at %s for product %s, more info => %w",
			cdlerrors.CONNECTION_ERROR,
			purchaseUrl,
			productId,
			err,
		)
	}
	return res, nil
}

func getProduct(productId string, dlOptions *FantiaDlOptions) ([]*httpfuncs.ToDownload, error) {
	var cacheKey string
	productUrl := fmt.Sprintf("%s/products/%s", constants.FANTIA_URL, productId)
	if dlOptions.UseCacheDb {
		if database.PostCacheExists(productUrl, constants.FANTIA) {
			return nil, nil
		}
		cacheKey = database.ParsePostKey(productUrl, constants.FANTIA)
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.FANTIA, false)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Method:    "GET",
			Url:       productUrl,
			Cookies:   dlOptions.SessionCookies,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			UserAgent: dlOptions.Configs.UserAgent,
			Context:   dlOptions.GetContext(),
		},
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, fmt.Errorf(
			"fantia error %d: failed to get product details for %s, more info => %w",
			cdlerrors.CONNECTION_ERROR,
			productUrl,
			err,
		)
	}
	return processProductPage(cacheKey, productId, dlOptions, res)
}

func (f *FantiaDl) GetProducts(dlOptions *FantiaDlOptions) ([]*httpfuncs.ToDownload, []error) {
	productFanclubIdsLen := len(f.ProductFanclubIds)
	if productFanclubIdsLen == 0 {
		return nil, nil
	}

	if productFanclubIdsLen != len(f.ProductFanclubPageNums) {
		return nil, []error{
			fmt.Errorf(
				"fantia error %d: fanclubs IDs and page numbers slices are not the same length",
				cdlerrors.DEV_ERROR,
			),
		}
	}

	// Now that we have the post ID, we can query Fantia's API
	// to get the post's contents from the JSON response.
	progress := dlOptions.MainProgBar
	progress.SetToSpinner()
	progress.UpdateBaseMsg(
		"Getting product contents from Fanclubs(s) on Fantia [%d/" + fmt.Sprintf("%d]...", productFanclubIdsLen),
	)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting product contents from %d Fanclubs(s) on Fantia!",
			productFanclubIdsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting product contents from %d Fanclubs(s) on Fantia.\nPlease refer to the logs for more details.",
			productFanclubIdsLen,
		),
	)
	progress.Start()
	defer progress.SnapshotTask()

	var wg sync.WaitGroup
	maxConcurrency := constants.FANTIA_PRODUCT_MAX_CONCURRENCY
	if productFanclubIdsLen < maxConcurrency {
		maxConcurrency = productFanclubIdsLen
	}
	queue := make(chan struct{}, maxConcurrency)
	resChan := make(chan []*httpfuncs.ToDownload, productFanclubIdsLen)
	errChan := make(chan error, productFanclubIdsLen)

	for idx, productId := range f.ProductFanclubIds {
		wg.Add(1)
		go func(productId string, pageNumIdx int) {
			defer func() {
				wg.Done()
				<-queue
			}()

			queue <- struct{}{}
			productToDownload, err := getProduct(productId, dlOptions)
			if err != nil {
				errChan <- err
			} else {
				resChan <- productToDownload
			}

			progress.Increment()
		}(productId, idx)
	}
	wg.Wait()
	close(queue)
	close(resChan)
	close(errChan)

	var errorSlice []error
	hasErr := len(errChan) > 0
	if hasErr {
		var hasCancelled bool
		if hasCancelled, errorSlice = logger.LogChanErrors(logger.ERROR, errChan); hasCancelled {
			dlOptions.CancelCtx()
			progress.StopInterrupt("Stopped getting product(s) from Fantia...")
			return nil, nil
		}
	}
	progress.Stop(hasErr)

	var productUrls []*httpfuncs.ToDownload
	for productToDownload := range resChan {
		productUrls = append(productUrls, productToDownload...)
	}
	return productUrls, errorSlice
}

// Parse the HTML response from the creator's page to get the post or product IDs.
func parseCreatorHtml(res *http.Response, creatorId, contentType string) ([]string, error) {
	// parse the response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	res.Body.Close()
	if err != nil {
		err = fmt.Errorf(
			"fantia error %d, failed to parse response body when getting %s for Fantia Fanclub %s, more info => %w",
			cdlerrors.HTML_ERROR,
			contentType, // posts or products
			creatorId,
			err,
		)
		return nil, err
	}

	// get the post ids similar to using the xpath of //a[@class='link-block']
	hasHtmlErr := false
	var postIds []string
	doc.Find("a.link-block").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			postIds = append(postIds, httpfuncs.GetLastPartOfUrl(href))
		} else if !hasHtmlErr {
			hasHtmlErr = true
		}
	})

	if hasHtmlErr {
		return nil, fmt.Errorf(
			"fantia error %d, failed to get href attribute for Fantia Fanclub %s, please report this issue",
			cdlerrors.HTML_ERROR,
			creatorId,
		)
	}
	return postIds, nil
}

const (
	POSTS    = "posts"
	PRODUCTS = "products"
)

// Get all the creator's posts by using goquery to parse the HTML response to get the post IDs
func getCreatorContent(creatorId, pageNum string, dlOptions *FantiaDlOptions, contentType string) ([]string, error) {
	var contentIds []string
	minPage, maxPage, hasMax, err := api.GetMinMaxFromStr(pageNum)
	if err != nil {
		return nil, err
	}

	var url string
	if contentType == PRODUCTS {
		url = fmt.Sprintf("%s/fanclubs/%s/products", constants.FANTIA_URL, creatorId)
	} else {
		url = fmt.Sprintf("%s/fanclubs/%s/posts", constants.FANTIA_URL, creatorId)
	}
	useHttp3 := httpfuncs.IsHttp3Supported(constants.FANTIA, false)
	curPage := minPage
	for {
		var params map[string]string
		if contentType == PRODUCTS {
			params = map[string]string{
				"q[name_cont]": "",      // query string
				"q[s]":         "newer", // sort by newest
				"q[tag]":       "",      // #tag
			}
		} else {
			params = map[string]string{
				"q[s]":   "newer",
				"q[tag]": "",
			}
		}
		params["page"] = strconv.Itoa(curPage)

		// note that even if the max page is more than
		// the actual number of pages, the response will still be 200 OK.
		res, err := httpfuncs.CallRequest(
			&httpfuncs.RequestArgs{
				Method:      "GET",
				Url:         url,
				Cookies:     dlOptions.SessionCookies,
				Params:      params,
				Http2:       !useHttp3,
				Http3:       useHttp3,
				CheckStatus: true,
				UserAgent:   dlOptions.Configs.UserAgent,
				Context:     dlOptions.GetContext(),
			},
		)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				err = fmt.Errorf(
					"fantia error %d: failed to get fanclub's %s pages for %s, more info => %w",
					cdlerrors.CONNECTION_ERROR,
					contentType,
					url,
					err,
				)
			}
			return nil, err
		}

		creatorContentIds, err := parseCreatorHtml(res, creatorId, contentType)
		if err != nil {
			return nil, err
		}
		contentIds = append(contentIds, creatorContentIds...)

		// if there are no more posts, break
		if len(creatorContentIds) == 0 || (hasMax && curPage >= maxPage) {
			break
		}
		curPage++
	}
	return contentIds, nil
}

// Retrieves all the posts based on the slice of creator IDs and updates its PostIds slice
func (f *FantiaDl) GetCreatorsContents(creatorIds []string, pageNums []string, contentType string, dlOptions *FantiaDlOptions) []error {
	if contentType != POSTS && contentType != PRODUCTS {
		return []error{
			fmt.Errorf(
				"fantia error %d: invalid content type %s, must be either %q or %q",
				cdlerrors.DEV_ERROR,
				contentType,
				POSTS,
				PRODUCTS,
			),
		}
	}

	creatorIdsLen := len(creatorIds)
	if creatorIdsLen == 0 {
		return nil
	}

	if creatorIdsLen != len(pageNums) {
		return []error{
			fmt.Errorf(
				"fantia error %d: fanclubs IDs and page numbers slices are not the same length",
				cdlerrors.DEV_ERROR,
			),
		}
	}

	var wg sync.WaitGroup
	maxConcurrency := constants.FANTIA_MAX_CONCURRENCY
	if creatorIdsLen < maxConcurrency {
		maxConcurrency = creatorIdsLen
	}
	queue := make(chan struct{}, maxConcurrency)
	resChan := make(chan []string, creatorIdsLen)
	errChan := make(chan error, creatorIdsLen)

	progress := dlOptions.MainProgBar
	if creatorIdsLen > 1 {
		baseMsg := "Getting " + contentType + " ID(s) from Fanclubs(s) on Fantia [%d/" + fmt.Sprintf("%d]...", creatorIdsLen)
		progress.SetToProgressBar()
		progress.UpdateBaseMsg(baseMsg)
		progress.UpdateSuccessMsg(
			fmt.Sprintf(
				"Finished getting "+contentType+" ID(s) from %d Fanclubs(s) on Fantia!",
				creatorIdsLen,
			),
		)
		progress.UpdateErrorMsg(
			fmt.Sprintf(
				"Something went wrong while getting "+contentType+" IDs from %d Fanclubs(s) on Fantia.\nPlease refer to the logs for more details.",
				creatorIdsLen,
			),
		)
		progress.UpdateMax(creatorIdsLen)
	} else {
		progress.SetToSpinner()
		fanclubId := f.FanclubIds[0]
		progress.UpdateBaseMsg("Getting " + contentType + " ID(s) from Fanclub, " + fanclubId + ", on Fantia...")
		progress.UpdateSuccessMsg("Finished getting " + contentType + " ID(s) from Fanclub, " + fanclubId + ", on Fantia!")
		progress.UpdateErrorMsg("Something went wrong while getting " + contentType + " ID(s) from Fanclub, " + fanclubId + ", on Fantia.\nPlease refer to the logs for more details.")
	}

	progress.Start()
	for idx, creatorId := range f.FanclubIds {
		wg.Add(1)
		go func(creatorId string, pageNumIdx int) {
			defer func() {
				wg.Done()
				<-queue
			}()

			queue <- struct{}{}
			contentIds, err := getCreatorContent(
				creatorId,
				f.FanclubPageNums[pageNumIdx],
				dlOptions,
				contentType,
			)
			if err != nil {
				errChan <- err
			} else {
				resChan <- contentIds
			}

			progress.Increment()
		}(creatorId, idx)
	}
	wg.Wait()
	close(queue)
	close(resChan)
	close(errChan)

	var errorSlice []error
	hasErr, hasCancelled := false, false
	if len(errChan) > 0 {
		hasErr = true
		if errCtxCancelled, errSlice := logger.LogChanErrors(logger.ERROR, errChan); !hasCancelled && errCtxCancelled {
			hasCancelled = true
		} else {
			errorSlice = append(errorSlice, errSlice...)
		}
	}
	if hasCancelled {
		dlOptions.CancelCtx()
		progress.StopInterrupt("Stopped getting " + contentType + " ID(s) from Fanclub(s) on Fantia...")
		progress.SnapshotTask()
		return nil
	}
	progress.Stop(hasErr)
	progress.SnapshotTask()

	for contentIdsRes := range resChan {
		if contentType == PRODUCTS {
			f.ProductIds = append(f.ProductIds, contentIdsRes...)
		} else {
			f.PostIds = append(f.PostIds, contentIdsRes...)
		}
	}
	if contentType == PRODUCTS {
		f.ProductIds = api.RemoveSliceDuplicates(f.ProductIds)
	} else {
		f.PostIds = api.RemoveSliceDuplicates(f.PostIds)
	}
	return errorSlice
}

func (f *FantiaDl) GetCreatorsPosts(dlOptions *FantiaDlOptions) []error {
	return f.GetCreatorsContents(f.FanclubIds, f.FanclubPageNums, POSTS, dlOptions)
}

func (f *FantiaDl) GetCreatorsProducts(dlOptions *FantiaDlOptions) []error {
	return f.GetCreatorsContents(f.ProductFanclubIds, f.ProductFanclubPageNums, PRODUCTS, dlOptions)
}
