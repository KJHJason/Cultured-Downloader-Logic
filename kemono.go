package cdlogic

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api/kemono"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

func KemonoDownloadProcess(kemonoDl *kemono.KemonoDl, dlOptions *kemono.KemonoDlOptions) []error {
	defer dlOptions.CancelCtx()
	if !dlOptions.DlAttachments && !dlOptions.DlGdrive {
		return nil
	}

	var errSlice []error
	var toDownload, gdriveLinks []*httpfuncs.ToDownload
	if kemonoDl.DlFav {
		prog := dlOptions.MainProgBar
		prog.SetToSpinner()
		prog.UpdateBaseMsg("Getting favourites from Kemono...")
		prog.UpdateSuccessMsg("Finished getting favourites from Kemono!")
		prog.UpdateErrorMsg("Something went wrong while getting favourites from Kemono.\nPlease refer to the logs for more details.")
		prog.Start()
		favToDl, favGdriveLinks, err := kemono.GetFavourites(dlOptions)
		hasErr := (err != nil)
		if hasErr {
			cancel := logger.LogErrors(logger.ERROR, err...)
			if cancel {
				return nil
			}
			errSlice = append(errSlice, err...)
		} else {
			toDownload = favToDl
			gdriveLinks = favGdriveLinks
		}
		prog.Stop(hasErr)
		prog.SnapshotTask()
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
		creatorsToDl, gdriveLinksToDl, err := kemono.GetMultipleCreators(kemonoDl.CreatorsToDl, dlOptions)
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
				MaxConcurrency: constants.KEMONO_MAX_CONCURRENCY,
				Cookies:        dlOptions.SessionCookies,
				UseHttp3:       httpfuncs.IsHttp3Supported(constants.KEMONO, false),
				SupportRange:   constants.KEMONO_RANGE_SUPPORTED,
				HeadReqTimeout: constants.KEMONO_HEAD_REQ_TIMEOUT,
				RetryDelay:     &httpfuncs.RetryDelay{Min: 25, Max: 35},
				ProgressBarInfo: &progress.ProgressBarInfo{
					MainProgressBar:      dlOptions.MainProgBar,
					DownloadProgressBars: dlOptions.DownloadProgressBars,
				},
			},
			dlOptions.Configs,
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
