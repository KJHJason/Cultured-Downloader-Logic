package cdlogic

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api/fantia"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

// Start the download process for Fantia
func FantiaDownloadProcess(fantiaDl *fantia.FantiaDl, fantiaDlOptions *fantia.FantiaDlOptions) []error {
	defer fantiaDlOptions.CancelCtx()
	if !fantiaDlOptions.DlThumbnails && !fantiaDlOptions.DlImages && !fantiaDlOptions.DlAttachments {
		return nil
	}

	var errorSlice []error
	if len(fantiaDl.FanclubIds) > 0 {
		if errSlice := fantiaDl.GetCreatorsPosts(fantiaDlOptions); len(errSlice) > 0 {
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
		if errSlice := fantiaDl.GetCreatorsProducts(fantiaDlOptions); len(errSlice) > 0 {
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
				Context:        fantiaDlOptions.GetContext(),
				MaxConcurrency: constants.FANTIA_MAX_CONCURRENCY,
				Cookies:        fantiaDlOptions.SessionCookies,
				UseHttp3:       constants.FANTIA_PRODUCT_USE_HTTP3,
				SupportRange:   constants.FANTIA_RANGE_SUPPORTED,
				HeadReqTimeout: constants.DEFAULT_HEAD_REQ_TIMEOUT,
				ProgressBarInfo: &progress.ProgressBarInfo{
					MainProgressBar:      fantiaDlOptions.MainProgBar,
					DownloadProgressBars: fantiaDlOptions.DownloadProgressBars,
				},
			},
			fantiaDlOptions.Configs,
		)
		if cancelled {
			return nil
		}
		if len(errSlice) > 0 {
			errorSlice = append(errorSlice, errSlice...)
		}
	}

	if fantiaDlOptions.GdriveClient != nil && len(gdriveLinks) > 0 && fantiaDlOptions.CtxIsActive() {
		gdriveErrs := fantiaDlOptions.GdriveClient.DownloadGdriveUrls(
			gdriveLinks,
			&progress.ProgressBarInfo{
				MainProgressBar:      fantiaDlOptions.MainProgBar,
				DownloadProgressBars: fantiaDlOptions.DownloadProgressBars,
			},
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
