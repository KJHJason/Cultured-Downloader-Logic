package cdlogic

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixivfanbox"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

// Start the download process for Pixiv Fanbox
func PixivFanboxDownloadProcess(pixivFanboxDl *pixivfanbox.PixivFanboxDl, pixivFanboxDlOptions *pixivfanbox.PixivFanboxDlOptions) []error {
	defer pixivFanboxDlOptions.CancelCtx()
	if !pixivFanboxDlOptions.Base.DlThumbnails && !pixivFanboxDlOptions.Base.DlImages && !pixivFanboxDlOptions.Base.DlAttachments && !pixivFanboxDlOptions.Base.DlGdrive {
		return nil
	}

	var errSlice []error
	if len(pixivFanboxDl.CreatorIds) > 0 {
		if err := pixivFanboxDl.GetCreatorsPosts(pixivFanboxDlOptions); len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
	}

	var urlsToDownload, gdriveUrlsToDownload []*httpfuncs.ToDownload
	if len(pixivFanboxDl.PostIds) > 0 && pixivFanboxDlOptions.CtxIsActive() {
		var err []error
		urlsToDownload, gdriveUrlsToDownload, err = pixivFanboxDl.GetPostDetails(
			pixivFanboxDlOptions,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
	}

	var downloadedPosts bool
	if len(urlsToDownload) > 0 && pixivFanboxDlOptions.CtxIsActive() {
		downloadedPosts = true
		cancelled, err := httpfuncs.DownloadUrls(
			urlsToDownload,
			&httpfuncs.DlOptions{
				Context:         pixivFanboxDlOptions.GetContext(),
				MaxConcurrency:  constants.PIXIV_FANBOX_MAX_CONCURRENCY,
				Headers:         pixivfanbox.GetPixivFanboxHeaders(),
				Cookies:         pixivFanboxDlOptions.Base.SessionCookies,
				UseHttp3:        false,
				HeadReqTimeout:  constants.DEFAULT_HEAD_REQ_TIMEOUT,
				SupportRange:    constants.PIXIV_FANBOX_RANGE_SUPPORTED,
				ProgressBarInfo: pixivFanboxDlOptions.Base.ProgressBarInfo,
			},
			pixivFanboxDlOptions.Base.Configs,
		)
		if cancelled {
			return nil
		}
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
	}
	if pixivFanboxDlOptions.Base.GdriveClient != nil && len(gdriveUrlsToDownload) > 0 && pixivFanboxDlOptions.CtxIsActive() {
		downloadedPosts = true
		err := pixivFanboxDlOptions.Base.GdriveClient.DownloadGdriveUrls(
			gdriveUrlsToDownload,
			pixivFanboxDlOptions.Base.ProgressBarInfo,
			pixivFanboxDlOptions.Base.Filters,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
	}

	notifier := pixivFanboxDlOptions.Base.Notifier
	if downloadedPosts {
		notifier.Alert("Downloaded all posts from Pixiv Fanbox!")
	} else {
		notifier.Alert("No posts to download from Pixiv Fanbox!")
	}
	return errSlice
}
