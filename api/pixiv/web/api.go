package pixivweb

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	"github.com/KJHJason/Cultured-Downloader-Logic/filters"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/metadata"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

func getArtworkDetailsLogic(artworkId string, reqArgs *httpfuncs.RequestArgs) (*ArtworkDetails, error) {
	artworkDetailsRes, err := httpfuncs.CallRequest(reqArgs)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
		}

		return nil, fmt.Errorf(
			"pixiv web error %d: failed to get artwork details for ID %s from %s",
			cdlerrors.CONNECTION_ERROR,
			artworkId,
			reqArgs.Url,
		)
	}

	if artworkDetailsRes.Resp.StatusCode != 200 {
		artworkDetailsRes.Close()
		return nil, fmt.Errorf(
			"pixiv web error %d: failed to get details for artwork ID %s due to %s response from %s",
			cdlerrors.RESPONSE_ERROR,
			artworkId,
			artworkDetailsRes.Resp.Status,
			reqArgs.Url,
		)
	}

	var artworkDetailsJsonRes ArtworkDetails
	if err := httpfuncs.LoadJsonFromResponse(artworkDetailsRes.Resp, &artworkDetailsJsonRes); err != nil {
		return nil, fmt.Errorf(
			"%w\ndetails: failed to read response body for Pixiv artwork ID %s",
			err,
			artworkId,
		)
	}
	return &artworkDetailsJsonRes, nil
}

func getArtworkUrlsToDlLogic(artworkType int, artworkId string, reqArgs *httpfuncs.RequestArgs) (*http.Response, error) {
	url, err := getDownloadableUrls(artworkType, artworkId)
	if err != nil {
		return nil, err
	}

	reqArgs.Url = url
	artworkUrlsRes, err := httpfuncs.CallRequest(reqArgs)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
		}

		return nil, fmt.Errorf(
			"pixiv web error %d: failed to get artwork URLs for ID %s from %s due to %w",
			cdlerrors.CONNECTION_ERROR,
			artworkId,
			url,
			err,
		)
	}

	if artworkUrlsRes.Resp.StatusCode != 200 {
		artworkUrlsRes.Close()
		return nil, fmt.Errorf(
			"pixiv web error %d: failed to get artwork URLs for ID %s due to %s response from %s",
			cdlerrors.RESPONSE_ERROR,
			artworkId,
			artworkUrlsRes.Resp.Status,
			url,
		)
	}
	return artworkUrlsRes.Resp, nil
}

