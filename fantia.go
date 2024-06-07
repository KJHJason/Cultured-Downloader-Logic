package cdlogic

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api/fantia"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

// Start the download process for Fantia
func FantiaDownloadProcess(fantiaDl *fantia.FantiaDl, fantiaDlOptions *fantia.FantiaDlOptions) []error {
	defer fantiaDlOptions.CancelCtx()
	if !fantiaDlOptions.Base.DlThumbnails && !fantiaDlOptions.Base.DlImages && !fantiaDlOptions.Base.DlAttachments {
		return nil
	}

	var errorSlice []error
	if len(fantiaDl.FanclubIds) > 0 {
		if errSlice := fantiaDl.GetFanclubsPosts(fantiaDlOptions); len(errSlice) > 0 {
			errorSlice = append(errorSlice, errSlice...)
		}
	}

	var gdriveLinks []*httpfuncs.ToDownload
	var downloadedPosts bool
	if len(fantiaDl.PostIds) > 0 && fantiaDlOptions.CtxIsActive() {
		gdriveUrls, errSlice := fantiaDl.DlFantiaPosts(fantiaDlOptions)
		if len(errSlice) > 0 {
			errorSlice = append(errorSlice, errSlice...)
		} else {
			gdriveLinks = append(gdriveLinks, gdriveUrls...)
		}
		downloadedPosts = true
	}

	if len(fantiaDl.ProductFanclubIds) > 0 && fantiaDlOptions.CtxIsActive() {
		if errSlice := fantiaDl.GetFanclubsProducts(fantiaDlOptions); len(errSlice) > 0 {
			errorSlice = append(errorSlice, errSlice...)
		}
	}

	if len(fantiaDl.ProductIds) > 0 && fantiaDlOptions.CtxIsActive() {
		productContents, errSlice := fantiaDl.GetProducts(fantiaDlOptions)
		if len(errSlice) > 0 {
			errorSlice = append(errorSlice, errSlice...)
		}

		downloadedPosts = true
		cancelled, errSlice := httpfuncs.DownloadUrls(
			productContents,
			&httpfuncs.DlOptions{
				Context:         fantiaDlOptions.GetContext(),
				MaxConcurrency:  constants.FANTIA_MAX_CONCURRENCY,
				Cookies:         fantiaDlOptions.Base.SessionCookies,
				UseHttp3:        constants.FANTIA_PRODUCT_USE_HTTP3,
				SupportRange:    constants.FANTIA_RANGE_SUPPORTED,
				HeadReqTimeout:  constants.DEFAULT_HEAD_REQ_TIMEOUT,
				SetMetadata:     fantiaDlOptions.Base.SetMetadata,
				Filters:         fantiaDlOptions.Base.Filters,
				ProgressBarInfo: fantiaDlOptions.Base.ProgressBarInfo,
			},
			fantiaDlOptions.Base.Configs,
		)
		if cancelled {
			return nil
		}
		if len(errSlice) > 0 {
			errorSlice = append(errorSlice, errSlice...)
		}
	}

	if fantiaDlOptions.Base.GdriveClient != nil && len(gdriveLinks) > 0 && fantiaDlOptions.CtxIsActive() {
		gdriveErrs := fantiaDlOptions.Base.GdriveClient.DownloadGdriveUrls(
			gdriveLinks,
			fantiaDlOptions.Base.ProgressBarInfo,
			fantiaDlOptions.Base.Filters,
		)
		errorSlice = append(errorSlice, gdriveErrs...)
		downloadedPosts = true
	}

	notifier := fantiaDlOptions.GetNotifier()
	if downloadedPosts {
		notifier.Alert("Downloaded all posts from Fantia!")
	} else {
		notifier.Alert("No posts to download from Fantia!")
	}
	return errorSlice
}
