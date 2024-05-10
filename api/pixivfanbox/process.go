package pixivfanbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

// Pixiv Fanbox permitted file extensions based on
// https://fanbox.pixiv.help/hc/en-us/articles/360011057793-What-types-of-attachments-can-I-post-
var pixivFanboxAllowedImageExt = []string{"jpg", "jpeg", "png", "gif"}

func detectUrlsAndPasswordsInPost(text, postFolderPath string, articleBlocks FanboxArticleBlocks, dlOptions *PixivFanboxDlOptions) ([]*httpfuncs.ToDownload, bool) {
	loggedPassword := false
	if api.DetectPasswordInText(text) {
		// Log the entire post text if it contains a password
		filePath := filepath.Join(postFolderPath, constants.PASSWORD_FILENAME)
		if !iofuncs.PathExists(filePath) {
			loggedPassword = true
			postBodyStr := "Found potential password in the post:\n\n"
			for _, articleContent := range articleBlocks {
				articleText := articleContent.Text
				if articleText != "" {
					postBodyStr += articleText + "\n"
				}
			}
			logger.LogMessageToPath(
				postBodyStr,
				filePath,
				logger.ERROR,
			)
		}
	}

	var gdriveLinks []*httpfuncs.ToDownload
	if dlOptions.Configs.LogUrls {
		api.DetectOtherExtDLLink(text, postFolderPath)
	}
	if api.DetectGDriveLinks(text, postFolderPath, false, dlOptions.Configs.LogUrls) && dlOptions.DlGdrive {
		gdriveLinks = append(gdriveLinks, &httpfuncs.ToDownload{
			Url:      text,
			FilePath: filepath.Join(postFolderPath, constants.GDRIVE_FOLDER),
		})
	}
	return gdriveLinks, loggedPassword
}

func processFanboxArticlePost(postBody json.RawMessage, postFolderPath string, dlOptions *PixivFanboxDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	var articleJson FanboxArticleJson
	if err := httpfuncs.LoadJsonFromBytes(postBody, &articleJson); err != nil {
		return nil, nil, err
	}

	var urlsSlice []*httpfuncs.ToDownload
	var gdriveLinks []*httpfuncs.ToDownload
	// retrieve images and attachments url(s)
	imageMap := articleJson.ImageMap
	if imageMap != nil && dlOptions.DlImages {
		for _, imageInfo := range imageMap {
			urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
				Url:      imageInfo.OriginalUrl,
				FilePath: filepath.Join(postFolderPath, constants.IMAGES_FOLDER),
			})
		}
	}

	attachmentMap := articleJson.FileMap
	if attachmentMap != nil && dlOptions.DlAttachments {
		for _, attachmentInfo := range attachmentMap {
			attachmentUrl := attachmentInfo.Url
			filename := attachmentInfo.Name + "." + attachmentInfo.Extension
			urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
				Url:      attachmentUrl,
				FilePath: filepath.Join(postFolderPath, constants.ATTACHMENT_FOLDER, filename),
			})
		}
	}

	articleBlocks := articleJson.Blocks
	if len(articleBlocks) == 0 {
		return urlsSlice, gdriveLinks, nil
	}

	loggedPassword := false
	for _, articleBlock := range articleBlocks {
		text := articleBlock.Text
		if text != "" && !loggedPassword {
			var detectedGdriveUrls []*httpfuncs.ToDownload
			detectedGdriveUrls, loggedPassword = detectUrlsAndPasswordsInPost(
				text,
				postFolderPath,
				articleBlocks,
				dlOptions,
			)
			gdriveLinks = append(gdriveLinks, detectedGdriveUrls...)
		}

		articleLinks := articleBlock.Links
		if len(articleLinks) > 0 {
			for _, articleLink := range articleLinks {
				linkUrl := articleLink.Url
				api.DetectOtherExtDLLink(linkUrl, postFolderPath)
				if api.DetectGDriveLinks(linkUrl, postFolderPath, true, dlOptions.Configs.LogUrls) && dlOptions.DlGdrive {
					gdriveLinks = append(gdriveLinks, &httpfuncs.ToDownload{
						Url:      linkUrl,
						FilePath: filepath.Join(postFolderPath, constants.GDRIVE_FOLDER),
					})
					continue
				}
			}
		}
	}

	return urlsSlice, gdriveLinks, nil
}

