package gdrive

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

func md5HashFile(file *os.File) (string, error) {
	md5Checksum := md5.New()
	_, err := io.Copy(md5Checksum, file)
	if err != nil {
		return "", fmt.Errorf(
			"gdrive error %d: failed to calculate file's md5 checksum, more info => %w",
			cdlerrors.OS_ERROR,
			err,
		)
	}
	return fmt.Sprintf("%x", md5Checksum.Sum(nil)), nil
}

func checkIfCanSkipDl(filePath string, fileInfo *GdriveFileToDl) (bool, int64, error) {
	if !iofuncs.PathExists(filePath) {
		return false, 0, nil
	}

	// check the md5 checksum and the file size
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return false, 0, fmt.Errorf(
			"gdrive error %d: failed to open file %q, more info => %w",
			cdlerrors.OS_ERROR,
			filePath,
			err,
		)
	}
	defer file.Close()

	fileStatInfo, err := file.Stat()
	if err != nil {
		return false, 0, fmt.Errorf(
			"gdrive error %d: failed to get file stat info of %q, more info => %w",
			cdlerrors.OS_ERROR,
			filePath,
			err,
		)
	}

	fileSize := fileStatInfo.Size()
	if strconv.FormatInt(fileSize, 10) != fileInfo.Size {
		return false, fileSize, nil
	}

	md5Checksum, err := md5HashFile(file)
	if err != nil {
		return false, 0, err
	}

	matchChecksum := md5Checksum == fileInfo.Md5Checksum
	if !matchChecksum {
		fileSize = 0 // overwrite the file
	}
	return matchChecksum, fileSize, nil
}

// Downloads the given GDrive file using GDrive API v3
//
// If the md5Checksum has a mismatch, the file will be overwritten and downloaded again
// Downloads the given GDrive file using GDrive API v3
//
// If the md5Checksum has a mismatch, the file will be overwritten and downloaded again
func (gdrive *GDrive) DownloadFile(ctx context.Context, fileInfo *GdriveFileToDl, filePath string, config *configs.Config, progBarInfo *progress.ProgressBarInfo, queue chan struct{}) error {
	skipDl, writtenBytes, err := checkIfCanSkipDl(filePath, fileInfo)
	if skipDl || err != nil {
		return err
	}

	queue <- struct{}{}

	var dlProgBar *progress.DownloadProgressBar
	if progBarInfo.DownloadProgressBars != nil {
		dlProgBar = progress.NewDlProgressBar(ctx, progress.Messages{
			Msg:        "Downloading GDrive file...",
			ErrMsg:     "Failed to download GDrive file!",
			SuccessMsg: "Finished downloading GDrive file!",
		})
		(*dlProgBar).UpdateFilename(filepath.Base(filePath))
		progBarInfo.AppendDlProgBar(dlProgBar)
	}

	var res *http.Response
	url := fmt.Sprintf("%s/%s", gdrive.apiUrl, fileInfo.Id)
	if gdrive.client != nil {
		fileCall := gdrive.client.Files.Get(fileInfo.Id).AcknowledgeAbuse(true).Context(ctx)
		if writtenBytes > 0 {
			// If the file has been partially downloaded, resume the download from where it left off
			fileCall.Header().Add("Range", fmt.Sprintf("bytes=%d-", writtenBytes))
		}
		res, err = fileCall.Download()
	} else {
		params := map[string]string{
			"key":              gdrive.apiKey,
			"alt":              "media", // to tell Google that we are downloading the file
			"acknowledgeAbuse": "true",  // If the files are marked as abusive, download them anyway
		}
		headers := map[string]string{}
		if writtenBytes > 0 {
			headers["Range"] = fmt.Sprintf("bytes=%d-", writtenBytes)
		}
		res, err = httpfuncs.CallRequest(
			&httpfuncs.RequestArgs{
				Url:       url,
				Method:    "GET",
				Timeout:   gdrive.downloadTimeout,
				Params:    params,
				Context:   ctx,
				UserAgent: config.UserAgent,
				Http2:     !constants.GDRIVE_HTTP3_SUPPORTED,
				Http3:     constants.GDRIVE_HTTP3_SUPPORTED,
				Headers:   headers,
			},
		)
	}
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return getFailedApiCallErr(res)
	}

	dlReqInfo := &httpfuncs.DlRequestInfo{
		Ctx: ctx,
		Url: url,
	}
	dlPartialInfo := httpfuncs.PartialDlInfo{
		DownloadPartial:  true,
		DownloadedBytes:  writtenBytes,
		ExpectedFileSize: fileInfo.GetIntSize(),
	}
	return httpfuncs.DlToFile(res, dlReqInfo, filePath, dlPartialInfo, dlProgBar)
}

