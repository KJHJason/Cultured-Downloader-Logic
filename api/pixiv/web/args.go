package pixivweb

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
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
type PixivWebDlOptions struct {
	ctx    context.Context
	cancel context.CancelFunc

	// Sort order of the results. Can be "date_desc" or "date_asc".
	SortOrder   string
	SearchMode  string
	RatingMode  string
	ArtworkType string

	Configs *configs.Config

	SessionCookies  []*http.Cookie
	SessionCookieId string

	Notifier notify.Notifier

	// Prog Bar
	TagSearchProgBar           progress.Progress
	GetPostsDetailProgBar      progress.Progress
	GetIllustratorPostsProgBar progress.Progress
}

func (p *PixivWebDlOptions) GetContext() context.Context {
	return p.ctx
}

func (p *PixivWebDlOptions) GetCancel() context.CancelFunc {
	return p.cancel
}

func (p *PixivWebDlOptions) SetContext(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
}

// ValidateArgs validates the arguments of the Pixiv download options.
//
// Should be called after initialising the struct.
func (p *PixivWebDlOptions) ValidateArgs(userAgent string) error {
	if p.GetContext() == nil {
		p.SetContext(context.Background())
	}

	p.SortOrder = strings.ToLower(p.SortOrder)
	_, err := api.ValidateStrArgs(
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
	if err != nil {
		return err
	}

	p.SearchMode = strings.ToLower(p.SearchMode)
	_, err = api.ValidateStrArgs(
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
	if err != nil {
		return err
	}

	p.RatingMode = strings.ToLower(p.RatingMode)
	_, err = api.ValidateStrArgs(
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
	if err != nil {
		return err
	}

	p.ArtworkType = strings.ToLower(p.ArtworkType)
	_, err = api.ValidateStrArgs(
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
	if err != nil {
		return err
	}

	if p.SessionCookieId != "" {
		if cookie, err := api.VerifyAndGetCookie(constants.PIXIV, p.SessionCookieId, userAgent); err != nil {
			return err
		} else {
			p.SessionCookies = []*http.Cookie{cookie}
		}
	}
	return nil
}
