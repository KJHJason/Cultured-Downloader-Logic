package pixivmobile

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/models"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

type offsetArgs struct {
	minOffset int
	maxOffset int
	hasMax    bool
}

// Returns the Ugoira structure with the necessary information to download the ugoira
//
// Will return an error which has been logged if unexpected error occurs like connection error, json marshal error, etc.
func (pixiv *PixivMobile) getUgoiraMetadata(illustId, dlFilePath string) (*models.Ugoira, error) {
	ugoiraUrl := pixiv.baseUrl + "/v1/ugoira/metadata"
	params := map[string]string{"illust_id": illustId}
	additionalHeaders := pixiv.getHeaders(
		map[string]string{"Referer": pixiv.baseUrl},
	)

	res, err := pixiv.SendRequest(
		&httpfuncs.RequestArgs{
			Url:         ugoiraUrl,
			CheckStatus: true,
			Headers:     additionalHeaders,
			Params:      params,
			Context:     pixiv.ctx,
		},
	)
	if err != nil {
		if err != context.Canceled {
			err = fmt.Errorf(
				"pixiv movile error %d: Failed to get ugoira metadata for %s",
				constants.CONNECTION_ERROR,
				illustId,
			)
		}
		return nil, err 
	}

	var ugoiraJson models.UgoiraJson
	if err := httpfuncs.LoadJsonFromResponse(res, &ugoiraJson); err != nil {
		return nil, err
	}

	ugoiraMetadata := ugoiraJson.Metadata
	ugoiraDlUrl := ugoiraMetadata.ZipUrls.Medium
	ugoiraDlUrl = strings.Replace(ugoiraDlUrl, "600x600", "1920x1080", 1)

	// map the files to their delay
	frameInfoMap := ugoira.MapDelaysToFilename(ugoiraMetadata.Frames)
	return &models.Ugoira{
		Url:      ugoiraDlUrl,
		Frames:   frameInfoMap,
		FilePath: dlFilePath,
	}, nil
}

// Query Pixiv's API (mobile) to get the JSON of an artwork ID
func (pixiv *PixivMobile) getArtworkDetails(artworkId, downloadPath string) ([]*httpfuncs.ToDownload, *models.Ugoira, error) {
	artworkUrl := pixiv.baseUrl + "/v1/illust/detail"
	params := map[string]string{"illust_id": artworkId}

	res, err := pixiv.SendRequest(
		&httpfuncs.RequestArgs{
			Context:     pixiv.ctx,
			Url:         artworkUrl,
			Params:      params,
			CheckStatus: true,
		},
	)
	if err != nil {
		if err != context.Canceled {
			err = fmt.Errorf(
				"pixiv mobile error %d: failed to get artwork details for %s, more info => %v",
				constants.CONNECTION_ERROR,
				artworkId,
				err,
			)
		}
		return nil, nil, err
	}

	var artworkJson models.PixivMobileArtworkJson
	if err := httpfuncs.LoadJsonFromResponse(res, &artworkJson); err != nil {
		return nil, nil, err
	}

	artworkDetails, ugoiraToDl, err := pixiv.processArtworkJson(
		artworkJson.Illust,
		downloadPath,
	)
	return artworkDetails, ugoiraToDl, err
}

func (pixiv *PixivMobile) GetMultipleArtworkDetails(artworkIds []string, downloadPath string) ([]*httpfuncs.ToDownload, []*models.Ugoira) {
	var artworksToDownload []*httpfuncs.ToDownload
	var ugoiraSlice []*models.Ugoira
	artworkIdsLen := len(artworkIds)
	lastIdx := artworkIdsLen - 1

	var errSlice []error
	baseMsg := "Getting and processing artwork details from Pixiv's Mobile API [%d/" + fmt.Sprintf("%d]...", artworkIdsLen)
	progress := pixiv.ArtworkProgress
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting and processing %d artwork details from Pixiv's Mobile API!",
			artworkIdsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting and processing %d artwork details from Pixiv's Mobile API!\nPlease refer to the logs for more details.",
			artworkIdsLen,
		),
	)
	progress.UpdateMax(artworkIdsLen)
	progress.Start()
	for idx, artworkId := range artworkIds {
		artworkDetails, ugoiraInfo, err := pixiv.getArtworkDetails(artworkId, downloadPath)
		if err != nil {
			errSlice = append(errSlice, err)
			if err == context.Canceled {
				break
			}
			progress.Increment()
			continue
		}

		if ugoiraInfo != nil {
			ugoiraSlice = append(ugoiraSlice, ugoiraInfo)
		} else {
			artworksToDownload = append(artworksToDownload, artworkDetails...)
		}

		if idx != lastIdx {
			pixiv.Sleep()
		}
		progress.Increment()
	}

	hasErr := false
	if len(errSlice) > 0 {
		hasErr = true
		if hasCancelled := logger.LogErrors(false, logger.ERROR, errSlice...); hasCancelled {
			progress.StopInterrupt("Stopped getting and processing artwork details from Pixiv's Mobile API!")
			return nil, nil
		}
	}
	progress.Stop(hasErr)
	return artworksToDownload, ugoiraSlice
}

