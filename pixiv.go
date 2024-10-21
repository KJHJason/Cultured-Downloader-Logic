package cdlogic

import (
	"fmt"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv"
	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	pixivmobile "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/mobile"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
	pixivweb "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/web"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

func alertUser(artworksToDl []*httpfuncs.ToDownload, ugoiraToDl []*ugoira.Ugoira, notifier notify.Notifier) {
	if len(artworksToDl) > 0 || len(ugoiraToDl) > 0 {
		notifier.Alert("Finished downloading artworks from Pixiv!")
	} else {
		notifier.Alert("No artworks to download from Pixiv!")
	}
}

// Start the download process for Pixiv
func PixivWebDownloadProcess(pixivDl *pixiv.PixivDl, pixivDlOptions *pixivweb.PixivWebDlOptions, pixivUgoiraOptions *ugoira.UgoiraOptions, catchInterrupt bool) []error {
	defer pixivDlOptions.CancelCtx()
	var errSlice []error
	var ugoiraToDl []*ugoira.Ugoira
	var artworksToDl []*httpfuncs.ToDownload

	if catchInterrupt {
		stopSignal := catchInterruptSignal(pixivDlOptions.CancelCtx)
		defer stopSignal()
	}

	tagNameLen := len(pixivDl.TagNames)
	if tagNameLen > 0 && pixivDlOptions.CtxIsActive() {
		// loop through each tag and page number
		baseMsg := "Searching for artworks based on tag names on Pixiv [%d/" + fmt.Sprintf("%d]...", tagNameLen)
		prog := pixivDlOptions.Base.MainProgBar()
		prog.UpdateBaseMsg(baseMsg)
		prog.UpdateSuccessMsg(
			fmt.Sprintf(
				"Finished searching for artworks based on %d tag names on Pixiv!",
				tagNameLen,
			),
		)
		prog.UpdateErrorMsg(
			fmt.Sprintf(
				"Finished with some errors while searching for artworks based on %d tag names on Pixiv!\nPlease refer to the logs for more details...",
				tagNameLen,
			),
		)
		prog.SetToProgressBar()
		prog.UpdateMax(tagNameLen)
		prog.Start()
		hasErr := false
		for idx, tagName := range pixivDl.TagNames {
			artworkIds, err, hasCancelled := pixivweb.TagSearch(
				tagName,
				pixivDl.TagNamesPageNums[idx],
				pixivDlOptions,
			)

			if len(err) > 0 {
				errSlice = append(errSlice, err...)
				hasErr = true
			}
			if hasCancelled {
				prog.StopInterrupt("Stopped searching for artworks based on tag names on Pixiv!")
				prog.SnapshotTask()
				return errSlice
			}

			pixivDl.ArtworkIds = append(pixivDl.ArtworkIds, artworkIds...)
			prog.Increment()
		}
		prog.Stop(hasErr)
		prog.SnapshotTask()
	}

	if len(pixivDl.ArtistIds) > 0 {
		artworkIdsSlice, err := pixivweb.GetMultipleArtistsPosts(
			pixivDl.ArtistIds,
			pixivDl.ArtistPageNums,
			pixivDlOptions,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		} else {
			pixivDl.ArtworkIds = append(pixivDl.ArtworkIds, artworkIdsSlice...)
		}
	}

	if len(pixivDl.ArtworkIds) > 0 && pixivDlOptions.CtxIsActive() {
		pixivDl.ArtworkIds = utils.RemoveDuplicatesFromSlice(pixivDl.ArtworkIds)
		artworkSlice, ugoiraSlice, err := pixivweb.GetMultipleArtworkDetails(
			pixivDl.ArtworkIds,
			pixivDlOptions,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		} else {
			artworksToDl = append(artworksToDl, artworkSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
		}
	}

	captchaHandler := pixivDlOptions.GetCaptchaHandler()
	if len(artworksToDl) > 0 && pixivDlOptions.CtxIsActive() {
		httpfuncs.DownloadUrls(
			artworksToDl,
			&httpfuncs.DlOptions{
				Context:         pixivDlOptions.GetContext(),
				MaxConcurrency:  constants.PIXIV_MAX_DOWNLOAD_CONCURRENCY,
				Headers:         pixivcommon.GetPixivRequestHeaders(),
				Cookies:         pixivDlOptions.Base.SessionCookies,
				UseHttp3:        false,
				HeadReqTimeout:  constants.DEFAULT_HEAD_REQ_TIMEOUT,
				SupportRange:    constants.PIXIV_RANGE_SUPPORTED,
				Filters:         pixivDlOptions.Base.Filters,
				ProgressBarInfo: pixivDlOptions.Base.ProgressBarInfo,
				CaptchaHandler:  captchaHandler,
			},
			pixivDlOptions.Base.Configs,
		)
	}
	if len(ugoiraToDl) > 0 && pixivDlOptions.CtxIsActive() {
		ugoiraArgs := &ugoira.UgoiraArgs{
			UseMobileApi:   false,
			ToDownload:     ugoiraToDl,
			Filters:        pixivDlOptions.Base.Filters,
			CaptchaHandler: captchaHandler,
			Cookies:        pixivDlOptions.Base.SessionCookies,
			MainProgBar:    pixivDlOptions.Base.MainProgBar(),
		}
		ugoiraArgs.SetContext(pixivDlOptions.GetContext())
		err := ugoira.DownloadMultipleUgoira(
			ugoiraArgs,
			pixivUgoiraOptions,
			pixivDlOptions.Base.Configs,
			httpfuncs.CallRequest,
			pixivDlOptions.Base.ProgressBarInfo,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
	}

	alertUser(artworksToDl, ugoiraToDl, pixivDlOptions.Base.Notifier)
	return errSlice
}

// Start the download process for Pixiv
func PixivMobileDownloadProcess(pixivDl *pixiv.PixivDl, pixivMobile *pixivmobile.PixivMobile, pixivUgoiraOptions *ugoira.UgoiraOptions, catchInterrupt bool) []error {
	defer pixivMobile.CancelCtx()
	var errSlice []error
	var ugoiraToDl []*ugoira.Ugoira
	var artworksToDl []*httpfuncs.ToDownload

	if catchInterrupt {
		stopSignal := catchInterruptSignal(pixivMobile.CancelCtx)
		defer stopSignal()
	}

	if len(pixivDl.ArtistIds) > 0 {
		artworkSlice, ugoiraSlice, err := pixivMobile.GetMultipleArtistsPosts(
			pixivDl.ArtistIds,
			pixivDl.ArtistPageNums,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		} else {
			artworksToDl = artworkSlice
			ugoiraToDl = ugoiraSlice
		}
	}

	if len(pixivDl.ArtworkIds) > 0 && pixivMobile.CtxIsActive() {
		artworkSlice, ugoiraSlice, err := pixivMobile.GetMultipleArtworkDetails(pixivDl.ArtworkIds)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		} else {
			artworksToDl = append(artworksToDl, artworkSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
		}
	}

	tagNamesLen := len(pixivDl.TagNames)
	if tagNamesLen > 0 && pixivMobile.CtxIsActive() {
		// loop through each tag and page number
		baseMsg := "Searching for artworks based on tag names on Pixiv [%d/" + fmt.Sprintf("%d]...", tagNamesLen)
		prog := pixivMobile.Base.MainProgBar()
		prog.UpdateBaseMsg(baseMsg)
		prog.UpdateSuccessMsg(
			fmt.Sprintf(
				"Finished searching for artworks based on %d tag names on Pixiv!",
				tagNamesLen,
			),
		)
		prog.UpdateErrorMsg(
			fmt.Sprintf(
				"Finished with some errors while searching for artworks based on %d tag names on Pixiv!\nPlease refer to the logs for more details...",
				tagNamesLen,
			),
		)
		prog.SetToProgressBar()
		prog.UpdateMax(tagNamesLen)
		prog.Start()
		hasErr := false
		for idx, tagName := range pixivDl.TagNames {
			var artworksSlice []*httpfuncs.ToDownload
			var ugoiraSlice []*ugoira.Ugoira
			artworksSlice, ugoiraSlice, err, hasCancelled := pixivMobile.TagSearch(
				tagName,
				pixivDl.TagNamesPageNums[idx],
			)

			if len(err) > 0 {
				errSlice = append(errSlice, err...)
				hasErr = true
			}
			if hasCancelled {
				prog.StopInterrupt("Stopped searching for artworks based on tag names on Pixiv!")
				prog.SnapshotTask()
				return errSlice
			}

			artworksToDl = append(artworksToDl, artworksSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
			prog.Increment()
		}
		prog.Stop(hasErr)
		prog.SnapshotTask()
	}

	captchaHandler := pixivMobile.GetCaptchaHandler()
	if len(artworksToDl) > 0 && pixivMobile.CtxIsActive() {
		cancelled, err := httpfuncs.DownloadUrlsWithHandler(
			artworksToDl,
			&httpfuncs.DlOptions{
				Context:         pixivMobile.GetContext(),
				MaxConcurrency:  constants.PIXIV_MAX_DOWNLOAD_CONCURRENCY,
				Headers:         pixivcommon.GetPixivRequestHeaders(),
				UseHttp3:        false,
				HeadReqTimeout:  constants.DEFAULT_HEAD_REQ_TIMEOUT,
				SupportRange:    constants.PIXIV_RANGE_SUPPORTED,
				Filters:         pixivMobile.Base.Filters,
				ProgressBarInfo: pixivMobile.Base.ProgressBarInfo,
				CaptchaHandler:  captchaHandler,
			},
			pixivMobile.Base.Configs,
			pixivMobile.SendRequest,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
		if cancelled {
			return errSlice
		}
	}
	if len(ugoiraToDl) > 0 && pixivMobile.CtxIsActive() {
		ugoiraArgs := &ugoira.UgoiraArgs{
			UseMobileApi:   true,
			ToDownload:     ugoiraToDl,
			Filters:        pixivMobile.Base.Filters,
			CaptchaHandler: captchaHandler,
			Cookies:        nil,
			MainProgBar:    pixivMobile.Base.MainProgBar(),
		}
		ugoiraArgs.SetContext(pixivMobile.GetContext())
		err := ugoira.DownloadMultipleUgoira(
			ugoiraArgs,
			pixivUgoiraOptions,
			pixivMobile.Base.Configs,
			pixivMobile.SendRequest,
			pixivMobile.Base.ProgressBarInfo,
		)
		if len(err) > 0 {
			errSlice = append(errSlice, err...)
		}
	}

	alertUser(artworksToDl, ugoiraToDl, pixivMobile.Base.Notifier)
	return errSlice
}
