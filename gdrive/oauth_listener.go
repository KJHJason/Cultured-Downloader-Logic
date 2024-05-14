package gdrive

import (
	"context"
	"net"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

// Example OAuth callback URL:
// http://localhost:8080/?state=state-token&code=<code>&scope=https://www.googleapis.com/auth/drive.readonly%20https://www.googleapis.com/auth/drive.metadata.readonly

func getOAuthCode(w http.ResponseWriter, r *http.Request) {
	// get the code from the URL
	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	updateOauthCode(code)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("The authentication flow has completed. You may close this window."))
}

func startOAuthServer(ctx context.Context) error {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("/", getOAuthCode)

	server := &http.Server{
		Addr: ":8080",
		BaseContext: func(listener net.Listener) context.Context {
			return childCtx
		},
		Handler: mux,
	}

	var startUpErr error
	go func() {
		logger.MainLogger.Info("Starting OAuth listener server...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// start up error
			startUpErr = err
			logger.MainLogger.Error("OAuth listener server error: " + err.Error())
			cancel()
		}
	}()
	<-childCtx.Done()

	if startUpErr != nil {
		return startUpErr
	}
	logger.MainLogger.Info("Shutting down OAuth listener server...")
	return server.Shutdown(ctx)
}

var (
	oauthCode string
	oauthMu   sync.RWMutex
)
func updateOauthCode(code string) {
	oauthMu.Lock()
	oauthCode = code
	oauthMu.Unlock()
}
func getOauthCode() string {
	oauthMu.RLock()
	defer oauthMu.RUnlock()
	return oauthCode
}

func StartOAuthListener(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	updateOauthCode("")
	srvCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer updateOauthCode("")

	var srvErr error
	go func() {
		srvErr = startOAuthServer(srvCtx)
	}()

codeLoop:
	for {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		case <-srvCtx.Done():
			if srvErr != nil {
				return nil, srvErr
			}
		default:
			if getOauthCode() != "" {
				break codeLoop
			}
		}
	}

	// process the code
	token, err := ProcessAuthCode(ctx, getOauthCode(), config)
	if err != nil {
		return nil, err
	}
	return token, nil
}
