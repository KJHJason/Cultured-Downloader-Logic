package pixivweb

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/models"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

func getArtworkDetailsLogic(artworkId string, reqArgs *httpfuncs.RequestArgs) (*models.ArtworkDetails, error) {
	artworkDetailsRes, err := httpfuncs.CallRequest(reqArgs)
	if err != nil {
		return nil, fmt.Errorf(
			"pixiv error %d: failed to get artwork details for ID %v from %s",
			constants.CONNECTION_ERROR,
			artworkId,
			reqArgs.Url,
		)
	}

	if artworkDetailsRes.StatusCode != 200 {
		artworkDetailsRes.Body.Close()
		return nil, fmt.Errorf(
			"pixiv error %d: failed to get details for artwork ID %s due to %s response from %s",
			constants.RESPONSE_ERROR,
			artworkId,
			artworkDetailsRes.Status,
			reqArgs.Url,
		)
	}

	var artworkDetailsJsonRes models.ArtworkDetails
	if err := httpfuncs.LoadJsonFromResponse(artworkDetailsRes, &artworkDetailsJsonRes); err != nil {
		return nil, fmt.Errorf(
			"%v\ndetails: failed to read response body for Pixiv artwork ID %s",
			err,
			artworkId,
		)
	}
	return &artworkDetailsJsonRes, nil
}

func getArtworkUrlsToDlLogic(artworkType int64, artworkId string, reqArgs *httpfuncs.RequestArgs) (*http.Response, error) {
	var url string
	switch artworkType {
	case ILLUST, MANGA: // illustration or manga
		url = fmt.Sprintf("%s/illust/%s/pages", constants.PIXIV_API_URL, artworkId)
	case UGOIRA: // ugoira
		url = fmt.Sprintf("%s/illust/%s/ugoira_meta", constants.PIXIV_API_URL, artworkId)
	default:
		return nil, fmt.Errorf(
			"pixiv error %d: unsupported artwork type %d for artwork ID %s",
			constants.JSON_ERROR,
			artworkType,
			artworkId,
		)
	}

	reqArgs.Url = url
	artworkUrlsRes, err := httpfuncs.CallRequest(reqArgs)
	if err != nil {
		return nil, fmt.Errorf(
			"pixiv error %d: failed to get artwork URLs for ID %s from %s due to %v",
			constants.CONNECTION_ERROR,
			artworkId,
			url,
			err,
		)
	}

	if artworkUrlsRes.StatusCode != 200 {
		artworkUrlsRes.Body.Close()
		return nil, fmt.Errorf(
			"pixiv error %d: failed to get artwork URLs for ID %s due to %s response from %s",
			constants.RESPONSE_ERROR,
			artworkId,
			artworkUrlsRes.Status,
			url,
		)
	}
	return artworkUrlsRes, nil
}

// Retrieves details of an artwork ID and returns
// the folder path to download the artwork to, the JSON response, and the artwork type
func getArtworkDetails(artworkId, downloadPath string, dlOptions *PixivWebDlOptions) ([]*httpfuncs.ToDownload, *models.Ugoira, error) {
	if artworkId == "" {
		return nil, nil, nil
	}

	url := fmt.Sprintf("%s/illust/%s", constants.PIXIV_API_URL, artworkId)
	headers := pixivcommon.GetPixivRequestHeaders()
	headers["Referer"] = pixivcommon.GetUserUrl(artworkId)

	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV, true)
	reqArgs := &httpfuncs.RequestArgs{
		Url:       url,
		Method:    "GET",
		Cookies:   dlOptions.SessionCookies,
		Headers:   headers,
		UserAgent: dlOptions.Configs.UserAgent,
		Http2:     !useHttp3,
		Http3:     useHttp3,
		Context:   dlOptions.GetContext(),
	}
	artworkDetailsJsonRes, err := getArtworkDetailsLogic(artworkId, reqArgs)
	if err != nil {
		return nil, nil, err
	}

	artworkJsonBody := artworkDetailsJsonRes.Body
	illustratorName := artworkJsonBody.UserName
	artworkName := artworkJsonBody.Title
	artworkPostDir := iofuncs.GetPostFolder(
		filepath.Join(downloadPath, constants.PIXIV_TITLE),
		illustratorName,
		artworkId,
		artworkName,
	)

	artworkType := artworkJsonBody.IllustType
	artworkUrlsRes, err := getArtworkUrlsToDlLogic(artworkType, artworkId, reqArgs)
	if err != nil {
		return nil, nil, err
	}

	urlsToDl, ugoiraInfo, err := processArtworkJson(
		artworkUrlsRes,
		artworkType,
		artworkPostDir,
	)
	if err != nil {
		return nil, nil, err
	}
	return urlsToDl, ugoiraInfo, nil
}

