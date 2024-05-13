package gdrive

import (
	"context"
	"fmt"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type GDrive struct {
	ctx                context.Context
	cancel             context.CancelFunc
	client             *drive.Service // Google Drive service client (if using service account credentials)
	apiUrl             string         // https://www.googleapis.com/drive/v3/files
	timeout            int            // timeout in seconds for GDrive API v3
	downloadTimeout    int            // timeout in seconds for GDrive file downloads
	maxDownloadWorkers int            // max concurrent workers for downloading files
}

type CredsInputs struct {
	ApiKey              string
	SrvAccJson          []byte
	ClientSecretJson    []byte 
	UserOauthTokenJson  []byte
}

// Returns a GDrive structure with the given API key and max download workers
func GetNewGDrive(ctx context.Context, creds *CredsInputs, maxDownloadWorkers int) (*GDrive, error) {
	if creds == nil {
		return nil, fmt.Errorf(
			"gdrive error %d: CredsInputs is nil in GetNewGDrive()",
			cdlerrors.DEV_ERROR,
		)
	}

	hasApiKey := creds.ApiKey != ""
	hasSrvAccJson := len(creds.SrvAccJson) > 0
	hasUserOauthTokenJson := len(creds.UserOauthTokenJson) > 0
	if !hasApiKey && !hasSrvAccJson && !hasUserOauthTokenJson {
		return nil, fmt.Errorf(
			"gdrive error %d: Google Drive API key, service account credentials, or user's OAuth token is required",
			cdlerrors.INPUT_ERROR,
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

	var err error
	var srv *drive.Service
	if hasApiKey {
		srv, err = parseApiKey(ctx, creds.ApiKey)
	} else if len(creds.SrvAccJson) > 0 {
		srv, err = parseSrvAccJson(ctx, creds.SrvAccJson)
	} else {
		srv, err = parseUserOauthJson(ctx, creds.ClientSecretJson, creds.UserOauthTokenJson)
	}
	gdrive.client = srv
	return gdrive, err
}

func (gdrive *GDrive) Release() {
	gdrive.cancel()
}

func parseApiKey(ctx context.Context, apiKey string) (*drive.Service, error) {
	if !constants.GDRIVE_API_KEY_REGEX.MatchString(apiKey) {
		return nil, fmt.Errorf(
			"gdrive error %d: Google Drive API key is invalid",
			cdlerrors.INPUT_ERROR,
		)
	}

	srv, err := drive.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: unable to create service with API key due to %w",
			cdlerrors.CONNECTION_ERROR,
			err,
		)
	}

	_, err = srv.Files.List().PageSize(1).Do()
	if err == nil {
		return nil, fmt.Errorf(
			"gdrive error %d: expecting error when using API key, but got none",
			cdlerrors.RESPONSE_ERROR,
		)
	}

	e, ok := err.(*googleapi.Error)
	if !ok {
		return nil, err
	}

	if e.Code == 400 && e.Message == "API key not valid. Please pass a valid API key." {
		return nil, fmt.Errorf(
			"gdrive error %d: Google Drive API key is invalid",
			cdlerrors.INPUT_ERROR,
		)
	} 

	// other errors should be due to insufficient permissions 
	//as we're using the API key instead of using OAuth
	return srv, nil
}

func parseSrvAccJson(ctx context.Context, srvAccJson []byte) (*drive.Service, error) {
	srv, err := drive.NewService(ctx, option.WithCredentialsJSON(srvAccJson))
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: unable to parse credentials JSON file due to %w",
			cdlerrors.INPUT_ERROR,
			err,
		)
	}

	_, err = srv.About.Get().Fields("user").Do()
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: service account credentials possibly invalid, more info => %w",
			cdlerrors.CONNECTION_ERROR,
			err,
		)
	}
	return srv, nil
}

func parseUserOauthJson(ctx context.Context, clientSecretJson []byte, tokenJson []byte) (*drive.Service, error) {
	config, err := ParseConfigFromClientJson(clientSecretJson)
	if err != nil {
		return nil, err
	}

	token, err := ParseTokenJson(tokenJson)
	if err != nil {
		return nil, err
	}

	client := config.Client(ctx, token)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: unable to parse credentials JSON file due to %w",
			cdlerrors.INPUT_ERROR,
			err,
		)
	}

	_, err = srv.Files.List().PageSize(1).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: Unable to verify user's OAuth JSON files. More info => %w",
			cdlerrors.RESPONSE_ERROR,
			err,
		)
	}
	return srv, nil
}
