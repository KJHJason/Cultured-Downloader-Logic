package pixivmobile

import (
	"context"
	"fmt"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

// PixivToDl is the struct that contains the arguments of Pixiv download options.
type PixivMobileDlOptions struct {
	ctx    context.Context
	cancel context.CancelFunc

	UseCacheDb          bool
	BaseDownloadDirPath string

	// Sort order of the results. Can be "date_desc" or "date_asc".
	SortOrder    string
	SearchMode   string
	SearchAiMode int // 0: filter AI works, 1: Display AI works
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
func (p *PixivMobileDlOptions) ValidateArgs() error {
	if p.GetContext() == nil {
		p.SetContext(context.Background())
	}

	if p.MainProgBar == nil {
		return fmt.Errorf(
			"pixiv mobile error %d: main progress bar is empty",
			cdlerrors.DEV_ERROR,
		)
	}

	if p.Configs == nil {
		return fmt.Errorf(
			"pixiv mobile error %d, configs is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	if p.UseCacheDb && p.Configs.OverwriteFiles {
		p.UseCacheDb = false
	}

	if dlDirPath, err := api.ValidateDlDirPath(p.BaseDownloadDirPath, constants.PIXIV_MOBILE_TITLE); err != nil {
		return err
	} else {
		p.BaseDownloadDirPath = dlDirPath
	}

	if p.Notifier == nil {
		return fmt.Errorf(
			"pixiv mobile error %d: notifier is nil",
			cdlerrors.DEV_ERROR,
		)
	}

	p.SortOrder = strings.ToLower(p.SortOrder)
	_, err := api.ValidateStrArgs(
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
	_, err = api.ValidateStrArgs(
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
	_, err = api.ValidateStrArgs(
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
	_, err = api.ValidateStrArgs(
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

	if p.RefreshToken != "" {
		p.MobileClient, err = NewPixivMobile(p.RefreshToken, 10, p.ctx, p.cancel)
		if err != nil {
			return err
		}
		p.MobileClient.SetMainProgBar(p.MainProgBar)
		p.MobileClient.SetBaseDlDirPath(p.BaseDownloadDirPath)
		p.MobileClient.SetUseCacheDb(p.UseCacheDb)

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
			return fmt.Errorf(
				"pixiv mobile error %d: invalid search mode %q",
				cdlerrors.DEV_ERROR,
				p.SearchMode,
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