func processFanboxFilePost(postBody json.RawMessage, postFolderPath string, dlOptions *PixivFanboxDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	var filePostJson FanboxFilePostJson
	if err := httpfuncs.LoadJsonFromBytes(postBody, &filePostJson); err != nil {
		return nil, nil, err
	}

	// process the text in the post
	var urlsSlice, gdriveLinks []*httpfuncs.ToDownload
	detectedGdriveLinks := gdrive.ProcessPostText(
		filePostJson.Text,
		postFolderPath,
		dlOptions.DlGdrive,
		dlOptions.Configs.LogUrls,
	)
	if len(detectedGdriveLinks) > 0 {
		gdriveLinks = append(gdriveLinks, detectedGdriveLinks...)
	}

	imageAndAttachmentUrls := filePostJson.Files
	if !dlOptions.DlImages && !dlOptions.DlAttachments {
		return nil, nil, nil
	}

	for _, fileInfo := range imageAndAttachmentUrls {
		fileUrl := fileInfo.Url
		extension := fileInfo.Extension
		filename := fileInfo.Name + "." + extension

		var filePath string
		isImage := api.SliceContains(pixivFanboxAllowedImageExt, extension)
		if isImage {
			filePath = filepath.Join(postFolderPath, constants.IMAGES_FOLDER, filename)
		} else {
			filePath = filepath.Join(postFolderPath, constants.ATTACHMENT_FOLDER, filename)
		}

		if (isImage && dlOptions.DlImages) || (!isImage && dlOptions.DlAttachments) {
			urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
				Url:      fileUrl,
				FilePath: filePath,
			})
		}
	}
	return urlsSlice, gdriveLinks, nil
}

func processFanboxImagePost(postBody json.RawMessage, postFolderPath string, dlOptions *PixivFanboxDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	var imagePostJson FanboxImagePostJson
	if err := httpfuncs.LoadJsonFromBytes(postBody, &imagePostJson); err != nil {
		return nil, nil, err
	}

	// process the text in the post
	var urlsSlice, gdriveLinks []*httpfuncs.ToDownload
	detectedGdriveLinks := gdrive.ProcessPostText(
		imagePostJson.Text,
		postFolderPath,
		dlOptions.DlGdrive,
		dlOptions.Configs.LogUrls,
	)
	if len(detectedGdriveLinks) > 0 {
		gdriveLinks = append(gdriveLinks, detectedGdriveLinks...)
	}

	// retrieve images and attachments url(s)
	imageAndAttachmentUrls := imagePostJson.Images
	if !dlOptions.DlImages && !dlOptions.DlAttachments {
		return nil, nil, nil
	}

	for _, fileInfo := range imageAndAttachmentUrls {
		fileUrl := fileInfo.OriginalUrl
		extension := fileInfo.Extension
		filename := httpfuncs.GetLastPartOfUrl(fileUrl)

		var filePath string
		isImage := api.SliceContains(pixivFanboxAllowedImageExt, extension)
		if isImage {
			filePath = filepath.Join(postFolderPath, constants.IMAGES_FOLDER, filename)
		} else {
			filePath = filepath.Join(postFolderPath, constants.ATTACHMENT_FOLDER, filename)
		}

		if (isImage && dlOptions.DlImages) || (!isImage && dlOptions.DlAttachments) {
			urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
				Url:      fileUrl,
				FilePath: filePath,
			})
		}
	}
	return urlsSlice, gdriveLinks, nil
}

