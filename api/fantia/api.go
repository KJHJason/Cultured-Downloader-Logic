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
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
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
	progress := dlOptions.PostProgBar
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
		errCode := constants.CONNECTION_ERROR
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
				"%s, more info => %v",
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

const fantiaPostUrl = constants.FANTIA_URL + "/api/v1/posts/"

func DlFantiaPost(count, maxCount int, postId string, dlOptions *FantiaDlOptions) ([]*httpfuncs.ToDownload, error) {
	msgSuffix := fmt.Sprintf(
		"[%d/%d]",
		count,
		maxCount,
	)

	res, err := getFantiaPostDetails(
		&fantiaPostArgs{
			msgSuffix: msgSuffix,
			postId:    postId,
			url:       fantiaPostUrl,
		},
		dlOptions,
	)
	if err != nil {
		return nil, err
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
	if err == errRecaptcha {
		err = api.SolveCaptcha(dlOptions, constants.FANTIA)
		if err != nil {
			if err := api.HandleCaptchaErr(err, dlOptions, constants.FANTIA); err != nil {
				// if the user doesn't want to solve the captcha manually, stop the download
				dlOptions.Cancel()
				return nil, err
			}
		}

		return DlFantiaPost(count, maxCount, postId, dlOptions)
	} else if err != nil {
		return nil, err
	}

	// Download the urls
	err = httpfuncs.DownloadUrls(
		urlsToDownload,
		&httpfuncs.DlOptions{
			Context:		dlOptions.GetContext(),
			MaxConcurrency: constants.MAX_CONCURRENT_DOWNLOADS,
			Headers:        nil,
			Cookies:        dlOptions.SessionCookies,
			UseHttp3:       false,
		},
		dlOptions.Configs,
	)
	fmt.Println()

	if err == context.Canceled {
		return nil, err
	}
	return postGdriveUrls, nil
}

// Query Fantia's API based on the slice of post IDs and get a map of urls to download from.
//
// Note that only the downloading of the URL(s) is/are executed concurrently
// to reduce the chance of the signed AWS S3 URL(s) from expiring before the download is
// executed or completed due to a download queue to avoid resource exhaustion of the user's system.
func (f *FantiaDl) DlFantiaPosts(dlOptions *FantiaDlOptions) []*httpfuncs.ToDownload {
	var errSlice []error
	var gdriveLinks []*httpfuncs.ToDownload
	postIdsLen := len(f.PostIds)
	for i, postId := range f.PostIds {
		postGdriveLinks, err := DlFantiaPost(i+1, postIdsLen, postId, dlOptions)

		if err != nil {
			if err == context.Canceled {
				return nil
			}
			errSlice = append(errSlice, err)
			continue
		}
		if len(postGdriveLinks) > 0 {
			gdriveLinks = append(gdriveLinks, postGdriveLinks...)
		}
	}

	if len(errSlice) > 0 {
		logger.LogErrors(false, logger.ERROR, errSlice...)
	}
	return gdriveLinks
}

// Parse the HTML response from the creator's page to get the post IDs.
func parseCreatorHtml(res *http.Response, creatorId string) ([]string, error) {
	// parse the response
	doc, err := goquery.NewDocumentFromReader(res.Body)
	res.Body.Close()
	if err != nil {
		err = fmt.Errorf(
			"fantia error %d, failed to parse response body when getting posts for Fantia Fanclub %s, more info => %v",
			constants.HTML_ERROR,
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
			constants.HTML_ERROR,
			creatorId,
		)
	}
	return postIds, nil
}

// Get all the creator's posts by using goquery to parse the HTML response to get the post IDs
func getCreatorPosts(creatorId, pageNum string, dlOptions *FantiaDlOptions) ([]string, error) {
	var postIds []string
	minPage, maxPage, hasMax, err := api.GetMinMaxFromStr(pageNum)
	if err != nil {
		return nil, err
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.FANTIA, false)
	curPage := minPage
	for {
		url := fmt.Sprintf("%s/fanclubs/%s/posts", constants.FANTIA_URL, creatorId)
		params := map[string]string{
			"page":   strconv.Itoa(curPage),
			"q[s]":   "newer",
			"q[tag]": "",
		}

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
			if err != context.Canceled {
				err = fmt.Errorf(
					"fantia error %d: failed to get creator's pages for %s, more info => %v",
					constants.CONNECTION_ERROR,
					url,
					err,
				)
			}
			return nil, err
		}

		creatorPostIds, err := parseCreatorHtml(res, creatorId)
		if err != nil {
			return nil, err
		}
		postIds = append(postIds, creatorPostIds...)

		// if there are no more posts, break
		if len(creatorPostIds) == 0 || (hasMax && curPage >= maxPage) {
			break
		}
		curPage++
	}
	return postIds, nil
}

// Retrieves all the posts based on the slice of creator IDs and updates its PostIds slice
func (f *FantiaDl) GetCreatorsPosts(dlOptions *FantiaDlOptions) {
	creatorIdsLen := len(f.FanclubIds)
	if creatorIdsLen != len(f.FanclubPageNums) {
		panic(
			fmt.Errorf(
				"fantia error %d: creator IDs and page numbers slices are not the same length",
				constants.DEV_ERROR,
			),
		)
	}

	var wg sync.WaitGroup
	maxConcurrency := constants.MAX_API_CALLS
	if creatorIdsLen < maxConcurrency {
		maxConcurrency = creatorIdsLen
	}
	queue := make(chan struct{}, maxConcurrency)
	resChan := make(chan []string, creatorIdsLen)
	errChan := make(chan error, creatorIdsLen)

	baseMsg := "Getting post ID(s) from Fanclubs(s) on Fantia [%d/" + fmt.Sprintf("%d]...", creatorIdsLen)
	progress := dlOptions.GetFanclubPostsProgBar
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting post ID(s) from %d Fanclubs(s) on Fantia!",
			creatorIdsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting post IDs from %d Fanclubs(s) on Fantia.\nPlease refer to the logs for more details.",
			creatorIdsLen,
		),
	)
	progress.UpdateMax(creatorIdsLen)
	progress.Start()
	for idx, creatorId := range f.FanclubIds {
		wg.Add(1)
		go func(creatorId string, pageNumIdx int) {
			defer func() {
				wg.Done()
				<-queue
			}()

			queue <- struct{}{}
			postIds, err := getCreatorPosts(
				creatorId,
				f.FanclubPageNums[pageNumIdx],
				dlOptions,
			)
			if err != nil {
				errChan <- err
			} else {
				resChan <- postIds
			}

			progress.Increment()
		}(creatorId, idx)
	}
	wg.Wait()
	close(queue)
	close(resChan)
	close(errChan)

	hasErr, hasCancelled := false, false
	if len(errChan) > 0 {
		hasErr = true
		if errCtxCancelled := logger.LogChanErrors(false, logger.ERROR, errChan); !hasCancelled && errCtxCancelled {
			hasCancelled = true
		} 
	}
	if hasCancelled {
		progress.StopInterrupt("Stopped getting post ID(s) from Fanclub(s) on Fantia...")
		return
	}
	progress.Stop(hasErr)

	for postIdsRes := range resChan {
		f.PostIds = append(f.PostIds, postIdsRes...)
	}
	f.PostIds = api.RemoveSliceDuplicates(f.PostIds)
}
