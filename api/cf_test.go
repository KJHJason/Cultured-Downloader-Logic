package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

func HasBypassed(website string, ctx context.Context) (bool, error) {
	if website == "https://www.fanbox.cc/" {
		// Fanbox has custom Cloudflare page
		anchorCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		var nodes []*cdp.Node // check if the page has <a href="/"> in the html content
		if err := chromedp.Nodes(`//a[@href="/"]`, &nodes).Do(anchorCtx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return false, nil
			}
			return false, fmt.Errorf("CF Solver - failed to get nodes: %w", err)
		}
		return len(nodes) == 1, nil
	}

	// Note: this won't work for custom Cloudflare pages
	var title string
	if err := chromedp.Run(ctx, chromedp.Title(&title)); err != nil {
		return false, fmt.Errorf("CF Solver - failed to get title: %w", err)
	}
	return !strings.Contains(strings.ToLower(title), "just a moment"), nil
}

// Doesn't work as it is detected by Cloudflare as a webdriver
func DemoChromedp(t *testing.T) {
	// Set up Chrome options
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgent),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("enable-automation", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	website := "https://nopecha.com/demo/cloudflare"
	for i := range 3 {
		if i == 0 {
			resp, err := chromedp.RunResponse(ctx, chromedp.Navigate(website))
			if err != nil {
				t.Fatal(err)
			}

			statusCode := resp.Status
			t.Log("Status code:", statusCode)
			if statusCode == 403 {
				t.Log("Cloudflare detected")
			} else if statusCode == 200 {
				t.Log("No Cloudflare detected")
				return // pass
			} else {
				t.Fatalf("Unknown status code: %d", statusCode)
			}
		} else {
			time.Sleep(5 * time.Second)
			if hasBypassed, err := HasBypassed(website, ctx); err != nil {
				t.Fatal(err)
			} else if hasBypassed {
				t.Log("Cloudflare bypassed")
				return // pass
			}
		}
		time.Sleep(5 * time.Second)

		t.Log("Solving Cloudflare challenge...")
		targets, err := chromedp.Targets(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var iframeTarget *target.Info
		for _, target := range targets {
			if target.Type == "iframe" {
				iframeTarget = target
				break
			}
		}
		if iframeTarget == nil {
			continue
		}

		iframeCtx, iframeCancel := chromedp.NewContext(ctx, chromedp.WithTargetID(iframeTarget.TargetID))
		defer iframeCancel()

		checkboxCtx, checkboxCancel := context.WithTimeout(iframeCtx, 3*time.Second)
		defer checkboxCancel()

		const checkboxSelector = `input[type='checkbox']`
		err = chromedp.Run(checkboxCtx,
			chromedp.WaitVisible(checkboxSelector, chromedp.ByQuery),
			chromedp.Click(checkboxSelector, chromedp.ByQuery),
		)
		if err != nil {
			t.Fatal(err)
		}

		err = chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
			time.Sleep(5 * time.Second)

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
				t.Log("Cloudflare cookies not found")
			}
			return nil
		}))
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Log("Done")

	// Catch SIGINT/SIGTERM signal and cancel the context when received
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()
	defer signal.Stop(sigs)

	var pause = make(chan struct{})
	<-pause
}
