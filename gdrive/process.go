package gdrive

import (
	"path/filepath"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/api"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

// Process and detects for any external download links from the post's text content
func ProcessPostText(postBodyStr, postFolderPath string, downloadGdrive bool, logUrls bool) []*httpfuncs.ToDownload {
	if postBodyStr == "" {
		return nil
	}

	// split the text by newlines
	postBodySlice := strings.FieldsFunc(
		postBodyStr,
		func(c rune) bool {
			return c == '\n'
		},
	)
	loggedPassword := false
	var detectedGdriveLinks []*httpfuncs.ToDownload
	for _, text := range postBodySlice {
		if api.DetectPasswordInText(text) && !loggedPassword {
			// Log the entire post text if it contains a password
			filePath := filepath.Join(postFolderPath, constants.PASSWORD_FILENAME)
			if !iofuncs.PathExists(filePath) {
				loggedPassword = true
				logger.LogMessageToPath(
					"Found potential password in the post:\n\n"+postBodyStr,
					filePath,
					logger.ERROR,
				)
			}
		}

		if logUrls {
			api.DetectOtherExtDLLink(text, postFolderPath)
		}
		if api.DetectGDriveLinks(text, postFolderPath, false, logUrls) && downloadGdrive {
			detectedGdriveLinks = append(detectedGdriveLinks, &httpfuncs.ToDownload{
				Url:      text,
				FilePath: filepath.Join(postFolderPath, constants.GDRIVE_FOLDER),
			})
		}
	}
	return detectedGdriveLinks
}
