package constants

import (
	"fmt"
	"regexp"
)

const (
	DEBUG_MODE             = false // Will save a copy of all JSON response from the API
	DEFAULT_PERMS          = 0755  // Owner: rwx, Group: rx, Others: rx
	VERSION                = "1.1.5"
	MAX_RETRY_DELAY        = 3
	MIN_RETRY_DELAY        = 1
	HTTP3_MAX_RETRY        = 2
	RETRY_COUNTER          = 4
	GITHUB_API_URL_FORMAT  = "https://api.github.com/repos/%s/releases/latest"
	MAIN_REPO_NAME         = "KJHJason/Cultured-Downloader"
	CLI_REPO_NAME          = "KJHJason/Cultured-Downloader-CLI"
	LOGIC_REPO_NAME        = "KJHJason/Cultured-Downloader-Logic"
	EN                     = "en"
	JP                     = "ja"
	FFMPEG_MAX_CONCURRENCY = 2

	ERR_RECAPTCHA_STR = "recaptcha detected for the current session"

	PAGE_NUM_REGEX_STR            = `[1-9]\d*(?:-[1-9]\d*)?`
	PAGE_NUM_IDX_NAME             = "pageNum"
	PAGE_NUM_WITH_INPUT_REGEX_STR = `(?:;(?P<pageNum>[1-9]\d*(?:-[1-9]\d*)?))?`

	DEFAULT_HEAD_REQ_TIMEOUT = 15      // in seconds
	DOWNLOAD_TIMEOUT         = 25 * 60 // 25 minutes in seconds as downloads
	// can take quite a while for large files (especially for Pixiv)
	// However, the average max file size on these platforms is around 300MB.
	// Note: Fantia do have a max file size per post of 3GB if one paid extra for it.

	FANTIA                      = "fantia"
	FANTIA_TITLE                = "Fantia"
	FANTIA_URL                  = "https://fantia.jp"
	FANTIA_RECAPTCHA_URL        = "https://fantia.jp/recaptcha"
	FANTIA_RANGE_SUPPORTED      = true
	FANTIA_MAX_CONCURRENT       = 5
	FANTIA_POST_API_URL         = "https://fantia.jp/api/v1/posts/"
	FANTIA_CAPTCHA_BTN_SELECTOR = `//input[@name='commit']`
	FANTIA_CAPTCHA_TIMEOUT      = 45
	FANTIA_POST_BLOG_DIR_NAME   = "blog_contents"

	PIXIV                          = "pixiv"
	PIXIV_MOBILE                   = "pixiv_mobile"
	PIXIV_TITLE                    = "Pixiv"
	PIXIV_MOBILE_TITLE             = "Pixiv (Mobile)"
	PIXIV_PER_PAGE                 = 60
	PIXIV_MOBILE_PER_PAGE          = 30
	PIXIV_URL                      = "https://www.pixiv.net"
	PIXIV_API_URL                  = "https://www.pixiv.net/ajax"
	PIXIV_MOBILE_URL               = "https://app-api.pixiv.net"
	PIXIV_RANGE_SUPPORTED          = true
	PIXIV_MAX_CONCURRENCY          = 1 // Not used rn as the Pixiv download is being done sequentially instead of concurrently
	PIXIV_MAX_DOWNLOAD_CONCURRENCY = 2 // https://i.pixiv.net not using Cloudflare's proxy

	PIXIV_MOBILE_BASE_URL       = PIXIV_MOBILE_URL
	PIXIV_MOBILE_CLIENT_ID      = "MOBrBDS8blbauoSck0ZfDbtuzpyT"
	PIXIV_MOBILE_CLIENT_SECRET  = "lsACyCD94FhDUtGTXi3QzcFE2uU1hqtDaKeqrdwj"
	PIXIV_MOBILE_USER_AGENT     = "PixivIOSApp/7.13.3 (iOS 14.6; iPhone13,2)"
	PIXIV_MOBILE_AUTH_TOKEN_URL = "https://oauth.secure.pixiv.net/auth/token"
	PIXIV_MOBILE_LOGIN_URL      = PIXIV_MOBILE_BASE_URL + "/web/v1/login"
	PIXIV_MOBILE_REDIRECT_URL   = PIXIV_MOBILE_BASE_URL + "/web/v1/users/auth/pixiv/callback"

	PIXIV_MOBILE_UGOIRA_URL        = PIXIV_MOBILE_BASE_URL + "/v1/ugoira/metadata"
	PIXIV_MOBILE_ARTWORK_URL       = PIXIV_MOBILE_BASE_URL + "/v1/illust/detail"
	PIXIV_MOBILE_ARTIST_POSTS_URL  = PIXIV_MOBILE_BASE_URL + "/v1/user/illusts"
	PIXIV_MOBILE_ILLUST_SEARCH_URL = PIXIV_MOBILE_BASE_URL + "/v1/search/illust"

	PIXIV_FANBOX                      = "fanbox"
	PIXIV_FANBOX_TITLE                = "Pixiv Fanbox"
	PIXIV_FANBOX_CREATOR_ID_REGEX_STR = `[\w&.-]+`
	PIXIV_FANBOX_URL                  = "https://www.fanbox.cc"
	PIXIV_FANBOX_API_URL              = "https://api.fanbox.cc"
	PIXIV_FANBOX_RANGE_SUPPORTED      = false
	PIXIV_FANBOX_MAX_CONCURRENT       = 2 // Pixiv Fanbox throttles the download speed

	KEMONO                       = "kemono"
	KEMONO_TITLE                 = "Kemono"
	KEMONO_PER_PAGE              = 50
	KEMONO_URL                   = "https://kemono.su"
	KEMONO_API_URL               = "https://kemono.su/api/v1"
	KEMONO_RANGE_SUPPORTED       = true
	KEMONO_HEAD_REQ_TIMEOUT      = 60 // to avoid the common issue of context deadline exceeded (Client.Timeout exceeded while awaiting headers) as the default 15s is too short
	KEMONO_BASE_REGEX_STR        = `https://kemono\.su/(?P<service>patreon|fanbox|gumroad|subscribestar|dlsite|fantia|boosty)/user/(?P<creatorId>[\w-]+)`
	KEMONO_POST_SUFFIX_REGEX_STR = `/post/(?P<postId>\d+)`
	KEMONO_SERVICE_GROUP_NAME    = "service"
	KEMONO_CREATOR_ID_GROUP_NAME = "creatorId"
	KEMONO_POST_ID_GROUP_NAME    = "postId"
	KEMONO_MAX_CONCURRENT        = 1 // Since Kemono server is very slow as of April 2024

	PASSWORD_FILENAME = "detected_passwords.txt"
	ATTACHMENT_FOLDER = "attachments"
	IMAGES_FOLDER     = "images"

	KEMONO_EMBEDS_FOLDER  = "embeds"
	KEMONO_CONTENT_FOLDER = "post_content"

	GDRIVE_URL                    = "https://drive.google.com"
	GDRIVE_FILE_API_URL           = "https://www.googleapis.com/drive/v3/files"
	GDRIVE_FOLDER                 = "gdrive"
	GDRIVE_FILENAME               = "detected_gdrive_links.txt"
	GDRIVE_HTTP3_SUPPORTED        = true
	GDRIVE_ERROR_FILENAME         = "gdrive_download.log"
	GDRIVE_BASE_API_KEY_REGEX_STR = `AIza[\w-]{35}`
	GDRIVE_MAX_CONCURRENCY        = 2
	GDRIVE_OAUTH_MAX_CONCURRENCY  = 4

	// file fields to fetch from GDrive API:
	// https://developers.google.com/drive/api/v3/reference/files
	GDRIVE_FILE_FIELDS   = "id,name,size,mimeType,md5Checksum"
	GDRIVE_FOLDER_FIELDS = "nextPageToken,files(id,name,size,mimeType,md5Checksum)"

	OTHER_LINKS_FILENAME = "detected_external_links.txt"
)

