package cdlsolvers

import (
	"context"
	"testing"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cdlsolvers/cdldocker"
)

func TestDockerImageForCf(t *testing.T) {
	const website = "https://nopecha.com/demo/cloudflare"
	const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36"
	cookies, err := cdldocker.CallDockerImageForCf(context.Background(), userAgent, website)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Cookies: %v", cookies)
}