func (pixiv *PixivMobile) getIllustratorPostMainLogic(params map[string]string, userId, downloadPath string, offsetArg *offsetArgs) ([]*httpfuncs.ToDownload, []*models.Ugoira, []error) {
	var errSlice []error
	var ugoiraSlice []*models.Ugoira
	var artworksToDownload []*httpfuncs.ToDownload
	nextUrl := pixiv.baseUrl + "/v1/user/illusts"

	curOffset := offsetArg.minOffset
	for nextUrl != "" {
		res, err := pixiv.SendRequest(
			&httpfuncs.RequestArgs{
				Context:     pixiv.ctx,
				Url:         nextUrl,
				Params:      params,
				CheckStatus: true,
			},
		)
		if err != nil {
			if err == context.Canceled {
				errSlice = append(errSlice, err)
				return nil, nil, errSlice
			}
			err = fmt.Errorf(
				"pixiv mobile error %d: failed to get illustrator posts for %s, more info => %v",
				constants.CONNECTION_ERROR,
				userId,
				err,
			)
			errSlice = append(errSlice, err)
			return nil, nil, errSlice
		}

		var resJson models.PixivMobileArtworksJson
		if err := httpfuncs.LoadJsonFromResponse(res, &resJson); err != nil {
			errSlice = append(errSlice, err)
			return nil, nil, errSlice
		}

		artworks, ugoira, errS := pixiv.processMultipleArtworkJson(&resJson, downloadPath)
		if len(errS) > 0 {
			errSlice = append(errSlice, errS...)
		}
		artworksToDownload = append(artworksToDownload, artworks...)
		ugoiraSlice = append(ugoiraSlice, ugoira...)

		curOffset += 30
		params["offset"] = strconv.Itoa(curOffset)
		jsonNextUrl := resJson.NextUrl
		if jsonNextUrl == nil || (offsetArg.hasMax && curOffset >= offsetArg.maxOffset) {
			nextUrl = ""
		} else {
			nextUrl = *jsonNextUrl
			pixiv.Sleep()
		}
	}
	return artworksToDownload, ugoiraSlice, errSlice
}

// Query Pixiv's API (mobile) to get all the posts JSON(s) of a user ID
func (pixiv *PixivMobile) getIllustratorPosts(userId, pageNum, downloadPath, artworkType string) ([]*httpfuncs.ToDownload, []*models.Ugoira, []error) {
	minPage, maxPage, hasMax, err := api.GetMinMaxFromStr(pageNum)
	if err != nil {
		return nil, nil, []error{err}
	}
	minOffset, maxOffset := pixivcommon.ConvertPageNumToOffset(minPage, maxPage, constants.PIXIV_PER_PAGE, false)

	params := map[string]string{
		"user_id": userId,
		"filter":  "for_ios",
		"offset":  strconv.Itoa(minOffset),
		"type":    artworkType,
	}
	if artworkType == "all" {
		params["type"] = "illust"
	}

	offsetArgs := &offsetArgs{
		minOffset: minOffset,
		maxOffset: maxOffset,
		hasMax:    hasMax,
	}
	artworksToDl, ugoiraSlice, errSlice := pixiv.getIllustratorPostMainLogic(
		params,
		userId,
		downloadPath,
		offsetArgs,
	)

	if params["type"] == "illust" && artworkType == "all" {
		// if the user is downloading both
		// illust and manga, loop again to get the manga
		params["type"] = "manga"
		artworksToDl2, ugoiraSlice2, errSlice2 := pixiv.getIllustratorPostMainLogic(
			params,
			userId,
			downloadPath,
			offsetArgs,
		)
		artworksToDl = append(artworksToDl, artworksToDl2...)
		ugoiraSlice = append(ugoiraSlice, ugoiraSlice2...)
		errSlice = append(errSlice, errSlice2...)
	}
	return artworksToDl, ugoiraSlice, errSlice
}

