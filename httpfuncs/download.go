package httpfuncs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	ctxio "github.com/jbenet/go-context/io"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/filters"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/metadata"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils/threadsafe"
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
			cdlerrors.UNEXPECTED_ERROR,
			err,
			res.Request.URL.String(),
		)
	}
	filename = GetLastPartOfUrl(filename)
	filenameWithoutExt := iofuncs.RemoveExtFromFilename(filename)
	filePath = filepath.Join(
		filePath,
		filenameWithoutExt+strings.ToLower(filepath.Ext(filename)),
	)
	return filePath, nil
}

type skipDlArgs struct {
	ctx            context.Context
	filePath       string
	fileSize       int64
	contentLength  int64
	forceOverwrite bool
	supportRange   bool
	setMetadata    bool
}

// check if the file size matches the content length
// if not, then the file does not exist or is corrupted and should be re-downloaded
func checkIfCanSkipDl(skipDlArgsVal skipDlArgs) bool {
	if skipDlArgsVal.forceOverwrite {
		return false
	}
	if skipDlArgsVal.fileSize == -1 {
		return false // file does not exist
	}

	if skipDlArgsVal.setMetadata && skipDlArgsVal.fileSize > skipDlArgsVal.contentLength {
		fileSizeWithoutMetadata, err := metadata.GetFileSizeWithoutExifData(skipDlArgsVal.ctx, skipDlArgsVal.filePath)
		if err != nil {
			logger.LogError(
				fmt.Errorf(
					"error %d: failed to get file size without metadata, more info => %w",
					cdlerrors.UNEXPECTED_ERROR,
					err,
				),
				logger.ERROR,
			)
			return false
		}
		skipDlArgsVal.fileSize = fileSizeWithoutMetadata
	}

	if skipDlArgsVal.fileSize == skipDlArgsVal.contentLength {
		// If the file already exists and the file size
		// is equal to the expected file size in the Content-Length header,
		// then skip the download process.
		return true
	}

	if skipDlArgsVal.fileSize > 0 && !skipDlArgsVal.supportRange {
		// If the file already exists and have more than 0 bytes
		// but the server doesn't have Content-Length headers or doesn't
		// support range requests, we will assume that the file is already downloaded
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
	Ctx     context.Context
	Url     string
	Filters *filters.Filters
}

type PartialDlInfo struct {
	DownloadPartial  bool
	DownloadedBytes  int64
	ExpectedFileSize int64
}

func writeDlDetailsToProgBar(dlProgBar *progress.DownloadProgressBar, startTime time.Time, bytesOnDisk, reqWrittenBytes, expectedFileSize int64) {
	durationInSec := time.Since(startTime).Seconds()
	var downloadSpeed float64
	if durationInSec > 0 {
		downloadSpeed = float64(reqWrittenBytes) / durationInSec
		(*dlProgBar).UpdateDownloadSpeed(downloadSpeed / 1024 / 1024)
	} else {
		downloadSpeed = 0
	}

	var estimatedTime float64
	var progressPercentage float64
	if expectedFileSize <= 0 { // not present in the response or the time elapsed is too short
		estimatedTime = -1 // -1 indicates that the ETA is unknown
		progressPercentage = 0
	} else {
		// Calculate the total progress made so far, including initial progress from bytesOnDisk
		totalProgress := reqWrittenBytes + bytesOnDisk

		remainingBytesLenToDl := expectedFileSize - totalProgress
		if remainingBytesLenToDl <= 0 {
			// As a fallback, set the estimated
			// time to -1 to avoid negative values.
			estimatedTime = -1
			progressPercentage = 0
		} else {
			// Calculate the progress percentage based on total bytes and written bytes
			progressPercentage = float64(totalProgress) / float64(expectedFileSize) * 100

			// Calculate the estimated time based on remaining bytes to be downloaded and download speed
			estimatedTime = float64(remainingBytesLenToDl) / downloadSpeed
		}
	}
	(*dlProgBar).UpdateDownloadETA(estimatedTime)

	if progressPercentage > 100 {
		progressPercentage = 100
	} else if progressPercentage < 0 {
		progressPercentage = 0
	}
	(*dlProgBar).UpdatePercentage(int(progressPercentage))
}

func DlToFile(res *http.Response, dlRequestInfo *DlRequestInfo, filePath string, partialDlInfo PartialDlInfo, dlProgBar *progress.DownloadProgressBar) error {
	fileFlags := os.O_CREATE | os.O_WRONLY
	if partialDlInfo.DownloadPartial {
		fileFlags |= os.O_APPEND
	} else {
		fileFlags |= os.O_TRUNC
	}

	filters := dlRequestInfo.Filters
	if partialDlInfo.ExpectedFileSize != -1 {
		if !filters.IsFileSizeInRange(partialDlInfo.ExpectedFileSize) {
			return nil
		}
	}

	if !filters.IsFilePathExtValid(filePath) || !filters.IsFilePathFileNameValid(filePath) {
		return nil
	}

	file, err := os.OpenFile(filePath, fileFlags, 0644)
	if err != nil {
		return fmt.Errorf(
			"error %d: failed to open/create file, more info => %w\nfile path: %s",
			cdlerrors.OS_ERROR,
			err,
			filePath,
		)
	}
	defer file.Close()

	var reqWrittenBytes int64
	expectedFileSize := partialDlInfo.ExpectedFileSize
	bytesOnDisk := partialDlInfo.DownloadedBytes
	if bytesOnDisk == -1 { // since the iofuncs.GetFileSize function returns -1 if the file does not exist
		bytesOnDisk = 0
	}

	progressTicker := time.NewTicker(100 * time.Millisecond)
	dlInfoCtx, cancelDlInfoCtx := context.WithCancel(dlRequestInfo.Ctx)
	hasDlProgBar := dlProgBar != nil
	if hasDlProgBar {
		// Measure download speed and ETA
		startTime := time.Now()
		go func() {
			for {
				select {
				case <-dlInfoCtx.Done():
					(*dlProgBar).UpdateDownloadETA(0)
					(*dlProgBar).UpdateDownloadSpeed(0)
					return
				case <-progressTicker.C:
					writeDlDetailsToProgBar(dlProgBar, startTime, bytesOnDisk, reqWrittenBytes, expectedFileSize)
				}
			}
		}()
	}

	// write the body to file
	respReader := ctxio.NewReader(dlRequestInfo.Ctx, res.Body)
	_, err = io.Copy(io.MultiWriter(file, &totalBytesWriter{&reqWrittenBytes}), respReader)
	progressTicker.Stop()
	cancelDlInfoCtx()

	if err == nil && hasDlProgBar {
		(*dlProgBar).UpdatePercentage(100)
		(*dlProgBar).Stop(false)
	} else {
		if !partialDlInfo.DownloadPartial {
			// Due to the checkIfCanSkipDl check before downloading,
			// remove the file if the download process failed or was cancelled
			// to prevent incomplete files from being kept when the server does not support range requests.
			fileErr := os.Remove(filePath)
			if fileErr != nil {
				logger.LogError(
					fmt.Errorf(
						"error %d: failed to remove file %s, more info => %w",
						cdlerrors.OS_ERROR,
						filePath,
						fileErr,
					),
					logger.ERROR,
				)
			}
		}

		if errors.Is(err, context.Canceled) {
			if hasDlProgBar {
				(*dlProgBar).UpdateErrMsg("Download process was cancelled!")
				(*dlProgBar).Stop(true)
			}
			return nil
		}

		if hasDlProgBar {
			(*dlProgBar).Stop(true)
		}
		logger.LogError(
			fmt.Errorf(
				"failed to download %s due to %w",
				dlRequestInfo.Url,
				err,
			),
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

	res, err := reqArgs.RequestHandler(reqArgs)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			err = fmt.Errorf(
				"error %d: failed to download file, more info => %w\nurl: %s",
				cdlerrors.DOWNLOAD_ERROR,
				err,
				reqArgs.Url,
			)
		}
		return err
	}
	defer res.Close()
	fileReqContentLength := res.Resp.ContentLength

	filePath, err = getFullFilePath(res.Resp, filePath)
	if err != nil {
		return err
	}

	downloadedBytes, err := iofuncs.GetFileSize(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// if the error wasn't because the file does not exist,
			// then log the error and continue with the download process
			logger.LogError(err, logger.ERROR)
		}
	}

	downloadPartial := false
	if dlOptions.SupportRange {
		if downloadedBytes > 0 && downloadedBytes < fileReqContentLength {
			downloadPartial = true
			reqArgs.EditMu.Lock()
			if reqArgs.Headers == nil {
				reqArgs.Headers = make(map[string]string)
			}
			reqArgs.Headers["Range"] = fmt.Sprintf("bytes=%d-", downloadedBytes)
			reqArgs.EditMu.Unlock()
		}
	}

	var dlProgBar *progress.DownloadProgressBar
	hasDlProgBar := dlOptions.ProgressBarInfo.DownloadProgressBars != nil
	if hasDlProgBar {
		dlProgBar = progress.NewDlProgressBar(reqArgs.Context, progress.Messages{
			Msg:        "Downloading file...",
			ErrMsg:     "Failed to download file!",
			SuccessMsg: "Finished downloading file!",
		})
		dlProgBar.UpdateTotalBytes(fileReqContentLength)
		dlProgBar.UpdateFilename(filepath.Base(filePath))
		dlOptions.ProgressBarInfo.AppendDlProgBar(dlProgBar)
	}

	skipDlArgsVal := skipDlArgs{
		ctx:            reqArgs.Context,
		filePath:       filePath,
		fileSize:       downloadedBytes,
		contentLength:  fileReqContentLength,
		forceOverwrite: overwriteExistingFile,
		supportRange:   dlOptions.SupportRange,
		setMetadata:    dlOptions.SetMetadata,
	}
	if !checkIfCanSkipDl(skipDlArgsVal) {
		dlReqInfo := &DlRequestInfo{
			Ctx:     reqArgs.Context,
			Url:     reqArgs.Url,
			Filters: dlOptions.Filters,
		}
		dlPartialInfo := PartialDlInfo{
			DownloadPartial:  downloadPartial,
			DownloadedBytes:  downloadedBytes,
			ExpectedFileSize: fileReqContentLength,
		}
		err = DlToFile(res.Resp, dlReqInfo, filePath, dlPartialInfo, dlProgBar)
	} else {
		if hasDlProgBar {
			dlProgBar.UpdateTotalBytes(downloadedBytes)
			dlProgBar.UpdateSuccessMsg("File already exists!")
			dlProgBar.Stop(false)
		}
	}
	return err
}