func filterDownloads(files []*GdriveFileToDl) []*GdriveFileToDl {
	var notAllowedForDownload []*GdriveFileToDl
	allowedForDownload := make([]*GdriveFileToDl, 0, len(files))
	for _, file := range files {
		if strings.Contains(file.MimeType, "application/vnd.google-apps") {
			notAllowedForDownload = append(notAllowedForDownload, file)
		} else {
			allowedForDownload = append(allowedForDownload, file)
		}
	}

	if len(notAllowedForDownload) > 0 {
		noticeMsg := "The following files are not allowed for download:\n"
		for _, file := range notAllowedForDownload {
			noticeMsg += fmt.Sprintf(
				"Filename: %s (ID: %s, MIME Type: %s)\n",
				file.Name, file.Id, file.MimeType,
			)
		}
		logger.LogError(errors.New(noticeMsg), false, logger.INFO)
	}
	return allowedForDownload
}

func (gdrive *GDrive) processGdriveDlError(errChan chan *GdriveError, prog progress.ProgressBar) []error {
	defer prog.SnapshotTask()
	if len(errChan) == 0 {
		return nil
	}

	killProgram := false
	errSlice := make([]error, 0, len(errChan))
	for errInfo := range errChan {
		if errors.Is(errInfo.Err, context.Canceled) {
			if !killProgram {
				killProgram = true
			}
			continue
		}

		errSlice = append(errSlice, errInfo.Err)
		errMsg := censorApiKeyFromStr(errInfo.Err.Error())
		logger.LogMessageToPath(
			censorApiKeyFromStr(errMsg),
			errInfo.FilePath,
			logger.ERROR,
		)
	}

	if killProgram {
		gdrive.cancel()
		prog.StopInterrupt(
			"Stopped downloading GDrive files (incomplete downloads may be deleted)...",
		)
		return nil
	}
	return errSlice
}

// Downloads the multiple GDrive file in parallel using GDrive API v3
func (gdrive *GDrive) DownloadMultipleFiles(files []*GdriveFileToDl, config *configs.Config, progBarInfo *progress.ProgressBarInfo) []error {
	allowedForDownload := filterDownloads(files)
	dlLen := len(allowedForDownload)
	if dlLen == 0 {
		return nil
	}

	// Create a context that can be cancelled when SIGINT/SIGTERM signal is received
	ctx, cancel := context.WithCancel(gdrive.ctx)
	defer cancel()

	// Catch SIGINT/SIGTERM signal and cancel the context when received
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()
	defer signal.Stop(sigs)

	maxConcurrency := gdrive.maxDownloadWorkers
	if dlLen < maxConcurrency {
		maxConcurrency = dlLen
	}
	var wg sync.WaitGroup
	queue := make(chan struct{}, maxConcurrency)
	errChan := make(chan *GdriveError, dlLen)

	baseMsg := "Downloading GDrive files [%d/" + fmt.Sprintf("%d]...", dlLen)
	prog := progBarInfo.MainProgressBar
	prog.SetToProgressBar()
	prog.UpdateBaseMsg(baseMsg)
	prog.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished downloading %d GDrive files!",
			dlLen,
		),
	)
	prog.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while downloading %d GDrive files!\nPlease refer to the generated log files for more details.",
			dlLen,
		),
	)
	prog.UpdateMax(dlLen)
	prog.Start()
	for _, file := range allowedForDownload {
		wg.Add(1)
		go func(file *GdriveFileToDl) {
			defer func() {
				wg.Done()
				<-queue
			}()

			os.MkdirAll(file.FilePath, 0755)
			filePath := filepath.Join(file.FilePath, file.Name)

			err := gdrive.DownloadFile(ctx, file, filePath, config, progBarInfo, queue)
			if err != nil && err != context.Canceled {
				err = fmt.Errorf(
					"failed to download file: %s (ID: %s, MIME Type: %s)\nRefer to error details below:\n%w",
					file.Name, file.Id, file.MimeType, err,
				)
				errChan <- &GdriveError{
					Err: err,
					FilePath: filepath.Join(
						file.FilePath,
						constants.GDRIVE_ERROR_FILENAME,
					),
				}
			}
			prog.Increment()
		}(file)
	}
	wg.Wait()
	close(queue)
	close(errChan)

	hasErr := false
	if len(errChan) > 0 {
		hasErr = true
	}
	prog.Stop(hasErr)
	return gdrive.processGdriveDlError(errChan, prog)
}

// Uses regex to extract the file ID and the file type (type: file, folder) from the given URL
func GetFileIdAndTypeFromUrl(url string) (string, string) {
	matched := constants.GDRIVE_URL_REGEX.FindStringSubmatch(url)
	if matched == nil {
		return "", ""
	}

	var fileType string
	matchedFileType := matched[constants.GDRIVE_REGEX_TYPE_INDEX]
	if strings.Contains(matchedFileType, "folder") {
		fileType = "folder"
	} else if strings.Contains(matchedFileType, "file") {
		fileType = "file"
	} else {
		err := fmt.Errorf(
			"gdrive error %d: could not determine file type from URL, %q",
			cdlerrors.DEV_ERROR,
			url,
		)
		logger.LogError(err, false, logger.ERROR)
		return "", ""
	}
	return matched[constants.GDRIVE_REGEX_ID_INDEX], fileType
}

