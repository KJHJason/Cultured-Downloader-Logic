package httpfuncs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/jbenet/go-context/io"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

func getFullFilePath(res *http.Response, filePath string) (string, error) {
	// check if filepath already have a filename attached
	if filepath.Ext(filePath) != "" {
		filePathDir := filepath.Dir(filePath)
		os.MkdirAll(filePathDir, 0755)
		filePathWithoutExt := iofuncs.RemoveExtFromFilename(filePath)
		return filePathWithoutExt + strings.ToLower(filepath.Ext(filePath)), nil
	}

	os.MkdirAll(filePath, 0755)
	filename, err := url.PathUnescape(res.Request.URL.String())
	if err != nil {
		// should never happen but just in case
		return "", fmt.Errorf(
			"error %d: failed to unescape URL, more info => %w\nurl: %s",
			errs.UNEXPECTED_ERROR,
			err,
			res.Request.URL.String(),
		)
	}
	filename = GetLastPartOfUrl(filename)
	filenameWithoutExt := iofuncs.RemoveExtFromFilename(filename)
	filePath = filepath.Join(
		filePath,
		filenameWithoutExt + strings.ToLower(filepath.Ext(filename)),
	)
	return filePath, nil
}

// check if the file size matches the content length
// if not, then the file does not exist or is corrupted and should be re-downloaded
func checkIfCanSkipDl(fileSize, contentLength int64, forceOverwrite bool) bool {
	if fileSize == contentLength {
		// If the file already exists and the file size
		// matches the expected file size in the Content-Length header,
		// then skip the download process.
		return true
	} else if !forceOverwrite && fileSize > 0 {
		// If the file already exists and have more than 0 bytes
		// but the Content-Length header does not exist in the response,
		// we will assume that the file is already downloaded
		// and skip the download process if the overwrite flag is false.
		return true
	}
	return false
}

func dlToFile(ctx context.Context, res *http.Response, url, filePath string, downloadPartial bool) error {
	fileFlags := os.O_CREATE | os.O_WRONLY
	if downloadPartial {
		fileFlags |= os.O_APPEND
	} else {
		fileFlags |= os.O_TRUNC
	}

	file, err := os.OpenFile(filePath, fileFlags, 0644)
	if err != nil {
		return fmt.Errorf(
			"error %d: failed to open/create file, more info => %w\nfile path: %s",
			errs.OS_ERROR,
			err,
			filePath,
		)
	}
	defer file.Close()

	// write the body to file
	respReader := ctxio.NewReader(ctx, res.Body)
	_, err = io.Copy(file, respReader)
	if err != nil {
		if err == context.Canceled {
			return nil
		}

		logger.LogError(
			fmt.Errorf(
				"failed to download %s due to %w",
				url,
				err,
			), 
			false, 
			logger.ERROR,
		)
		return err
	}
	return nil
}

// DownloadUrl is used to download a file from a URL
//
// Note: If the file already exists, the download process will be skipped
func downloadUrl(filePath string, queue chan struct{}, reqArgs *RequestArgs, overwriteExistingFile, supportRange bool) error {
	queue <- struct{}{}
	// Send a HEAD request first to get the expected file size from the Content-Length header.
	// A GET request might work but most of the time
	// as the Content-Length header may not present due to chunked encoding.
	headRes, err := reqArgs.RequestHandler(
		&RequestArgs{
			Url:         reqArgs.Url,
			Method:      "HEAD",
			Timeout:     15,
			Cookies:     reqArgs.Cookies,
			Headers:     reqArgs.Headers,
			UserAgent:   reqArgs.UserAgent,
			CheckStatus: true,
			RetryDelay:  reqArgs.RetryDelay,
			Http3:       reqArgs.Http3,
			Http2:       reqArgs.Http2,
			Context:     reqArgs.Context,
		},
	)
	if err != nil {
		return err
	}
	fileReqContentLength := headRes.ContentLength
	headRes.Body.Close()

	downloadedBytes, err := iofuncs.GetFileSize(filePath)
	if err != nil {
		if err != os.ErrNotExist {
			// if the error wasn't because the file does not exist,
			// then log the error and continue with the download process
			logger.LogError(err, false, logger.ERROR)
		}
	}

	downloadPartial := false
	if supportRange {
		if downloadedBytes > 0 && downloadedBytes < fileReqContentLength {
			downloadPartial = true
			reqArgs.Headers["Range"] = fmt.Sprintf("bytes=%d-", downloadedBytes)
		}
	}

	res, err := reqArgs.RequestHandler(reqArgs)
	if err != nil {
		if err != context.Canceled {
			err = fmt.Errorf(
				"error %d: failed to download file, more info => %w\nurl: %s",
				errs.DOWNLOAD_ERROR,
				err,
				reqArgs.Url,
			)
		}
		return err
	}
	defer res.Body.Close()

	filePath, err = getFullFilePath(res, filePath)
	if err != nil {
		return err
	}

	if !checkIfCanSkipDl(downloadedBytes, fileReqContentLength, overwriteExistingFile) {
		err = dlToFile(reqArgs.Context, res, reqArgs.Url, filePath, downloadPartial)
	}
	return err
}

