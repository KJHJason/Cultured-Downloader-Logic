package kemono

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/metadata"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

func getInlineImages(content, postFolderPath string) []*httpfuncs.ToDownload {
	matches := constants.KEMONO_IMG_SRC_TAG_REGEX.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	toDownload := make([]*httpfuncs.ToDownload, len(matches))
	for idx, match := range matches {
		imgSrc := match[constants.KEMONO_IMG_SRC_TAG_REGEX_IDX]
		if imgSrc == "" {
			continue
		}
		toDownload[idx] = &httpfuncs.ToDownload{
			Url:      constants.KEMONO_URL + imgSrc,
			FilePath: filepath.Join(postFolderPath, constants.IMAGES_FOLDER, httpfuncs.GetLastPartOfUrl(imgSrc)),
		}
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

// Convert "2024-05-24T15:00:00" string to time.Time
//
// Note: The value returned by Kemono is UTC+0
func parsePublishedDate(publishedDate string) time.Time {
	datePublished, err := time.Parse(constants.KEMONO_PUBLISHED_DATE_LAYOUT, publishedDate)
	if err != nil {
		errMsg := fmt.Errorf(
			"kemono error %d: failed to parse published date %q, more info => %w",
			cdlerrors.UNEXPECTED_ERROR,
			publishedDate,
			err,
		)
		logger.LogError(errMsg, logger.ERROR)
		return time.Time{}
	}
	return datePublished.In(constants.KEMONO_DATETIME_OFFSET) // convert to UTC+9/JST
}

func processJson(resJson *MainKemonoJson, dlOptions *KemonoDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload) {
	publishedDate := parsePublishedDate(resJson.Published)
	if !dlOptions.Base.Filters.IsPostDateValid(publishedDate) {
		return nil, nil
	}

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
		logger.LogError(err, logger.ERROR)
		creatorNamePath = resJson.User
	} else {
		creatorNamePath = fmt.Sprintf("%s [%s]", creatorName, resJson.User)
	}

	postFolderPath := iofuncs.GetPostFolder(
		filepath.Join(dlOptions.Base.DownloadDirPath, resJson.Service),
		creatorNamePath,
		resJson.Id,
		resJson.Title,
	)

	if dlOptions.Base.SetMetadata {
		postMetadata := metadata.KemonoPost{
			PostId: resJson.Id,
			Url: fmt.Sprintf(
				"%s/%s/user/%s/post/%s",
				constants.KEMONO_URL,
				resJson.Service,
				resJson.User,
				resJson.Id,
			),
			Title:        resJson.Title,
			Service:      resJson.Service,
			Content:      resJson.Content,
			PublishedUTC: resJson.Published,
			EmbedContent: metadata.KemonoPostEmbeddedContent{
				Description: resJson.Embed.Description,
				Subject:     resJson.Embed.Subject,
				Url:         resJson.Embed.Url,
			},
		}
		if err := metadata.WriteMetadata(postMetadata, postFolderPath); err != nil {
			return nil, nil
		}
	}

	var gdriveLinks []*httpfuncs.ToDownload
	var toDownload []*httpfuncs.ToDownload
	if dlOptions.Base.DlAttachments {
		toDownload = getInlineImages(resJson.Content, postFolderPath)
		for _, attachment := range resJson.Attachments {
			if !dlOptions.Base.Filters.IsFileNameValid(attachment.Name) || !dlOptions.Base.Filters.IsFilePathExtValid(attachment.Name) {
				continue
			}
			toDownload = append(toDownload, &httpfuncs.ToDownload{
				Url:      constants.KEMONO_URL + attachment.Path,
				FilePath: getKemonoFilePath(postFolderPath, constants.KEMONO_CONTENT_FOLDER, attachment.Name),
			})
		}

		if resJson.Embed.Url != "" {
			embedsDirPath := filepath.Join(postFolderPath, constants.KEMONO_EMBEDS_FOLDER)
			if dlOptions.Base.Configs.LogUrls {
				utils.DetectOtherExtDLLink(resJson.Embed.Url, embedsDirPath)
			}
			if dlOptions.Base.DlGdrive && utils.DetectGDriveLinks(resJson.Embed.Url, postFolderPath, true, dlOptions.Base.Configs.LogUrls) {
				gdriveLinks = append(gdriveLinks, &httpfuncs.ToDownload{
					Url:      resJson.Embed.Url,
					FilePath: embedsDirPath,
				})
			}
		}

		if resJson.File.Path != "" {
			if dlOptions.Base.Filters.IsFileNameValid(resJson.File.Name) && dlOptions.Base.Filters.IsFilePathExtValid(resJson.File.Name) {
				// usually is the thumbnail of the post
				toDownload = append(toDownload, &httpfuncs.ToDownload{
					Url:      constants.KEMONO_URL + resJson.File.Path,
					FilePath: getKemonoFilePath(postFolderPath, "", resJson.File.Name),
				})
			}
		}
	}

	contentGdriveLinks := gdrive.ProcessPostText(
		resJson.Content,
		postFolderPath,
		dlOptions.Base.DlGdrive,
		dlOptions.Base.Configs.LogUrls,
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
