package fantia

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

func getFantiaProductPaidContent(purchaseRelativeUrl, productId string, dlOptions *FantiaDlOptions) (*http.Response, error) {
	useHttp3 := httpfuncs.IsHttp3Supported(constants.FANTIA, false)
	purchaseUrl := constants.FANTIA_URL + purchaseRelativeUrl // https://fantia.jp/mypage/users/purchases/123456
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Method:    "GET",
			Url:       purchaseUrl,
			Cookies:   dlOptions.Base.SessionCookies,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			UserAgent: dlOptions.Base.Configs.UserAgent,
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
	if dlOptions.Base.UseCacheDb {
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
			Cookies:   dlOptions.Base.SessionCookies,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			UserAgent: dlOptions.Base.Configs.UserAgent,
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
	productIdsLen := len(f.ProductIds)
	if productIdsLen == 0 {
		return nil, nil
	}

	// Now that we have the post ID, we can query Fantia's API
	// to get the post's contents from the JSON response.
	progress := dlOptions.Base.MainProgBar()
	progress.SetToProgressBar()
	progress.UpdateBaseMsg(
		"Getting product contents from Fantia [%d/" + fmt.Sprintf("%d]...", productIdsLen),
	)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting %d products from Fantia!",
			productIdsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting %d product contents from Fantia.\nPlease refer to the logs for more details.",
			productIdsLen,
		),
	)
	progress.UpdateMax(productIdsLen)
	progress.Start()
	defer progress.SnapshotTask()

	var wg sync.WaitGroup
	maxConcurrency := constants.FANTIA_PRODUCT_MAX_CONCURRENCY
	if productIdsLen < maxConcurrency {
		maxConcurrency = productIdsLen
	}
	queue := make(chan struct{}, maxConcurrency)
	resChan := make(chan []*httpfuncs.ToDownload, productIdsLen)
	errChan := make(chan error, productIdsLen)

	for idx, productId := range f.ProductIds {
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
			progress.StopInterrupt(
				fmt.Sprintf("Stopped getting %d product content from Fantia...", productIdsLen),
			)
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

func (f *FantiaDl) GetFanclubsProducts(dlOptions *FantiaDlOptions) []error {
	return f.GetFanclubsContents(f.ProductFanclubIds, f.ProductFanclubPageNums, PRODUCTS, dlOptions)
}