func (pixiv *PixivMobile) GetMultipleIllustratorPosts(userIds, pageNums []string, downloadPath, artworkType string) ([]*httpfuncs.ToDownload, []*models.Ugoira) {
	userIdsLen := len(userIds)
	lastIdx := userIdsLen - 1

	var errSlice []error
	var ugoiraSlice []*models.Ugoira
	var artworksToDownload []*httpfuncs.ToDownload
	baseMsg := "Getting artwork details from illustrator(s) on Pixiv [%d/" + fmt.Sprintf("%d]...", userIdsLen)
	progress := pixiv.IllustratorProgress
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting artwork details from %d illustrator(s) on Pixiv!",
			userIdsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting artwork details from %d illustrator(s) on Pixiv!\nPlease refer to the logs for more details.",
			userIdsLen,
		),
	)
	progress.UpdateMax(userIdsLen)
	progress.Start()
	for idx, userId := range userIds {
		artworkDetails, ugoiraInfo, err := pixiv.getIllustratorPosts(
			userId,
			pageNums[idx],
			downloadPath,
			artworkType,
		)
		if err != nil {
			errSlice = append(errSlice, err...)
			progress.Increment()
			continue
		}

		artworksToDownload = append(artworksToDownload, artworkDetails...)
		ugoiraSlice = append(ugoiraSlice, ugoiraInfo...)
		if idx != lastIdx {
			pixiv.Sleep()
		}
		progress.Increment()
	}

	hasErr := false
	if len(errSlice) > 0 {
		hasErr = true
		if hasCancelled := logger.LogErrors(false, logger.ERROR, errSlice...); hasCancelled {
			progress.StopInterrupt("Stopped getting artwork details from illustrator(s) on Pixiv!")
			return nil, nil
		}
	}
	progress.Stop(hasErr)
	return artworksToDownload, ugoiraSlice
}

func (pixiv *PixivMobile) tagSearchLogic(tagName, downloadPath string, dlOptions *PixivMobileDlOptions, offsetArg *offsetArgs) ([]*httpfuncs.ToDownload, []*models.Ugoira, []error) {
	var errSlice []error
	var ugoiraSlice []*models.Ugoira
	var artworksToDownload []*httpfuncs.ToDownload
	params := map[string]string{
		"word":          tagName,
		"search_target": dlOptions.SearchMode,
		"sort":          dlOptions.SortOrder,
		"filter":        "for_ios",
		"offset":        strconv.Itoa(offsetArg.minOffset),
	}
	curOffset := offsetArg.minOffset
	nextUrl := pixiv.baseUrl + "/v1/search/illust"
	for nextUrl != "" {
		res, err := pixiv.SendRequest(
			&httpfuncs.RequestArgs{
				Url:         nextUrl,
				Params:      params,
				CheckStatus: true,
				Context:     dlOptions.GetContext(),
			},
		)
		if err != nil {
			if err == context.Canceled {
				errSlice = append(errSlice, err)
				return nil, nil, errSlice
			}

			err = fmt.Errorf(
				"pixiv mobile error %d: failed to search for %q, more info => %v",
				constants.CONNECTION_ERROR,
				tagName,
				err,
			)
			errSlice = append(errSlice, err)
			return nil, nil, errSlice
		}

		var resJson models.PixivMobileArtworksJson
		if err := httpfuncs.LoadJsonFromResponse(res, &resJson); err != nil {
			errSlice = append(errSlice, err)
			continue
		}

		artworks, ugoira, errS := pixiv.processMultipleArtworkJson(&resJson, downloadPath)
		errSlice = append(errSlice, errS...)
		artworksToDownload = append(artworksToDownload, artworks...)
		ugoiraSlice = append(ugoiraSlice, ugoira...)

		curOffset += 30
		params["offset"] = strconv.Itoa(curOffset)
		jsonNextUrl := resJson.NextUrl
		if jsonNextUrl == nil || (offsetArg.hasMax && curOffset >= offsetArg.maxOffset) {
			nextUrl = ""
		} else {
			nextUrl = *jsonNextUrl
			pixiv.Sleep()
		}
	}
	return artworksToDownload, ugoiraSlice, errSlice
}

// Query Pixiv's API (mobile) to get the JSON of a search query
// Returns the ToDownload slice, Ugoira slice, boolean indicating if there was an error, and boolean indicating if the context was cancelled
func (pixiv *PixivMobile) TagSearch(tagName, downloadPath, pageNum string, dlOptions *PixivMobileDlOptions) ([]*httpfuncs.ToDownload, []*models.Ugoira, bool, bool) {
	minPage, maxPage, hasMax, err := api.GetMinMaxFromStr(pageNum)
	if err != nil {
		logger.LogError(err, false, logger.ERROR)
		return nil, nil, true, false
	}
	minOffset, maxOffset := pixivcommon.ConvertPageNumToOffset(minPage, maxPage, constants.PIXIV_PER_PAGE, false)

	artworksToDl, ugoiraSlice, errSlice := pixiv.tagSearchLogic(
		tagName,
		downloadPath,
		dlOptions,
		&offsetArgs{
			minOffset: minOffset,
			maxOffset: maxOffset,
			hasMax:    hasMax,
		},
	)
	if len(errSlice) > 0 {
		if hasCancelled := logger.LogErrors(false, logger.ERROR, errSlice...); hasCancelled {
			return nil, nil, false, true
		}
	}
	return artworksToDl, ugoiraSlice, len(errSlice) > 0, false
}