// DownloadUrls is used to download multiple files from URLs concurrently
//
// Note: If the file already exists, the download process will be skipped
func DownloadUrlsWithHandler(urlInfoSlice []*ToDownload, dlOptions *DlOptions, config *configs.Config, reqHandler RequestHandler) error {
	urlsLen := len(urlInfoSlice)
	if urlsLen == 0 {
		return nil
	}
	if urlsLen < dlOptions.MaxConcurrency {
		dlOptions.MaxConcurrency = urlsLen
	}

	var wg sync.WaitGroup
	queue := make(chan struct{}, dlOptions.MaxConcurrency)
	errChan := make(chan error, urlsLen)

	// Create a context that can be cancelled when SIGINT/SIGTERM signal is received
	ctx, cancel := context.WithCancel(dlOptions.Context)
	defer cancel()

	// Catch SIGINT/SIGTERM signal and cancel the context when received
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()
	defer signal.Stop(sigs)

	baseMsg := "Downloading files [%d/" + fmt.Sprintf("%d]...", urlsLen)
	progress := dlOptions.DownloadProgressBar
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished downloading %d files",
			urlsLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while downloading %d files.\nPlease refer to the logs for more details.",
			urlsLen,
		),
	)
	progress.UpdateMax(urlsLen)
	progress.Start()
	for _, urlInfo := range urlInfoSlice {
		wg.Add(1)
		go func(fileUrl, filePath string) {
			defer func() {
				wg.Done()
				<-queue
			}()
			err := downloadUrl(
				filePath,
				queue,
				&RequestArgs{
					Url:            fileUrl,
					Method:         "GET",
					Timeout:        constants.DOWNLOAD_TIMEOUT,
					Cookies:        dlOptions.Cookies,
					Headers:        dlOptions.Headers,
					Http2:          !dlOptions.UseHttp3,
					Http3:          dlOptions.UseHttp3,
					RetryDelay:     dlOptions.RetryDelay,
					UserAgent:      config.UserAgent,
					RequestHandler: reqHandler,
					Context:        ctx,
				},
				config.OverwriteFiles,
				dlOptions.SupportRange,
			)
			if err != nil {
				errChan <- err
			}

			if err != context.Canceled {
				progress.Increment()
			}
		}(urlInfo.Url, urlInfo.FilePath)
	}
	wg.Wait()
	close(queue)
	close(errChan)

	hasErr := false
	if len(errChan) > 0 {
		hasErr = true
		if hasCancelled := logger.LogChanErrors(false, logger.ERROR, errChan); hasCancelled {
			progress.StopInterrupt("Stopped downloading files (incomplete downloads will be deleted)...")
			return context.Canceled
		} 
	}
	progress.Stop(hasErr)
	return nil
}

// Same as DownloadUrlsWithHandler but uses the default request handler (CallRequest)
func DownloadUrls(urlInfoSlice []*ToDownload, dlOptions *DlOptions, config *configs.Config) error {
	return DownloadUrlsWithHandler(urlInfoSlice, dlOptions, config, CallRequest)
}
