package httpfuncs

import (
	"context"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/filters"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

type RequestHandler func(reqArgs *RequestArgs) (*http.Response, error)

type versionInfo struct {
	Major int
	Minor int
	Patch int
}

type RetryDelay struct {
	Max float32
	Min float32
}

type ToDownload struct {
	CacheKey string
	CacheFn  func(key string)
	Url      string
	FilePath string
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

	Filters *filters.Filters

	ProgressBarInfo *progress.ProgressBarInfo
}

type GithubApiRes struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
}