// Retrieves details of an artwork ID and returns
// the folder path to download the artwork to, the JSON response, and the artwork type
func getArtworkDetails(artworkId string, dlOptions *PixivWebDlOptions) ([]*httpfuncs.ToDownload, *ugoira.Ugoira, error) {
	if artworkId == "" {
		return nil, nil, nil
	}

	var artworkCacheKey string
	var ugoiraCacheKey string
	url := getArtworkDetailsApi(artworkId) // API URL
	webUrl := fmt.Sprintf("https://www.pixiv.net/artworks/%s", artworkId)
	if dlOptions.Base.UseCacheDb {
		if database.PostCacheExists(webUrl, constants.PIXIV) || database.UgoiraCacheExists(webUrl) {
			// either the artwork or the ugoira cache exists
			return nil, nil, nil
		}
		ugoiraCacheKey = webUrl
		artworkCacheKey = database.ParsePostKey(webUrl, constants.PIXIV)
	}

	headers := pixivcommon.GetPixivRequestHeaders()
	headers["Referer"] = pixivcommon.GetUserUrl(artworkId)

	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV, true)
	reqArgs := &httpfuncs.RequestArgs{
		Url:            url,
		Method:         "GET",
		Cookies:        dlOptions.Base.SessionCookies,
		Headers:        headers,
		UserAgent:      dlOptions.Base.Configs.UserAgent,
		Http2:          !useHttp3,
		Http3:          useHttp3,
		Context:        dlOptions.GetContext(),
		CaptchaHandler: dlOptions.GetCaptchaHandler(),
	}
	artworkDetailsJsonRes, err := getArtworkDetailsLogic(artworkId, reqArgs)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			dlOptions.CancelCtx()
		}
		return nil, nil, err
	}

	artworkJsonBody := artworkDetailsJsonRes.Body
	if !dlOptions.Base.Filters.IsPostDateValid(artworkJsonBody.UploadDate) {
		return nil, nil, nil
	}

	illustratorName := artworkJsonBody.UserName
	artworkName := artworkJsonBody.Title
	artworkPostDir := iofuncs.GetPostFolder(
		dlOptions.Base.DownloadDirPath,
		illustratorName,
		artworkId,
		artworkName,
	)

	if dlOptions.Base.SetMetadata {
		var readableIllustType string
		switch artworkJsonBody.IllustType {
		case ILLUST:
			readableIllustType = "Illustration"
		case MANGA:
			readableIllustType = "Manga"
		case UGOIRA:
			readableIllustType = "Ugoira"
		default:
			readableIllustType = "Unknown"
		}
		artworkMetadata := metadata.PixivPost{
			Url:   webUrl,
			Title: artworkName,
			Type:  readableIllustType,
		}
		if err := metadata.WriteMetadata(artworkMetadata, artworkPostDir); err != nil {
			return nil, nil, err
		}
	}

	artworkType := artworkJsonBody.IllustType
	artworkUrlsRes, err := getArtworkUrlsToDlLogic(artworkType, artworkId, reqArgs)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			dlOptions.CancelCtx()
		}
		return nil, nil, err
	}

	urlsToDl, ugoiraInfo, err := processArtworkJson(
		ugoiraCacheKey,
		artworkCacheKey,
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
func GetMultipleArtworkDetails(artworkIds []string, dlOptions *PixivWebDlOptions) ([]*httpfuncs.ToDownload, []*ugoira.Ugoira, []error) {
	var errSlice []error
	var ugoiraDetails []*ugoira.Ugoira
	var artworkDetails []*httpfuncs.ToDownload
	artworkIdsLen := len(artworkIds)
	lastArtworkId := artworkIds[artworkIdsLen-1]

	var progress progress.ProgressBar
	baseMsg := "Getting and processing artwork details from Pixiv [%d/" + fmt.Sprintf("%d]...", artworkIdsLen)
	progress = dlOptions.Base.MainProgBar()
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
	progress.SetToProgressBar()
	progress.UpdateMax(artworkIdsLen)
	progress.Start()
	defer progress.SnapshotTask()
	for _, artworkId := range artworkIds {
		artworksToDl, ugoiraInfo, err := getArtworkDetails(
			artworkId,
			dlOptions,
		)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				dlOptions.CancelCtx()
				progress.StopInterrupt("Stopped getting and processing artwork details from Pixiv!")
				return nil, nil, errSlice
			}

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
		if hasCancelled := logger.LogErrors(logger.ERROR, errSlice...); hasCancelled {
			dlOptions.CancelCtx()
			progress.StopInterrupt("Stopped getting and processing artwork details from Pixiv!")
			return nil, nil, errSlice
		}
	}

	progress.Stop(hasErr)
	return artworkDetails, ugoiraDetails, errSlice
}

// Query Pixiv's API for all the illustrator's posts
func getArtistPosts(illustratorId, pageNum string, dlOptions *PixivWebDlOptions) ([]string, error) {
	headers := pixivcommon.GetPixivRequestHeaders()
	headers["Referer"] = pixivcommon.GetIllustUrl(illustratorId)
	url := getArtistArtworksApi(illustratorId)

	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV, true)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Url:            url,
			Method:         "GET",
			Cookies:        dlOptions.Base.SessionCookies,
			Headers:        headers,
			UserAgent:      dlOptions.Base.Configs.UserAgent,
			Http2:          !useHttp3,
			Http3:          useHttp3,
			Context:        dlOptions.GetContext(),
			CaptchaHandler: dlOptions.GetCaptchaHandler(),
		},
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, fmt.Errorf(
			"pixiv web error %d: failed to get illustrator's posts with an ID of %s due to %w",
			cdlerrors.CONNECTION_ERROR,
			illustratorId,
			err,
		)
	}
	if res.Resp.StatusCode != 200 {
		res.Close()
		return nil, fmt.Errorf(
			"pixiv web error %d: failed to get illustrator's posts with an ID of %s due to %s response",
			cdlerrors.RESPONSE_ERROR,
			illustratorId,
			res.Resp.Status,
		)
	}

	var jsonBody IllustratorJson
	if err := httpfuncs.LoadJsonFromResponse(res.Resp, &jsonBody); err != nil {
		return nil, err
	}
	artworkIds, err := processArtistPostsJson(&jsonBody, pageNum, dlOptions)
	return artworkIds, err
}

