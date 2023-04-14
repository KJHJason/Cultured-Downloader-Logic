package pixivfanbox

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/notifier"
)

// Start the download process for Pixiv Fanbox
func PixivFanboxDownloadProcess(pixivFanboxDl *PixivFanboxDl, pixivFanboxDlOptions *PixivFanboxDlOptions, notifTitle string) {
	if !pixivFanboxDlOptions.DlThumbnails && !pixivFanboxDlOptions.DlImages && !pixivFanboxDlOptions.DlAttachments && !pixivFanboxDlOptions.DlGdrive {
		return
	}

	if len(pixivFanboxDl.CreatorIds) > 0 {
		pixivFanboxDl.getCreatorsPosts(
			pixivFanboxDlOptions,
		)
	}

	var urlsToDownload, gdriveUrlsToDownload []*httpfuncs.ToDownload
	if len(pixivFanboxDl.PostIds) > 0 {
		urlsToDownload, gdriveUrlsToDownload = pixivFanboxDl.getPostDetails(
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
				Headers:        GetPixivFanboxHeaders(),
				Cookies:        pixivFanboxDlOptions.SessionCookies,
				UseHttp3:       false,
			},
			pixivFanboxDlOptions.Configs,
		)
	}
	if pixivFanboxDlOptions.GdriveClient != nil && len(gdriveUrlsToDownload) > 0 {
		downloadedPosts = true
		pixivFanboxDlOptions.GdriveClient.DownloadGdriveUrls(gdriveUrlsToDownload, pixivFanboxDlOptions.Configs)
	}

	if downloadedPosts {
		notifier.AlertWithoutErr(notifTitle, "Downloaded all posts from Pixiv Fanbox!")
	} else {
		notifier.AlertWithoutErr(notifTitle, "No posts to download from Pixiv Fanbox!")
	}
}
