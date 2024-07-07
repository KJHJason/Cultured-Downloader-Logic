package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
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

func TestFanboxCF(t *testing.T) {
	// Set up Chrome options
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(userAgent),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("headless", false),
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
			if hasBypassed, err := HasBypassed(website, ctx); err != nil {
				t.Fatal(err)
			} else if hasBypassed {
				t.Log("Cloudflare bypassed")
				return // pass
			}
		}

		t.Log("Solving Cloudflare challenge...")
		err := chromedp.Run(ctx,
			// chromedp.Navigate(website),
			chromedp.ActionFunc(func(ctx context.Context) error {
				time.Sleep(4 * time.Second)

				// translate
				// if self.driver.wait.ele_displayed('xpath://div/iframe',timeout=1.5):
				// time.sleep(1.5)
				// self.driver('xpath://div/iframe').ele("Verify you are human", timeout=2.5).click()
				// # The location of the button may vary time to time. I sometimes check the button's location and update the code.

				iframeCtx, iframeCancel := context.WithTimeout(ctx, 3*time.Second)
				defer iframeCancel()

				var iframeNodes []*cdp.Node
				if err := chromedp.Nodes(`//div/iframe`, &iframeNodes).Do(iframeCtx); err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						return nil
					}
					t.Fatal(err)
				} else if len(iframeNodes) != 1 {
					t.Logf("Expected 1 iframe, got %d", len(iframeNodes))
					return nil
				}

				// TODO: click on the checkbox since chromedp doesn't really have good support with iframes
				// go-rod is another option but its settings doesn't work with cloudflare iframes like chromedp with disable-site-isolation-trials enabled
				// Might consider another language like Python with Drission or some Rust webdrivers with iframe support
				err := chromedp.Evaluate(`document.querySelector("iframe").contentWindow.document.querySelector("input[type='checkbox']").click()`, nil).Do(ctx)
				if err != nil {
					t.Fatal(err)
				}

				// // get html content from iframe
				// var iframeHTML string
				// err = chromedp.InnerHTML("html", &iframeHTML).Do(ictx)
				// if err != nil {
				// 	t.Fatal(err)
				// }
				// t.Log("iframe HTML:", iframeHTML)

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
					t.Log("Cloudflare cookies not found")
				}
				return nil
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Log("Done")

	var pause = make(chan struct{})
	<-pause
}

// package api

// import (
// 	"context"
// 	"fmt"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/go-rod/rod"
// 	"github.com/go-rod/rod/lib/launcher"
// 	"github.com/go-rod/rod/lib/proto"
// )

// func HasBypassed(website string, p *rod.Page, ctx context.Context) (bool, error) {
// 	if website == "https://www.fanbox.cc/" {
// 		// Fanbox has custom Cloudflare page
// 		anchorCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
// 		defer cancel()

// 		// var nodes []*cdp.Node // check if the page has <a href="/"> in the html content
// 		// if err := chromedp.Nodes(`//a[@href="/"]`, &nodes).Do(anchorCtx); err != nil {
// 		// 	if errors.Is(err, context.DeadlineExceeded) {
// 		// 		return false, nil
// 		// 	}
// 		// 	return false, fmt.Errorf("CF Solver - failed to get nodes: %w", err)
// 		// }
// 		// return len(nodes) == 1, nil

// 		nodes, err := p.Context(anchorCtx).Elements("a[href='/']")
// 		if err != nil {
// 			return false, fmt.Errorf("CF Solver - failed to get nodes: %w", err)
// 		}
// 		return len(nodes) == 1, nil
// 	}

// 	// Note: this won't work for custom Cloudflare pages
// 	// var title string
// 	// if err := chromedp.Run(ctx, chromedp.Title(&title)); err != nil {
// 	// 	return false, fmt.Errorf("CF Solver - failed to get title: %w", err)
// 	// }

// 	info, err := p.Info()
// 	if err != nil {
// 		return false, fmt.Errorf("CF Solver - failed to get title: %w", err)
// 	}

// 	title := info.Title
// 	return !strings.Contains(strings.ToLower(title), "just a moment"), nil
// }

// func TestFanboxCF(t *testing.T) {
// 	// Set up Chrome options
// 	// userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

// 	l := launcher.New().
// 		Leakless(false).
// 		Headless(false)
// 	defer l.Cleanup()

