package httpfuncs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/quic-go/quic-go/http3"
)

func GetHttp2Client(reqArgs *RequestArgs) *http.Client {
	return &http.Client{
		Transport: &http.Transport{},
		Timeout:   time.Duration(reqArgs.Timeout) * time.Second,
	}
}

func GetHttp3Client(reqArgs *RequestArgs) *http.Client {
	return &http.Client{
		Transport: &http3.RoundTripper{},
		Timeout:   time.Duration(reqArgs.Timeout) * time.Second,
	}
}

// Get a new HTTP/2 or HTTP/3 client based on the request arguments
func GetHttpClient(reqArgs *RequestArgs) *http.Client {
	if reqArgs.Http3 {
		return GetHttp3Client(reqArgs)
	}
	return GetHttp2Client(reqArgs)
}

// add headers to the request
func AddHeaders(headers map[string]string, defaultUserAgent string, req *http.Request) {
	if len(headers) == 0 {
		return
	}

	if userAgent, ok := headers["User-Agent"]; !ok || userAgent == "" {
		headers["User-Agent"] = defaultUserAgent
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}
}

// add cookies to the request
func AddCookies(reqUrl string, cookies []*http.Cookie, req *http.Request) {
	if len(cookies) == 0 {
		return
	}

	for _, cookie := range cookies {
		if strings.Contains(reqUrl, cookie.Domain) {
			req.AddCookie(cookie)
		}
	}
}

// add params to the request
func AddParams(params map[string]string, req *http.Request) {
	if len(params) == 0 {
		return
	}

	query := req.URL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()
}

func Http2FallbackLogic(isUsingHttp3 *bool, failedHttp3Req *int, retryCount *int, err error, reqArgs *RequestArgs, client *http.Client) {
	if *isUsingHttp3 {
		if *failedHttp3Req < constants.HTTP3_MAX_RETRY {
			*failedHttp3Req++
		} else {
			// if the request failed too many times,
			// switch to HTTP/2 in the event that the server does not support HTTP/3
			*client = *GetHttp2Client(reqArgs)
			*isUsingHttp3 = false
		}
	} else {
		// only start incrementing the retry count
		// if the request failed and is not using HTTP/3
		*retryCount++
	}
	logger.MainLogger.Errorf(
		"error %d: request to %s failed, more info => %v",
		cdlerrors.CONNECTION_ERROR,
		reqArgs.Url,
		err,
	)
}

// send the request to the target URL and retries if the request was not successful
func sendRequest(req *http.Request, reqArgs *RequestArgs) (*http.Response, error) {
	reqArgs.EditMu.Lock()
	AddCookies(reqArgs.Url, reqArgs.Cookies, req)
	AddHeaders(reqArgs.Headers, reqArgs.UserAgent, req)
	AddParams(reqArgs.Params, req)
	reqArgs.EditMu.Unlock()

	var err error
	var res *http.Response

	retryCount := 1
	failedHttp3Req := 0
	isUsingHttp3 := reqArgs.Http3
	client := GetHttpClient(reqArgs)
	for retryCount <= constants.RETRY_COUNTER {
		res, err = client.Do(req)
		if errors.Is(err, context.Canceled) {
			return nil, context.Canceled
		}

		if err == nil {
			if res.StatusCode == 200 || !reqArgs.CheckStatus {
				return res, nil
			}
			res.Body.Close()
			retryCount++
			goto retry
		}

		Http2FallbackLogic(
			&isUsingHttp3,
			&failedHttp3Req,
			&retryCount,
			err,
			reqArgs,
			client,
		)

	retry:
		if retryCount < constants.RETRY_COUNTER {
			time.Sleep(GetRandomDelay(reqArgs.RetryDelay))
		}
	}

	errMsg := fmt.Sprintf(
		"the request to %s failed after %d retries",
		reqArgs.Url,
		constants.RETRY_COUNTER,
	)
	if err != nil {
		err = fmt.Errorf("%s, more info => %w",
			errMsg,
			err,
		)
	} else if res != nil {
		err = fmt.Errorf("%s, status code => %s",
			errMsg,
			res.Status,
		)
	} else {
		err = errors.New(errMsg)
	}
	return nil, err
}

// CallRequest is used to make a request to a URL and return the response
//
// If the request fails, it will retry the request again up
// to the defined max retries in the constants.go in utils package
func CallRequest(reqArgs *RequestArgs) (*http.Response, error) {
	err := reqArgs.ValidateArgs()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		reqArgs.Context,
		reqArgs.Method,
		reqArgs.Url,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error %d: unable to create a new request, more info => %w",
			cdlerrors.DEV_ERROR,
			err,
		)
	}

	return sendRequest(req, reqArgs)
}

// Check for active internet connection (To be used at the start of the program)
func CheckInternetConnection() error {
	_, err := CallRequest(
		&RequestArgs{
			Url:         "https://www.google.com",
			Method:      "HEAD",
			Timeout:     10,
			CheckStatus: false,
			Http3:       true,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"error %d: unable to connect to the internet, more info => %w",
			cdlerrors.DEV_ERROR,
			err,
		)
	}
	return nil
}

// Sends a request with the given data
func CallRequestWithData(reqArgs *RequestArgs, data map[string]string) (*http.Response, error) {
	reqArgs.EditMu.Lock()
	err := reqArgs.ValidateArgs()
	if err != nil {
		reqArgs.EditMu.Unlock()
		return nil, err
	}

	form := url.Values{}
	for key, value := range data {
		form.Add(key, value)
	}
	if len(data) > 0 {
		reqArgs.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	}
	reqArgs.EditMu.Unlock()

	req, err := http.NewRequestWithContext(
		reqArgs.Context,
		reqArgs.Method,
		reqArgs.Url,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}

	return sendRequest(req, reqArgs)
}
