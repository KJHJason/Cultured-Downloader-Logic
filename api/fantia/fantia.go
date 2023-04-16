package fantia

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notifier"
	"fyne.io/fyne/v2"
)

// Start the download process for Fantia
func FantiaDownloadProcess(fantiaDl *FantiaDl, fantiaDlOptions *FantiaDlOptions, notifTitle string, app fyne.App) {
	if !fantiaDlOptions.DlThumbnails && !fantiaDlOptions.DlImages && !fantiaDlOptions.DlAttachments {
		return
	}

	if len(fantiaDl.FanclubIds) > 0 {
		fantiaDl.getCreatorsPosts(fantiaDlOptions)
	}

	var gdriveLinks []*httpfuncs.ToDownload
	var downloadedPosts bool
	if len(fantiaDl.PostIds) > 0 {
		fantiaDl.dlFantiaPosts(fantiaDlOptions, notifTitle, app)
		downloadedPosts = true
	}

	if fantiaDlOptions.GdriveClient != nil && len(gdriveLinks) > 0 {
		fantiaDlOptions.GdriveClient.DownloadGdriveUrls(gdriveLinks, fantiaDlOptions.Configs)
		downloadedPosts = true
	}

	if downloadedPosts {
		notifier.AlertWithoutErr(notifTitle, "Downloaded all posts from Fantia!", app)
	} else {
		notifier.AlertWithoutErr(notifTitle, "No posts to download from Fantia!", app)
	}
}
