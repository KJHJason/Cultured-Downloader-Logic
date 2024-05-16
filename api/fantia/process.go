package fantia

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

type postContentId struct {
	postId    int
	commentId int
}

func dlImagesFromPost(content *FantiaContent, postFolderPath string, organise bool, id *postContentId) []*httpfuncs.ToDownload {
	postContentPhotos := content.PostContentPhotos
	matchedUrlInComments := constants.FANTIA_COMMENT_IMAGE_URL_REGEX.FindAllStringSubmatch(content.Comment, -1)
	commentsLen := len(matchedUrlInComments)
	postContentLen := len(postContentPhotos)
	if commentsLen == 0 && postContentLen == 0 {
		return make([]*httpfuncs.ToDownload, 0)
	}
	urlsSlice := make([]*httpfuncs.ToDownload, 0, commentsLen+postContentLen)

	// for images that are embedded in the post content
	commentCount := 0
	commentFolderId := strconv.Itoa(id.commentId)
	for _, matched := range matchedUrlInComments {
		imageUrl := constants.FANTIA_URL + matched[constants.FANTIA_COMMENT_REGEX_URL_IDX]
		filePath := filepath.Join(postFolderPath, constants.FANTIA_POST_BLOG_DIR_NAME)

		if organise {
			commentCount++
			fileExt := matched[constants.FANTIA_COMMENT_REGEX_EXT_IDX]
			filePath = filepath.Join(filePath, commentFolderId, fmt.Sprintf("%d.%s", commentCount, fileExt))
		}

		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      imageUrl,
			FilePath: filePath,
		})
	}
	if organise && commentsLen > 0 {
		id.commentId++
	}

	// download images that are uploaded to their own section
	postCount := 0
	postFolderId := strconv.Itoa(id.postId)
	for _, image := range postContentPhotos {
		imageUrl := image.URL.Original
		filePath := filepath.Join(postFolderPath, constants.IMAGES_FOLDER)

		if organise {
			postCount++
			matched := constants.FANTIA_IMAGE_URL_REGEX.FindStringSubmatch(imageUrl)
			if len(matched) > 0 {
				fileExt := matched[constants.FANTIA_IMAGE_URL_REGEX_EXT_IDX]
				filePath = filepath.Join(filePath, postFolderId, fmt.Sprintf("%d.%s", postCount, fileExt))
			} else {
				err := fmt.Errorf(
					"fantia error %d: failed to match image url %q when trying to organise images",
					cdlerrors.UNEXPECTED_ERROR,
					imageUrl,
				)
				logger.LogError(err, logger.ERROR)
			}
		}

		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      imageUrl,
			FilePath: filePath,
		})
	}
	if organise && postContentLen > 0 {
		id.postId++
	}
	return urlsSlice
}

func dlAttachmentsFromPost(content *FantiaContent, postFolderPath string) []*httpfuncs.ToDownload {
	var urlsSlice []*httpfuncs.ToDownload

	// get the attachment url string if it exists
	attachmentUrl := content.AttachmentURI
	if attachmentUrl != "" {
		attachmentUrlStr := constants.FANTIA_URL + attachmentUrl
		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      attachmentUrlStr,
			FilePath: filepath.Join(postFolderPath, constants.ATTACHMENT_FOLDER),
		})
	} else if content.DownloadUri != "" {
		// if the attachment url string does not exist,
		// then get the download url for the file
		downloadUrl := constants.FANTIA_URL + content.DownloadUri
		filename := content.Filename
		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      downloadUrl,
			FilePath: filepath.Join(postFolderPath, constants.ATTACHMENT_FOLDER, filename),
		})
	}
	return urlsSlice
}

