package pixiv

import (
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/mobile"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/models"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/web"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notifier"
	"github.com/KJHJason/Cultured-Downloader-Logic/spinner"
)

func alertUser(artworksToDl []*httpfuncs.ToDownload, ugoiraToDl []*models.Ugoira, notifTitle string, app fyne.App) {
	if len(artworksToDl) > 0 || len(ugoiraToDl) > 0 {
		notifier.AlertWithoutErr(notifTitle, "Finished downloading artworks from Pixiv!", app)
	} else {
		notifier.AlertWithoutErr(notifTitle, "No artworks to download from Pixiv!", app)
	}
}

// Start the download process for Pixiv
func PixivWebDownloadProcess(pixivDl *PixivDl, pixivDlOptions *pixivweb.PixivWebDlOptions, pixivUgoiraOptions *ugoira.UgoiraOptions, notifTitle string, app fyne.App) {
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

	if len(pixivDl.TagNames) > 0 {
		// loop through each tag and page number
		baseMsg := "Searching for artworks based on tag names on Pixiv [%d/" + fmt.Sprintf("%d]...", len(pixivDl.TagNames))
		progress := spinner.New(
			"pong",
			"fgHiYellow",
			fmt.Sprintf(
				baseMsg,
				0,
			),
			fmt.Sprintf(
				"Finished searching for artworks based on %d tag names on Pixiv!",
				len(pixivDl.TagNames),
			),
			fmt.Sprintf(
				"Finished with some errors while searching for artworks based on %d tag names on Pixiv!\nPlease refer to the logs for more details...",
				len(pixivDl.TagNames),
			),
			len(pixivDl.TagNames),
		)
		progress.Start()
		hasErr := false
		for idx, tagName := range pixivDl.TagNames {
			var artworksSlice []*httpfuncs.ToDownload
			var ugoiraSlice []*models.Ugoira
			artworksSlice, ugoiraSlice, hasErr = pixivweb.TagSearch(
				tagName,
				iofuncs.DOWNLOAD_PATH,
				pixivDl.TagNamesPageNums[idx],
				pixivDlOptions,
			)
			artworksToDl = append(artworksToDl, artworksSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
			progress.MsgIncrement(baseMsg)
		}
		progress.Stop(hasErr)
	}

	if len(artworksToDl) > 0 {
		httpfuncs.DownloadUrls(
			artworksToDl,
			&httpfuncs.DlOptions{
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
				UseMobileApi: false,
				ToDownload:   ugoiraToDl,
				Cookies:      pixivDlOptions.SessionCookies,
			},
			pixivUgoiraOptions,
			pixivDlOptions.Configs,
			httpfuncs.CallRequest,
		)
	}

	alertUser(artworksToDl, ugoiraToDl, notifTitle, app)
}

// Start the download process for Pixiv
func PixivMobileDownloadProcess(pixivDl *PixivDl, pixivDlOptions *pixivmobile.PixivMobileDlOptions, pixivUgoiraOptions *ugoira.UgoiraOptions, notifTitle string, app fyne.App) {
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

	if len(pixivDl.TagNames) > 0 {
		// loop through each tag and page number
		baseMsg := "Searching for artworks based on tag names on Pixiv [%d/" + fmt.Sprintf("%d]...", len(pixivDl.TagNames))
		progress := spinner.New(
			"pong",
			"fgHiYellow",
			fmt.Sprintf(
				baseMsg,
				0,
			),
			fmt.Sprintf(
				"Finished searching for artworks based on %d tag names on Pixiv!",
				len(pixivDl.TagNames),
			),
			fmt.Sprintf(
				"Finished with some errors while searching for artworks based on %d tag names on Pixiv!\nPlease refer to the logs for more details...",
				len(pixivDl.TagNames),
			),
			len(pixivDl.TagNames),
		)
		progress.Start()
		hasErr := false
		for idx, tagName := range pixivDl.TagNames {
			var artworksSlice []*httpfuncs.ToDownload
			var ugoiraSlice []*models.Ugoira
			artworksSlice, ugoiraSlice, hasErr = pixivDlOptions.MobileClient.TagSearch(
				tagName,
				iofuncs.DOWNLOAD_PATH,
				pixivDl.TagNamesPageNums[idx],
				pixivDlOptions,
			)
			artworksToDl = append(artworksToDl, artworksSlice...)
			ugoiraToDl = append(ugoiraToDl, ugoiraSlice...)
			progress.MsgIncrement(baseMsg)
		}
		progress.Stop(hasErr)
	}

	if len(artworksToDl) > 0 {
		httpfuncs.DownloadUrls(
			artworksToDl,
			&httpfuncs.DlOptions{
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

	alertUser(artworksToDl, ugoiraToDl, notifTitle, app)
}