// Retrieves multiple artwork details based on the given slice of artwork IDs
// and returns a map to use for downloading and a slice of Ugoira structures
func GetMultipleArtworkDetails(artworkIds []string, downloadPath string, dlOptions *PixivWebDlOptions) ([]*httpfuncs.ToDownload, []*models.Ugoira) {
	var errSlice []error
	var ugoiraDetails []*models.Ugoira
	var artworkDetails []*httpfuncs.ToDownload
	artworkIdsLen := len(artworkIds)
	lastArtworkId := artworkIds[artworkIdsLen-1]

	baseMsg := "Getting and processing artwork details from Pixiv [%d/" + fmt.Sprintf("%d]...", artworkIdsLen)
	progress := dlOptions.GetPostsDetailProgBar
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting and processing %d artwork details from Pixiv!",
			artworkIdsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting and processing %d artwork details from Pixiv!\nPlease refer to the logs for more details.",
			artworkIdsLen,
		),
	)
	progress.UpdateMax(artworkIdsLen)
	progress.Start()
	for _, artworkId := range artworkIds {
		artworksToDl, ugoiraInfo, err := getArtworkDetails(
			artworkId,
			downloadPath,
			dlOptions,
		)
		if err != nil {
			errSlice = append(errSlice, err)
			progress.Increment()
			continue
		}

		if ugoiraInfo != nil {
			ugoiraDetails = append(ugoiraDetails, ugoiraInfo)
		} else {
			artworkDetails = append(artworkDetails, artworksToDl...)
		}

		progress.Increment()
		if artworkId != lastArtworkId {
			pixivSleep()
		}
	}

	hasErr := false
	if len(errSlice) > 0 {
		hasErr = true
		if hasCancelled := logger.LogErrors(false, logger.ERROR, errSlice...); hasCancelled {
			progress.StopInterrupt("Stopped getting and processing artwork details from Pixiv!")
			return nil, nil
		}
	}
	progress.Stop(hasErr)
	return artworkDetails, ugoiraDetails
}

// Query Pixiv's API for all the illustrator's posts
func getIllustratorPosts(illustratorId, pageNum string, dlOptions *PixivWebDlOptions) ([]string, error) {
	headers := pixivcommon.GetPixivRequestHeaders()
	headers["Referer"] = pixivcommon.GetIllustUrl(illustratorId)
	url := fmt.Sprintf("%s/user/%s/profile/all", constants.PIXIV_API_URL, illustratorId)

	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV, true)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Url:       url,
			Method:    "GET",
			Cookies:   dlOptions.SessionCookies,
			Headers:   headers,
			UserAgent: dlOptions.Configs.UserAgent,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			Context:   dlOptions.GetContext(),
		},
	)
	if err != nil {
		if err == context.Canceled {
			return nil, err
		}
		return nil, fmt.Errorf(
			"pixiv error %d: failed to get illustrator's posts with an ID of %s due to %v",
			constants.CONNECTION_ERROR,
			illustratorId,
			err,
		)
	}
	if res.StatusCode != 200 {
		res.Body.Close()
		return nil, fmt.Errorf(
			"pixiv error %d: failed to get illustrator's posts with an ID of %s due to %s response",
			constants.RESPONSE_ERROR,
			illustratorId,
			res.Status,
		)
	}

	var jsonBody models.PixivWebIllustratorJson
	if err := httpfuncs.LoadJsonFromResponse(res, &jsonBody); err != nil {
		return nil, err
	}
	artworkIds, err := processIllustratorPostJson(&jsonBody, pageNum, dlOptions)
	return artworkIds, err
}

// Get posts from multiple illustrators and returns a slice of artwork IDs
func GetMultipleIllustratorPosts(illustratorIds, pageNums []string, downloadPath string, dlOptions *PixivWebDlOptions) []string {
	var errSlice []error
	var artworkIdsSlice []string
	illustratorIdsLen := len(illustratorIds)
	lastIllustratorIdx := illustratorIdsLen - 1

	baseMsg := "Getting artwork details from illustrator(s) on Pixiv [%d/" + fmt.Sprintf("%d]...", illustratorIdsLen)
	progress := dlOptions.GetIllustratorPostsProgBar
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting artwork details from %d illustrator(s) on Pixiv!",
			illustratorIdsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting artwork details from %d illustrator(s) on Pixiv!\nPlease refer to the logs for more details.",
			illustratorIdsLen,
		),
	)
	progress.UpdateMax(illustratorIdsLen)
	progress.Start()
	for idx, illustratorId := range illustratorIds {
		artworkIds, err := getIllustratorPosts(
			illustratorId,
			pageNums[idx],
			dlOptions,
		)
		if err != nil {
			errSlice = append(errSlice, err)
		} else {
			artworkIdsSlice = append(artworkIdsSlice, artworkIds...)
		}

		if idx != lastIllustratorIdx {
			pixivSleep()
		}
		progress.Increment()
	}

	hasErr := false
	if len(errSlice) > 0 {
		hasErr = true
		if hasCancelled := logger.LogErrors(false, logger.ERROR, errSlice...); hasCancelled {
			progress.StopInterrupt("Stopped getting artwork details from illustrator(s) on Pixiv!")
			return nil
		}
	}
	progress.Stop(hasErr)

	return artworkIdsSlice
}

