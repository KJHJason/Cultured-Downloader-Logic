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
	"time"

	"github.com/jbenet/go-context/io"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
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
	}

	if !forceOverwrite && fileSize > 0 {
		// If the file already exists and have more than 0 bytes
		// but the Content-Length header does not exist in the response,
		// we will assume that the file is already downloaded
		// and skip the download process if the overwrite flag is false.
		return true
	}
	return false
}

// totalBytesWriter is a custom type that implements io.Writer interface to accumulate totalBytes.
type totalBytesWriter struct {
	totalBytes *int64
}

// Write writes len(p) bytes from p to accumulate the total bytes written.
func (tbw *totalBytesWriter) Write(p []byte) (int, error) {
	n := len(p)
	*tbw.totalBytes += int64(n)
	return n, nil
}

type DlRequestInfo struct {
	Ctx context.Context
	Url string
}

type PartialDlInfo struct {
	DownloadPartial  bool
	DownloadedBytes  int64
	ExpectedFileSize int64
} 

func DlToFile(res *http.Response, dlRequestInfo *DlRequestInfo, filePath string, partialDlInfo PartialDlInfo, dlProgBar *progress.DownloadProgressBar) error {
	fileFlags := os.O_CREATE | os.O_WRONLY
	if partialDlInfo.DownloadPartial {
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

	expectedFileSize := partialDlInfo.ExpectedFileSize
	writtenBytes := partialDlInfo.DownloadedBytes
	if writtenBytes == -1 { // since the iofuncs.GetFileSize function returns -1 if the file does not exist
		writtenBytes = 0
	}

	var progressTicker *time.Ticker
	dlInfoCtx, cancelDlInfoCtx := context.WithCancel(dlRequestInfo.Ctx)
	if dlProgBar != nil {
		// Measure download speed and ETA
		startTime := time.Now()
		derefDlProgBar := *dlProgBar
		progressTicker := time.NewTicker(100 * time.Millisecond)
		go func() {
			for {
				select {
				case <-dlInfoCtx.Done():
					return
				case <-progressTicker.C:
					duration := time.Since(startTime)
					downloadSpeed := float64(writtenBytes) / duration.Seconds()

					var estimatedTime float64
					if expectedFileSize == -1 { // not present in the response
						estimatedTime = -1 // -1 indicates that the ETA is unknown
					} else {
						estimatedTime = float64(expectedFileSize-writtenBytes) / downloadSpeed
						derefDlProgBar.UpdatePercentage(int(float64(writtenBytes) / float64(expectedFileSize) * 100))
					}
					derefDlProgBar.UpdateDownloadSpeed(downloadSpeed/1024/1024)
					derefDlProgBar.UpdateDownloadETA(estimatedTime)
					// fmt.Printf("\rDownload speed: %.2f MB/s | ETA: %.2f seconds", downloadSpeed/1024/1024, estimatedTime)
				}
			}
		}()
	}

	// write the body to file
	respReader := ctxio.NewReader(dlRequestInfo.Ctx, res.Body)
	_, err = io.Copy(io.MultiWriter(file, &totalBytesWriter{&writtenBytes}), respReader)
	if dlProgBar != nil {
		progressTicker.Stop()
	}
	cancelDlInfoCtx()

	if err != nil {
		if !partialDlInfo.DownloadPartial {
			// Due to the checkIfCanSkipDl check before downloading, 
			// remove the file if the download process failed or was cancelled
			// to prevent incomplete files from being kept when the server does not support range requests.
			err := os.Remove(filePath)
			if err != nil {
				logger.LogError(
					fmt.Errorf(
						"error %d: failed to remove file %s, more info => %w",
						errs.OS_ERROR,
						filePath,
						err,
					),
					false,
					logger.ERROR,
				)
			}
		}

		if err == context.Canceled {
			if dlProgBar != nil {
				(*dlProgBar).UpdateErrMsg("Download process was cancelled!")
				(*dlProgBar).Stop(true)
			}
			return nil
		}

		if dlProgBar != nil {
			(*dlProgBar).Stop(true)
		}
		logger.LogError(
			fmt.Errorf(
				"failed to download %s due to %w",
				dlRequestInfo.Url,
				err,
			), 
			false, 
			logger.ERROR,
		)
		return err
	}
	return nil
}

// DownloadUrl is used to download a file from a URL.
// Note: If the file already exists, the download process will be skipped
func downloadUrl(filePath string, queue chan struct{}, reqArgs *RequestArgs, overwriteExistingFile bool, dlOptions *DlOptions) error {
	queue <- struct{}{}

	var dlProgBar *progress.DownloadProgressBar
	if dlOptions.ProgressBarInfo.DownloadProgressBars != nil {
		dlProgBar = progress.NewDlProgressBar(reqArgs.Context, progress.Messages{
			Msg:        "Downloading file...",
			ErrMsg:     "Failed to download file!",
			SuccessMsg: "Finished downloading file!",
		})
		(*dlProgBar).UpdateFilename(filepath.Base(filePath))
		dlOptions.ProgressBarInfo.AppendDlProgBar(dlProgBar)
	}

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
	if dlOptions.SupportRange {
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
		dlReqInfo := &DlRequestInfo{
			Ctx: reqArgs.Context,
			Url: reqArgs.Url,
		}
		dlPartialInfo := PartialDlInfo{
			DownloadPartial:  downloadPartial,
			DownloadedBytes:  downloadedBytes,
			ExpectedFileSize: fileReqContentLength,
		}
		err = DlToFile(res, dlReqInfo, filePath, dlPartialInfo, dlProgBar)
	}
	return err
}

// DownloadUrls is used to download multiple files from URLs concurrently
//
// Note: If the file already exists, the download process will be skipped
func DownloadUrlsWithHandler(urlInfoSlice []*ToDownload, dlOptions *DlOptions, config *configs.Config, reqHandler RequestHandler) (cancelled bool, errors []*error) {
	urlsLen := len(urlInfoSlice)
	if urlsLen == 0 {
		return false, nil
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
	progress := dlOptions.ProgressBarInfo.MainProgressBar
	progress.SetToProgressBar()
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
	defer progress.SnapshotTask()
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
				dlOptions,
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
	var errorSlice []*error
	if len(errChan) > 0 {
		hasErr = true
		if hasCancelled, errSlice := logger.LogChanErrors(false, logger.ERROR, errChan); hasCancelled {
			progress.StopInterrupt("Stopped downloading files (incomplete downloads will be resumed later or be deleted)...")
			return true, errSlice
		} else {
			errorSlice = errSlice
		}
	}
	progress.Stop(hasErr)
	return false, errorSlice
}

// Same as DownloadUrlsWithHandler but uses the default request handler (CallRequest)
func DownloadUrls(urlInfoSlice []*ToDownload, dlOptions *DlOptions, config *configs.Config) (cancelled bool, errors []*error) {
	return DownloadUrlsWithHandler(urlInfoSlice, dlOptions, config, CallRequest)
}
