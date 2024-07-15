package ugoira

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/extractor"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

// Map the Ugoira frame delays to their respective filenames
func MapDelaysToFilename(ugoiraFramesJson UgoiraFramesJson) map[string]int64 {
	frameInfoMap := map[string]int64{}
	for _, frame := range ugoiraFramesJson {
		frameInfoMap[frame.File] = int64(frame.Delay)
	}
	return frameInfoMap
}

type UgoiraFfmpegArgs struct {
	context       context.Context
	ffmpegPath    string
	outputPath    string
	ugoiraQuality int
}

// Converts the Ugoira to the desired output path using FFmpeg
func ConvertUgoira(ugoiraInfo *Ugoira, imagesFolderPath string, ugoiraFfmpeg *UgoiraFfmpegArgs) error {
	outputExt := filepath.Ext(ugoiraFfmpeg.outputPath)
	if !utils.SliceContains(UGOIRA_ACCEPTED_EXT, outputExt) {
		return fmt.Errorf(
			"pixiv error %d: Output extension %s is not allowed for ugoira conversion",
			cdlerrors.INPUT_ERROR,
			outputExt,
		)
	}

	concatDelayFilePath, sortedFilenames, err := writeDelays(ugoiraInfo, imagesFolderPath)
	if err != nil {
		return err
	}

	args, err := getFfmpegFlagsForUgoira(
		&ffmpegOptions{
			ugoiraArgs:          ugoiraFfmpeg,
			outputExt:           outputExt,
			concatDelayFilePath: concatDelayFilePath,
			sortedFilenames:     sortedFilenames,
		},
		imagesFolderPath,
	)
	if err != nil {
		return err
	}

	// convert the frames to a gif or a video
	cmd := exec.CommandContext(ugoiraFfmpeg.context, ugoiraFfmpeg.ffmpegPath, args...)
	utils.PrepareCmdForBgTask(cmd)
	// cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		os.Remove(ugoiraFfmpeg.outputPath)
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf(
			"pixiv error %d: failed to convert ugoira to %s, more info => %w",
			cdlerrors.CMD_ERROR,
			ugoiraFfmpeg.outputPath,
			err,
		)
	}

	// delete unzipped folder which contains
	// the frames images and the delays text file
	os.RemoveAll(imagesFolderPath)
	return nil
}

// Returns the ugoira's zip file path and the ugoira's converted file path
func GetUgoiraFilePaths(ugoireFilePath, ugoiraUrl, outputFormat string) (string, string) {
	filePath := filepath.Join(ugoireFilePath, httpfuncs.GetLastPartOfUrl(ugoiraUrl))
	outputFilePath := iofuncs.RemoveExtFromFilename(filePath) + outputFormat
	return filePath, outputFilePath
}

func convertUgoira(ctx context.Context, ugoira *Ugoira, ugoiraOptions *UgoiraOptions, config *configs.Config) error {
	zipFilePath, outputPath := GetUgoiraFilePaths(ugoira.FilePath, ugoira.Url, ugoiraOptions.OutputFormat)
	if iofuncs.PathExists(outputPath) || !iofuncs.PathExists(zipFilePath) {
		return nil
	}

	unzipFolderPath := filepath.Join(
		filepath.Dir(zipFilePath),
		"unzipped",
	)
	err := extractor.ExtractFiles(ctx, zipFilePath, unzipFolderPath, true)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		err := fmt.Errorf(
			"pixiv error %d: failed to unzip file %s, more info => %w",
			cdlerrors.OS_ERROR,
			zipFilePath,
			err,
		)
		return err
	}

	err = ConvertUgoira(
		ugoira,
		unzipFolderPath,
		&UgoiraFfmpegArgs{
			context:       ctx,
			ffmpegPath:    config.FfmpegPath,
			outputPath:    outputPath,
			ugoiraQuality: ugoiraOptions.Quality,
		},
	)
	if err == nil {
		if ugoiraOptions.DeleteZip {
			os.Remove(zipFilePath)
		}
		if ugoiraOptions.UseCacheDb {
			database.CacheUgoira(ugoira.CacheKey)
		}
	}
	return err
}

