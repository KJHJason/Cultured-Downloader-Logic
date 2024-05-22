package pixivweb

import (
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

func processIllustratorPostJson(resJson *IllustratorJson, pageNum string, pixivDlOptions *PixivWebDlOptions) ([]string, error) {
	minPage, maxPage, hasMax, err := api.GetMinMaxFromStr(pageNum)
	if err != nil {
		return nil, err
	}
	minOffset, maxOffset := pixivcommon.ConvertPageNumToOffset(minPage, maxPage, constants.PIXIV_PER_PAGE, false)

	var artworkIds []string
	if pixivDlOptions.ArtworkType == "all" || pixivDlOptions.ArtworkType == "illust_and_ugoira" {
		illusts := resJson.Body.Illusts
		switch t := illusts.(type) {
		case map[string]interface{}:
			curOffset := 0
			for illustId := range t {
				curOffset++
				if curOffset < minOffset {
					continue
				}
				if hasMax && curOffset > maxOffset {
					break
				}

				artworkIds = append(artworkIds, illustId)
			}
		default: // where there are no posts or has an unknown type
			break
		}
	}

	if pixivDlOptions.ArtworkType == "all" || pixivDlOptions.ArtworkType == "manga" {
		manga := resJson.Body.Manga
		switch t := manga.(type) {
		case map[string]interface{}:
			curOffset := 0
			for mangaId := range t {
				curOffset++
				if curOffset < minOffset {
					continue
				}
				if hasMax && curOffset > maxOffset {
					break
				}

				artworkIds = append(artworkIds, mangaId)
			}
		default: // where there are no posts or has an unknown type
			break
		}
	}
	return artworkIds, nil
}

// Process the artwork details JSON and returns a map of urls
// with its file path or a Ugoira struct (One of them will be null depending on the artworkType)
func processArtworkJson(ugoiraCacheKey, artworkCacheKey string, res *http.Response, artworkType int64, postDownloadDir string) ([]*httpfuncs.ToDownload, *ugoira.Ugoira, error) {
	if artworkType == UGOIRA {
		var ugoiraJson ArtworkUgoiraJson
		if err := httpfuncs.LoadJsonFromResponse(res, &ugoiraJson); err != nil {
			return nil, nil, err
		}

		ugoiraMap := ugoiraJson.Body
		originalUrl := ugoiraMap.OriginalSrc
		ugoiraInfo := &ugoira.Ugoira{
			CacheKey: ugoiraCacheKey,
			Url:      originalUrl,
			FilePath: postDownloadDir,
			Frames:   ugoira.MapDelaysToFilename(ugoiraMap.Frames),
		}
		return nil, ugoiraInfo, nil
	}

	var artworkUrls ArtworkJson
	if err := httpfuncs.LoadJsonFromResponse(res, &artworkUrls); err != nil {
		return nil, nil, err
	}

	var urlsToDownload []*httpfuncs.ToDownload
	for _, artworkUrl := range artworkUrls.Body {
		urlsToDownload = append(urlsToDownload, &httpfuncs.ToDownload{
			CacheKey: artworkCacheKey,
			CacheFn:  database.CachePost,
			Url:      artworkUrl.Urls.Original,
			FilePath: postDownloadDir,
		})
	}
	return urlsToDownload, nil, nil
}

// Process the tag search results JSON and returns a slice of artwork IDs
func processTagJsonResults(res *http.Response) ([]string, error) {
	var pixivTagJson PixivTag
	if err := httpfuncs.LoadJsonFromResponse(res, &pixivTagJson); err != nil {
		return nil, err
	}

	artworksSlice := []string{}
	for _, illust := range pixivTagJson.Body.IllustManga.Data {
		artworksSlice = append(artworksSlice, illust.Id)
	}
	return artworksSlice, nil
}