// Process the JSON response from Pixiv Fanbox's API and
// returns a map of urls and a map of GDrive urls to download from
func processFanboxPostJson(res *http.Response, dlOptions *PixivFanboxDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	var post FanboxPostJson
	if err := httpfuncs.LoadJsonFromResponse(res, &post); err != nil {
		return nil, nil, err
	}

	postJson := post.Body
	postId := postJson.Id
	postTitle := postJson.Title
	creatorId := postJson.CreatorId
	postFolderPath := iofuncs.GetPostFolder(
		dlOptions.BaseDownloadDirPath,
		creatorId,
		postId,
		postTitle,
	)

	var urlsSlice []*httpfuncs.ToDownload
	thumbnail := postJson.CoverImageUrl
	if dlOptions.DlThumbnails && thumbnail != "" {
		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      thumbnail,
			FilePath: postFolderPath,
		})
	}

	// Note that Pixiv Fanbox posts have 3 types of formatting (as of now):
	//	1. With proper formatting and mapping of post content elements ("article")
	//	2. With a simple formatting that obly contains info about the text and files ("file", "image")
	postType := postJson.Type
	postBody := postJson.Body
	if postBody == nil {
		return urlsSlice, nil, nil
	}

	var err error
	var newUrlsSlice []*httpfuncs.ToDownload
	var gdriveLinks []*httpfuncs.ToDownload
	switch postType {
	case "file":
		newUrlsSlice, gdriveLinks, err = processFanboxFilePost(postBody, postFolderPath, dlOptions)
	case "image":
		newUrlsSlice, gdriveLinks, err = processFanboxImagePost(postBody, postFolderPath, dlOptions)
	case "article":
		newUrlsSlice, gdriveLinks, err = processFanboxArticlePost(postBody, postFolderPath, dlOptions)
	case "text": // text post
		// Usually has no content but try to detect for any external download links
		var textContent FanboxTextPostJson
		if err = httpfuncs.LoadJsonFromBytes(postBody, &textContent); err == nil {
			gdriveLinks = gdrive.ProcessPostText(
				textContent.Text,
				postFolderPath,
				dlOptions.DlGdrive,
				dlOptions.Configs.LogUrls,
			)
		}
	default: // unknown post type
		jsonBytes, _ := json.MarshalIndent(post, "", "\t")
		return nil, nil, fmt.Errorf(
			"pixiv fanbox error %d: unknown post type, %q\nPixiv Fanbox post content:\n%s",
			cdlerrors.JSON_ERROR,
			postType,
			string(jsonBytes),
		)
	}

	if err != nil {
		return nil, nil, err
	}
	urlsSlice = append(urlsSlice, newUrlsSlice...)
	return urlsSlice, gdriveLinks, nil
}

func processMultiplePostJson(resChan chan *http.Response, dlOptions *PixivFanboxDlOptions) (urlsSlice []*httpfuncs.ToDownload, gdriveUrls []*httpfuncs.ToDownload) {
	var errSlice []error
	resChanLen := len(resChan)
	baseMsg := "Processing received JSON(s) from Pixiv Fanbox [%d/" + fmt.Sprintf("%d]...", resChanLen)
	progress := dlOptions.MainProgBar
	progress.UpdateBaseMsg(baseMsg)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished processing %d JSON(s) from Pixiv Fanbox!",
			resChanLen,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while processing %d JSON(s) from Pixiv Fanbox.\nPlease refer to the logs for more details.",
			resChanLen,
		),
	)
	progress.SetToProgressBar()
	progress.UpdateMax(resChanLen)
	progress.Start()
	defer progress.SnapshotTask()
	for res := range resChan {
		postUrls, postGdriveLinks, err := processFanboxPostJson(res, dlOptions)
		if err != nil {
			errSlice = append(errSlice, err)
		} else {
			urlsSlice = append(urlsSlice, postUrls...)
			gdriveUrls = append(gdriveUrls, postGdriveLinks...)
		}
		progress.Increment()
	}

	hasErr := false
	if len(errSlice) > 0 {
		hasErr = true
		logger.LogErrors(false, logger.ERROR, errSlice...)
	}
	progress.Stop(hasErr)
	return urlsSlice, gdriveUrls
}
