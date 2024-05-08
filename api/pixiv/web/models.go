package pixivweb

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
)

type ArtworkDetails struct {
	Body struct {
		UserName   string `json:"userName"`
		Title      string `json:"title"`
		IllustType int64  `json:"illustType"`
	}
}

type ArtworkUgoiraJson struct {
	Body struct {
		Src         string `json:"src"`
		OriginalSrc string `json:"originalSrc"`
		MimeType    string `json:"mime_type"`
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
				Id string `json:"id"`
			} `json:"data"`
		} `json:"illustManga"`
	} `json:"body"`
}

type IllustratorJson struct {
    Body struct {
        Illusts interface{} `json:"illusts"`
        Manga   interface{} `json:"manga"`
    } `json:"body"`
}
