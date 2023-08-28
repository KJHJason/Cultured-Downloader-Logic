package constants

import (
	"fmt"
	"regexp"
	"net/http"
)

const (
	DEBUG_MODE                     = false // Will save a copy of all JSON response from the API
	VERSION                        = "1.0.0"
	MAX_RETRY_DELAY                = 3
	MIN_RETRY_DELAY                = 1
	RETRY_COUNTER                  = 4
	MAX_CONCURRENT_DOWNLOADS       = 4
	PIXIV_MAX_CONCURRENT_DOWNLOADS = 3
	MAX_API_CALLS                  = 10

	PAGE_NUM_REGEX_STR = `[1-9]\d*(-[1-9]\d*)?`
	DOWNLOAD_TIMEOUT   = 25 * 60 // 25 minutes in seconds as downloads
	// can take quite a while for large files (especially for Pixiv)
	// However, the average max file size on these platforms is around 300MB.
	// Note: Fantia do have a max file size per post of 3GB if one paid extra for it.

	FANTIA               = "fantia"
	FANTIA_TITLE         = "Fantia"
	FANTIA_URL           = "https://fantia.jp"
	FANTIA_RECAPTCHA_URL = "https://fantia.jp/recaptcha"

	PIXIV            = "pixiv"
	PIXIV_MOBILE     = "pixiv_mobile"
	PIXIV_TITLE      = "Pixiv"
	PIXIV_PER_PAGE   = 60
	PIXIV_URL        = "https://www.pixiv.net"
	PIXIV_API_URL    = "https://www.pixiv.net/ajax"
	PIXIV_MOBILE_URL = "https://app-api.pixiv.net"

	PIXIV_FANBOX         = "fanbox"
	PIXIV_FANBOX_TITLE   = "Pixiv Fanbox"
	PIXIV_FANBOX_URL     = "https://www.fanbox.cc"
	PIXIV_FANBOX_API_URL = "https://api.fanbox.cc"

	KEMONO                      = "kemono"
	KEMONO_SESSION_COOKIE_NAME  = "session"
	KEMONO_COOKIE_DOMAIN        = "kemono.party"
	KEMONO_BACKUP               = "kemono_backup"
	KEMONO_COOKIE_BACKUP_DOMAIN = "kemono.su"
	KEMONO_TITLE                = "Kemono Party"
	KEMONO_PER_PAGE             = 50
	KEMONO_TLD                  = "party"
	KEMONO_BACKUP_TLD           = "su"
	KEMONO_URL                  = "https://kemono.party"
	KEMONO_API_URL              = "https://kemono.party/api"
	BACKUP_KEMONO_URL           = "https://kemono.su"
	BACKUP_KEMONO_API_URL       = "https://kemono.su/api"

	PASSWORD_FILENAME = "detected_passwords.txt"
	ATTACHMENT_FOLDER = "attachments"
	IMAGES_FOLDER     = "images"

	KEMONO_EMBEDS_FOLDER   = "embeds"
	KEMONO_CONTENT_FOLDER  = "post_content"

	GDRIVE_URL 	         = "https://drive.google.com"
	GDRIVE_FOLDER        = "gdrive"
	GDRIVE_FILENAME      = "detected_gdrive_links.txt"
	OTHER_LINKS_FILENAME = "detected_external_links.txt"

	// Progress Bar Map Key
	CAPTCHA_SOLVER_PROG_BAR = "captcha_solver_progress_bar"
	FANTIA_POST_PROG_BAR = "fantia_post_progress_bar"
	FANTIA_GET_POST_ID_PROG_BAR = "fantia_get_post_id_progress_bar"
	FANTIA_PROCESS_JSON_PROG_BAR = "fantia_process_json_progress_bar"
)

// For Fantia so far but can be used for other websites if required
type CAPTCHA_FN func(useHttp3 bool, sessionCookies []*http.Cookie, userAgent, url string) error

// Although the variables below are not
// constants, they are not supposed to be changed
var (
	USER_AGENT string

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

	// For Pixiv Fanbox
	PASSWORD_TEXTS              = []string{"パス", "Pass", "pass", "密码"}
	EXTERNAL_DOWNLOAD_PLATFORMS = []string{"mega", "gigafile", "dropbox", "mediafire"}
)
