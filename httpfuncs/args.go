package httpfuncs

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

type RequestArgs struct {
	// Main Request Options
	Method      string
	Url         string
	Timeout     int

	// Additional Request Options
	EditMu             sync.Mutex
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
	RetryDelay  *RetryDelay

	// Context is used to cancel the request if needed.
	// E.g. if the user presses Ctrl+C, we can use context.WithCancel(context.Background())
	Context context.Context

	// RequestHandler is the main function that will be called to make the request.
	RequestHandler RequestHandler
}

func (args *RequestArgs) validateHttp3Arg() error {
	if !args.Http2 && !args.Http3 {
		// if http2 and http3 are not enabled,
		// do a check to determine which protocol to use.
		if constants.FANTIA_DOWNLOAD_URL.MatchString(args.Url) || constants.FANTIA_ALBUM_URL.MatchString(args.Url) {
			args.Http2 = true
		} else {
			// check if the URL supports HTTP/3 first
			// before falling back to the default HTTP/2.
			for _, domain := range constants.HTTP3_SUPPORT_ARR {
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
		return fmt.Errorf(
			"error %d: http2 and http3 cannot be enabled at the same time",
			cdlerrors.DEV_ERROR,
		)
	}
	return nil
}

func (args *RequestArgs) getDefaultArgs() {
	if args.RetryDelay == nil {
		args.RetryDelay = &RetryDelay{
			Min: constants.MIN_RETRY_DELAY,
			Max: constants.MAX_RETRY_DELAY,
		}
	}

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
		args.UserAgent = DEFAULT_USER_AGENT
	}

	if args.Context == nil {
		args.Context = context.Background()
	}
}

// ValidateArgs validates the arguments of the request
//
// Will panic if the arguments are invalid as this is a developer error
func (args *RequestArgs) ValidateArgs() error {
	args.getDefaultArgs()
	err := args.validateHttp3Arg()
	if err != nil {
		return err
	}

	if args.Method == "" {
		return fmt.Errorf(
			"error %d: method cannot be empty",
			cdlerrors.DEV_ERROR,
		)
	}

	if args.Url == "" {
		return fmt.Errorf(
			"error %d: url cannot be empty",
			cdlerrors.DEV_ERROR,
		)
	}

	if args.Timeout < 0 {
		return fmt.Errorf(
			"error %d: timeout cannot be negative",
			cdlerrors.DEV_ERROR,
		)
	} else if args.Timeout == 0 {
		args.Timeout = 15
	}
	return nil
}
