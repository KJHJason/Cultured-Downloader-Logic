package cdlogic

import (
	"fmt"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv"
	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	pixivmobile "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/mobile"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/models"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
	pixivweb "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/web"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
)

func alertUser(artworksToDl []*httpfuncs.ToDownload, ugoiraToDl []*models.Ugoira, notifier notify.Notifier) {
	if len(artworksToDl) > 0 || len(ugoiraToDl) > 0 {
		notifier.Alert("Finished downloading artworks from Pixiv!")
	} else {
		notifier.Alert("No artworks to download from Pixiv!")
	}
}

// Start the download process for Pixiv
func PixivWebDownloadProcess(pixivDl *pixiv.PixivDl, pixivDlOptions *pixivweb.PixivWebDlOptions, pixivUgoiraOptions *ugoira.UgoiraOptions) {
	var ugoiraToDl []*models.Ugoira
	var artworksToDl []*httpfuncs.ToDownload
	if len(pixivDl.IllustratorIds) > 0 {
		artworkIdsSlice := pixivweb.GetMultipleIllustratorPosts(
			pixivDl.IllustratorIds,
			pixivDl.IllustratorPageNums,
			iofuncs.DOWNLOAD_PATH,
			pixivDlOptions,
		)
		pixivDl.ArtworkIds = append(pixivDl.ArtworkIds, artworkIdsSlice...)
		pixivDl.ArtworkIds = api.RemoveSliceDuplicates(pixivDl.ArtworkIds)
	}

	if len(pixivDl.ArtworkIds) > 0 {
		artworkSlice, ugoiraSlice := pixivweb.GetMultipleArtworkDetails(
			pixivDl.ArtworkIds,
			iofuncs.DOWNLOAD_PATH,
			pixivDlOptions,
		)
		artworksToDl = append(artworksToDl, artworkSlice...)
		ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
	}

	tagNameLen := len(pixivDl.TagNames)
	if tagNameLen > 0 {
		// loop through each tag and page number
		baseMsg := "Searching for artworks based on tag names on Pixiv [%d/" + fmt.Sprintf("%d]...", tagNameLen)
		progress := pixivDlOptions.TagSearchProgBar
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
		progress.UpdateMax(tagNameLen)
		progress.Start()
		hasErr, hasCancelled := false, false
		for idx, tagName := range pixivDl.TagNames {
			var artworksSlice []*httpfuncs.ToDownload
			var ugoiraSlice []*models.Ugoira
			artworksSlice, ugoiraSlice, hasErr, hasCancelled = pixivweb.TagSearch(
				tagName,
				iofuncs.DOWNLOAD_PATH,
				pixivDl.TagNamesPageNums[idx],
				pixivDlOptions,
			)
			if hasCancelled {
				progress.StopInterrupt("Stopped searching for artworks based on tag names on Pixiv!")
				return
			}
			artworksToDl = append(artworksToDl, artworksSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
			progress.Increment()
		}
		progress.Stop(hasErr)
	}

	if len(artworksToDl) > 0 {
		httpfuncs.DownloadUrls(
			artworksToDl,
			&httpfuncs.DlOptions{
				Context:        pixivDlOptions.GetContext(),
				MaxConcurrency: constants.PIXIV_MAX_CONCURRENT_DOWNLOADS,
				Headers:        pixivcommon.GetPixivRequestHeaders(),
				Cookies:        pixivDlOptions.SessionCookies,
				UseHttp3:       false,
			},
			pixivDlOptions.Configs,
		)
	}
	if len(ugoiraToDl) > 0 {
		ugoira.DownloadMultipleUgoira(
			&ugoira.UgoiraArgs{
				Context:      pixivDlOptions.GetContext(),
				UseMobileApi: false,
				ToDownload:   ugoiraToDl,
				Cookies:      pixivDlOptions.SessionCookies,
			},
			pixivUgoiraOptions,
			pixivDlOptions.Configs,
			httpfuncs.CallRequest,
		)
	}

	alertUser(artworksToDl, ugoiraToDl, pixivDlOptions.Notifier)
}

// Start the download process for Pixiv
func PixivMobileDownloadProcess(pixivDl *pixiv.PixivDl, pixivDlOptions *pixivmobile.PixivMobileDlOptions, pixivUgoiraOptions *ugoira.UgoiraOptions, notifTitle string) {
	var ugoiraToDl []*models.Ugoira
	var artworksToDl []*httpfuncs.ToDownload
	if len(pixivDl.IllustratorIds) > 0 {
		artworkSlice, ugoiraSlice := pixivDlOptions.MobileClient.GetMultipleIllustratorPosts(
			pixivDl.IllustratorIds,
			pixivDl.IllustratorPageNums,
			iofuncs.DOWNLOAD_PATH,
			pixivDlOptions.ArtworkType,
		)
		artworksToDl = artworkSlice
		ugoiraToDl = ugoiraSlice
	}

	if len(pixivDl.ArtworkIds) > 0 {
		artworkSlice, ugoiraSlice := pixivDlOptions.MobileClient.GetMultipleArtworkDetails(
			pixivDl.ArtworkIds,
			iofuncs.DOWNLOAD_PATH,
		)
		artworksToDl = append(artworksToDl, artworkSlice...)
		ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
	}

	tagNamesLen := len(pixivDl.TagNames)
	if tagNamesLen > 0 {
		// loop through each tag and page number
		baseMsg := "Searching for artworks based on tag names on Pixiv [%d/" + fmt.Sprintf("%d]...", tagNamesLen)
		progress := pixivDlOptions.TagSearchProgBar
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
		progress.UpdateMax(tagNamesLen)
		progress.Start()
		hasErr, hasCancelled := false, false
		for idx, tagName := range pixivDl.TagNames {
			var artworksSlice []*httpfuncs.ToDownload
			var ugoiraSlice []*models.Ugoira
			artworksSlice, ugoiraSlice, hasErr, hasCancelled = pixivDlOptions.MobileClient.TagSearch(
				tagName,
				iofuncs.DOWNLOAD_PATH,
				pixivDl.TagNamesPageNums[idx],
				pixivDlOptions,
			)
			if hasCancelled {
				progress.StopInterrupt("Stopped searching for artworks based on tag names on Pixiv!")
				return
			}
			artworksToDl = append(artworksToDl, artworksSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
			progress.Increment()
		}
		progress.Stop(hasErr)
	}

	if len(artworksToDl) > 0 {
		httpfuncs.DownloadUrls(
			artworksToDl,
			&httpfuncs.DlOptions{
				Context:        pixivDlOptions.GetContext(),
				MaxConcurrency: constants.PIXIV_MAX_CONCURRENT_DOWNLOADS,
				Headers:        pixivcommon.GetPixivRequestHeaders(),
				UseHttp3:       false,
			},
			pixivDlOptions.Configs,
		)
	}
	if len(ugoiraToDl) > 0 {
		ugoira.DownloadMultipleUgoira(
			&ugoira.UgoiraArgs{
				UseMobileApi: true,
				ToDownload:   ugoiraToDl,
				Cookies:      nil,
			},
			pixivUgoiraOptions,
			pixivDlOptions.Configs,
			pixivDlOptions.MobileClient.SendRequest,
		)
	}

	alertUser(artworksToDl, ugoiraToDl, pixivDlOptions.Notifier)
}
