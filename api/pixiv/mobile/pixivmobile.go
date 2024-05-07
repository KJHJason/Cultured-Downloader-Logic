package pixivmobile

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

const (
	BASE_URL       = constants.PIXIV_MOBILE_URL
	CLIENT_ID      = "MOBrBDS8blbauoSck0ZfDbtuzpyT"
	CLIENT_SECRET  = "lsACyCD94FhDUtGTXi3QzcFE2uU1hqtDaKeqrdwj"
	USER_AGENT     = "PixivIOSApp/7.13.3 (iOS 14.6; iPhone13,2)"
	AUTH_TOKEN_URL = "https://oauth.secure.pixiv.net/auth/token"
	LOGIN_URL      = BASE_URL + "/web/v1/login"
	REDIRECT_URL   = BASE_URL + "/web/v1/users/auth/pixiv/callback"
)
