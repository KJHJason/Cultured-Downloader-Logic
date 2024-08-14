package pixivmobile

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

type PixivMobile struct {
	user   *UserDetails
	ctx    context.Context
	cancel context.CancelFunc

	Base     *api.BaseDl
	pFilters *pixivcommon.PixivFilters

	// API information and its endpoints
	refreshToken string

	// User given arguments
	apiTimeout int

	// Access token information
	accessTokenMu  sync.Mutex
	accessTokenMap OAuthTokenInfo
}

func (p *PixivMobile) GetCaptchaHandler() httpfuncs.CaptchaHandler {
	return pixivcommon.NewHttpCaptchaHandler(
		p.ctx,
		constants.PIXIV_URL, // not using constants.PIXIV_MOBILE_URL as it's under the same domain
		p.Base.Configs.UserAgent,
		p.Base.Notifier,
	)
}

func (p *PixivMobile) GetContext() context.Context {
	return p.ctx
}

func (p *PixivMobile) GetCancel() context.CancelFunc {
	return p.cancel
}

func (p *PixivMobile) SetContext(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
}

// CancelCtx releases the resources used and cancels the context of the PixivMobile struct.
func (p *PixivMobile) CancelCtx() {
	p.cancel()
}

func (p *PixivMobile) CtxIsActive() bool {
	return p.ctx.Err() == nil
}

func (pixiv *PixivMobile) SetPixivFilters(filters pixivcommon.PixivFilters) error {
	if pixiv.user == nil {
		panic("pixiv user is nil, did you forget to use NewPixivMobile() or forgot to refresh the access token first?")
	}

	if err := filters.ValidateForMobileApi(pixiv.user.IsPremium); err != nil {
		return err
	}
	pixiv.pFilters = &filters
	return nil
}

// Get a new PixivMobile structure
func NewPixivMobile(refreshToken string, timeout int, ctx context.Context) (*PixivMobile, error) {
	pixivMobile := &PixivMobile{
		refreshToken: refreshToken,
		apiTimeout:   timeout,
	}
	pixivMobile.SetContext(ctx)
	if refreshToken != "" {
		// refresh the access token and verify it
		if err := pixivMobile.refreshTokenField(); err != nil {
			return nil, err
		}
	}
	return pixivMobile, nil
}

// This is due to Pixiv's strict rate limiting.
//
// Without delays, the user might get 429 too many requests
// or the user's account might get suspended.
//
// Additionally, pixiv.net is protected by cloudflare, so
// to prevent the user's IP reputation from going down, delays are added.
func (pixiv *PixivMobile) Sleep() {
	time.Sleep(httpfuncs.GetRandomTimeIntMs(1000, 1500))
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
func (pixiv *PixivMobile) SendRequest(reqArgs *httpfuncs.RequestArgs) (*httpfuncs.ResponseWrapper, error) {
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

	if reqArgs.CaptchaHandler.IsNotConfigured() {
		reqArgs.CaptchaHandler = pixiv.GetCaptchaHandler()
	}

	if reqArgs.CaptchaHandler.CallBeforeReq {
		if err := reqArgs.CaptchaHandler.ReqModifier(req); err != nil {
			return nil, err
		}
	}

	var res *http.Response
	retryCount := 1
	failedHttp3Req := 0
	isUsingHttp3 := reqArgs.Http3
	client := httpfuncs.GetHttpClient(reqArgs)
	for retryCount <= constants.RETRY_COUNTER {
		res, err = client.Do(req)
		if errors.Is(err, context.Canceled) {
			return nil, context.Canceled
		}

		if err == nil {
			if refreshed {
				continue
			}

			respWrapper, ok, captchaErr := httpfuncs.CaptchaHandlerLogic(req, res, reqArgs)
			if captchaErr != nil {
				return nil, captchaErr
			}
			if !ok {
				continue
			}
			if res.StatusCode == 200 || !reqArgs.CheckStatus {
				return respWrapper, nil
			}
			retryCount++
		} else {
			httpfuncs.Http2FallbackLogic(
				&isUsingHttp3,
				&failedHttp3Req,
				&retryCount,
				err,
				reqArgs,
				client,
			)
		}

		time.Sleep(httpfuncs.GetDefaultRandomDelay())
	}
	return nil, fmt.Errorf(
		"request to %s failed after %d retries",
		reqArgs.Url,
		constants.RETRY_COUNTER,
	)
}
