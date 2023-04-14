package httpfuncs

import (
	"context"
	"net/http"
)

type RequestHandler func (reqArgs *RequestArgs) (*http.Response, error)

type RequestArgs struct {
	// Main Request Options
	Method string
	Url string
	Timeout int

	// Additional Request Options
	Headers            map[string]string
	Params             map[string]string
	Cookies            []*http.Cookie
	UserAgent          string
	DisableCompression bool

	// HTTP/2 and HTTP/3 Options
	Http2 bool
	Http3 bool

	// Check status will check the status code of the response for 200 OK.
	// If the status code is not 200 OK, it will retry several times and 
	// if the status code is still not 200 OK, it will return an error.
	// Otherwise, it will return the response regardless of the status code.
	CheckStatus bool

	// Context is used to cancel the request if needed.
	// E.g. if the user presses Ctrl+C, we can use context.WithCancel(context.Background())
	Context context.Context

	// RequestHandler is the main function that will be called to make the request.
	RequestHandler RequestHandler
}

type ToDownload struct {
	Url      string
	FilePath string
}

type DlOptions struct {
	// MaxConcurrency is the maximum number of concurrent downloads
	MaxConcurrency int

	// Cookies is a list of cookies to be used in the download process
	Cookies []*http.Cookie

	// Headers is a map of headers to be used in the download process
	Headers map[string]string

	// UseHttp3 is a flag to enable HTTP/3
	// Otherwise, HTTP/2 will be used by default
	UseHttp3 bool
}

type GithubApiRes struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
}
