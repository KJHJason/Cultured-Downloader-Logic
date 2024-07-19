package pixivweb

import (
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
)

type ArtworkDetails struct {
	// Error   bool   `json:"error"`
	// Message string `json:"message"`
	Body struct {
		// IllustID      string    `json:"illustId"`
		// IllustTitle   string    `json:"illustTitle"`
		// IllustComment string    `json:"illustComment"`
		// ID            string    `json:"id"`
		Title string `json:"title"` // should be same as IllustTitle
		// Description   string    `json:"description"`
		IllustType int `json:"illustType"`
		// CreateDate    time.Time `json:"createDate"`
		UploadDate time.Time `json:"uploadDate"` // 2024-05-12T11:17:00+00:00
		// Restrict      int       `json:"restrict"`
		// XRestrict     int       `json:"xRestrict"`
		// Sl            int       `json:"sl"`
		// Urls          struct {
		// 	Mini     string `json:"mini"`
		// 	Thumb    string `json:"thumb"`
		// 	Small    string `json:"small"`
		// 	Regular  string `json:"regular"`
		// 	Original string `json:"original"`
		// } `json:"urls"`
		// Tags struct {
		// 	AuthorID string `json:"authorId"`
		// 	IsLocked bool   `json:"isLocked"`
		// 	Tags     []struct {
		// 		Tag         string `json:"tag"`
		// 		Locked      bool   `json:"locked"`
		// 		Deletable   bool   `json:"deletable"`
		// 		UserID      string `json:"userId,omitempty"`
		// 		Romaji      string `json:"romaji"`
		// 		Translation struct {
		// 			En string `json:"en"`
		// 		} `json:"translation,omitempty"`
		// 		UserName string `json:"userName,omitempty"`
		// 	} `json:"tags"`
		// 	Writable bool `json:"writable"`
		// } `json:"tags"`
		// Alt         string `json:"alt"`
		// UserID      string `json:"userId"`
		UserName string `json:"userName"`
		// UserAccount string `json:"userAccount"`
		// LikeData             bool  `json:"likeData"`
		// Width                int   `json:"width"`
		// Height               int   `json:"height"`
		// PageCount            int   `json:"pageCount"`
		// BookmarkCount        int   `json:"bookmarkCount"`
		// LikeCount            int   `json:"likeCount"`
		// CommentCount         int   `json:"commentCount"`
		// ResponseCount        int   `json:"responseCount"`
		// ViewCount            int   `json:"viewCount"`
		// BookStyle            int   `json:"bookStyle"`
		// IsHowto              bool  `json:"isHowto"`
		// IsOriginal           bool  `json:"isOriginal"`
		// ImageResponseOutData []any `json:"imageResponseOutData"`
		// ImageResponseData    []any `json:"imageResponseData"`
		// ImageResponseCount   int   `json:"imageResponseCount"`
		// PollData             any   `json:"pollData"`
		// SeriesNavData        any   `json:"seriesNavData"`
		// DescriptionBoothID   any   `json:"descriptionBoothId"`
		// DescriptionYoutubeID any   `json:"descriptionYoutubeId"`
		// ComicPromotion       any   `json:"comicPromotion"`
		// FanboxPromotion      any   `json:"fanboxPromotion"`
		// ContestBanners       []any `json:"contestBanners"`
		// IsBookmarkable       bool  `json:"isBookmarkable"`
		// BookmarkData         any   `json:"bookmarkData"`
		// ContestData          any   `json:"contestData"`
		// ZoneConfig           struct {
		// 	Responsive struct {
		// 		URL string `json:"url"`
		// 	} `json:"responsive"`
		// 	Rectangle struct {
		// 		URL string `json:"url"`
		// 	} `json:"rectangle"`
		// 	Five00X500 struct {
		// 		URL string `json:"url"`
		// 	} `json:"500x500"`
		// 	Header struct {
		// 		URL string `json:"url"`
		// 	} `json:"header"`
		// 	Footer struct {
		// 		URL string `json:"url"`
		// 	} `json:"footer"`
		// 	ExpandedFooter struct {
		// 		URL string `json:"url"`
		// 	} `json:"expandedFooter"`
		// 	Logo struct {
		// 		URL string `json:"url"`
		// 	} `json:"logo"`
		// 	Relatedworks struct {
		// 		URL string `json:"url"`
		// 	} `json:"relatedworks"`
		// } `json:"zoneConfig"`
		// ExtraData struct {
		// 	Meta struct {
		// 		Title              string `json:"title"`
		// 		Description        string `json:"description"`
		// 		Canonical          string `json:"canonical"`
		// 		AlternateLanguages struct {
		// 			Ja string `json:"ja"`
		// 			En string `json:"en"`
		// 		} `json:"alternateLanguages"`
		// 		DescriptionHeader string `json:"descriptionHeader"`
		// 		Ogp               struct {
		// 			Description string `json:"description"`
		// 			Image       string `json:"image"`
		// 			Title       string `json:"title"`
		// 			Type        string `json:"type"`
		// 		} `json:"ogp"`
		// 		Twitter struct {
		// 			Description string `json:"description"`
		// 			Image       string `json:"image"`
		// 			Title       string `json:"title"`
		// 			Card        string `json:"card"`
		// 		} `json:"twitter"`
		// 	} `json:"meta"`
		// } `json:"extraData"`
		// TitleCaptionTranslation struct {
		// 	WorkTitle   any `json:"workTitle"`
		// 	WorkCaption any `json:"workCaption"`
		// } `json:"titleCaptionTranslation"`
		// IsUnlisted               bool `json:"isUnlisted"`
		// Request                  any  `json:"request"`
		// CommentOff               int  `json:"commentOff"`
		// AiType                   int  `json:"aiType"`
		// ReuploadDate             any  `json:"reuploadDate"`
		// LocationMask             bool `json:"locationMask"`
		// CommissionIllustHaveRisk bool `json:"commissionIllustHaveRisk"`
	} `json:"body"`
}

