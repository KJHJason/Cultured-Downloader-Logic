package cdlogic

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api/fantia"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

// Start the download process for Fantia
func FantiaDownloadProcess(fantiaDl *fantia.FantiaDl, fantiaDlOptions *fantia.FantiaDlOptions, dlOptions *httpfuncs.DlOptions) []*error {
	if !fantiaDlOptions.DlThumbnails && !fantiaDlOptions.DlImages && !fantiaDlOptions.DlAttachments {
		return nil
	}

	if len(fantiaDl.FanclubIds) > 0 {
		fantiaDl.GetCreatorsPosts(fantiaDlOptions)
	}

	var errorSlice []*error
	var gdriveLinks []*httpfuncs.ToDownload
	var downloadedPosts bool
	if len(fantiaDl.PostIds) > 0 {
		fantiaDl.DlFantiaPosts(fantiaDlOptions)
		downloadedPosts = true
	}

	if fantiaDlOptions.GdriveClient != nil && len(gdriveLinks) > 0 {
		gdriveErrs := fantiaDlOptions.GdriveClient.DownloadGdriveUrls(
			gdriveLinks, 
			fantiaDlOptions.Configs, 
			dlOptions,
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
