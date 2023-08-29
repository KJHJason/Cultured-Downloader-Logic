package kemono

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/KJHJason/Cultured-Downloader-Logic/notify"
	"github.com/fatih/color"
)

const (
	BASE_REGEX_STR             = `https://kemono\.(?P<topLevelDomain>party|su)/(?P<service>patreon|fanbox|gumroad|subscribestar|dlsite|fantia|boosty)/user/(?P<creatorId>[\w-]+)`
	BASE_POST_SUFFIX_REGEX_STR = `/post/(?P<postId>\d+)`
	TLD_GROUP_NAME             = "topLevelDomain"
	SERVICE_GROUP_NAME         = "service"
	CREATOR_ID_GROUP_NAME      = "creatorId"
	POST_ID_GROUP_NAME         = "postId"
	API_MAX_CONCURRENT         = 3
)

var (
	POST_URL_REGEX = regexp.MustCompile(
		fmt.Sprintf(
			`^%s%s$`,
			BASE_REGEX_STR,
			BASE_POST_SUFFIX_REGEX_STR,
		),
	)
	POST_URL_REGEX_TLD_INDEX = POST_URL_REGEX.SubexpIndex(TLD_GROUP_NAME)
	POST_URL_REGEX_SERVICE_INDEX    = POST_URL_REGEX.SubexpIndex(SERVICE_GROUP_NAME)
	POST_URL_REGEX_CREATOR_ID_INDEX = POST_URL_REGEX.SubexpIndex(CREATOR_ID_GROUP_NAME)
	POST_URL_REGEX_POST_ID_INDEX    = POST_URL_REGEX.SubexpIndex(POST_ID_GROUP_NAME)

	CREATOR_URL_REGEX = regexp.MustCompile(
		fmt.Sprintf(
			`^%s$`,
			BASE_REGEX_STR,
		),
	)
	CREATOR_URL_REGEX_TLD_INDEX = CREATOR_URL_REGEX.SubexpIndex(TLD_GROUP_NAME)
	CREATOR_URL_REGEX_SERVICE_INDEX    = CREATOR_URL_REGEX.SubexpIndex(SERVICE_GROUP_NAME)
	CREATOR_URL_REGEX_CREATOR_ID_INDEX = CREATOR_URL_REGEX.SubexpIndex(CREATOR_ID_GROUP_NAME)
)

type KemonoDl struct {
	CreatorUrls     []string
	CreatorPageNums []string
	CreatorsToDl    []*KemonoCreatorToDl

	PostUrls  []string
	PostsToDl []*KemonoPostToDl

	DlFav bool
}

func ProcessCreatorUrls(creatorUrls []string, pageNums []string) []*KemonoCreatorToDl {
	creatorsToDl := make([]*KemonoCreatorToDl, len(creatorUrls))
	for i, creatorUrl := range creatorUrls {
		matched := CREATOR_URL_REGEX.FindStringSubmatch(creatorUrl)
		creatorsToDl[i] = &KemonoCreatorToDl{
			Service:   matched[CREATOR_URL_REGEX_SERVICE_INDEX],
			CreatorId: matched[CREATOR_URL_REGEX_CREATOR_ID_INDEX],
			PageNum:   pageNums[i],
			Tld:       matched[CREATOR_URL_REGEX_TLD_INDEX],
		}
	}

	return creatorsToDl
}

func ProcessPostUrls(postUrls []string) []*KemonoPostToDl {
	postsToDl := make([]*KemonoPostToDl, len(postUrls))
	for i, postUrl := range postUrls {
		matched := POST_URL_REGEX.FindStringSubmatch(postUrl)
		postsToDl[i] = &KemonoPostToDl{
			Service:   matched[POST_URL_REGEX_SERVICE_INDEX],
			CreatorId: matched[POST_URL_REGEX_CREATOR_ID_INDEX],
			PostId:    matched[POST_URL_REGEX_POST_ID_INDEX],
			Tld:       matched[POST_URL_REGEX_TLD_INDEX],
		}
	}

	return postsToDl
}

