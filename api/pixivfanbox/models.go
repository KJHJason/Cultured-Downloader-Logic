package pixivfanbox

import (
	"encoding/json"
	"time"
)

// Contains the paginated URL(s)
//
// e.g.
//
//	{
//	    "body": [
//	        "https://api.fanbox.cc/post.listCreator?creatorId=gmkj0324&maxPublishedDatetime=2024-05-25%2000%3A00%3A00&maxId=7970693&limit=10",
//	        "https://api.fanbox.cc/post.listCreator?creatorId=gmkj0324&maxPublishedDatetime=2024-02-17%2021%3A47%3A47&maxId=7474888&limit=10",
//	        "https://api.fanbox.cc/post.listCreator?creatorId=gmkj0324&maxPublishedDatetime=2024-01-06%2023%3A30%3A00&maxId=7262829&limit=10",
//	        "https://api.fanbox.cc/post.listCreator?creatorId=gmkj0324&maxPublishedDatetime=2023-11-18%2000%3A00%3A00&maxId=6999578&limit=10",
//	        "https://api.fanbox.cc/post.listCreator?creatorId=gmkj0324&maxPublishedDatetime=2023-09-09%2000%3A00%3A00&maxId=6649219&limit=10",
//	        "https://api.fanbox.cc/post.listCreator?creatorId=gmkj0324&maxPublishedDatetime=2019-11-15%2000%3A45%3A08&maxId=657958&limit=10"
//	    ]
//	}
type CreatorPaginatedPostsJson struct {
	Body []string `json:"body"`
}

type FanboxCreatorPostsJson struct {
	Body struct {
		Items []struct {
			ID                string    `json:"id"`
			Title             string    `json:"title"`
			FeeRequired       int       `json:"feeRequired"`
			PublishedDatetime time.Time `json:"publishedDatetime"` // "2023-03-15T14:08:23+09:00",
			UpdatedDatetime   time.Time `json:"updatedDatetime"`   // "2023-03-15T14:08:23+09:00",
			Tags              []string  `json:"tags"`
			IsLiked           bool      `json:"isLiked"`
			LikeCount         int       `json:"likeCount"`
			CommentCount      int       `json:"commentCount"`
			IsRestricted      bool      `json:"isRestricted"`
			User              struct {
				UserID  string `json:"userId"`
				Name    string `json:"name"`
				IconURL string `json:"iconUrl"`
			} `json:"user"`
			CreatorID       string `json:"creatorId"`
			HasAdultContent bool   `json:"hasAdultContent"`
			Cover           struct {
				Type string `json:"type"`
				URL  string `json:"url"`
			} `json:"cover"`
			Excerpt string `json:"excerpt"`
		} `json:"items"`
		NextURL *string `json:"nextUrl"` // can be null value
	} `json:"body"`
}

type FanboxPostJson struct {
	Body struct {
		ID                string    `json:"id"`
		Title             string    `json:"title"`
		FeeRequired       int       `json:"feeRequired"`
		PublishedDatetime time.Time `json:"publishedDatetime"` // "2023-03-15T14:08:23+09:00",
		UpdatedDatetime   time.Time `json:"updatedDatetime"`   // "2023-03-15T14:08:23+09:00",
		Tags              []any     `json:"tags"`
		IsLiked           bool      `json:"isLiked"`
		LikeCount         int       `json:"likeCount"`
		CommentCount      int       `json:"commentCount"`
		IsRestricted      bool      `json:"isRestricted"`
		User              struct {
			UserID  string `json:"userId"`
			Name    string `json:"name"`
			IconURL string `json:"iconUrl"`
		} `json:"user"`
		CreatorID       string          `json:"creatorId"`
		HasAdultContent bool            `json:"hasAdultContent"`
		Type            string          `json:"type"`
		CoverImageURL   string          `json:"coverImageUrl"`
		Body            json.RawMessage `json:"body"` // can be one of FanboxFilePostJson, FanboxImagePostJson, FanboxTextPostJson, FanboxArticleJson
	} `json:"body"`
}

type FanboxFilePostJson struct {
	Text  string `json:"text"`
	Files []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Extension string `json:"extension"`
		Size      int    `json:"size"`
		Url       string `json:"url"`
	} `json:"files"`
}

type FanboxImagePostJson struct {
	Text   string `json:"text"`
	Images []struct {
		ID           string `json:"id"`
		Extension    string `json:"extension"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		OriginalUrl  string `json:"originalUrl"`
		ThumbnailUrl string `json:"thumbnailUrl"`
	} `json:"images"`
}

type FanboxTextPostJson struct {
	Text string `json:"text"`
}

type FanboxArticleBlocks []struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	ImageID string `json:"imageId,omitempty"`
	Styles  []struct {
		Type   string `json:"type"`
		Offset int    `json:"offset"`
		Length int    `json:"length"`
	} `json:"styles,omitempty"`
	Links []struct {
		Offset int    `json:"offset"`
		Length int    `json:"length"`
		Url    string `json:"url"`
	} `json:"links,omitempty"`
	FileID string `json:"fileId,omitempty"`
}

type FanboxArticleJson struct {
	Blocks   FanboxArticleBlocks `json:"blocks"`
	ImageMap map[string]struct {
		ID           string `json:"id"`
		Extension    string `json:"extension"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		OriginalUrl  string `json:"originalUrl"`
		ThumbnailUrl string `json:"thumbnailUrl"`
	} `json:"imageMap"`
	FileMap map[string]struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		Extension string `json:"extension"`
		Size      int    `json:"size"`
		Url       string `json:"url"`
	} `json:"fileMap"`
}
