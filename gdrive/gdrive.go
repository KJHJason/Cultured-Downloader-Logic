package gdrive

import (
	"fmt"
	"regexp"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"google.golang.org/api/drive/v3"
)

const (
	HTTP3_SUPPORTED        = true
	GDRIVE_ERROR_FILENAME  = "gdrive_download.log"
	BASE_API_KEY_REGEX_STR = `AIza[\w-]{35}`

	// file fields to fetch from GDrive API:
	// https://developers.google.com/drive/api/v3/reference/files
	GDRIVE_FILE_FIELDS = "id,name,size,mimeType,md5Checksum"
	GDRIVE_FOLDER_FIELDS = "nextPageToken,files(id,name,size,mimeType,md5Checksum)"
)

var (
	API_KEY_REGEX       = regexp.MustCompile(fmt.Sprintf(`^%s$`, BASE_API_KEY_REGEX_STR))
	API_KEY_PARAM_REGEX = regexp.MustCompile(fmt.Sprintf(`key=%s`, BASE_API_KEY_REGEX_STR))
)

type GDrive struct {
	apiKey             string         // Google Drive API key to use
	client             *drive.Service // Google Drive service client (if using service account credentials)
	apiUrl             string         // https://www.googleapis.com/drive/v3/files
	timeout            int            // timeout in seconds for GDrive API v3
	downloadTimeout    int            // timeout in seconds for GDrive file downloads
	maxDownloadWorkers int            // max concurrent workers for downloading files
}

func (gdrive *GDrive) SetApiKey(apiKey string) {
	gdrive.apiKey = apiKey
}

func (gdrive *GDrive) SetClient(client *drive.Service) {
	gdrive.client = client
}

func (gdrive *GDrive) SetApiUrl(apiUrl string) {
	gdrive.apiUrl = apiUrl
}

func (gdrive *GDrive) SetTimeout(timeout int) {
	gdrive.timeout = timeout
}

func (gdrive *GDrive) SetDownloadTimeout(downloadTimeout int) {
	gdrive.downloadTimeout = downloadTimeout
}

func (gdrive *GDrive) SetMaxDownloadWorkers(maxDownloadWorkers int) {
	gdrive.maxDownloadWorkers = maxDownloadWorkers
}

// Checks if the given Google Drive API key is valid
//
// Will return true if the given Google Drive API key is valid
func (gdrive *GDrive) GDriveKeyIsValid(userAgent string) (bool, error) {
	match := API_KEY_REGEX.MatchString(gdrive.apiKey)
	if !match {
		return false, nil
	}

	params := map[string]string{"key": gdrive.apiKey}
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Url:       gdrive.apiUrl,
			Method:    "GET",
			Timeout:   gdrive.timeout,
			Params:    params,
			UserAgent: userAgent,
			Http2:     !HTTP3_SUPPORTED,
			Http3:     HTTP3_SUPPORTED,
		},
	)
	if err != nil {
		return false, fmt.Errorf(
			"gdrive error %d: failed to check if Google Drive API key is valid, more info => %v",
			constants.CONNECTION_ERROR,
			err,
		)
	}
	res.Body.Close()
	return res.StatusCode != 400, nil
}
