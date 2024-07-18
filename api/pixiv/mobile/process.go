package pixivmobile

import (
	"fmt"
	"strconv"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/metadata"
)

func getUgoiraUrl(artworkId string) string {
	return fmt.Sprintf("%s?illust_id=%s", constants.PIXIV_MOBILE_UGOIRA_URL, artworkId)
}

func getUgoiraUrlFromInt(artworkId int) string {
	return getUgoiraUrl(strconv.Itoa(artworkId))
}

// Process the artwork JSON and returns a slice of map that contains the urls of the images and the file path
func (pixiv *PixivMobile) processArtworkJson(ugoiraCacheKey string, artworkJson *IllustJson) ([]*httpfuncs.ToDownload, *ugoira.Ugoira, error) {
	if artworkJson == nil {
		return nil, nil, nil
	}

	artworkId := strconv.Itoa(artworkJson.Id)
	artworkTitle := artworkJson.Title
	artworkType := artworkJson.Type
	artistName := artworkJson.User.Name
	artworkFolderPath := iofuncs.GetPostFolder(
		pixiv.baseDownloadDirPath, artistName, artworkId, artworkTitle,
	)

	if pixiv.setMetadata {
		postMetadata := metadata.PixivPost{
			Url:   fmt.Sprintf("https://www.pixiv.net/artworks/%s", artworkId),
			Title: artworkTitle,
			Type:  artworkType,
		}
		if err := metadata.WriteMetadata(postMetadata, artworkFolderPath); err != nil {
			return nil, nil, err
		}
	}

	if artworkType == "ugoira" {
		ugoiraInfo, err := pixiv.getUgoiraMetadata(ugoiraCacheKey, artworkId, artworkFolderPath)
		if err != nil {
			return nil, nil, err
		}
		return nil, ugoiraInfo, nil
	}

	var artworksToDownload []*httpfuncs.ToDownload
	singlePageImageUrl := artworkJson.MetaSinglePage.OriginalImageUrl
	if singlePageImageUrl != "" {
		artworksToDownload = append(artworksToDownload, &httpfuncs.ToDownload{
			Url:      singlePageImageUrl,
			FilePath: artworkFolderPath,
		})
	} else {
		for _, image := range artworkJson.MetaPages {
			imageUrl := image.ImageUrls.Original
			artworksToDownload = append(artworksToDownload, &httpfuncs.ToDownload{
				Url:      imageUrl,
				FilePath: artworkFolderPath,
			})
		}
	}
	return artworksToDownload, nil, nil
}

// The same as the processArtworkJson function but for multiple JSONs at once
// (Those with the "illusts" key which holds a slice of maps containing the artwork JSON)
func (pixiv *PixivMobile) processMultipleArtworkJson(resJson *ArtworksJson) ([]*httpfuncs.ToDownload, []*ugoira.Ugoira, []error) {
	if resJson == nil {
		return nil, nil, nil
	}

	artworksMaps := resJson.Illusts
	if len(artworksMaps) == 0 {
		return nil, nil, nil
	}

	var errSlice []error
	var ugoiraToDl []*ugoira.Ugoira
	var artworksToDl []*httpfuncs.ToDownload
	for _, artwork := range artworksMaps {
		artworks, ugoiraVal, err := pixiv.processArtworkJson(
			getUgoiraUrlFromInt(artwork.Id), artwork,
		)
		if err != nil {
			errSlice = append(errSlice, err)
			continue
		}
		if ugoiraVal != nil {
			ugoiraToDl = append(ugoiraToDl, ugoiraVal)
			continue
		}
		artworksToDl = append(artworksToDl, artworks...)
	}
	return artworksToDl, ugoiraToDl, errSlice
}