// Although the variables below are not
// constants, they are not supposed to be changed
var (
	// General
	GITHUB_VER_REGEX = regexp.MustCompile(`\d+\.\d+\.\d+`)

	PAGE_NUM_REGEX = regexp.MustCompile(
		fmt.Sprintf(`^%s$`, PAGE_NUM_REGEX_STR),
	)
	NUMBER_REGEX     = regexp.MustCompile(`^\d+$`)
	PASSWORD_TEXTS   = [...]string{"パス", "Pass", "pass", "密码"}
	PASSWORD_REGEXES = [...]*regexp.Regexp{
		regexp.MustCompile(`ダウンロード(?:<\/span>)?<\/a><\/p><p>[\w-]+<\/p>`),
		regexp.MustCompile(`ダウンロード\n([\w-]+)\n`),
	}
	EXTERNAL_DOWNLOAD_PLATFORMS = [...]string{"mega", "gigafile", "dropbox", "mediafire"}

	// For GDrive
	GDRIVE_URL_REGEX = regexp.MustCompile(
		`https://drive\.google\.com/(?P<type>file/d|drive/(u/\d+/)?folders)/(?P<id>[\w-]+)`,
	)
	GDRIVE_REGEX_ID_IDX   = GDRIVE_URL_REGEX.SubexpIndex("id")
	GDRIVE_REGEX_TYPE_IDX = GDRIVE_URL_REGEX.SubexpIndex("type")

	GDRIVE_API_KEY_REGEX = regexp.MustCompile(
		fmt.Sprintf(`^%s$`, GDRIVE_BASE_API_KEY_REGEX_STR),
	)
	GDRIVE_API_KEY_PARAM_REGEX = regexp.MustCompile(
		fmt.Sprintf(`key=%s`, GDRIVE_BASE_API_KEY_REGEX_STR),
	)

	// For Fantia
	FANTIA_COMMENT_IMAGE_URL_REGEX = regexp.MustCompile(
		// Note: the "original_url" field points to the "url" field in the JSON response
		`"url":"(?P<url>https://cc\.fantia\.jp/uploads/album_image/file/[\d]+/[\w-]+\.(?P<ext>[a-z]+)\?[^"]+)"`,
	)
	FANTIA_COMMENT_REGEX_EXT_IDX = FANTIA_IMAGE_URL_REGEX.SubexpIndex("ext")
	FANTIA_COMMENT_REGEX_URL_IDX = FANTIA_IMAGE_URL_REGEX.SubexpIndex("url")

	FANTIA_IMAGE_URL_REGEX = regexp.MustCompile(
		`^https://cc\.fantia\.jp/uploads/post_content_photo/file/[\d]+/[\w-]+\.(?P<ext>[a-z]+)\?`,
	)
	FANTIA_IMAGE_URL_REGEX_EXT_IDX = FANTIA_IMAGE_URL_REGEX.SubexpIndex("ext")

	// Since the URLs below will be redirected to Fantia's AWS S3 URL,
	// we need to use HTTP/2 as it is not supported by HTTP/3 yet.
	FANTIA_ALBUM_URL = regexp.MustCompile(
		`^https://fantia.jp/posts/[\d]+/album_image`,
	)
	FANTIA_DOWNLOAD_URL = regexp.MustCompile(
		`^https://fantia.jp/posts/[\d]+/download/[\d]+`,
	)

	HTTP3_SUPPORT_ARR = [...]string{
		PIXIV_URL,
		PIXIV_MOBILE_URL,

		"https://www.google.com",
		GDRIVE_URL,
	}

	// For Pixiv
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

	// For Kemono
	KEMONO_IMG_SRC_TAG_REGEX     = regexp.MustCompile(`(?i)<img[^>]+src=(?:\\)?"(?P<imgSrc>[^">]+)(?:\\)?"[^>]*>`)
	KEMONO_IMG_SRC_TAG_REGEX_IDX = KEMONO_IMG_SRC_TAG_REGEX.SubexpIndex("imgSrc")

	// For Kemono input validations
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
			// ^https://kemono\.su/(?P<service>patreon|fanbox|gumroad|subscribestar|dlsite|fantia|boosty)/user/(?P<creatorId>[\w-]+)(?:;(?P<pageNum>[1-9]\d*(?:-[1-9]\d*)?))?$
			`^%s%s$`,
			KEMONO_BASE_REGEX_STR,
			PAGE_NUM_WITH_INPUT_REGEX_STR,
		),
	)
	KEMONO_CREATOR_URL_REGEX_SERVICE_IDX    = KEMONO_CREATOR_URL_REGEX.SubexpIndex(KEMONO_SERVICE_GROUP_NAME)
	KEMONO_CREATOR_URL_REGEX_CREATOR_ID_IDX = KEMONO_CREATOR_URL_REGEX.SubexpIndex(KEMONO_CREATOR_ID_GROUP_NAME)
	KEMONO_CREATOR_URL_REGEX_PAGE_NUM_IDX   = KEMONO_CREATOR_URL_REGEX.SubexpIndex(PAGE_NUM_IDX_NAME)

	// For Fantia input validations
	FANTIA_POST_URL_REGEX = regexp.MustCompile(
		`^https://fantia.jp/posts/(?P<id>\d+)$`,
	)
	FANTIA_POST_ID_IDX = FANTIA_POST_URL_REGEX.SubexpIndex("id")

	FANTIA_CREATOR_URL_REGEX = regexp.MustCompile(
		fmt.Sprintf(
			// ^https://fantia\.jp/fanclubs/(?P<id>\d+)(?:/posts)?(?:;(?P<pageNum>[1-9]\d*(?:-[1-9]\d*)?))?$
			`^https://fantia\.jp/fanclubs/(?P<id>\d+)(?:/posts)?%s$`,
			PAGE_NUM_WITH_INPUT_REGEX_STR,
		),
	)
	FANTIA_CREATOR_ID_IDX       = FANTIA_CREATOR_URL_REGEX.SubexpIndex("id")
	FANTIA_CREATOR_PAGE_NUM_IDX = FANTIA_CREATOR_URL_REGEX.SubexpIndex(PAGE_NUM_IDX_NAME)

	// For Pixiv Fanbox input validations
	PIXIV_FANBOX_CREATOR_ID_REGEX = regexp.MustCompile(
		fmt.Sprintf(`^%s$`, PIXIV_FANBOX_CREATOR_ID_REGEX_STR),
	)

	PIXIV_FANBOX_POST_URL_REGEX1 = regexp.MustCompile(
		fmt.Sprintf(
			`^https://www\.fanbox\.cc/@%s/posts/(?P<id>\d+)$`,
			PIXIV_FANBOX_CREATOR_ID_REGEX_STR,
		),
	)
	PIXIV_FANBOX_POST_ID_IDX1 = PIXIV_FANBOX_POST_URL_REGEX1.SubexpIndex("id")

	PIXIV_FANBOX_POST_URL_REGEX2 = regexp.MustCompile(
		fmt.Sprintf(
			`^https://%s\.fanbox\.cc/posts/(?P<id>\d+)$`,
			PIXIV_FANBOX_CREATOR_ID_REGEX_STR,
		),
	)
	PIXIV_FANBOX_POST_ID_IDX2 = PIXIV_FANBOX_POST_URL_REGEX2.SubexpIndex("id")

	PIXIV_FANBOX_CREATOR_URL_REGEX1 = regexp.MustCompile(
		fmt.Sprintf(
			// ^https://(?P<id>[\w&.-]+)\.fanbox\.cc(?:/(?:posts)?)?(?:;(?P<pageNum>[1-9]\d*(?:-[1-9]\d*)?))?$
			`^https://(?P<id>%s)\.fanbox\.cc(?:/(?:posts)?)?%s$`,
			PIXIV_FANBOX_CREATOR_ID_REGEX_STR,
			PAGE_NUM_WITH_INPUT_REGEX_STR,
		),
	)
	PIXIV_FANBOX_CREATOR_ID_IDX1       = PIXIV_FANBOX_CREATOR_URL_REGEX1.SubexpIndex("id")
	PIXIV_FANBOX_CREATOR_PAGE_NUM_IDX1 = PIXIV_FANBOX_CREATOR_URL_REGEX1.SubexpIndex(PAGE_NUM_IDX_NAME)

	PIXIV_FANBOX_CREATOR_URL_REGEX2 = regexp.MustCompile(
		fmt.Sprintf(
			// ^https://www\.fanbox\.cc/@(?P<id>[\w&.-]+)(?:/posts)?(?:;(?P<pageNum>[1-9]\d*(?:-[1-9]\d*)?))?$
			`^https://www\.fanbox\.cc/@(?P<id>%s)(?:/posts)?%s$`,
			PIXIV_FANBOX_CREATOR_ID_REGEX_STR,
			PAGE_NUM_WITH_INPUT_REGEX_STR,
		),
	)
	PIXIV_FANBOX_CREATOR_ID_IDX2       = PIXIV_FANBOX_CREATOR_URL_REGEX2.SubexpIndex("id")
	PIXIV_FANBOX_CREATOR_PAGE_NUM_IDX2 = PIXIV_FANBOX_CREATOR_URL_REGEX2.SubexpIndex(PAGE_NUM_IDX_NAME)

	// For Pixiv input validations
	// can be illust or manga
	PIXIV_ARTWORK_URL_REGEX = regexp.MustCompile(
		`^https://www\.pixiv\.net/(?:en/)?artworks/(?P<id>\d+)$`,
	)
	PIXIV_ARTWORK_ID_IDX = PIXIV_ARTWORK_URL_REGEX.SubexpIndex("id")

	PIXIV_ARTIST_URL_REGEX = regexp.MustCompile(
		fmt.Sprintf(
			// ^https://www\.pixiv\.net/(?:en/)?users/(?P<id>\d+)(?:;(?P<pageNum>[1-9]\d*(?:-[1-9]\d*)?))?$
			`^https://www\.pixiv\.net/(?:en/)?users/(?P<id>\d+)%s$`,
			PAGE_NUM_WITH_INPUT_REGEX_STR,
		),
	)
	PIXIV_ARTIST_ID_IDX       = PIXIV_ARTIST_URL_REGEX.SubexpIndex("id")
	PIXIV_ARTIST_PAGE_NUM_IDX = PIXIV_ARTIST_URL_REGEX.SubexpIndex(PAGE_NUM_IDX_NAME)

	PIXIV_OAUTH_CODE_REGEX = regexp.MustCompile(`^[\w-]{43}$`)
)
