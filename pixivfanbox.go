package cdlogic

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixivfanbox"
)

// Start the download process for Pixiv Fanbox
func PixivFanboxDownloadProcess(pixivFanboxDl *pixivfanbox.PixivFanboxDl, pixivFanboxDlOptions *pixivfanbox.PixivFanboxDlOptions) []error {
	defer pixivFanboxDlOptions.CancelCtx()
	if !pixivFanboxDlOptions.DlThumbnails && !pixivFanboxDlOptions.DlImages && !pixivFanboxDlOptions.DlAttachments && !pixivFanboxDlOptions.DlGdrive {
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
	if len(urlsToDownload) > 0 &&  pixivFanboxDlOptions.CtxIsActive() {
		downloadedPosts = true
		cancelled, err := httpfuncs.DownloadUrls(
			urlsToDownload,
			&httpfuncs.DlOptions{
				Context:        pixivFanboxDlOptions.GetContext(),
				MaxConcurrency: constants.PIXIV_FANBOX_MAX_CONCURRENT,
				Headers:        pixivfanbox.GetPixivFanboxHeaders(),
				Cookies:        pixivFanboxDlOptions.SessionCookies,
				UseHttp3:       false,
				SupportRange:   constants.PIXIV_FANBOX_RANGE_SUPPORTED,
				ProgressBarInfo: &progress.ProgressBarInfo{
					MainProgressBar:      pixivFanboxDlOptions.MainProgBar,
					DownloadProgressBars: pixivFanboxDlOptions.DownloadProgressBars,
				},
			},
			pixivFanboxDlOptions.Configs,
		)
		if cancelled {
			return nil
		}
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
	}
	if pixivFanboxDlOptions.GdriveClient != nil && len(gdriveUrlsToDownload) > 0 && pixivFanboxDlOptions.CtxIsActive() {
		downloadedPosts = true
		err := pixivFanboxDlOptions.GdriveClient.DownloadGdriveUrls(
			gdriveUrlsToDownload, 
			pixivFanboxDlOptions.Configs, 
			&progress.ProgressBarInfo{
				MainProgressBar:      pixivFanboxDlOptions.MainProgBar,
				DownloadProgressBars: pixivFanboxDlOptions.DownloadProgressBars,
			},
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
	}

	notifier := pixivFanboxDlOptions.Notifier
	if downloadedPosts {
		notifier.Alert("Downloaded all posts from Pixiv Fanbox!")
	} else {
		notifier.Alert("No posts to download from Pixiv Fanbox!")
	}
	return errSlice
}
