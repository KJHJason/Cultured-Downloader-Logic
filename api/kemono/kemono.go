package kemono

import (
	"fyne.io/fyne/v2"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

func KemonoDownloadProcess(config *configs.Config, kemonoDl *KemonoDl, dlOptions *KemonoDlOptions, notifTitle string, app fyne.App) {
	if !dlOptions.DlAttachments && !dlOptions.DlGdrive {
		return
	}

	var toDownload, gdriveLinks []*httpfuncs.ToDownload
	if kemonoDl.DlFav {
		progress := dlOptions.GetFavouritesPostProgBar
		progress.UpdateBaseMsg("Getting favourites from Kemono Party...")
		progress.UpdateSuccessMsg("Finished getting favourites from Kemono Party!")
		progress.UpdateErrorMsg("Something went wrong while getting favourites from Kemono Party.\nPlease refer to the logs for more details.")
		progress.Start()
		favToDl, favGdriveLinks, err := getFavourites(
			iofuncs.DOWNLOAD_PATH,
			dlOptions,
		)
		hasErr := (err != nil)
		if hasErr {
			logger.LogError(err, false, logger.ERROR)
		} else {
			toDownload = favToDl
			gdriveLinks = favGdriveLinks
		}
		progress.Stop(hasErr)
	}

	if len(kemonoDl.PostsToDl) > 0 {
		postsToDl, gdriveLinksToDl := getMultiplePosts(
			kemonoDl.PostsToDl,
			iofuncs.DOWNLOAD_PATH,
			dlOptions,
		)
		toDownload = append(toDownload, postsToDl...)
		gdriveLinks = append(gdriveLinks, gdriveLinksToDl...)
	}
	if len(kemonoDl.CreatorsToDl) > 0 {
		creatorsToDl, gdriveLinksToDl := getMultipleCreators(
			kemonoDl.CreatorsToDl,
			iofuncs.DOWNLOAD_PATH,
			dlOptions,
		)
		toDownload = append(toDownload, creatorsToDl...)
		gdriveLinks = append(gdriveLinks, gdriveLinksToDl...)
	}

	var downloadedPosts bool
	if len(toDownload) > 0 {
		downloadedPosts = true
		httpfuncs.DownloadUrls(
			toDownload,
			&httpfuncs.DlOptions{
				MaxConcurrency: constants.PIXIV_MAX_CONCURRENT_DOWNLOADS,
				Cookies:        dlOptions.SessionCookies,
				UseHttp3:       httpfuncs.IsHttp3Supported(constants.KEMONO, false),
				RetryDelay:     &httpfuncs.RetryDelay{Min: 25, Max: 35},
			},
			config,
		)
	}
	if dlOptions.GdriveClient != nil && len(gdriveLinks) > 0 {
		downloadedPosts = true
		dlOptions.GdriveClient.DownloadGdriveUrls(gdriveLinks, config, dlOptions.GdriveApiProgBar, dlOptions.GdriveDlProgBar)
	}

	notifier := dlOptions.Notifier
	if downloadedPosts {
		notifier.Alert("Downloaded all posts from Kemono Party!")
	} else {
		notifier.Alert("No posts to download from Kemono Party!")
	}
}
