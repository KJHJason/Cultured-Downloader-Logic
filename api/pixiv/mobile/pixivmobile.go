package pixivmobile

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

type PixivMobile struct {
	user   *UserDetails
	ctx    context.Context
	cancel context.CancelFunc

	useCacheDb          bool
	baseDownloadDirPath string

	// API information and its endpoints
	refreshToken string

	// User given arguments
	apiTimeout int

	// Access token information
	accessTokenMu  sync.Mutex
	accessTokenMap OAuthTokenInfo

	// Prog bar
	MainProgBar progress.ProgressBar
}

// Get a new PixivMobile structure
func NewPixivMobile(refreshToken string, timeout int, ctx context.Context, cancelFunc context.CancelFunc) (*PixivMobile, error) {
	pixivMobile := &PixivMobile{
		ctx:          ctx,
		cancel:       cancelFunc,
		refreshToken: refreshToken,
		apiTimeout:   timeout,
	}
	if refreshToken != "" {
		// refresh the access token and verify it
		if err := pixivMobile.refreshTokenField(); err != nil {
			return nil, err
		}
	}
	return pixivMobile, nil
}

func (pixiv *PixivMobile) SetBaseDlDirPath(dlDirPath string) {
	pixiv.baseDownloadDirPath = dlDirPath
}

func (pixiv *PixivMobile) SetUseCacheDb(useCacheDb bool) {
	pixiv.useCacheDb = useCacheDb
}

// This is due to Pixiv's strict rate limiting.
//
// Without delays, the user might get 429 too many requests
// or the user's account might get suspended.
//
// Additionally, pixiv.net is protected by cloudflare, so
// to prevent the user's IP reputation from going down, delays are added.
func (pixiv *PixivMobile) Sleep() {
	time.Sleep(httpfuncs.GetRandomTime(1.0, 1.5))
}

// Get the required headers to communicate with the Pixiv API
func (pixiv *PixivMobile) getHeaders(additional map[string]string) map[string]string {
	headers := make(map[string]string)
	for k, v := range additional {
		headers[k] = v
	}

	baseHeaders := map[string]string{
		"User-Agent":     constants.PIXIV_MOBILE_USER_AGENT,
		"App-OS":         "ios",
		"App-OS-Version": "14.6",
		"Authorization":  "Bearer " + pixiv.accessTokenMap.AccessToken,
	}
	for k, v := range baseHeaders {
		headers[k] = v
	}
	return headers
}

// Sends a request to the Pixiv API and refreshes the access token if required
//
// Returns the JSON interface and errors if any
func (pixiv *PixivMobile) SendRequest(reqArgs *httpfuncs.RequestArgs) (*http.Response, error) {
	if reqArgs.Method == "" {
		reqArgs.Method = "GET"
	}
	if reqArgs.Timeout == 0 {
		reqArgs.Timeout = pixiv.apiTimeout
	}
	useHttp3 := httpfuncs.IsHttp3Supported(constants.PIXIV_MOBILE, true)
	reqArgs.Http3 = useHttp3
	reqArgs.Http2 = !useHttp3
	err := reqArgs.ValidateArgs()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(reqArgs.Method, reqArgs.Url, nil)
	if err != nil {
		return nil, err
	}

	refreshed, err := pixiv.refreshTokenFieldIfReq()
	if err != nil {
		return nil, err
	}

	for k, v := range pixiv.getHeaders(reqArgs.Headers) {
		req.Header.Set(k, v)
	}
	httpfuncs.AddParams(reqArgs.Params, req)

	var res *http.Response
	client := httpfuncs.GetHttpClient(reqArgs)
	client.Timeout = time.Duration(reqArgs.Timeout) * time.Second
	for i := 1; i <= constants.RETRY_COUNTER; i++ {
		res, err = client.Do(req)
		if err == nil {
			if refreshed {
				continue
			} else if res.StatusCode == 200 || !reqArgs.CheckStatus {
				return res, nil
			}
		}
		time.Sleep(httpfuncs.GetDefaultRandomDelay())
	}
	return nil, fmt.Errorf(
		"request to %s failed after %d retries",
		reqArgs.Url,
		constants.RETRY_COUNTER,
	)
}