// Get posts from multiple illustrators and returns a slice of artwork IDs
func GetMultipleArtistsPosts(illustratorIds, pageNums []string, dlOptions *PixivWebDlOptions) ([]string, []error) {
	var errSlice []error
	var artworkIdsSlice []string
	illustratorIdsLen := len(illustratorIds)
	lastIllustratorIdx := illustratorIdsLen - 1

	baseMsg := "Getting artwork details from artist(s) on Pixiv [%d/" + fmt.Sprintf("%d]...", illustratorIdsLen)
	progress := dlOptions.Base.MainProgBar()
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting artwork details from %d artist(s) on Pixiv!",
			illustratorIdsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting artwork details from %d artist(s) on Pixiv!\nPlease refer to the logs for more details.",
			illustratorIdsLen,
		),
	)
	progress.SetToProgressBar()
	progress.UpdateMax(illustratorIdsLen)
	progress.Start()
	defer progress.SnapshotTask()
	for idx, illustratorId := range illustratorIds {
		artworkIds, err := getArtistPosts(
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
		if hasCancelled := logger.LogErrors(logger.ERROR, errSlice...); hasCancelled {
			dlOptions.CancelCtx()
			progress.StopInterrupt("Stopped getting artwork details from artist(s) on Pixiv!")
			return nil, errSlice
		}
	}
	progress.Stop(hasErr)

	return artworkIdsSlice, errSlice
}

type pageNumArgs struct {
	minPage int
	maxPage int
	hasMax  bool
}

func tagSearchLogic(filters *filters.Filters, tagName string, reqArgs *httpfuncs.RequestArgs, pageNumArgs *pageNumArgs) ([]string, []error) {
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
			if errors.Is(err, context.Canceled) {
				errSlice = append(errSlice, err)
				return nil, errSlice
			}
			err = fmt.Errorf(
				"pixiv web error %d: failed to get tag search results for %s due to %w",
				cdlerrors.CONNECTION_ERROR,
				tagName,
				err,
			)
			errSlice = append(errSlice, err)
			continue
		}

		tagArtworkIds, err := processTagJsonResults(filters, res.Resp)
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
func TagSearch(tagName, pageNum string, dlOptions *PixivWebDlOptions) ([]string, []error, bool) {
	minPage, maxPage, hasMax, err := utils.GetMinMaxFromStr(pageNum)
	if err != nil {
		logger.LogError(err, logger.ERROR)
		return nil, []error{err}, false
	}

	url := getTagArtworksApi(tagName)
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

		// 0: display AI works, 1: hide AI works
		"ai_type": strconv.Itoa(dlOptions.SearchAiMode),
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV, true)
	headers := pixivcommon.GetPixivRequestHeaders()
	headers["Referer"] = fmt.Sprintf("%s/tags/%s/artworks", constants.PIXIV_URL, tagName)
	artworkIds, errSlice := tagSearchLogic(
		dlOptions.Base.Filters,
		tagName,
		&httpfuncs.RequestArgs{
			Url:            url,
			Method:         "GET",
			Cookies:        dlOptions.Base.SessionCookies,
			Headers:        headers,
			Params:         params,
			CheckStatus:    true,
			UserAgent:      dlOptions.Base.Configs.UserAgent,
			Http2:          !useHttp3,
			Http3:          useHttp3,
			Context:        dlOptions.GetContext(),
			CaptchaHandler: dlOptions.GetCaptchaHandler(),
		},
		&pageNumArgs{
			minPage: minPage,
			maxPage: maxPage,
			hasMax:  hasMax,
		},
	)

	if len(errSlice) > 0 {
		if cancelled := logger.LogErrors(logger.ERROR, errSlice...); cancelled {
			dlOptions.CancelCtx()
			return nil, errSlice, true
		}
	}

	return artworkIds, errSlice, false
}
