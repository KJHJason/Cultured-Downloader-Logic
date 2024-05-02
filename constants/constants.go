package constants

import (
	"fmt"
	"regexp"
	"runtime"

	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

const (
	DEBUG_MODE                     = false // Will save a copy of all JSON response from the API
	VERSION                        = "1.0.3"
	MAX_RETRY_DELAY                = 3
	MIN_RETRY_DELAY                = 1
	RETRY_COUNTER                  = 4
	MAX_CONCURRENT_DOWNLOADS       = 4
	PIXIV_MAX_CONCURRENT_DOWNLOADS = 3
	MAX_API_CALLS                  = 10
	CLI_REPO_URL                   = "https://api.github.com/repos/KJHJason/Cultured-Downloader-CLI/releases/latest"
	LOGIC_REPO_URL                 = "https://api.github.com/repos/KJHJason/Cultured-Downloader-Logic/releases/latest"

	PAGE_NUM_REGEX_STR = `[1-9]\d*(-[1-9]\d*)?`
	DOWNLOAD_TIMEOUT   = 25 * 60 // 25 minutes in seconds as downloads
	// can take quite a while for large files (especially for Pixiv)
	// However, the average max file size on these platforms is around 300MB.
	// Note: Fantia do have a max file size per post of 3GB if one paid extra for it.

	FANTIA                 = "fantia"
	FANTIA_TITLE           = "Fantia"
	FANTIA_URL             = "https://fantia.jp"
	FANTIA_RECAPTCHA_URL   = "https://fantia.jp/recaptcha"
	FANTIA_RANGE_SUPPORTED = true
	FANTIA_MAX_CONCURRENT  = 5
	FANTIA_POST_API_URL    = "https://fantia.jp/api/v1/posts/"

	PIXIV                 = "pixiv"
	PIXIV_MOBILE          = "pixiv_mobile"
	PIXIV_TITLE           = "Pixiv"
	PIXIV_PER_PAGE        = 60
	PIXIV_MOBILE_PER_PAGE = 30
	PIXIV_URL             = "https://www.pixiv.net"
	PIXIV_API_URL         = "https://www.pixiv.net/ajax"
	PIXIV_MOBILE_URL      = "https://app-api.pixiv.net"
	PIXIV_RANGE_SUPPORTED = true
	PIXIV_MAX_CONCURRENT  = 5

	PIXIV_FANBOX                 = "fanbox"
	PIXIV_FANBOX_TITLE           = "Pixiv Fanbox"
	PIXIV_FANBOX_URL             = "https://www.fanbox.cc"
	PIXIV_FANBOX_API_URL         = "https://api.fanbox.cc"
	PIXIV_FANBOX_RANGE_SUPPORTED = false
	PIXIV_FANBOX_MAX_CONCURRENT  = 2 // Pixiv Fanbox throttles the download speed


	KEMONO                       = "kemono"
	KEMONO_TITLE                 = "Kemono Party"
	KEMONO_PER_PAGE              = 50
	KEMONO_URL                   = "https://kemono.su"
	KEMONO_API_URL               = "https://kemono.su/api/v1"
	KEMONO_RANGE_SUPPORTED       = true
	KEMONO_BASE_REGEX_STR        = `https://kemono\.(?:party|su)/(?P<service>patreon|fanbox|gumroad|subscribestar|dlsite|fantia|boosty)/user/(?P<creatorId>[\w-]+)`
	KEMONO_POST_SUFFIX_REGEX_STR = `/post/(?P<postId>\d+)`
	KEMONO_SERVICE_GROUP_NAME    = "service"
	KEMONO_CREATOR_ID_GROUP_NAME = "creatorId"
	KEMONO_POST_ID_GROUP_NAME    = "postId"
	KEMONO_MAX_CONCURRENT        = 1 // Since Kemono server is very slow as of April 2024

	PASSWORD_FILENAME = "detected_passwords.txt"
	ATTACHMENT_FOLDER = "attachments"
	IMAGES_FOLDER     = "images"

	KEMONO_EMBEDS_FOLDER   = "embeds"
	KEMONO_CONTENT_FOLDER  = "post_content"

	GDRIVE_URL 	         = "https://drive.google.com"
	GDRIVE_FOLDER        = "gdrive"
	GDRIVE_FILENAME      = "detected_gdrive_links.txt"
	OTHER_LINKS_FILENAME = "detected_external_links.txt"
)

// Although the variables below are not
// constants, they are not supposed to be changed
var (
	USER_AGENT string
	GITHUB_VER_REGEX = regexp.MustCompile(`\d+\.\d+\.\d+`)

	PAGE_NUM_REGEX = regexp.MustCompile(
		fmt.Sprintf(`^%s$`, PAGE_NUM_REGEX_STR),
	)
	NUMBER_REGEX             = regexp.MustCompile(`^\d+$`)

	GDRIVE_URL_REGEX         = regexp.MustCompile(
		`https://drive\.google\.com/(?P<type>file/d|drive/(u/\d+/)?folders)/(?P<id>[\w-]+)`,
	)
	GDRIVE_REGEX_ID_INDEX   = GDRIVE_URL_REGEX.SubexpIndex("id")
	GDRIVE_REGEX_TYPE_INDEX = GDRIVE_URL_REGEX.SubexpIndex("type")

	FANTIA_IMAGE_URL_REGEX  = regexp.MustCompile(
		`original_url\":\"(?P<url>/posts/\d+/album_image\?query=[\w%-]*)\"`,
	)
	FANTIA_REGEX_URL_INDEX = FANTIA_IMAGE_URL_REGEX.SubexpIndex("url")

	// Since the URLs below will be redirected to Fantia's AWS S3 URL, 
	// we need to use HTTP/2 as it is not supported by HTTP/3 yet.
	FANTIA_ALBUM_URL = regexp.MustCompile(
		`^https://fantia.jp/posts/[\d]+/album_image`,
	)
	FANTIA_DOWNLOAD_URL = regexp.MustCompile(
		`^https://fantia.jp/posts/[\d]+/download/[\d]+`,
	)

	HTTP3_SUPPORT_ARR = [...]string{
		"https://www.pixiv.net",
		"https://app-api.pixiv.net",

		"https://www.google.com",
		"https://drive.google.com",
	}

	// For Pixiv Fanbox
	PASSWORD_TEXTS              = []string{"パス", "Pass", "pass", "密码"}
	EXTERNAL_DOWNLOAD_PLATFORMS = []string{"mega", "gigafile", "dropbox", "mediafire"}

	// For Kemono
	KEMONO_POST_URL_REGEX = regexp.MustCompile(
		fmt.Sprintf(
			`^%s%s$`,
			KEMONO_BASE_REGEX_STR,
			KEMONO_POST_SUFFIX_REGEX_STR,
		),
	)
	KEMONO_POST_URL_REGEX_SERVICE_IDX    = KEMONO_POST_URL_REGEX.SubexpIndex(KEMONO_SERVICE_GROUP_NAME)
	KEMONO_POST_URL_REGEX_CREATOR_ID_IDX = KEMONO_POST_URL_REGEX.SubexpIndex(KEMONO_CREATOR_ID_GROUP_NAME)
	KEMONO_POST_URL_REGEX_POST_ID_IDX    = KEMONO_POST_URL_REGEX.SubexpIndex(KEMONO_POST_ID_GROUP_NAME)

	KEMONO_CREATOR_URL_REGEX = regexp.MustCompile(
		fmt.Sprintf(
			`^%s$`,
			KEMONO_BASE_REGEX_STR,
		),
	)
	KEMONO_CREATOR_URL_REGEX_SERVICE_IDX    = KEMONO_CREATOR_URL_REGEX.SubexpIndex(KEMONO_SERVICE_GROUP_NAME)
	KEMONO_CREATOR_URL_REGEX_CREATOR_ID_IDX = KEMONO_CREATOR_URL_REGEX.SubexpIndex(KEMONO_CREATOR_ID_GROUP_NAME)
)

func init() {
	var userAgent = map[string]string{
		"linux":   "Mozilla/5.0 (X11; Linux x86_64)",
		"darwin":  "Mozilla/5.0 (Macintosh; Intel Mac OS X 12_6)",
		"windows": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	}
	userAgentOS, ok := userAgent[runtime.GOOS]
	if !ok {
		panic(
			fmt.Errorf(
				"error %d: Failed to get user agent OS as your OS, %q, is not supported",
				errs.OS_ERROR,
				runtime.GOOS,
			),
		)
	}
	USER_AGENT = fmt.Sprintf("%s AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36", userAgentOS)
}
