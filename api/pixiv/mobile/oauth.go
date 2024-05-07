package pixivmobile

import (
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/models"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

type PixivOAuthTokenInfo struct {
	AccessToken  string    // The access token that will be used to communicate with the Pixiv's Mobile API
	ExpiresAt    time.Time // The time when the access token expires
}

// Perform a S256 transformation method on a byte array
func S256(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

var pixivOauthCodeRegex = regexp.MustCompile(`^[\w-]{43}$`)

// Start the OAuth flow to get the refresh token
func GetOAuthURL() (url string, codeVerifier string) {
	// create a random 32 bytes that is cryptographically secure
	codeVerifierBytes := make([]byte, 32)
	_, err := cryptorand.Read(codeVerifierBytes)
	if err != nil {
		// should never happen but just in case
		panic(
			fmt.Sprintf(
				"pixiv mobile error %d: failed to generate random bytes, more info => %v",
				errs.DEV_ERROR,
				err,
			),
		)
	}
	codeVerifier = base64.RawURLEncoding.EncodeToString(codeVerifierBytes)
	codeChallenge := S256([]byte(codeVerifier))

	loginParams := map[string]string{
		"code_challenge":        codeChallenge,
		"code_challenge_method": "S256",
		"client":                "pixiv-android",
	}
	return LOGIN_URL + "?" + httpfuncs.ParamsToString(loginParams), codeVerifier
}

func VerifyOAuthCode(code, codeVerifier string, timeout int) (string, error) {
	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV_MOBILE, true)

	if !pixivOauthCodeRegex.MatchString(code) {
		return "", fmt.Errorf(
			"pixiv mobile error %d: invalid code format, please check if the code is correct",
			errs.INPUT_ERROR,
		)
	}

	res, err := httpfuncs.CallRequestWithData(
		&httpfuncs.RequestArgs{
			Url:         AUTH_TOKEN_URL,
			Method:      "POST",
			Timeout:     timeout,
			CheckStatus: true,
			UserAgent:   "PixivAndroidApp/5.0.234 (Android 11; Pixel 5)",
			Http2:       !useHttp3,
			Http3:       useHttp3,
		},
		map[string]string{
			"client_id":      CLIENT_ID,
			"client_secret":  CLIENT_SECRET,
			"code":           code,
			"code_verifier":  codeVerifier,
			"grant_type":     "authorization_code",
			"include_policy": "true",
			"redirect_uri":   REDIRECT_URL,
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"pixiv mobile error %d: failed to verify code, please ensure the code is correct and try again",
			errs.INPUT_ERROR,
		)
	}

	var oauthJson models.PixivOauthFlowJson
	if err := httpfuncs.LoadJsonFromResponse(res, &oauthJson); err != nil {
		return "", err
	}
	return oauthJson.RefreshToken, nil
}

// Refresh the access token
func RefreshAccessToken(timeout int, refreshToken string) (*PixivOAuthTokenInfo, error) {
	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV_MOBILE, true)
	res, err := httpfuncs.CallRequestWithData(
		&httpfuncs.RequestArgs{
			Url:       AUTH_TOKEN_URL,
			Method:    "POST",
			Timeout:   timeout,
			UserAgent: USER_AGENT,
			Http2:     !useHttp3,
			Http3:     useHttp3,
		},
		map[string]string{
			"client_id":      CLIENT_ID,
			"client_secret":  CLIENT_SECRET,
			"grant_type":     "refresh_token",
			"include_policy": "true",
			"refresh_token":   refreshToken,
		},
	)
	if err != nil || res.StatusCode != 200 {
		if err == nil {
			res.Body.Close()
			err = fmt.Errorf(
				"pixiv mobile error %d: failed to refresh token due to %s response from Pixiv\n"+
					"Please check your refresh token and try again or use the \"-pixiv_start_oauth\" flag to get a new refresh token",
				errs.RESPONSE_ERROR,
				res.Status,
			)
		} else {
			err = fmt.Errorf(
				"pixiv mobile error %d: failed to refresh token due to %v\n"+
					"Please check your internet connection and try again",
				errs.CONNECTION_ERROR,
				err,
			)
		}
		return nil, err
	}

	var oauthJson models.PixivOauthJson
	if err := httpfuncs.LoadJsonFromResponse(res, &oauthJson); err != nil {
		return nil, err
	}

	expiresIn := oauthJson.ExpiresIn - 15 // usually 3600 but minus 15 seconds to be safe
	oauthInfo := PixivOAuthTokenInfo{
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
		AccessToken: oauthJson.AccessToken,
	}
	return &oauthInfo, nil
}

// Reads the response JSON and checks if the access token has expired,
// if so, refreshes the access token for future requests.
//
// Returns a boolean indicating if the access token was refreshed.
// func (pixiv *PixivMobile) refreshTokenIfReq() (bool, error) {
// 	if pixiv.accessTokenMap.accessToken != "" && pixiv.accessTokenMap.expiresAt.After(time.Now()) {
// 		return false, nil
// 	}

// 	err := pixiv.refreshAccessToken()
// 	if err != nil {
// 		return true, err
// 	}
// 	return true, nil
// }
