package iofuncs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

var (
	APP_PATH      = getAppPath()
	DOWNLOAD_PATH = GetDefaultDownloadPath()
)

// Returns the path to the application's config directory
func getAppPath() string {
	appPath, err := os.UserConfigDir()
	if err != nil {
		panic(
			fmt.Errorf(
				"error %d, failed to get user's config directory: %v",
				errs.OS_ERROR,
				err,
			),
		)
	}
	return filepath.Join(appPath, "Cultured-Downloader")
}
