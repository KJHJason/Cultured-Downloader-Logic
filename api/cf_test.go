package api

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func TestFanboxCF(t *testing.T) {
	// Set up Chrome options
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgent),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("headless", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Start browser and open the page
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://fanbox.cc/"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			t.Log("Waiting for the page to load...")
			err := chromedp.WaitVisible(`a[href="/"]`).Do(ctx)
			if err != nil {
				t.Fatal(err)
			}

			// get the cookies
			time.Sleep(4 * time.Second)
			t.Log("Getting cookies...")
			browserCookies, err := network.GetCookies().Do(ctx)
			if err != nil {
				t.Fatal(err)
			}

			hasCfBeam, hasCfClearance := false, false
			for _, cookie := range browserCookies {
				if cookie.Name == "__cf_bm" {
					hasCfBeam = true
					t.Log("__cf_bm cookie found:", cookie.Value)
				}
				if cookie.Name == "cf_clearance" {
					hasCfClearance = true
					t.Log("cf_clearance cookie found:", cookie.Value)
				}
			}
			if !hasCfBeam || !hasCfClearance {
				t.Fatal("Cloudflare cookies not found")
			}
			return nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Done")

	var pause = make(chan struct{})
	<-pause
}
