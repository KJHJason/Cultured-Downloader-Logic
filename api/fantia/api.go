package fantia

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils/threadsafe"
	"github.com/PuerkitoBio/goquery"
)

// Parse the HTML response from the Fanclub's page to get the post or product IDs.
func parseFanclubHtml(resBody *bytes.Reader, fanclubId, contentType string) ([]string, error) {
	// parse the response
	doc, err := goquery.NewDocumentFromReader(resBody)
	if err != nil {
		err = fmt.Errorf(
			"fantia error %d, failed to parse response body when getting %s for Fantia Fanclub %s, more info => %w",
			cdlerrors.HTML_ERROR,
			contentType, // posts or products
			fanclubId,
			err,
		)
		return nil, err
	}

	// get the post ids similar to using the xpath of //a[@class='link-block']
	hasHtmlErr := false
	var contentIds []string
	doc.Find("a.link-block").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			contentIds = append(contentIds, httpfuncs.GetLastPartOfUrl(href))
		} else if !hasHtmlErr {
			hasHtmlErr = true
		}
	})

	if hasHtmlErr {
		return nil, fmt.Errorf(
			"fantia error %d, failed to get href attribute for Fantia Fanclub %s, please report this issue",
			cdlerrors.HTML_ERROR,
			fanclubId,
		)
	}
	return contentIds, nil
}

const (
	POSTS    = "posts"
	PRODUCTS = "products"
)

// Get all the Fanclub's posts by using goquery to parse the HTML response to get the post IDs
func getFanclubContent(fanclubId, pageNum string, dlOptions *FantiaDlOptions, contentType string) ([]string, error) {
	var contentIds []string
	minPage, maxPage, hasMax, err := utils.GetMinMaxFromStr(pageNum)
	if err != nil {
		return nil, err
	}

	var url string
	if contentType == PRODUCTS {
		url = fmt.Sprintf("%s/fanclubs/%s/products", constants.FANTIA_URL, fanclubId)
	} else {
		url = fmt.Sprintf("%s/fanclubs/%s/posts", constants.FANTIA_URL, fanclubId)
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
				Cookies:     dlOptions.Base.SessionCookies,
				Params:      params,
				Http2:       !useHttp3,
				Http3:       useHttp3,
				CheckStatus: true,
				UserAgent:   dlOptions.Base.Configs.UserAgent,
				Context:     dlOptions.GetContext(),
				CaptchaHandler: httpfuncs.CaptchaHandler{
					Check:                CaptchaChecker,
					Handler:              newCaptchaHandler(dlOptions),
					InjectCaptchaCookies: nil,
				},
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

		resBody, err := res.GetBodyReader()
		if err != nil {
			return nil, err
		}

		fanclubContentIds, err := parseFanclubHtml(resBody, fanclubId, contentType)
		if err != nil {
			return nil, err
		}
		contentIds = append(contentIds, fanclubContentIds...)

		// if there are no more posts, break
		if len(fanclubContentIds) == 0 || (hasMax && curPage >= maxPage) {
			break
		}
		curPage++
	}
	return contentIds, nil
}

// Retrieves all the posts based on the slice of Fanclub IDs and updates its PostIds slice
func (f *FantiaDl) GetFanclubsContents(fanclubIds []string, pageNums []string, contentType string, dlOptions *FantiaDlOptions) []error {
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

	fanclubIdsLen := len(fanclubIds)
	if fanclubIdsLen == 0 {
		return nil
	}

	if fanclubIdsLen != len(pageNums) {
		return []error{
			fmt.Errorf(
				"fantia error %d: fanclubs IDs and page numbers slices are not the same length",
				cdlerrors.DEV_ERROR,
			),
		}
	}

	var wg sync.WaitGroup
	maxConcurrency := constants.FANTIA_MAX_CONCURRENCY
	if fanclubIdsLen < maxConcurrency {
		maxConcurrency = fanclubIdsLen
	}
	queue := make(chan struct{}, maxConcurrency)
	resTsSlice := threadsafe.NewSliceWithCapacity[[]string](fanclubIdsLen)
	errTsSlice := threadsafe.NewSlice[error]()

	progress := dlOptions.Base.MainProgBar()
	if fanclubIdsLen > 1 {
		baseMsg := "Getting " + contentType + " ID(s) from Fanclubs(s) on Fantia [%d/" + fmt.Sprintf("%d]...", fanclubIdsLen)
		progress.SetToProgressBar()
		progress.UpdateBaseMsg(baseMsg)
		progress.UpdateSuccessMsg(
			fmt.Sprintf(
				"Finished getting %s ID(s) from %d Fanclubs(s) on Fantia!",
				contentType,
				fanclubIdsLen,
			),
		)
		progress.UpdateErrorMsg(
			fmt.Sprintf(
				"Something went wrong while getting %s IDs from %d Fanclubs(s) on Fantia.\nPlease refer to the logs for more details.",
				contentType,
				fanclubIdsLen,
			),
		)
		progress.UpdateMax(fanclubIdsLen)
	} else {
		progress.SetToSpinner()
		fanclubId := fanclubIds[0]
		progress.UpdateBaseMsg("Getting " + contentType + " ID(s) from Fanclub, " + fanclubId + ", on Fantia...")
		progress.UpdateSuccessMsg("Finished getting " + contentType + " ID(s) from Fanclub, " + fanclubId + ", on Fantia!")
		progress.UpdateErrorMsg("Something went wrong while getting " + contentType + " ID(s) from Fanclub, " + fanclubId + ", on Fantia.\nPlease refer to the logs for more details.")
	}
	progress.Start()
	for pageNumIdx, fanclubId := range fanclubIds {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				<-queue
			}()

			queue <- struct{}{}
			contentIds, err := getFanclubContent(
				fanclubId,
				pageNums[pageNumIdx],
				dlOptions,
				contentType,
			)
			if err != nil {
				errTsSlice.Append(err)
			} else {
				resTsSlice.Append(contentIds)
			}

			progress.Increment()
		}()
	}
	wg.Wait()
	close(queue)

	var errorSlice []error
	hasErr, hasCancelled := false, false
	if errTsSlice.LenUnsafe() > 0 {
		hasErr = true
		hasCancelled, errorSlice = logger.LogSliceErrors(logger.ERROR, errTsSlice)
	}
	if hasCancelled {
		dlOptions.CancelCtx()
		progress.StopInterrupt("Stopped getting " + contentType + " ID(s) from Fanclub(s) on Fantia...")
		progress.SnapshotTask()
		return nil
	}
	progress.Stop(hasErr)
	progress.SnapshotTask()

	resIter := resTsSlice.NewIter()
	for resIter.Next() {
		contentIdsRes := resIter.Item()
		if contentType == PRODUCTS {
			f.ProductIds = append(f.ProductIds, contentIdsRes...)
		} else {
			f.PostIds = append(f.PostIds, contentIdsRes...)
		}
	}
	if contentType == PRODUCTS {
		f.ProductIds = utils.RemoveSliceDuplicates(f.ProductIds)
	} else {
		f.PostIds = utils.RemoveSliceDuplicates(f.PostIds)
	}
	return errorSlice
}