// 	t.Log("Launching browser...")
// 	url, err := l.Launch()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	t.Log("Connecting to browser...")
// 	browser := rod.New().ControlURL(url)
// 	if err := browser.Connect(); err != nil {
// 		t.Fatal(err)
// 	}
// 	defer browser.MustClose()

// 	page := browser.MustPage("")

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	page = page.Context(ctx)

// 	website := "https://nopecha.com/demo/cloudflare"
// 	for i := range 3 {
// 		t.Log("Checking if Cloudflare is detected...")
// 		if i == 0 {
// 			var e proto.NetworkResponseReceived
// 			wait := browser.WaitEvent(&e)
// 			if err = page.Navigate(website); err != nil {
// 				t.Fatal(err)
// 			}
// 			wait()

// 			statusCode := e.Response.Status
// 			t.Log("Status code:", statusCode)
// 			if statusCode == 403 {
// 				t.Log("Cloudflare detected")
// 			} else if statusCode == 200 {
// 				t.Log("No Cloudflare detected")
// 				return // pass
// 			} else {
// 				t.Fatalf("Unknown status code: %d", statusCode)
// 			}
// 		} else {
// 			if hasBypassed, err := HasBypassed(website, page, ctx); err != nil {
// 				t.Fatal(err)
// 			} else if hasBypassed {
// 				t.Log("Cloudflare bypassed")
// 				return // pass
// 			}
// 		}
// 		break
// 		t.Log("Solving Cloudflare challenge...")
// 		time.Sleep(4 * time.Second)
// 	}

// 	// 	err := chromedp.Run(ctx,
// 	// 		// chromedp.Navigate(website),
// 	// 		chromedp.ActionFunc(func(ctx context.Context) error {

// 	// 			// translate
// 	// 			// if self.driver.wait.ele_displayed('xpath://div/iframe',timeout=1.5):
// 	// 			// time.sleep(1.5)
// 	// 			// self.driver('xpath://div/iframe').ele("Verify you are human", timeout=2.5).click()
// 	// 			// # The location of the button may vary time to time. I sometimes check the button's location and update the code.

// 	// 			iframeCtx, iframeCancel := context.WithTimeout(ctx, 3*time.Second)
// 	// 			defer iframeCancel()

// 	// 			var iframeNodes []*cdp.Node
// 	// 			if err := chromedp.Nodes(`//div/iframe`, &iframeNodes).Do(iframeCtx); err != nil {
// 	// 				if errors.Is(err, context.DeadlineExceeded) {
// 	// 					return nil
// 	// 				}
// 	// 				t.Fatal(err)
// 	// 			} else if len(iframeNodes) != 1 {
// 	// 				t.Logf("Expected 1 iframe, got %d", len(iframeNodes))
// 	// 				return nil
// 	// 			}

// 	// 			err := chromedp.Evaluate(`document.querySelector("iframe").contentWindow.document.querySelector("input[type='checkbox']").click()`, nil).Do(ctx)
// 	// 			if err != nil {
// 	// 				t.Fatal(err)
// 	// 			}

// 	// 			time.Sleep(4 * time.Second)

// 	// 			t.Log("Getting cookies...")
// 	// 			browserCookies, err := network.GetCookies().Do(ctx)
// 	// 			if err != nil {
// 	// 				t.Fatal(err)
// 	// 			}

// 	// 			hasCfBeam, hasCfClearance := false, false
// 	// 			for _, cookie := range browserCookies {
// 	// 				if cookie.Name == "__cf_bm" {
// 	// 					hasCfBeam = true
// 	// 					t.Log("__cf_bm cookie found:", cookie.Value)
// 	// 				}
// 	// 				if cookie.Name == "cf_clearance" {
// 	// 					hasCfClearance = true
// 	// 					t.Log("cf_clearance cookie found:", cookie.Value)
// 	// 				}
// 	// 			}
// 	// 			if !hasCfBeam || !hasCfClearance {
// 	// 				t.Fatal("Cloudflare cookies not found")
// 	// 			}
// 	// 			return nil
// 	// 		}),
// 	// 	)
// 	// 	if err != nil {
// 	// 		t.Fatal(err)
// 	// 	}
// 	// }

// 	t.Log("Done")

// 	var pause = make(chan struct{})
// 	<-pause
// }