type cacheEl struct {
	hasErr   bool
	cacheKey string
	cacheFn  func(key string) // Note: no need to use batch call as the update is done sequentially
}

// DownloadUrls is used to download multiple files from URLs concurrently
//
// Note: If the file already exists, the download process will be skipped
func DownloadUrlsWithHandler(urlInfoSlice []*ToDownload, dlOptions *DlOptions, config *configs.Config, reqHandler RequestHandler) (cancelled bool, errorSlice []error) {
	urlsLen := len(urlInfoSlice)
	if urlsLen == 0 {
		return false, nil
	}
	if urlsLen < dlOptions.MaxConcurrency {
		dlOptions.MaxConcurrency = urlsLen
	}

	var wg sync.WaitGroup
	queue := make(chan struct{}, dlOptions.MaxConcurrency)
	errTsSlice := threadsafe.NewSlice[error]()
	cacheTsSlice := threadsafe.NewSliceWithCapacity[*cacheEl](urlsLen)

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
		go func() {
			defer func() {
				wg.Done()
				<-queue
			}()
			err := downloadUrl(
				urlInfo.FilePath,
				queue,
				&RequestArgs{
					Method:         "GET",
					Url:            urlInfo.Url,
					Timeout:        constants.DOWNLOAD_TIMEOUT,
					Headers:        dlOptions.Headers,
					Cookies:        dlOptions.Cookies,
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
			hasErr := err != nil
			if hasErr {
				errTsSlice.Append(err)
			}
			if urlInfo.CacheKey != "" {
				cacheTsSlice.Append(&cacheEl{
					hasErr:   hasErr,
					cacheKey: urlInfo.CacheKey,
					cacheFn:  urlInfo.CacheFn,
				})
			}

			if !errors.Is(err, context.Canceled) {
				progress.Increment()
			}
		}()
	}
	wg.Wait()
	close(queue)

	// since the CacheKey in the ToDownload struct can belong to multiple URLs
	// E.g. CacheKey of "https://example.com/post/123456" for all elements,
	// [
	// 	{url: "https://example.com/post/123456/image1.jpg", cacheKey: "https://example.com/post/123456"},
	// 	{url: "https://example.com/post/123456/image2.jpg", cacheKey: "https://example.com/post/123456"},
	// ]
	// we have to make sure all the request for that particular cache key has no errors to assume that the download was successful
	var hasSeenCacheKey map[string]*cacheEl
	if cacheTsSlice.LenUnsafe() > 0 {
		hasSeenCacheKey = make(map[string]*cacheEl)
		cacheElIter := cacheTsSlice.NewIter()
		for cacheElIter.Next() {
			cacheEl := cacheElIter.Item()
			if _, ok := hasSeenCacheKey[cacheEl.cacheKey]; ok {
				if cacheEl.hasErr {
					hasSeenCacheKey[cacheEl.cacheKey].hasErr = true
				}
				continue
			}
			hasSeenCacheKey[cacheEl.cacheKey] = cacheEl
		}

		for cacheKey, el := range hasSeenCacheKey {
			if el.hasErr {
				continue
			}
			if el.cacheFn != nil {
				el.cacheFn(cacheKey)
			} else { // default to database.CachePost
				database.CachePost(cacheKey)
			}
		}
	}

	hasErr := false
	if errTsSlice.LenUnsafe() > 0 {
		hasErr = true
		var hasCancelled bool
		if hasCancelled, errorSlice = logger.LogSliceErrors(logger.ERROR, errTsSlice); hasCancelled {
			progress.StopInterrupt("Stopped downloading files (incomplete downloads will be resumed later or be deleted)...")
			return true, errorSlice
		}
	}
	progress.Stop(hasErr)
	return false, errorSlice
}

// Same as DownloadUrlsWithHandler but uses the default request handler (CallRequest)
func DownloadUrls(urlInfoSlice []*ToDownload, dlOptions *DlOptions, config *configs.Config) (cancelled bool, errors []error) {
	return DownloadUrlsWithHandler(urlInfoSlice, dlOptions, config, CallRequest)
}
