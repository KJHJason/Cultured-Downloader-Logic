package kemono

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

func getInlineImages(content, postFolderPath string) []*httpfuncs.ToDownload {
	var toDownload []*httpfuncs.ToDownload
	for _, match := range constants.KEMONO_IMG_SRC_TAG_REGEX.FindAllStringSubmatch(content, -1) {
		imgSrc := match[constants.KEMONO_IMG_SRC_TAG_REGEX_IDX]
		if imgSrc == "" {
			continue
		}
		toDownload = append(toDownload, &httpfuncs.ToDownload{
			Url:      constants.KEMONO_URL + imgSrc,
			FilePath: filepath.Join(postFolderPath, constants.IMAGES_FOLDER, httpfuncs.GetLastPartOfUrl(imgSrc)),
		})
	}
	return toDownload
}

// Since the name of each attachment or file is not always the filename of the file as it could be a URL,
// we need to check if the returned name value is a URL and if it is, we just return the postFolderPath as the file path.
func getKemonoFilePath(postFolderPath, childDir, fileName string) string {
	if strings.HasPrefix(fileName, "http://") || strings.HasPrefix(fileName, "https://") {
		return filepath.Join(postFolderPath, childDir)
	}
	return filepath.Join(postFolderPath, childDir, fileName)
}

func processJson(resJson *MainKemonoJson, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload) {
	var creatorNamePath string
	if creatorName, err := getCreatorName(resJson.Service, resJson.User, dlOptions); err != nil {
		if errors.Is(err, context.Canceled) {
			dlOptions.CancelCtx()
			return nil, nil
		}
		err = fmt.Errorf(
			"error getting creator name for %q (%s)... falling back to creator ID! (Details below)\n%w",
			resJson.User,
			resJson.Service,
			err,
		)
		logger.LogError(err, false, logger.ERROR)
		creatorNamePath = resJson.User
	} else {
		creatorNamePath = fmt.Sprintf("%s [%s]", creatorName, resJson.User)
	}

	postFolderPath := iofuncs.GetPostFolder(
		filepath.Join(dlOptions.BaseDownloadDirPath, resJson.Service),
		creatorNamePath,
		resJson.Id,
		resJson.Title,
	)

	var gdriveLinks []*httpfuncs.ToDownload
	var toDownload []*httpfuncs.ToDownload
	if dlOptions.DlAttachments {
		toDownload = getInlineImages(resJson.Content, postFolderPath)
		for _, attachment := range resJson.Attachments {
			toDownload = append(toDownload, &httpfuncs.ToDownload{
				Url:      constants.KEMONO_URL + attachment.Path,
				FilePath: getKemonoFilePath(postFolderPath, constants.KEMONO_CONTENT_FOLDER, attachment.Name),
			})
		}

		if resJson.Embed.Url != "" {
			embedsDirPath := filepath.Join(postFolderPath, constants.KEMONO_EMBEDS_FOLDER)
			if dlOptions.Configs.LogUrls {
				api.DetectOtherExtDLLink(resJson.Embed.Url, embedsDirPath)
			}
			if api.DetectGDriveLinks(resJson.Embed.Url, postFolderPath, true, dlOptions.Configs.LogUrls) && dlOptions.DlGdrive {
				gdriveLinks = append(gdriveLinks, &httpfuncs.ToDownload{
					Url:      resJson.Embed.Url,
					FilePath: embedsDirPath,
				})
			}
		}

		if resJson.File.Path != "" {
			// usually is the thumbnail of the post
			toDownload = append(toDownload, &httpfuncs.ToDownload{
				Url:      constants.KEMONO_URL + resJson.File.Path,
				FilePath: getKemonoFilePath(postFolderPath, "", resJson.File.Name),
			})
		}
	}

	contentGdriveLinks := gdrive.ProcessPostText(
		resJson.Content,
		postFolderPath,
		dlOptions.DlGdrive,
		dlOptions.Configs.LogUrls,
	)
	gdriveLinks = append(gdriveLinks, contentGdriveLinks...)
	return toDownload, gdriveLinks
}

func processMultipleJson(resJson KemonoJson, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload) {
	var urlsToDownload, gdriveLinks []*httpfuncs.ToDownload
	for _, post := range resJson {
		toDownload, foundGdriveLinks := processJson(post, dlOptions)
		urlsToDownload = append(urlsToDownload, toDownload...)
		gdriveLinks = append(gdriveLinks, foundGdriveLinks...)
	}
	return urlsToDownload, gdriveLinks
}
