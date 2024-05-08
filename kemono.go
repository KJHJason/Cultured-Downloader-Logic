package cdlogic

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/kemono"
)

func KemonoDownloadProcess(config *configs.Config, kemonoDl *kemono.KemonoDl, dlOptions *kemono.KemonoDlOptions) []error {
	defer dlOptions.CancelCtx()
	if !dlOptions.DlAttachments && !dlOptions.DlGdrive {
		return nil
	}

	var errSlice []error
	var toDownload, gdriveLinks []*httpfuncs.ToDownload
	if kemonoDl.DlFav {
		progress := dlOptions.MainProgBar
		progress.SetToSpinner()
		progress.UpdateBaseMsg("Getting favourites from Kemono...")
		progress.UpdateSuccessMsg("Finished getting favourites from Kemono!")
		progress.UpdateErrorMsg("Something went wrong while getting favourites from Kemono.\nPlease refer to the logs for more details.")
		progress.Start()
		favToDl, favGdriveLinks, err := kemono.GetFavourites(dlOptions)
		hasErr := (err != nil)
		if hasErr {
			cancel := logger.LogErrors(false, logger.ERROR, err...)
			if cancel {
				return nil
			}
			errSlice = append(errSlice, err...)
		} else {
			toDownload = favToDl
			gdriveLinks = favGdriveLinks
		}
		progress.Stop(hasErr)
		progress.SnapshotTask()
	}

	if len(kemonoDl.PostsToDl) > 0 && dlOptions.CtxIsActive() {
		postsToDl, gdriveLinksToDl, err := kemono.GetMultiplePosts(kemonoDl.PostsToDl, dlOptions)
		if err != nil {
			errSlice = append(errSlice, err...)
		} else {
			toDownload = append(toDownload, postsToDl...)
			gdriveLinks = append(gdriveLinks, gdriveLinksToDl...)
		}
	}
	if len(kemonoDl.CreatorsToDl) > 0 && dlOptions.CtxIsActive() {
		creatorsToDl, gdriveLinksToDl, err := kemono.GetMultipleCreators(kemonoDl.CreatorsToDl, dlOptions,)
		if err != nil {
			errSlice = append(errSlice, err...)
		} else {
			toDownload = append(toDownload, creatorsToDl...)
			gdriveLinks = append(gdriveLinks, gdriveLinksToDl...)
		}
	}

	var downloadedPosts bool
	if len(toDownload) > 0 && dlOptions.CtxIsActive() {
		downloadedPosts = true
		cancelled, err := httpfuncs.DownloadUrls(
			toDownload,
			&httpfuncs.DlOptions{
				Context:        dlOptions.GetContext(),
				MaxConcurrency: constants.PIXIV_MAX_CONCURRENT_DOWNLOADS,
				Cookies:        dlOptions.SessionCookies,
				UseHttp3:       httpfuncs.IsHttp3Supported(constants.KEMONO, false),
				RetryDelay:     &httpfuncs.RetryDelay{Min: 25, Max: 35},
			},
			config,
		)
		if cancelled {
			return nil
		}
		if err != nil {
			errSlice = append(errSlice, err...)
		}
	}
	if dlOptions.GdriveClient != nil && len(gdriveLinks) > 0 && dlOptions.CtxIsActive() {
		downloadedPosts = true
		err := dlOptions.GdriveClient.DownloadGdriveUrls(
			gdriveLinks, 
			config, 
			&progress.ProgressBarInfo{
				MainProgressBar:      dlOptions.MainProgBar,
				DownloadProgressBars: dlOptions.DownloadProgressBars,
			},
		)
		if err != nil {
			errSlice = append(errSlice, err...)
		}
	}

	notifier := dlOptions.Notifier
	if downloadedPosts {
		notifier.Alert("Downloaded all posts from Kemono!")
	} else {
		notifier.Alert("No posts to download from Kemono!")
	}
	return errSlice
}