type pageNumArgs struct {
	minPage int
	maxPage int
	hasMax  bool
}

func tagSearchLogic(tagName string, reqArgs *httpfuncs.RequestArgs, pageNumArgs *pageNumArgs) ([]string, []error) {
	var errSlice []error
	var artworkIds []string
	page := 0
	for {
		page++
		if page < pageNumArgs.minPage {
			continue
		}
		if pageNumArgs.hasMax && page > pageNumArgs.maxPage {
			break
		}

		reqArgs.Params["p"] = strconv.Itoa(page) // page number
		res, err := httpfuncs.CallRequest(reqArgs)
		if err != nil {
			if err == context.Canceled {
				errSlice = append(errSlice, err)
				return nil, errSlice
			}
			err = fmt.Errorf(
				"pixiv error %d: failed to get tag search results for %s due to %v",
				constants.CONNECTION_ERROR,
				tagName,
				err,
			)
			errSlice = append(errSlice, err)
			continue
		}

		tagArtworkIds, err := processTagJsonResults(res)
		if err != nil {
			errSlice = append(errSlice, err)
			continue
		}

		if len(tagArtworkIds) == 0 {
			break
		}

		artworkIds = append(artworkIds, tagArtworkIds...)
		if page != pageNumArgs.maxPage {
			pixivSleep()
		}
	}
	return artworkIds, errSlice
}

// Query Pixiv's API and search for posts based on the supplied tag name
// which will return a map and a slice of Ugoira structures for downloads
// Returns the map, the slice, a boolean indicating if there was an error, and a boolean indicating if the user cancelled the operation
func TagSearch(tagName, downloadPath, pageNum string, dlOptions *PixivWebDlOptions) ([]*httpfuncs.ToDownload, []*models.Ugoira, bool, bool) {
	minPage, maxPage, hasMax, err := api.GetMinMaxFromStr(pageNum)
	if err != nil {
		logger.LogError(err, false, logger.ERROR)
		return nil, nil, true, false
	}

	url := fmt.Sprintf("%s/search/artworks/%s", constants.PIXIV_API_URL, tagName)
	params := map[string]string{
		// search term
		"word": tagName,

		// search mode: s_tag, s_tag_full, s_tc
		"s_mode": dlOptions.SearchMode,

		// sort order: date, popular, popular_male, popular_female
		// (add "_d" suffix for descending order, e.g. date_d)
		"order": dlOptions.SortOrder,

		//  r18, safe, or all for both
		"mode": dlOptions.RatingMode,

		// illust_and_ugoira, manga, all
		"type": dlOptions.ArtworkType,
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV, true)
	headers := pixivcommon.GetPixivRequestHeaders()
	headers["Referer"] = fmt.Sprintf("%s/tags/%s/artworks", constants.PIXIV_URL, tagName)
	artworkIds, errSlice := tagSearchLogic(
		tagName,
		&httpfuncs.RequestArgs{
			Url:         url,
			Method:      "GET",
			Cookies:     dlOptions.SessionCookies,
			Headers:     headers,
			Params:      params,
			CheckStatus: true,
			UserAgent:   dlOptions.Configs.UserAgent,
			Http2:       !useHttp3,
			Http3:       useHttp3,
			Context:     dlOptions.GetContext(),
		},
		&pageNumArgs{
			minPage: minPage,
			maxPage: maxPage,
			hasMax:  hasMax,
		},
	)

	hasErr := false
	if len(errSlice) > 0 {
		hasErr = true
		if cancelled := logger.LogErrors(false, logger.ERROR, errSlice...); cancelled {
			return nil, nil, false, true
		}
	}

	artworkSlice, ugoiraSlice := GetMultipleArtworkDetails(
		artworkIds,
		downloadPath,
		dlOptions,
	)
	return artworkSlice, ugoiraSlice, hasErr, false
}
