package pixivcommon

import (
	"fmt"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

type PixivFilters struct {
	// Sort order of the results. Can be "date_desc" or "date_asc".
	SortOrder  string
	SearchMode string

	// Web API:
	// 1: filter AI works, 0: Display AI works
	//
	// Mobile API:
	// 0: filter AI works, 1: Display AI works
	SearchAiMode int
	RatingMode   string
	ArtworkType  string
}

func (p *PixivFilters) ValidateForMobileApi(userIsPremium bool) error {
	p.SortOrder = strings.ToLower(p.SortOrder)
	_, err := utils.ValidateStrArgs(
		p.SortOrder,
		constants.ACCEPTED_SORT_ORDER,
		[]string{
			fmt.Sprintf(
				"pixiv mobile error %d: Sort order %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.SortOrder,
			),
		},
	)
	if err != nil {
		return err
	}

	p.SearchMode = strings.ToLower(p.SearchMode)
	_, err = utils.ValidateStrArgs(
		p.SearchMode,
		constants.ACCEPTED_SEARCH_MODE,
		[]string{
			fmt.Sprintf(
				"pixiv mobile error %d: Search order %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.SearchMode,
			),
		},
	)
	if err != nil {
		return err
	}

	p.RatingMode = strings.ToLower(p.RatingMode)
	_, err = utils.ValidateStrArgs(
		p.RatingMode,
		constants.ACCEPTED_RATING_MODE,
		[]string{
			fmt.Sprintf(
				"pixiv mobile error %d: Rating order %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.RatingMode,
			),
		},
	)
	if err != nil {
		return err
	}

	p.ArtworkType = strings.ToLower(p.ArtworkType)
	_, err = utils.ValidateStrArgs(
		p.ArtworkType,
		constants.ACCEPTED_ARTWORK_TYPE,
		[]string{
			fmt.Sprintf(
				"pixiv mobile error %d: Artwork type %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.ArtworkType,
			),
		},
	)
	if err != nil {
		return err
	}

	// The web API value is the opposite of the mobile API;
	// Mobile API:
	// - 0: Filter AI works
	// - 1: Display AI works
	// Web API:
	// - 0: Display AI works
	// - 1: Filter AI works
	// Hence, we will have to invert the value.
	if p.SearchAiMode == 1 {
		p.SearchAiMode = 0
	} else if p.SearchAiMode == 0 {
		p.SearchAiMode = 1
	} else { // invalid value
		p.SearchAiMode = 0 // default to filter AI works
	}

	// Now that we have the client,
	// we will have to update the ajax equivalent parameters to suit the mobile utils.
	if p.RatingMode != "all" {
		p.RatingMode = "all" // only supports "all"
	}

	if p.ArtworkType == "illust_and_ugoira" {
		// convert "illust_and_ugoira" to "illust"
		// since the mobile API does not support "illust_and_ugoira"
		// However, there will still be ugoira posts in the results
		p.ArtworkType = "illust"
	}

	// Convert search mode to the correct value
	// based on the Pixiv's ajax web API
	switch p.SearchMode {
	case "s_tag":
		p.SearchMode = "partial_match_for_tags"
	case "s_tag_full":
		p.SearchMode = "exact_match_for_tags"
	case "s_tc":
		p.SearchMode = "title_and_caption"
	default:
		return fmt.Errorf(
			"pixiv mobile error %d: invalid search mode %q",
			cdlerrors.DEV_ERROR,
			p.SearchMode,
		)
	}

	// Convert sort order to the correct value
	// based on the Pixiv's ajax web API
	var newSortOrder string
	if userIsPremium && strings.Contains(p.SortOrder, "popular") {
		newSortOrder = "popular_desc" // only supports popular_desc
	} else if p.SortOrder == "date" {
		newSortOrder = "date_asc"
	} else {
		newSortOrder = "date_desc"
	}
	p.SortOrder = newSortOrder
	return nil
}

func (p *PixivFilters) ValidateForWebApi() error {
	// Web API:
	// - 0: Display AI works
	// - 1: Filter AI works
	if p.SearchAiMode != 0 && p.SearchAiMode != 1 {
		p.SearchAiMode = 1 // Default to filter AI works
	}

	p.SortOrder = strings.ToLower(p.SortOrder)
	_, err := utils.ValidateStrArgs(
		p.SortOrder,
		constants.ACCEPTED_SORT_ORDER,
		[]string{
			fmt.Sprintf(
				"pixiv web error %d: Sort order %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.SortOrder,
			),
		},
	)
	if err != nil {
		return err
	}

	p.SearchMode = strings.ToLower(p.SearchMode)
	_, err = utils.ValidateStrArgs(
		p.SearchMode,
		constants.ACCEPTED_SEARCH_MODE,
		[]string{
			fmt.Sprintf(
				"pixiv web error %d: Search order %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.SearchMode,
			),
		},
	)
	if err != nil {
		return err
	}

	p.RatingMode = strings.ToLower(p.RatingMode)
	_, err = utils.ValidateStrArgs(
		p.RatingMode,
		constants.ACCEPTED_RATING_MODE,
		[]string{
			fmt.Sprintf(
				"pixiv web error %d: Rating order %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.RatingMode,
			),
		},
	)
	if err != nil {
		return err
	}

	p.ArtworkType = strings.ToLower(p.ArtworkType)
	_, err = utils.ValidateStrArgs(
		p.ArtworkType,
		constants.ACCEPTED_ARTWORK_TYPE,
		[]string{
			fmt.Sprintf(
				"pixiv web error %d: Artwork type %s is not allowed",
				cdlerrors.INPUT_ERROR,
				p.ArtworkType,
			),
		},
	)
	if err != nil {
		return err
	}
	return nil
}