func convertMultipleUgoira(ugoiraArgs *UgoiraArgs, ugoiraOptions *UgoiraOptions, config *configs.Config) []error {
	// Create a context that can be cancelled when SIGINT/SIGTERM signal is received
	ctx, cancel := context.WithCancel(ugoiraArgs.context)
	defer cancel()

	// Catch SIGINT/SIGTERM signal and cancel the context when received
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()
	defer signal.Stop(sigs)

	downloadInfoLen := len(ugoiraArgs.ToDownload)
	maxConcurrency := config.FfmpegWorkers
	if maxConcurrency <= 0 {
		maxConcurrency = constants.FFMPEG_MAX_CONCURRENCY
	}
	if downloadInfoLen < maxConcurrency {
		maxConcurrency = downloadInfoLen
	}
	var wg sync.WaitGroup
	queue := make(chan struct{}, maxConcurrency)
	errChan := make(chan error, downloadInfoLen)

	baseMsg := fmt.Sprintf("Converting Ugoira to %s ", ugoiraOptions.OutputFormat) + "[%d/" + fmt.Sprintf("%d]...", downloadInfoLen)
	prog := ugoiraArgs.MainProgBar
	prog.UpdateBaseMsg(baseMsg)
	prog.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished converting %d Ugoira to %s!",
			downloadInfoLen,
			ugoiraOptions.OutputFormat,
		),
	)
	prog.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while converting %d Ugoira to %s!\nPlease refer to the logs for more details.",
			downloadInfoLen,
			ugoiraOptions.OutputFormat,
		),
	)
	prog.SetToProgressBar()
	prog.UpdateMax(downloadInfoLen)
	defer prog.SnapshotTask()
	prog.Start()
	for _, ugoira := range ugoiraArgs.ToDownload {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				<-queue
			}()
			queue <- struct{}{}
			err := convertUgoira(ctx, ugoira, ugoiraOptions, config)
			if err != nil {
				errChan <- err
			}
			prog.Increment()
		}()
	}
	wg.Wait()
	close(queue)
	close(errChan)

	var errSlice []error
	hasErr := len(errChan) > 0
	if hasErr {
		var hasCancelled bool
		if hasCancelled, errSlice = logger.LogChanErrors(logger.ERROR, errChan); hasCancelled {
			prog.StopInterrupt(
				fmt.Sprintf("Stopped converting ugoira to %s!", ugoiraOptions.OutputFormat),
			)
			ugoiraArgs.cancel()
		}
	}
	prog.Stop(hasErr)
	return errSlice
}

type UgoiraArgs struct {
	context      context.Context
	cancel       context.CancelFunc
	UseMobileApi bool
	ToDownload   []*Ugoira
	Cookies      []*http.Cookie
	MainProgBar  progress.ProgressBar
}

func (u *UgoiraArgs) SetContext(ctx context.Context) {
	u.context, u.cancel = context.WithCancel(ctx)
}

// Downloads multiple Ugoira artworks and converts them based on the output format
func DownloadMultipleUgoira(ugoiraArgs *UgoiraArgs, ugoiraOptions *UgoiraOptions, config *configs.Config, reqHandler httpfuncs.RequestHandler, setMetadata bool, progBarInfo *progress.ProgressBarInfo) []error {
	if ugoiraOptions.UseCacheDb {
		filteredUgoira := make([]*Ugoira, 0, len(ugoiraArgs.ToDownload))
		for _, ugoira := range ugoiraArgs.ToDownload {
			if ugoira.CacheKey != "" && database.UgoiraCacheExists(ugoira.CacheKey) {
				continue
			}
			filteredUgoira = append(filteredUgoira, ugoira)
		}
		ugoiraArgs.ToDownload = filteredUgoira
	}

	var urlsToDownload []*httpfuncs.ToDownload
	for _, ugoira := range ugoiraArgs.ToDownload {
		filePath, outputFilePath := GetUgoiraFilePaths(
			ugoira.FilePath,
			ugoira.Url,
			ugoiraOptions.OutputFormat,
		)
		if !iofuncs.PathExists(outputFilePath) {
			urlsToDownload = append(urlsToDownload, &httpfuncs.ToDownload{
				Url:      ugoira.Url,
				FilePath: filePath,
			})
		}
	}

	var useHttp3 bool
	var headers map[string]string
	if ugoiraArgs.UseMobileApi {
		headers = map[string]string{
			"Referer": constants.PIXIV_MOBILE_URL,
		}
	} else {
		headers = pixivcommon.GetPixivRequestHeaders()
		useHttp3 = httpfuncs.IsHttp3Supported(constants.PIXIV, true)
	}

	cancelled, err := httpfuncs.DownloadUrlsWithHandler(
		urlsToDownload,
		&httpfuncs.DlOptions{
			Context:         ugoiraArgs.context,
			MaxConcurrency:  constants.PIXIV_MAX_DOWNLOAD_CONCURRENCY,
			SupportRange:    constants.PIXIV_RANGE_SUPPORTED,
			HeadReqTimeout:  constants.DEFAULT_HEAD_REQ_TIMEOUT,
			Headers:         headers,
			Cookies:         ugoiraArgs.Cookies,
			UseHttp3:        useHttp3,
			SetMetadata:     setMetadata,
			Filters:         nil,
			ProgressBarInfo: progBarInfo,
		},
		config, // Note: if isMobileApi is true, custom user-agent will be ignored
		reqHandler,
	)
	if cancelled {
		ugoiraArgs.cancel()
		return nil
	}
	if len(err) > 0 {
		return err
	}

	return convertMultipleUgoira(ugoiraArgs, ugoiraOptions, config)
}