func (gdrive *GDrive) getGdriveFileInfo(gdriveId *GDriveToDl, config *configs.Config) ([]*GdriveFileToDl, *GdriveError) {
	switch gdriveId.Type {
	case "file":
		fileInfo, err := gdrive.GetFileDetails(
			gdriveId,
			config,
		)
		if err != nil {
			return nil, &GdriveError{
				Err:      err,
				FilePath: gdriveId.FilePath,
			}
		}
		fileInfo.FilePath = gdriveId.FilePath
		return []*GdriveFileToDl{fileInfo}, nil
	case "folder":
		filesInfo, err := gdrive.GetNestedFolderContents(
			gdriveId.Id,
			gdriveId.FilePath,
			config,
		)
		if err != nil {
			return nil, &GdriveError{
				Err:      err,
				FilePath: gdriveId.FilePath,
			}
		}
		var gdriveFilesInfo []*GdriveFileToDl
		for _, fileInfo := range filesInfo {
			fileInfo.FilePath = gdriveId.FilePath
			gdriveFilesInfo = append(gdriveFilesInfo, fileInfo)
		}
		return gdriveFilesInfo, nil
	default:
		return nil, &GdriveError{
			Err: fmt.Errorf(
				"gdrive error %d: unknown Google Drive URL type, %q",
				cdlerrors.DEV_ERROR,
				gdriveId.Type,
			),
			FilePath: gdriveId.FilePath,
		}
	}
}

// Downloads multiple GDrive files based on a slice of GDrive URL strings in parallel
func (gdrive *GDrive) DownloadGdriveUrls(gdriveUrls []*httpfuncs.ToDownload, config *configs.Config, progBarInfo *progress.ProgressBarInfo) []error {
	if len(gdriveUrls) == 0 {
		return nil
	}

	// Retrieve the id from the url text
	var gdriveIds []*GDriveToDl
	for _, gdriveUrl := range gdriveUrls {
		fileId, fileType := GetFileIdAndTypeFromUrl(gdriveUrl.Url)
		if fileId != "" && fileType != "" {
			gdriveIds = append(gdriveIds, &GDriveToDl{
				Id:       fileId,
				Type:     fileType,
				FilePath: gdriveUrl.FilePath,
			})
		}
	}

	// Note: Can't do API calls concurrently as to avoid being blocked by Google's bot detection
	var gdriveErrSlice []*GdriveError
	var gdriveFilesInfo []*GdriveFileToDl
	gdriveIdsLen := len(gdriveIds)
	baseMsg := "Getting GDrive file information from GDrive ID(s) [%d/" + fmt.Sprintf("%d]...", len(gdriveIds))
	mainProg := progBarInfo.MainProgressBar
	mainProg.SetToProgressBar()
	mainProg.UpdateBaseMsg(baseMsg)
	mainProg.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished getting GDrive file information from %d GDrive ID(s)!",
			gdriveIdsLen,
		),
	)
	mainProg.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while getting GDrive file information from %d GDrive ID(s)!\nPlease refer to the generated log files for more details.",
			gdriveIdsLen,
		),
	)
	mainProg.UpdateMax(gdriveIdsLen)
	mainProg.Start()
	for _, gdriveId := range gdriveIds {
		fileInfo, err := gdrive.getGdriveFileInfo(gdriveId, config)
		if err != nil {
			if errors.Is(err.Err, context.Canceled) {
				gdrive.cancel()
				mainProg.StopInterrupt("Stopped getting GDrive file information...")
				mainProg.SnapshotTask()
				return nil
			}

			gdriveErrSlice = append(gdriveErrSlice, err)
		} else {
			gdriveFilesInfo = append(gdriveFilesInfo, fileInfo...)
		}
		mainProg.Increment()
	}

	hasErr := false
	errSlice := make([]error, 0, len(gdriveErrSlice))
	if len(gdriveErrSlice) > 0 {
		hasErr = true
		for _, err := range gdriveErrSlice {
			logger.LogMessageToPath(
				censorApiKeyFromStr(err.Err.Error()),
				err.FilePath,
				logger.ERROR,
			)
			errSlice = append(errSlice, err.Err)
		}
	}
	mainProg.Stop(hasErr)
	mainProg.SnapshotTask()

	errSlice = append(errSlice, gdrive.DownloadMultipleFiles(gdriveFilesInfo, config, progBarInfo)...)
	return errSlice
}
