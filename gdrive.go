package cdlogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/gdrive"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Returns a GDrive structure with the given API key and max download workers
func GetNewGDrive(apiKey, userAgent string, credsJson []byte, maxDownloadWorkers int, ctx context.Context) (*gdrive.GDrive, error) {
	if credsJson != nil && apiKey != "" {
		return nil, errors.New("both google drive API key and service account credentials cannot be used at the same time")
	} else if credsJson == nil && apiKey == "" {
		return nil, errors.New("google drive API key or service account credentials is required")
	}

	gdrive := gdrive.NewGDrive(
		ctx, 
		15,  // 15 secs for api timeout
		900, // 15 mins for download timeout
		maxDownloadWorkers,
	)
	if apiKey != "" {
		gdrive.SetApiKey(apiKey)
		gdriveIsValid, err := gdrive.GDriveKeyIsValid(userAgent)
		if err != nil {
			return nil, err
		} else if !gdriveIsValid {
			return nil, errors.New("google drive API key is invalid")
		}
		return gdrive, nil
	} 

	srv, err := drive.NewService(ctx, option.WithCredentialsJSON(credsJson))
	if err != nil {
		return nil, fmt.Errorf("unable to access drive API due to %v", err)
	}
	gdrive.SetClient(srv)
	return gdrive, nil
}
