package pixivmobile

import (
	"context"
	"fmt"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
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
	SortOrder    string
	SearchMode   string
	SearchAiType int // 0: filter AI works, 1: Display AI works
	RatingMode   string
	ArtworkType  string

	Configs *configs.Config

	MobileClient *PixivMobile
	RefreshToken string

	Notifier notify.Notifier

	// Progress indicators
	MainProgBar          progress.ProgressBar
	DownloadProgressBars *[]*progress.DownloadProgressBar

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

// CancelCtx releases the resources used and cancels the context of the PixivMobileDlOptions struct.
func (p *PixivMobileDlOptions) CancelCtx() {
	p.cancel()
}

func (p *PixivMobileDlOptions) CtxIsActive() bool {
	return p.ctx.Err() == nil
}

// ValidateArgs validates the arguments of the Pixiv download options.
//
// Should be called after initialising the struct.
func (p *PixivMobileDlOptions) ValidateArgs(userAgent string) error {
	if p.GetContext() == nil {
		p.SetContext(context.Background())
	}

	if p.MainProgBar == nil {
		return fmt.Errorf(
			"pixiv mobile error %d: main progress bar is empty",
			errs.DEV_ERROR,
		)
	}

	if p.Notifier == nil {
		return fmt.Errorf(
			"pixiv mobile error %d: notifier is nil",
			errs.DEV_ERROR,
		)
	}

	p.SortOrder = strings.ToLower(p.SortOrder)
	_, err := api.ValidateStrArgs(
		p.SortOrder,
		ACCEPTED_SORT_ORDER,
		[]string{
			fmt.Sprintf(
				"pixiv mobile error %d: Sort order %s is not allowed",
				errs.INPUT_ERROR,
				p.SortOrder,
			),
		},
	)
	if err != nil {
		return err
	}

	p.SearchMode = strings.ToLower(p.SearchMode)
	_, err = api.ValidateStrArgs(
		p.SearchMode,
		ACCEPTED_SEARCH_MODE,
		[]string{
			fmt.Sprintf(
				"pixiv mobile error %d: Search order %s is not allowed",
				errs.INPUT_ERROR,
				p.SearchMode,
			),
		},
	)
	if err != nil {
		return err
	}

	p.RatingMode = strings.ToLower(p.RatingMode)
	_, err = api.ValidateStrArgs(
		p.RatingMode,
		ACCEPTED_RATING_MODE,
		[]string{
			fmt.Sprintf(
				"pixiv mobile error %d: Rating order %s is not allowed",
				errs.INPUT_ERROR,
				p.RatingMode,
			),
		},
	)
	if err != nil {
		return err
	}

	p.ArtworkType = strings.ToLower(p.ArtworkType)
	_, err = api.ValidateStrArgs(
		p.ArtworkType,
		ACCEPTED_ARTWORK_TYPE,
		[]string{
			fmt.Sprintf(
				"pixiv mobile error %d: Artwork type %s is not allowed",
				errs.INPUT_ERROR,
				p.ArtworkType,
			),
		},
	)
	if err != nil {
		return err
	}

	if p.RefreshToken != "" {
		p.MobileClient, err = NewPixivMobile(p.RefreshToken, 10, p.ctx, p.cancel)
		if err != nil {
			return err
		}

		// The web API value is the opposite of the mobile API;
		// Movile API:
		// - 0: Filter AI works
		// - 1: Display AI works
		// Web API:
		// - 0: Display AI works
		// - 1: Filter AI works
		// Hence, we will have to invert the value.
		if p.SearchAiType == 1 { 
			p.SearchAiType = 0 
		} else if p.SearchAiType == 0 {
			p.SearchAiType = 1
		} else { // invalid value
			p.SearchAiType = 0 // default to filter AI works
		}

		// Now that we have the client, 
		// we will have to update the ajax equivalent parameters to suit the mobile API.
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
			panic(
				fmt.Sprintf(
					"pixiv mobile error %d: invalid search mode %q",
					errs.DEV_ERROR,
					p.SearchMode,
				),
			)
		}

		// Convert sort order to the correct value
		// based on the Pixiv's ajax web API
		var newSortOrder string
		isPremium := p.MobileClient.user.IsPremium
		if isPremium && strings.Contains(p.SortOrder, "popular") {
			newSortOrder = "popular_desc" // only supports popular_desc
		} else if p.SortOrder == "date" {
			newSortOrder = "date_asc"
		} else {
			newSortOrder = "date_desc"
		}
		p.SortOrder = newSortOrder
	}
	return nil
}
