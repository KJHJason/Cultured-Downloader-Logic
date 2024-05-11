package gdrive

import (
	"context"
	"fmt"

	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type GDrive struct {
	ctx                context.Context
	cancel             context.CancelFunc
	apiKey             string         // Google Drive API key to use
	client             *drive.Service // Google Drive service client (if using service account credentials)
	apiUrl             string         // https://www.googleapis.com/drive/v3/files
	timeout            int            // timeout in seconds for GDrive API v3
	downloadTimeout    int            // timeout in seconds for GDrive file downloads
	maxDownloadWorkers int            // max concurrent workers for downloading files
}

// Returns a GDrive structure with the given API key and max download workers
func GetNewGDrive(ctx context.Context, apiKey, userAgent string, jsonBytes []byte, maxDownloadWorkers int) (*GDrive, error) {
	if len(jsonBytes) == 0 && apiKey != "" {
		return nil, fmt.Errorf(
			"gdrive error %d: Both Google Drive API key and service account credentials file cannot be used at the same time",
			cdlerrors.DEV_ERROR,
		)
	} else if len(jsonBytes) == 0 && apiKey == "" {
		return nil, fmt.Errorf(
			"gdrive error %d: Google Drive API key or service account credentials file is required",
			cdlerrors.DEV_ERROR,
		)
	}

	gdriveCtx, cancel := context.WithCancel(ctx)
	gdrive := &GDrive{
		ctx:                gdriveCtx,
		cancel:             cancel,
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
				cdlerrors.INPUT_ERROR,
			)
		}
		return gdrive, nil
	}

	srv, err := drive.NewService(ctx, option.WithCredentialsJSON(jsonBytes))
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
	match := constants.GDRIVE_API_KEY_REGEX.MatchString(gdrive.apiKey)
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
			Http2:     !constants.GDRIVE_HTTP3_SUPPORTED,
			Http3:     constants.GDRIVE_HTTP3_SUPPORTED,
		},
	)
	if err != nil {
		return false, fmt.Errorf(
			"gdrive error %d: failed to check if Google Drive API key is valid, more info => %w",
			cdlerrors.CONNECTION_ERROR,
			err,
		)
	}
	res.Body.Close()
	return res.StatusCode != 400, nil
}
