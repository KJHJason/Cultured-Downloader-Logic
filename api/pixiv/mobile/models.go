package pixivmobile

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/ugoira"
)

type UserDetails struct {
	ProfileImageUrls struct {
		Px16X16   string `json:"px_16x16"`
		Px50X50   string `json:"px_50x50"`
		Px170X170 string `json:"px_170x170"`
	} `json:"profile_image_urls"`
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Account                string `json:"account"`
	MailAddress            string `json:"mail_address"`
	IsPremium              bool   `json:"is_premium"`
	XRestrict              int    `json:"x_restrict"` // 0: SFW, 1: R18, 2: R18/R18G
	IsMailAuthorized       bool   `json:"is_mail_authorized"`
	RequirePolicyAgreement bool   `json:"require_policy_agreement"`
}

type OauthJson struct {
	AccessToken  string      `json:"access_token"`
	ExpiresIn    int         `json:"expires_in"`
	TokenType    string      `json:"token_type"`
	Scope        string      `json:"scope"`
	RefreshToken string      `json:"refresh_token"`
	User         UserDetails `json:"user"`
	// Response struct {
	// 	AccessToken  string `json:"access_token"`
	// 	ExpiresIn    int    `json:"expires_in"`
	// 	TokenType    string `json:"token_type"`
	// 	Scope        string `json:"scope"`
	// 	RefreshToken string `json:"refresh_token"`
	// 	User         struct {
	// 		ProfileImageUrls struct {
	// 			Px16X16   string `json:"px_16x16"`
	// 			Px50X50   string `json:"px_50x50"`
	// 			Px170X170 string `json:"px_170x170"`
	// 		} `json:"profile_image_urls"`
	// 		ID                     string `json:"id"`
	// 		Name                   string `json:"name"`
	// 		Account                string `json:"account"`
	// 		MailAddress            string `json:"mail_address"`
	// 		IsPremium              bool   `json:"is_premium"`
	// 		XRestrict              int    `json:"x_restrict"` // 0: SFW, 1: R18, 2: R18/R18G
	// 		IsMailAuthorized       bool   `json:"is_mail_authorized"`
	// 		RequirePolicyAgreement bool   `json:"require_policy_agreement"`
	// 	} `json:"user"`
	// } `json:"response"`
}

type OauthFlowJson struct {
	RefreshToken string `json:"refresh_token"`
}

type UgoiraJson struct {
	Metadata struct {
		Frames ugoira.UgoiraFramesJson `json:"frames"`
		ZipUrls struct {
			Medium string `json:"medium"`
		} `json:"zip_urls"`
	} `json:"ugoira_metadata"`
}

type IllustJson struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`

	User struct {
		Name  string `json:"name"`
	} `json:"user"`

	MetaSinglePage struct {
		OriginalImageUrl string `json:"original_image_url"`
	} `json:"meta_single_page"`

	MetaPages []struct {
		ImageUrls struct {
			Original string `json:"original"`
		} `json:"image_urls"`
	} `json:"meta_pages"`
}

type ArtworkJson struct {
	Illust *IllustJson `json:"illust"`
}
type ArtworksJson struct {
	Illusts []*IllustJson `json:"illusts"`
	NextUrl *string                  `json:"next_url"`
}
