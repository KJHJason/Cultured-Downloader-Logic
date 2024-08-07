package ugoira

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/database"
	"github.com/KJHJason/Cultured-Downloader-Logic/extractor"
	"github.com/KJHJason/Cultured-Downloader-Logic/filters"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils/threadsafe"
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
	ctx, cancel := context.WithCancel(ugoiraArgs.context)
	defer cancel()

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
	errTsSlice := threadsafe.NewSlice[error]()

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
				errTsSlice.Append(err)
			}
			prog.Increment()
		}()
	}
	wg.Wait()
	close(queue)

	var errSlice []error
	hasErr := errTsSlice.LenUnsafe() > 0
	if hasErr {
		var hasCancelled bool
		if hasCancelled, errSlice = logger.LogSliceErrors(logger.ERROR, errTsSlice); hasCancelled {
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
	context        context.Context
	cancel         context.CancelFunc
	UseMobileApi   bool
	Filters        *filters.Filters
	CaptchaHandler httpfuncs.CaptchaHandler
	ToDownload     []*Ugoira
	Cookies        []*http.Cookie
	MainProgBar    progress.ProgressBar
}

func (u *UgoiraArgs) SetContext(ctx context.Context) {
	u.context, u.cancel = context.WithCancel(ctx)
}

// Downloads multiple Ugoira artworks and converts them based on the output format
func DownloadMultipleUgoira(ugoiraArgs *UgoiraArgs, ugoiraOptions *UgoiraOptions, config *configs.Config, reqHandler httpfuncs.RequestHandler, progBarInfo *progress.ProgressBarInfo) []error {
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

	// since we're mainly dealing with .zip, we have to add it into the filters if it's not there
	var filters *filters.Filters
	const ugoiraFileExt = ".zip"
	if ugoiraArgs.Filters.IsFileExtValid(ugoiraFileExt) {
		filters = ugoiraArgs.Filters
	} else {
		filters = ugoiraArgs.Filters.Copy()
		filters.FileExt = append(filters.FileExt, ugoiraFileExt)
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
			Filters:         filters,
			ProgressBarInfo: progBarInfo,
			CaptchaHandler:  ugoiraArgs.CaptchaHandler,
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
