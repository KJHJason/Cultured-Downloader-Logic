package api

import (
	"context"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
)

func SetChromedpAllocCookies(cookies []*http.Cookie) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for _, cookie := range cookies {
			var expr cdp.TimeSinceEpoch
			if cookie.Expires.IsZero() {
				expr = cdp.TimeSinceEpoch(time.Now().Add(365 * 24 * time.Hour))
			} else {
				expr = cdp.TimeSinceEpoch(cookie.Expires)
			}

			err := network.SetCookie(cookie.Name, cookie.Value).
				WithExpires(&expr).
				WithDomain(cookie.Domain).
				WithPath(cookie.Path).
				WithHTTPOnly(cookie.HttpOnly).
				WithSecure(cookie.Secure).
				Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func GetDefaultChromedpAlloc(userAgent string, ctx context.Context) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgent),
	)

	if ctx == nil {
		ctx = context.Background()
	}
	return chromedp.NewExecAllocator(ctx, opts...)
}

func ExecuteChromedpActions(allocCtx context.Context, allocCancelFn context.CancelFunc, actions ...chromedp.Action) error {
	if allocCtx == nil {
		allocCtx = context.Background()
	}

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	return chromedp.Run(taskCtx, actions...)
}
