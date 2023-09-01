package pixivmobile

import (
	"context"
	"fmt"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/fatih/color"
)

var (
	ACCEPTED_SORT_ORDER = []string{
		"date", "date_d",
		"popular", "popular_d",
		"popular_male", "popular_male_d",
		"popular_female", "popular_female_d",
	}
	ACCEPTED_SEARCH_MODE = []string{
		"s_tag",
		"s_tag_full",
		"s_tc",
	}
	ACCEPTED_RATING_MODE = []string{
		"safe",
		"r18",
		"all",
	}
	ACCEPTED_ARTWORK_TYPE = []string{
		"illust_and_ugoira",
		"manga",
		"all",
	}
)

// PixivToDl is the struct that contains the arguments of Pixiv download options.
type PixivMobileDlOptions struct {
	ctx    context.Context
	cancel context.CancelFunc

	// Sort order of the results. Can be "date_desc" or "date_asc".
	SortOrder   string
	SearchMode  string
	RatingMode  string
	ArtworkType string

	Configs *configs.Config

	MobileClient *PixivMobile
	RefreshToken string

	Notifier notify.Notifier

	// Prog bar
	TagSearchProgBar progress.Progress
}

func (p *PixivMobileDlOptions) GetContext() context.Context {
	return p.ctx
}

func (p *PixivMobileDlOptions) GetCancel() context.CancelFunc {
	return p.cancel
}

func (p *PixivMobileDlOptions) SetContext(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
}

// ValidateArgs validates the arguments of the Pixiv download options.
//
// Should be called after initialising the struct.
func (p *PixivMobileDlOptions) ValidateArgs(userAgent string) {
	if p.GetContext() == nil {
		p.SetContext(context.Background())
	}

	p.SortOrder = strings.ToLower(p.SortOrder)
	api.ValidateStrArgs(
		p.SortOrder,
		ACCEPTED_SORT_ORDER,
		[]string{
			fmt.Sprintf(
				"pixiv error %d: Sort order %s is not allowed",
				constants.INPUT_ERROR,
				p.SortOrder,
			),
		},
	)

	p.SearchMode = strings.ToLower(p.SearchMode)
	api.ValidateStrArgs(
		p.SearchMode,
		ACCEPTED_SEARCH_MODE,
		[]string{
			fmt.Sprintf(
				"pixiv error %d: Search order %s is not allowed",
				constants.INPUT_ERROR,
				p.SearchMode,
			),
		},
	)

	p.RatingMode = strings.ToLower(p.RatingMode)
	api.ValidateStrArgs(
		p.RatingMode,
		ACCEPTED_RATING_MODE,
		[]string{
			fmt.Sprintf(
				"pixiv error %d: Rating order %s is not allowed",
				constants.INPUT_ERROR,
				p.RatingMode,
			),
		},
	)

	p.ArtworkType = strings.ToLower(p.ArtworkType)
	api.ValidateStrArgs(
		p.ArtworkType,
		ACCEPTED_ARTWORK_TYPE,
		[]string{
			fmt.Sprintf(
				"pixiv error %d: Artwork type %s is not allowed",
				constants.INPUT_ERROR,
				p.ArtworkType,
			),
		},
	)

	if p.RefreshToken != "" {
		p.MobileClient = NewPixivMobile(p.RefreshToken, 10, p.ctx)
		if p.RatingMode != "all" {
			color.Red(
				api.CombineStringsWithNewline(
					fmt.Sprintf(
						"pixiv error %d: when using the refresh token, only \"all\" is supported for the --rating_mode flag.",
						constants.INPUT_ERROR,
					),
					fmt.Sprintf(
						"hence, the rating mode will be updated from %q to \"all\"...\n",
						p.RatingMode,
					),
				),
			)
			p.RatingMode = "all"
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
			panic(
				fmt.Sprintf(
					"pixiv mobile error %d: invalid search mode %q",
					constants.DEV_ERROR,
					p.SearchMode,
				),
			)
		}

		// Convert sort order to the correct value
		// based on the Pixiv's ajax web API
		var newSortOrder string
		if strings.Contains(p.SortOrder, "popular") {
			newSortOrder = "popular_desc" // only supports popular_desc
		} else if p.SortOrder == "date_d" {
			newSortOrder = "date_desc"
		} else {
			newSortOrder = "date_asc"
		}

		if p.SortOrder != "date" && p.SortOrder != "date_d" && p.SortOrder != "popular_d" {
			var ajaxEquivalent string
			switch newSortOrder {
			case "popular_desc":
				ajaxEquivalent = "popular_d"
			case "date_desc":
				ajaxEquivalent = "date_d"
			case "date_asc":
				ajaxEquivalent = "date"
			default:
				panic(
					fmt.Sprintf(
						"pixiv error %d: unknown sort order %q in PixivDlOptions.ValidateArgs()",
						constants.DEV_ERROR,
						newSortOrder,
					),
				)
			}

			color.Red(
				api.CombineStringsWithNewline(
					fmt.Sprintf(
						"pixiv error %d: when using the refresh token, only \"date\", \"date_d\", \"popular_d\" are supported for the --sort_order flag.",
						constants.INPUT_ERROR,
					),
					fmt.Sprintf(
						"hence, the sort order will be updated from %q to %q...\n",
						p.SortOrder,
						ajaxEquivalent,
					),
				),
			)
		}
		p.SortOrder = newSortOrder
	}
}
