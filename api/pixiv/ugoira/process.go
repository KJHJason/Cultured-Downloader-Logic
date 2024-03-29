package ugoira

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	pixivcommon "github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/common"
	"github.com/KJHJason/Cultured-Downloader-Logic/api/pixiv/models"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/extractor"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
)

// Map the Ugoira frame delays to their respective filenames
func MapDelaysToFilename(ugoiraFramesJson models.UgoiraFramesJson) map[string]int64 {
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
func ConvertUgoira(ugoiraInfo *models.Ugoira, imagesFolderPath string, ugoiraFfmpeg *UgoiraFfmpegArgs) error {
	outputExt := filepath.Ext(ugoiraFfmpeg.outputPath)
	if !api.SliceContains(UGOIRA_ACCEPTED_EXT, outputExt) {
		return fmt.Errorf(
			"pixiv error %d: Output extension %v is not allowed for ugoira conversion",
			constants.INPUT_ERROR,
			outputExt,
		)
	}

	concatDelayFilePath, sortedFilenames, err := writeDelays(ugoiraInfo, imagesFolderPath)
	if err != nil {
		return err
	}

	args, err := getFfmpegFlagsForUgoira(
		&ffmpegOptions{
			ffmpegPath:          ugoiraFfmpeg.ffmpegPath,
			outputExt:           outputExt,
			concatDelayFilePath: concatDelayFilePath,
			sortedFilenames:     sortedFilenames,
			outputPath:          ugoiraFfmpeg.outputPath,
			ugoiraQuality:       ugoiraFfmpeg.ugoiraQuality,
		},
		imagesFolderPath,
	)
	if err != nil {
		return err
	}

	// convert the frames to a gif or a video
	cmd := exec.CommandContext(ugoiraFfmpeg.context, ugoiraFfmpeg.ffmpegPath, args...)
	// cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		os.Remove(ugoiraFfmpeg.outputPath)
		if err == context.Canceled {
			return err
		}
		return fmt.Errorf(
			"pixiv error %d: failed to convert ugoira to %s, more info => %v",
			constants.CMD_ERROR,
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

func convertMultipleUgoira(ugoiraArgs *UgoiraArgs, ugoiraOptions *UgoiraOptions, config *configs.Config) {
	// Create a context that can be cancelled when SIGINT/SIGTERM signal is received
	ctx, cancel := context.WithCancel(ugoiraArgs.Context)
	defer cancel()

	// Catch SIGINT/SIGTERM signal and cancel the context when received
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()
	defer signal.Stop(sigs)

	var errSlice []error
	downloadInfoLen := len(ugoiraArgs.ToDownload)
	baseMsg := "Converting Ugoira to %s [%d/" + fmt.Sprintf("%d]...", downloadInfoLen)
	progress := ugoiraArgs.UgoiraProgBar
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished converting %d Ugoira to %s!",
			downloadInfoLen,
			ugoiraOptions.OutputFormat,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while converting %d Ugoira to %s!\nPlease refer to the logs for more details.",
			downloadInfoLen,
			ugoiraOptions.OutputFormat,
		),
	)
	progress.UpdateMax(downloadInfoLen)
	progress.Start()
	for i, ugoira := range ugoiraArgs.ToDownload {
		zipFilePath, outputPath := GetUgoiraFilePaths(ugoira.FilePath, ugoira.Url, ugoiraOptions.OutputFormat)
		if iofuncs.PathExists(outputPath) {
			progress.Increment()
			continue
		}
		if !iofuncs.PathExists(zipFilePath) {
			progress.Increment()
			continue
		}

		unzipFolderPath := filepath.Join(
			filepath.Dir(zipFilePath),
			"unzipped",
		)
		err := extractor.ExtractFiles(ctx, zipFilePath, unzipFolderPath, true)
		if err != nil {
			if err == context.Canceled {
				progress.StopInterrupt(
					fmt.Sprintf(
						"Stopped converting ugoira to %s [%d/%d]!",
						ugoiraOptions.OutputFormat,
						i,
						len(ugoiraArgs.ToDownload),
					),
				)
			}
			err := fmt.Errorf(
				"pixiv error %d: failed to unzip file %s, more info => %v",
				constants.OS_ERROR,
				zipFilePath,
				err,
			)
			errSlice = append(errSlice, err)
			progress.Increment()
			continue
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
		if err != nil {
			errSlice = append(errSlice, err)
		} else if ugoiraOptions.DeleteZip {
			os.Remove(zipFilePath)
		}
		progress.Increment()
	}

	hasErr := false
	if len(errSlice) > 0 {
		hasErr = true
		if hasCancelled := logger.LogErrors(false, logger.ERROR, errSlice...); hasCancelled {
			progress.StopInterrupt(
				fmt.Sprintf("Stopped converting ugoira to %s!", ugoiraOptions.OutputFormat),
			)
			return
		}
	}
	progress.Stop(hasErr)
}

type UgoiraArgs struct {
	Context       context.Context
	UseMobileApi  bool
	ToDownload    []*models.Ugoira
	Cookies       []*http.Cookie
	UgoiraProgBar progress.Progress
}

// Downloads multiple Ugoira artworks and converts them based on the output format
func DownloadMultipleUgoira(ugoiraArgs *UgoiraArgs, ugoiraOptions *UgoiraOptions, config *configs.Config, reqHandler httpfuncs.RequestHandler) {
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
			"Referer": "https://app-api.pixiv.net",
		}
	} else {
		headers = pixivcommon.GetPixivRequestHeaders()
		useHttp3 = httpfuncs.IsHttp3Supported(constants.PIXIV, true)
	}

	err := httpfuncs.DownloadUrlsWithHandler(
		urlsToDownload,
		&httpfuncs.DlOptions{
			Context:        ugoiraArgs.Context,
			MaxConcurrency: constants.PIXIV_MAX_CONCURRENT_DOWNLOADS,
			Headers:        headers,
			Cookies:        ugoiraArgs.Cookies,
			UseHttp3:       useHttp3,
		},
		config, // Note: if isMobileApi is true, custom user-agent will be ignored
		reqHandler,
	)
	if err != context.Canceled {
		convertMultipleUgoira(ugoiraArgs, ugoiraOptions, config)
	}
}
