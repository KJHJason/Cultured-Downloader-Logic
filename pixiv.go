package cdlogic

import (
	"fmt"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv"
	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	pixivmobile "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/mobile"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
	pixivweb "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/web"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

func alertUser(artworksToDl []*httpfuncs.ToDownload, ugoiraToDl []*ugoira.Ugoira, notifier notify.Notifier) {
	if len(artworksToDl) > 0 || len(ugoiraToDl) > 0 {
		notifier.Alert("Finished downloading artworks from Pixiv!")
	} else {
		notifier.Alert("No artworks to download from Pixiv!")
	}
}

// Start the download process for Pixiv
func PixivWebDownloadProcess(pixivDl *pixiv.PixivDl, pixivDlOptions *pixivweb.PixivWebDlOptions, pixivUgoiraOptions *ugoira.UgoiraOptions) []error {
	defer pixivDlOptions.CancelCtx()
	var errSlice []error
	var ugoiraToDl []*ugoira.Ugoira
	var artworksToDl []*httpfuncs.ToDownload

	if len(pixivDl.IllustratorIds) > 0 {
		artworkIdsSlice, err := pixivweb.GetMultipleArtistsPosts(
			pixivDl.IllustratorIds,
			pixivDl.IllustratorPageNums,
			iofuncs.DOWNLOAD_PATH,
			pixivDlOptions,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		} else {
			pixivDl.ArtworkIds = append(pixivDl.ArtworkIds, artworkIdsSlice...)
			pixivDl.ArtworkIds = api.RemoveSliceDuplicates(pixivDl.ArtworkIds)
		}
	}

	if len(pixivDl.ArtworkIds) > 0 && pixivDlOptions.CtxIsActive() {
		artworkSlice, ugoiraSlice, err := pixivweb.GetMultipleArtworkDetails(
			pixivDl.ArtworkIds,
			iofuncs.DOWNLOAD_PATH,
			pixivDlOptions,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		} else {
			artworksToDl = append(artworksToDl, artworkSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
		}
	}

	tagNameLen := len(pixivDl.TagNames)
	if tagNameLen > 0 && pixivDlOptions.CtxIsActive() {
		// loop through each tag and page number
		baseMsg := "Searching for artworks based on tag names on Pixiv [%d/" + fmt.Sprintf("%d]...", tagNameLen)
		progress := pixivDlOptions.MainProgBar
		progress.UpdateBaseMsg(baseMsg)
		progress.UpdateSuccessMsg(
			fmt.Sprintf(
				"Finished searching for artworks based on %d tag names on Pixiv!",
				tagNameLen,
			),
		)
		progress.UpdateErrorMsg(
			fmt.Sprintf(
				"Finished with some errors while searching for artworks based on %d tag names on Pixiv!\nPlease refer to the logs for more details...",
				tagNameLen,
			),
		)
		progress.SetToProgressBar()
		progress.UpdateMax(tagNameLen)
		progress.Start()
		hasErr := false
		for idx, tagName := range pixivDl.TagNames {
			artworksSlice, ugoiraSlice, err, hasCancelled := pixivweb.TagSearch(
				tagName,
				iofuncs.DOWNLOAD_PATH,
				pixivDl.TagNamesPageNums[idx],
				pixivDlOptions,
			)

			if len(err) > 0 {
				errSlice = append(errSlice, err...)
				hasErr = true
			}
			if hasCancelled {
				progress.StopInterrupt("Stopped searching for artworks based on tag names on Pixiv!")
				progress.SnapshotTask()
				return errSlice
			}

			artworksToDl = append(artworksToDl, artworksSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
			progress.Increment()
		}
		progress.Stop(hasErr)
		progress.SnapshotTask()
	}

	if len(artworksToDl) > 0 && pixivDlOptions.CtxIsActive() {
		httpfuncs.DownloadUrls(
			artworksToDl,
			&httpfuncs.DlOptions{
				Context:        pixivDlOptions.GetContext(),
				MaxConcurrency: constants.PIXIV_MAX_DOWNLOAD_CONCURRENCY,
				Headers:        pixivcommon.GetPixivRequestHeaders(),
				Cookies:        pixivDlOptions.SessionCookies,
				UseHttp3:       false,
				SupportRange:   constants.PIXIV_RANGE_SUPPORTED,
				ProgressBarInfo: &progress.ProgressBarInfo{
					MainProgressBar:      pixivDlOptions.MainProgBar,
					DownloadProgressBars: pixivDlOptions.DownloadProgressBars,
				},
			},
			pixivDlOptions.Configs,
		)
	}
	if len(ugoiraToDl) > 0 && pixivDlOptions.CtxIsActive() {
		ugoiraArgs := &ugoira.UgoiraArgs{
			UseMobileApi: false,
			ToDownload:   ugoiraToDl,
			Cookies:      pixivDlOptions.SessionCookies,
			MainProgBar:  pixivDlOptions.MainProgBar,
		}
		ugoiraArgs.SetContext(pixivDlOptions.GetContext())
		err := ugoira.DownloadMultipleUgoira(
			ugoiraArgs,
			pixivUgoiraOptions,
			pixivDlOptions.Configs,
			httpfuncs.CallRequest,
			&progress.ProgressBarInfo{
				MainProgressBar:      pixivDlOptions.MainProgBar,
				DownloadProgressBars: pixivDlOptions.DownloadProgressBars,
			},
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
	}

	alertUser(artworksToDl, ugoiraToDl, pixivDlOptions.Notifier)
	return errSlice
}

// Start the download process for Pixiv
func PixivMobileDownloadProcess(pixivDl *pixiv.PixivDl, pixivDlOptions *pixivmobile.PixivMobileDlOptions, pixivUgoiraOptions *ugoira.UgoiraOptions) []error {
	defer pixivDlOptions.CancelCtx()
	var errSlice []error
	var ugoiraToDl []*ugoira.Ugoira
	var artworksToDl []*httpfuncs.ToDownload

	if len(pixivDl.IllustratorIds) > 0 {
		artworkSlice, ugoiraSlice, err := pixivDlOptions.MobileClient.GetMultipleArtistsPosts(
			pixivDl.IllustratorIds,
			pixivDl.IllustratorPageNums,
			iofuncs.DOWNLOAD_PATH,
			pixivDlOptions.ArtworkType,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		} else {
			artworksToDl = artworkSlice
			ugoiraToDl = ugoiraSlice
		}
	}

	if len(pixivDl.ArtworkIds) > 0 && pixivDlOptions.CtxIsActive() {
		artworkSlice, ugoiraSlice, err := pixivDlOptions.MobileClient.GetMultipleArtworkDetails(
			pixivDl.ArtworkIds,
			iofuncs.DOWNLOAD_PATH,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		} else {
			artworksToDl = append(artworksToDl, artworkSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
		}
	}

	tagNamesLen := len(pixivDl.TagNames)
	if tagNamesLen > 0 && pixivDlOptions.CtxIsActive() {
		// loop through each tag and page number
		baseMsg := "Searching for artworks based on tag names on Pixiv [%d/" + fmt.Sprintf("%d]...", tagNamesLen)
		progress := pixivDlOptions.MainProgBar
		progress.UpdateBaseMsg(baseMsg)
		progress.UpdateSuccessMsg(
			fmt.Sprintf(
				"Finished searching for artworks based on %d tag names on Pixiv!",
				tagNamesLen,
			),
		)
		progress.UpdateErrorMsg(
			fmt.Sprintf(
				"Finished with some errors while searching for artworks based on %d tag names on Pixiv!\nPlease refer to the logs for more details...",
				tagNamesLen,
			),
		)
		progress.SetToProgressBar()
		progress.UpdateMax(tagNamesLen)
		progress.Start()
		hasErr := false
		for idx, tagName := range pixivDl.TagNames {
			var artworksSlice []*httpfuncs.ToDownload
			var ugoiraSlice []*ugoira.Ugoira
			artworksSlice, ugoiraSlice, err, hasCancelled := pixivDlOptions.MobileClient.TagSearch(
				tagName,
				iofuncs.DOWNLOAD_PATH,
				pixivDl.TagNamesPageNums[idx],
				pixivDlOptions,
			)

			if len(err) > 0 {
				errSlice = append(errSlice, err...)
				hasErr = true
			}
			if hasCancelled {
				progress.StopInterrupt("Stopped searching for artworks based on tag names on Pixiv!")
				progress.SnapshotTask()
				return errSlice
			}

			artworksToDl = append(artworksToDl, artworksSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
			progress.Increment()
		}
		progress.Stop(hasErr)
		progress.SnapshotTask()
	}

	if len(artworksToDl) > 0 && pixivDlOptions.CtxIsActive() {
		cancelled, err := httpfuncs.DownloadUrls(
			artworksToDl,
			&httpfuncs.DlOptions{
				Context:        pixivDlOptions.GetContext(),
				MaxConcurrency: constants.PIXIV_MAX_DOWNLOAD_CONCURRENCY,
				Headers:        pixivcommon.GetPixivRequestHeaders(),
				UseHttp3:       false,
				SupportRange:   constants.PIXIV_RANGE_SUPPORTED,
				ProgressBarInfo: &progress.ProgressBarInfo{
					MainProgressBar:      pixivDlOptions.MainProgBar,
					DownloadProgressBars: pixivDlOptions.DownloadProgressBars,
				},
			},
			pixivDlOptions.Configs,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
		if cancelled {
			return errSlice
		}
	}
	if len(ugoiraToDl) > 0 && pixivDlOptions.CtxIsActive() {
		err := ugoira.DownloadMultipleUgoira(
			&ugoira.UgoiraArgs{
				UseMobileApi: true,
				ToDownload:   ugoiraToDl,
				Cookies:      nil,
			},
			pixivUgoiraOptions,
			pixivDlOptions.Configs,
			pixivDlOptions.MobileClient.SendRequest,
			&progress.ProgressBarInfo{
				MainProgressBar:      pixivDlOptions.MainProgBar,
				DownloadProgressBars: pixivDlOptions.DownloadProgressBars,
			},
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
	}

	alertUser(artworksToDl, ugoiraToDl, pixivDlOptions.Notifier)
	return errSlice
}
