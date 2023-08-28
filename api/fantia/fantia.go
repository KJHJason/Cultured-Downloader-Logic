package fantia

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

// Start the download process for Fantia
func FantiaDownloadProcess(fantiaDl *FantiaDl, fantiaDlOptions *FantiaDlOptions, notifTitle string) {
	if !fantiaDlOptions.DlThumbnails && !fantiaDlOptions.DlImages && !fantiaDlOptions.DlAttachments {
		return
	}

	if len(fantiaDl.FanclubIds) > 0 {
		fantiaDl.getCreatorsPosts(fantiaDlOptions)
	}

	var gdriveLinks []*httpfuncs.ToDownload
	var downloadedPosts bool
	if len(fantiaDl.PostIds) > 0 {
		fantiaDl.dlFantiaPosts(fantiaDlOptions, notifTitle)
		downloadedPosts = true
	}

	if fantiaDlOptions.GdriveClient != nil && len(gdriveLinks) > 0 {
		fantiaDlOptions.GdriveClient.DownloadGdriveUrls(gdriveLinks, fantiaDlOptions.Configs)
		downloadedPosts = true
	}

	notifier := fantiaDlOptions.GetNotifier()
	if downloadedPosts {
		notifier.Alert("Downloaded all posts from Fantia!")
	} else {
		notifier.Alert("No posts to download from Fantia!")
	}
}
