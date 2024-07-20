package pixivweb

import (
	"fmt"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

func getDownloadableUrls(artworkType int, artworkId string) (string, error) {
	switch artworkType {
	case ILLUST, MANGA: // illustration or manga
		return fmt.Sprintf("%s/illust/%s/pages", constants.PIXIV_API_URL, artworkId), nil
	case UGOIRA: // ugoira
		return fmt.Sprintf("%s/illust/%s/ugoira_meta", constants.PIXIV_API_URL, artworkId), nil
	default:
		return "", fmt.Errorf(
			"pixiv web error %d: unsupported artwork type %d for artwork ID %s",
			cdlerrors.JSON_ERROR,
			artworkType,
			artworkId,
		)
	}
}

func getArtworkDetailsApi(artworkId string) string {
	return fmt.Sprintf("%s/illust/%s", constants.PIXIV_API_URL, artworkId)
}

func getArtistArtworksApi(artistId string) string {
	return fmt.Sprintf("%s/user/%s/profile/all", constants.PIXIV_API_URL, artistId)
}

func getTagArtworksApi(tag string) string {
	return fmt.Sprintf("%s/search/artworks/%s", constants.PIXIV_API_URL, tag)
}
