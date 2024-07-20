package pixivmobile

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

type OAuthTokenInfo struct {
	AccessToken string    // The access token that will be used to communicate with the Pixiv's Mobile API
	ExpiresAt   time.Time // The time when the access token expires
}

// Perform a S256 transformation method on a byte array
func S256(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

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
				cdlerrors.DEV_ERROR,
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
	return constants.PIXIV_MOBILE_LOGIN_URL + "?" + httpfuncs.ParamsToString(loginParams), codeVerifier
}

func VerifyOAuthCode(code, codeVerifier string, timeout int) (string, error) {
	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV_MOBILE, true)

	if !constants.PIXIV_OAUTH_CODE_REGEX.MatchString(code) {
		return "", fmt.Errorf(
			"pixiv mobile error %d: invalid code format, please check if the code is correct",
			cdlerrors.INPUT_ERROR,
		)
	}

	res, err := httpfuncs.CallRequestWithData(
		&httpfuncs.RequestArgs{
			Url:         constants.PIXIV_MOBILE_AUTH_TOKEN_URL,
			Method:      "POST",
			Timeout:     timeout,
			CheckStatus: true,
			UserAgent:   "PixivAndroidApp/5.0.234 (Android 11; Pixel 5)",
			Http2:       !useHttp3,
			Http3:       useHttp3,
		},
		map[string]string{
			"client_id":      constants.PIXIV_MOBILE_CLIENT_ID,
			"client_secret":  constants.PIXIV_MOBILE_CLIENT_SECRET,
			"code":           code,
			"code_verifier":  codeVerifier,
			"grant_type":     "authorization_code",
			"include_policy": "true",
			"redirect_uri":   constants.PIXIV_MOBILE_REDIRECT_URL,
		},
	)
	if err != nil {
		return "", fmt.Errorf(
			"pixiv mobile error %d: failed to verify code, please ensure the code is correct and try again",
			cdlerrors.INPUT_ERROR,
		)
	}

	var oauthJson OauthFlowJson
	if err := httpfuncs.LoadJsonFromResponse(res.Resp, &oauthJson); err != nil {
		return "", err
	}
	return oauthJson.RefreshToken, nil
}

// Refresh the access token
func RefreshAccessToken(ctx context.Context, timeout int, refreshToken string) (OAuthTokenInfo, *UserDetails, error) {
	if !constants.PIXIV_OAUTH_CODE_REGEX.MatchString(refreshToken) {
		return OAuthTokenInfo{}, nil, fmt.Errorf(
			"pixiv mobile error %d: invalid refresh token format, please check if the refresh token is correct",
			cdlerrors.INPUT_ERROR,
		)
	}

	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV_MOBILE, true)
	res, err := httpfuncs.CallRequestWithData(
		&httpfuncs.RequestArgs{
			Url:       constants.PIXIV_MOBILE_AUTH_TOKEN_URL,
			Method:    "POST",
			Timeout:   timeout,
			UserAgent: constants.PIXIV_MOBILE_USER_AGENT,
			Http2:     !useHttp3,
			Http3:     useHttp3,
			Context:   ctx,
		},
		map[string]string{
			"client_id":      constants.PIXIV_MOBILE_CLIENT_ID,
			"client_secret":  constants.PIXIV_MOBILE_CLIENT_SECRET,
			"grant_type":     "refresh_token",
			"include_policy": "true",
			"refresh_token":  refreshToken,
		},
	)
	if err != nil || res.Resp.StatusCode != 200 {
		if err == nil {
			res.Close()
			err = fmt.Errorf(
				"pixiv mobile error %d: failed to refresh token due to %s response from Pixiv\n"+
					"Please check your refresh token and try again or use the \"-pixiv_start_oauth\" flag to get a new refresh token",
				cdlerrors.RESPONSE_ERROR,
				res.Resp.Status,
			)
		} else {
			if errors.Is(err, context.Canceled) {
				return OAuthTokenInfo{}, nil, err
			}

			err = fmt.Errorf(
				"pixiv mobile error %d: failed to refresh token due to %w\n"+
					"Please check your internet connection and try again",
				cdlerrors.CONNECTION_ERROR,
				err,
			)
		}
		return OAuthTokenInfo{}, nil, err
	}

	var oauthJson OauthJson
	if err := httpfuncs.LoadJsonFromResponse(res.Resp, &oauthJson); err != nil {
		return OAuthTokenInfo{}, nil, err
	}

	expiresIn := oauthJson.ExpiresIn - 15 // usually 3600 but minus 15 seconds to be safe
	oauthInfo := OAuthTokenInfo{
		ExpiresAt:   time.Now().Add(time.Duration(expiresIn) * time.Second),
		AccessToken: oauthJson.AccessToken,
	}
	return oauthInfo, &oauthJson.User, nil
}

// Refreshes the access token for future requests.
//
// Note: this function is not thread-safe! Please use the refreshTokenFieldIfReq function instead.
func (pixiv *PixivMobile) refreshTokenField() error {
	oauthInfo, user, err := RefreshAccessToken(pixiv.ctx, pixiv.apiTimeout, pixiv.refreshToken)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			pixiv.cancel()
			return err
		}
		return err
	}
	pixiv.accessTokenMap = oauthInfo
	if pixiv.user == nil {
		pixiv.user = user
	}
	return nil
}

// Checks if the current access token has expired,
// if so, refreshes the access token for future requests.
//
// Returns a boolean indicating if the access token was refreshed.
func (pixiv *PixivMobile) refreshTokenFieldIfReq() (bool, error) {
	pixiv.accessTokenMu.Lock()
	defer pixiv.accessTokenMu.Unlock()
	if pixiv.accessTokenMap.AccessToken != "" && pixiv.accessTokenMap.ExpiresAt.After(time.Now()) {
		return false, nil
	}
	return true, pixiv.refreshTokenField()
}
