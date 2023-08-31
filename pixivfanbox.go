package cdlogic

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixivfanbox"
)

// Start the download process for Pixiv Fanbox
func PixivFanboxDownloadProcess(pixivFanboxDl *pixivfanbox.PixivFanboxDl, pixivFanboxDlOptions *pixivfanbox.PixivFanboxDlOptions) {
	if !pixivFanboxDlOptions.DlThumbnails && !pixivFanboxDlOptions.DlImages && !pixivFanboxDlOptions.DlAttachments && !pixivFanboxDlOptions.DlGdrive {
		return
	}

	if len(pixivFanboxDl.CreatorIds) > 0 {
		pixivFanboxDl.GetCreatorsPosts(
			pixivFanboxDlOptions,
		)
	}

	var urlsToDownload, gdriveUrlsToDownload []*httpfuncs.ToDownload
	if len(pixivFanboxDl.PostIds) > 0 {
		urlsToDownload, gdriveUrlsToDownload = pixivFanboxDl.GetPostDetails(
			pixivFanboxDlOptions,
		)
	}

	var downloadedPosts bool
	if len(urlsToDownload) > 0 {
		downloadedPosts = true
		httpfuncs.DownloadUrls(
			urlsToDownload,
			&httpfuncs.DlOptions{
				MaxConcurrency: constants.PIXIV_MAX_CONCURRENT_DOWNLOADS,
				Headers:        pixivfanbox.GetPixivFanboxHeaders(),
				Cookies:        pixivFanboxDlOptions.SessionCookies,
				UseHttp3:       false,
			},
			pixivFanboxDlOptions.Configs,
		)
	}
	if pixivFanboxDlOptions.GdriveClient != nil && len(gdriveUrlsToDownload) > 0 {
		downloadedPosts = true
		pixivFanboxDlOptions.GdriveClient.DownloadGdriveUrls(
			gdriveUrlsToDownload, 
			pixivFanboxDlOptions.Configs, 
			pixivFanboxDlOptions.GdriveApiProgBar, 
			pixivFanboxDlOptions.GdriveDlProgBar,
		)
	}

	notifier := pixivFanboxDlOptions.Notifier
	if downloadedPosts {
		notifier.Alert("Downloaded all posts from Pixiv Fanbox!")
	} else {
		notifier.Alert("No posts to download from Pixiv Fanbox!")
	}
}
