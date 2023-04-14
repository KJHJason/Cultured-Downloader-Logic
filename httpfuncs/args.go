package httpfuncs

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"regexp"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

var (
	// Since the URLs below will be redirected to Fantia's AWS S3 URL, 
	// we need to use HTTP/2 as it is not supported by HTTP/3 yet.
	FANTIA_ALBUM_URL = regexp.MustCompile(
		`^https://fantia.jp/posts/[\d]+/album_image`,
	)
	FANTIA_DOWNLOAD_URL = regexp.MustCompile(
		`^https://fantia.jp/posts/[\d]+/download/[\d]+`,
	)

	HTTP3_SUPPORT_ARR = [...]string{
		"https://www.pixiv.net",
		"https://app-api.pixiv.net",

		"https://www.google.com",
		"https://drive.google.com",
	}
)

func (args *RequestArgs) validateHttp3Arg() {
	if !args.Http2 && !args.Http3 {
		// if http2 and http3 are not enabled,
		// do a check to determine which protocol to use.
		if FANTIA_DOWNLOAD_URL.MatchString(args.Url) || FANTIA_ALBUM_URL.MatchString(args.Url) {
			args.Http2 = true
		} else {
			// check if the URL supports HTTP/3 first
			// before falling back to the default HTTP/2.
			for _, domain := range HTTP3_SUPPORT_ARR {
				if strings.HasPrefix(args.Url, domain) {
					args.Http3 = true
					break
				}
			}
			// if HTTP/3 is not supported, fall back to HTTP/2
			if !args.Http3 {
				args.Http2 = true
			}
		}
	} else if args.Http2 && args.Http3 {
		panic(
			fmt.Errorf(
				"error %d: http2 and http3 cannot be enabled at the same time",
				constants.DEV_ERROR,
			),
		)
	}
}

func (args *RequestArgs) getDefaultArgs() {
	if args.RequestHandler == nil {
		args.RequestHandler = CallRequest
	}

	if args.Headers == nil {
		args.Headers = make(map[string]string)
	}

	if args.Params == nil {
		args.Params = make(map[string]string)
	}

	if args.Cookies == nil {
		args.Cookies = make([]*http.Cookie, 0)
	}

	if args.UserAgent == "" {
		args.UserAgent = constants.USER_AGENT
	}

	if args.Context == nil {
		args.Context = context.Background()
	}
}

// ValidateArgs validates the arguments of the request
//
// Will panic if the arguments are invalid as this is a developer error
func (args *RequestArgs) ValidateArgs() {
	args.getDefaultArgs()
	args.validateHttp3Arg()

	if args.Method == "" {
		panic(
			fmt.Errorf(
				"error %d: method cannot be empty",
				constants.DEV_ERROR,
			),
		)
	}

	if args.Url == "" {
		panic(
			fmt.Errorf(
				"error %d: url cannot be empty",
				constants.DEV_ERROR,
			),
		)
	}

	if args.Timeout < 0 {
		panic(
			fmt.Errorf(
				"error %d: timeout cannot be negative",
				constants.DEV_ERROR,
			),
		)
	} else if args.Timeout == 0 {
		args.Timeout = 15
	}
}
