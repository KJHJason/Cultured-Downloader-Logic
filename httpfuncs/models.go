package httpfuncs

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/filters"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

type RequestHandler func(reqArgs *RequestArgs) (*ResponseWrapper, error)

type ResponseWrapper struct {
	Resp   *http.Response
	body   []byte
	closed bool
}

func NewResponseWrapper(resp *http.Response) *ResponseWrapper {
	return &ResponseWrapper{Resp: resp}
}

// Close the response body
func (rw *ResponseWrapper) Close() {
	if !rw.closed && rw.Resp != nil {
		rw.Resp.Body.Close()
	}
}

func (rw *ResponseWrapper) Url() string {
	return rw.Resp.Request.URL.String()
}

func (rw *ResponseWrapper) GetBody() ([]byte, error) {
	if rw.body == nil {
		rw.closed = true // since ReadResBody closes the body
		if body, err := ReadResBody(rw.Resp); err != nil {
			return nil, err
		} else {
			rw.body = body
		}
	}
	return rw.body, nil
}

func (rw *ResponseWrapper) GetBodyReader() (bodyReader *bytes.Reader, err error) {
	body, err := rw.GetBody()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(body), nil
}

type versionInfo struct {
	Major int
	Minor int
	Patch int
}

// Max and Min time in nano-seconds (refer to time.Duration)
type RetryDelay struct {
	Max time.Duration
	Min time.Duration
}

type ToDownload struct {
	CacheKey string
	CacheFn  func(key string)
	Url      string
	FilePath string
}

type CaptchaHandler struct {
	Check   func(*ResponseWrapper) (bool, error)
	Handler interface {
		Call(*http.Request) error
	}
	InjectCfCookies func() []*http.Cookie
}

func (ch CaptchaHandler) Call(req *http.Request) error {
	return ch.Handler.Call(req)
}

type DlOptions struct {
	// Parent context for the download process
	Context context.Context

	// MaxConcurrency is the maximum number of concurrent downloads
	MaxConcurrency int

	// Cookies is a list of cookies to be used in the download process
	Cookies []*http.Cookie

	// Headers is a map of headers to be used in the download process
	Headers map[string]string

	// UseHttp3 is a flag to enable HTTP/3
	// Otherwise, HTTP/2 will be used by default
	UseHttp3 bool

	// Since a HEAD request is sent to determine the expected
	// file size (if known), HeadReqTimeout is the timeout for the HEAD request
	HeadReqTimeout int

	// RetryDelay is the delay between retries
	RetryDelay *RetryDelay

	// Whether the server supports Accept-Ranges header value
	SupportRange bool

	SetMetadata bool

	Filters *filters.Filters

	ProgressBarInfo *progress.ProgressBarInfo

	CaptchaHandler CaptchaHandler
}

type GithubApiRes struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
}
