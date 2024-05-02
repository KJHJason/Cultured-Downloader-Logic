package cdlogic

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api/fantia"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

// Start the download process for Fantia
func FantiaDownloadProcess(fantiaDl *fantia.FantiaDl, fantiaDlOptions *fantia.FantiaDlOptions) []*error {
	if !fantiaDlOptions.DlThumbnails && !fantiaDlOptions.DlImages && !fantiaDlOptions.DlAttachments {
		return nil
	}

	var errorSlice []*error
	if len(fantiaDl.FanclubIds) > 0 {
		if errSlice := fantiaDl.GetCreatorsPosts(fantiaDlOptions); len(errSlice) > 0 {
			errorSlice = append(errorSlice, errSlice...)
		}
	}

	var gdriveLinks []*httpfuncs.ToDownload
	var downloadedPosts bool
	if len(fantiaDl.PostIds) > 0 {
		gdriveUrls, errSlice := fantiaDl.DlFantiaPosts(fantiaDlOptions)
		if len(errSlice) > 0 {
			errorSlice = append(errorSlice, errSlice...)
		} else {
			gdriveLinks = append(gdriveLinks, gdriveUrls...)
		}
		downloadedPosts = true
	}

	if fantiaDlOptions.GdriveClient != nil && len(gdriveLinks) > 0 {
		gdriveErrs := fantiaDlOptions.GdriveClient.DownloadGdriveUrls(
			gdriveLinks, 
			fantiaDlOptions.Configs, 
			&progress.ProgressBarInfo{
				MainProgressBar:        fantiaDlOptions.MainProgBar,
				DownloadProgressBars:   &fantiaDlOptions.DownloadProgressBars,
				NewDownloadProgressBar: fantiaDlOptions.NewDownloadProgressBar,
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
