package cdlogic

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api/fantia"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

// Start the download process for Fantia
func FantiaDownloadProcess(fantiaDl *fantia.FantiaDl, fantiaDlOptions *fantia.FantiaDlOptions) {
	if !fantiaDlOptions.DlThumbnails && !fantiaDlOptions.DlImages && !fantiaDlOptions.DlAttachments {
		return
	}

	if len(fantiaDl.FanclubIds) > 0 {
		fantiaDl.GetCreatorsPosts(fantiaDlOptions)
	}

	var gdriveLinks []*httpfuncs.ToDownload
	var downloadedPosts bool
	if len(fantiaDl.PostIds) > 0 {
		fantiaDl.DlFantiaPosts(fantiaDlOptions)
		downloadedPosts = true
	}

	if fantiaDlOptions.GdriveClient != nil && len(gdriveLinks) > 0 {
		fantiaDlOptions.GdriveClient.DownloadGdriveUrls(
			gdriveLinks, 
			fantiaDlOptions.Configs, 
			fantiaDlOptions.GdriveApiProgBar, 
			fantiaDlOptions.GdriveDlProgBar,
		)
		downloadedPosts = true
	}

	notifier := fantiaDlOptions.GetNotifier()
	if downloadedPosts {
		notifier.Alert("Downloaded all posts from Fantia!")
	} else {
		notifier.Alert("No posts to download from Fantia!")
	}
}