// RemoveDuplicates removes duplicate creators and posts from the slice
func (k *KemonoDl) RemoveDuplicates() {
	if len(k.CreatorsToDl) > 0 {
		newCreatorSlice := make([]*KemonoCreatorToDl, 0, len(k.CreatorsToDl))
		seen := make(map[string]struct{})
		for _, creator := range k.CreatorsToDl {
			key := fmt.Sprintf("%s/%s", creator.Service, creator.CreatorId)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			newCreatorSlice = append(newCreatorSlice, creator)
		}
		k.CreatorsToDl = newCreatorSlice
	}

	if len(k.PostsToDl) == 0 {
		return
	}
	newPostSlice := make([]*KemonoPostToDl, 0, len(k.PostsToDl))
	seen := make(map[string]struct{})
	for _, post := range k.PostsToDl {
		key := fmt.Sprintf("%s/%s/%s", post.Service, post.CreatorId, post.PostId)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		newPostSlice = append(newPostSlice, post)
	}
	k.PostsToDl = newPostSlice
}

func (k *KemonoDl) ValidateArgs() {
	valid, outlier := api.SliceMatchesRegex(CREATOR_URL_REGEX, k.CreatorUrls)
	if !valid {
		color.Red(
			fmt.Sprintf(
				"kemono error %d: invalid creator URL found for kemono party: %s",
				constants.INPUT_ERROR,
				outlier,
			),
		)
		os.Exit(1)
	}

	valid, outlier = api.SliceMatchesRegex(POST_URL_REGEX, k.PostUrls)
	if !valid {
		color.Red(
			fmt.Sprintf(
				"kemono error %d: invalid post URL found for kemono party: %s",
				constants.INPUT_ERROR,
				outlier,
			),
		)
		os.Exit(1)
	}

	if len(k.CreatorUrls) > 0 {
		if len(k.CreatorPageNums) == 0 {
			k.CreatorPageNums = make([]string, len(k.CreatorUrls))
		} else {
			api.ValidatePageNumInput(
				len(k.CreatorUrls),
				k.CreatorPageNums,
				[]string{
					"Number of creator URL(s) and page numbers must be equal.",
				},
			)
		}
		creatorsToDl := ProcessCreatorUrls(k.CreatorUrls, k.CreatorPageNums)
		k.CreatorsToDl = append(k.CreatorsToDl, creatorsToDl...)
		k.CreatorUrls = nil
		k.CreatorPageNums = nil
	}
	if len(k.PostUrls) > 0 {
		postsToDl := ProcessPostUrls(k.PostUrls)
		k.PostsToDl = append(k.PostsToDl, postsToDl...)
		k.PostUrls = nil
	}
	k.RemoveDuplicates()
}

// KemonoDlOptions is the struct that contains the arguments for Kemono download options.
type KemonoDlOptions struct {
	DlAttachments bool
	DlGdrive      bool

	Configs *configs.Config

	// GdriveClient is the Google Drive client to be
	// used in the download process for Pixiv Fanbox posts
	GdriveClient *gdrive.GDrive

	SessionCookieId string
	SessionCookies  []*http.Cookie

	Notifier       notify.Notifier

	// Prog bars
	PostProgBar  progress.Progress
	GetCreatorPostProgBar progress.Progress
	ProcessJsonProgBar progress.Progress
	GetFavouritesPostProgBar progress.Progress
}

// ValidateArgs validates the session cookie ID of the Kemono account to download from.
// It also validates the Google Drive client if the user wants to download to Google Drive.
//
// Should be called after initialising the struct.
func (k *KemonoDlOptions) ValidateArgs(userAgent string) error {
	if k.SessionCookieId != "" {
		if cookie, err := api.VerifyAndGetCookie(constants.KEMONO, k.SessionCookieId, userAgent); err != nil {
			return err
		} else {
			k.SessionCookies = []*http.Cookie{
				cookie,
			}
		}
	} else {
		return fmt.Errorf("kemono error %d: session cookie ID is required", constants.INPUT_ERROR)
	}

	if k.DlGdrive && k.GdriveClient == nil {
		k.DlGdrive = false
	} else if !k.DlGdrive && k.GdriveClient != nil {
		k.GdriveClient = nil
	}
	return nil
}
