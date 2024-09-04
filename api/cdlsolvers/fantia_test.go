package cdlsolvers

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cdlsolvers/cdldocker"
	"github.com/joho/godotenv"
)

func loadDotEnv(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatal("Error loading .env file")
	}
}

// go test -run ^TestDockerImageForFantia$ github.com/KJHJason/Cultured-Downloader-Logic/api/cdlsolvers
func TestDockerImageForFantia(t *testing.T) {
	loadDotEnv(t)

	sessionId := os.Getenv("FANTIA_SESSION_ID")
	cookies := []*http.Cookie{
		{
			Name:    "_session_id",
			Value:   sessionId,
			Domain:  ".fantia.jp",
			Path:    "/",
			Secure:  true,
			Expires: time.Now().Add(24 * time.Hour),
		},
	}

	const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36"
	err := cdldocker.CallDockerImageForFantia(context.Background(), userAgent, cookies)
	if err != nil {
		t.Fatal(err)
	}
}