type ArtworkUgoiraJson struct {
	Body struct {
		Src         string                  `json:"src"`
		OriginalSrc string                  `json:"originalSrc"`
		MimeType    string                  `json:"mime_type"`
		Frames      ugoira.UgoiraFramesJson `json:"frames"`
	} `json:"body"`
}

type ArtworkJson struct {
	Body []struct {
		Urls struct {
			ThumbMini string `json:"thumb_mini"`
			Small     string `json:"small"`
			Regular   string `json:"regular"`
			Original  string `json:"original"`
		} `json:"urls"`
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"body"`
}

type PixivTag struct {
	Body struct {
		IllustManga struct {
			Data []struct {
				ID string `json:"id"`
				// Title                   string   `json:"title"`
				// IllustType              int      `json:"illustType"`
				// XRestrict               int      `json:"xRestrict"`
				// Restrict                int      `json:"restrict"`
				// Sl                      int      `json:"sl"`
				// URL                     string   `json:"url"`
				// Description             string   `json:"description"`
				// Tags                    []string `json:"tags"`
				// UserID                  string   `json:"userId"`
				// UserName                string   `json:"userName"`
				// Width                   int      `json:"width"`
				// Height                  int      `json:"height"`
				// PageCount               int      `json:"pageCount"`
				// IsBookmarkable          bool     `json:"isBookmarkable"`
				// BookmarkData            any      `json:"bookmarkData"`
				// Alt                     string   `json:"alt"`
				// TitleCaptionTranslation struct {
				// 	WorkTitle   any `json:"workTitle"`
				// 	WorkCaption any `json:"workCaption"`
				// } `json:"titleCaptionTranslation"`
				CreateDate time.Time `json:"createDate"` // 2024-07-19T14:39:25+09:00
				// UpdateDate      time.Time `json:"updateDate"` // 2024-07-19T14:39:25+09:00
				// IsUnlisted      bool      `json:"isUnlisted"`
				// IsMasked        bool      `json:"isMasked"`
				// AiType          int       `json:"aiType"`
				// ProfileImageURL string    `json:"profileImageUrl"`
			} `json:"data"`
		} `json:"illustManga"`
	} `json:"body"`
}

// As to why Illust and Manga are any/interfaces
//
//	"illusts": {
//		"<id>": null,
//	}
type IllustratorJson struct {
	Body struct {
		Illusts any `json:"illusts"`
		Manga   any `json:"manga"`
	} `json:"body"`
}