// Process the JSON response from Fantia's API and
// returns a slice of urls and a slice of gdrive urls to download from
func processFantiaPost(res *http.Response, dlOptions *FantiaDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	// processes a fantia post
	// returns a map containing the post id and the url to download the file from
	var postJson FantiaPost
	if err := httpfuncs.LoadJsonFromResponse(res, &postJson); err != nil {
		return nil, nil, err
	}

	if postJson.Redirect != "" {
		if postJson.Redirect != "/recaptcha" {
			return nil, nil, fmt.Errorf(
				"fantia error %d: unknown redirect url, %q",
				cdlerrors.UNEXPECTED_ERROR,
				postJson.Redirect,
			)
		}
		return nil, nil, cdlerrors.ErrRecaptcha
	}

	post := postJson.Post
	postId := strconv.Itoa(post.ID)
	postTitle := post.Title
	creatorName := post.Fanclub.User.Name
	postFolderPath := iofuncs.GetPostFolder(dlOptions.BaseDownloadDirPath, creatorName, postId, postTitle)

	var urlsSlice []*httpfuncs.ToDownload
	thumbnail := post.Thumb.Original
	if dlOptions.DlThumbnails && thumbnail != "" {
		urlsSlice = append(urlsSlice, &httpfuncs.ToDownload{
			Url:      thumbnail,
			FilePath: postFolderPath,
		})
	}

	gdriveLinks := gdrive.ProcessPostText(
		post.Comment,
		postFolderPath,
		dlOptions.DlGdrive,
		dlOptions.Configs.LogUrls,
	)

	postContent := post.PostContents
	if postContent == nil {
		return urlsSlice, gdriveLinks, nil
	}

	contentIds := &postContentId{
		commentId: 1,
		postId:    1,
	}
	for _, content := range postContent {
		commentGdriveLinks := gdrive.ProcessPostText(
			content.Comment,
			postFolderPath,
			dlOptions.DlGdrive,
			dlOptions.Configs.LogUrls,
		)
		if len(commentGdriveLinks) > 0 {
			gdriveLinks = append(gdriveLinks, commentGdriveLinks...)
		}
		if dlOptions.DlImages {
			urlsSlice = append(urlsSlice, dlImagesFromPost(&content, postFolderPath, dlOptions.OrganiseImages, contentIds)...)
		}
		if dlOptions.DlAttachments {
			urlsSlice = append(urlsSlice, dlAttachmentsFromPost(&content, postFolderPath)...)
		}
	}
	return urlsSlice, gdriveLinks, nil
}

type processIllustArgs struct {
	res        *http.Response
	postId     string
	postIdsLen int
	msgSuffix  string
}

// Process the JSON response to get the urls to download
func processIllustDetailApiRes(illustArgs *processIllustArgs, dlOptions *FantiaDlOptions) ([]*httpfuncs.ToDownload, []*httpfuncs.ToDownload, error) {
	progress := dlOptions.MainProgBar
	progress.SetToSpinner()
	progress.UpdateBaseMsg(
		fmt.Sprintf(
			"Processing retrieved JSON for post %s from Fantia %s...",
			illustArgs.postId,
			illustArgs.msgSuffix,
		),
	)
	progress.UpdateSuccessMsg(
		fmt.Sprintf(
			"Finished processing retrieved JSON for post %s from Fantia %s!",
			illustArgs.postId,
			illustArgs.msgSuffix,
		),
	)
	progress.UpdateErrorMsg(
		fmt.Sprintf(
			"Something went wrong while processing retrieved JSON for post %s from Fantia %s.\nPlease refer to the logs for more details.",
			illustArgs.postId,
			illustArgs.msgSuffix,
		),
	)
	progress.Start()
	defer progress.SnapshotTask()

	urlsToDownload, gdriveLinks, err := processFantiaPost(
		illustArgs.res,
		dlOptions,
	)
	if err != nil {
		if errors.Is(err, cdlerrors.ErrRecaptcha) {
			progress.UpdateErrorMsg(constants.ERR_RECAPTCHA_STR)
		}
		progress.Stop(true)
		return nil, nil, err
	}
	progress.Stop(false)
	return urlsToDownload, gdriveLinks, nil
}
