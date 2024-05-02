package gdrive

import (
	"context"
	"fmt"
	"regexp"

	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	HTTP3_SUPPORTED        = true
	GDRIVE_ERROR_FILENAME  = "gdrive_download.log"
	BASE_API_KEY_REGEX_STR = `AIza[\w-]{35}`

	// file fields to fetch from GDrive API:
	// https://developers.google.com/drive/api/v3/reference/files
	GDRIVE_FILE_FIELDS   = "id,name,size,mimeType,md5Checksum"
	GDRIVE_FOLDER_FIELDS = "nextPageToken,files(id,name,size,mimeType,md5Checksum)"
)

var (
	API_KEY_REGEX = regexp.MustCompile(fmt.Sprintf(`^%s$`, BASE_API_KEY_REGEX_STR))
	API_KEY_PARAM_REGEX = regexp.MustCompile(fmt.Sprintf(`key=%s`, BASE_API_KEY_REGEX_STR))
)

type GDrive struct {
	ctx                context.Context
	apiKey             string         // Google Drive API key to use
	client             *drive.Service // Google Drive service client (if using service account credentials)
	apiUrl             string         // https://www.googleapis.com/drive/v3/files
	timeout            int            // timeout in seconds for GDrive API v3
	downloadTimeout    int            // timeout in seconds for GDrive file downloads
	maxDownloadWorkers int            // max concurrent workers for downloading files
}

// Returns a GDrive structure with the given API key and max download workers
func GetNewGDrive(ctx context.Context, apiKey, jsonPath, userAgent string, maxDownloadWorkers int) (*GDrive, error) {
	if jsonPath != "" && apiKey != "" {
		return nil, fmt.Errorf(
			"gdrive error %d: Both Google Drive API key and service account credentials file cannot be used at the same time",
			errs.DEV_ERROR,
		)
	} else if jsonPath == "" && apiKey == "" {
		return nil, fmt.Errorf(
			"gdrive error %d: Google Drive API key or service account credentials file is required",
			errs.DEV_ERROR,
		)
	}

	gdrive := &GDrive{
		ctx:                ctx,
		apiUrl:             "https://www.googleapis.com/drive/v3/files",
		timeout:            15,
		downloadTimeout:    900, // 15 minutes
		maxDownloadWorkers: maxDownloadWorkers,
	}
	if apiKey != "" {
		gdrive.apiKey = apiKey
		gdriveIsValid, err := gdrive.GDriveKeyIsValid(userAgent)
		if err != nil {
			return nil, err
		} else if !gdriveIsValid {
			return nil, fmt.Errorf(
				"gdrive error %d: Google Drive API key is invalid",
				errs.INPUT_ERROR,
			)
		}
		return gdrive, nil
	} 

	if !iofuncs.PathExists(jsonPath) {
		return nil, fmt.Errorf(
			"unable to access Drive API due to missing credentials file: %s",
			jsonPath,
		)
	}
	srv, err := drive.NewService(context.Background(), option.WithCredentialsFile(jsonPath))
	if err != nil {
		return nil, fmt.Errorf(
			"unable to access Drive API due to %w",
			err,
		)
	}
	gdrive.client = srv
	return gdrive, nil
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
			errs.CONNECTION_ERROR,
			err,
		)
	}
	res.Body.Close()
	return res.StatusCode != 400, nil
}
