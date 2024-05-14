package gdrive

import (
	"context"
	"encoding/json"
	"fmt"

	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// https://developers.google.com/identity/protocols/oauth2/scopes#drive
var scopes = []string{
	drive.DriveMetadataReadonlyScope,
	drive.DriveReadonlyScope,
}

func ParseConfigFromClientJson(credsJson []byte) (*oauth2.Config, error) {
	config, err := google.ConfigFromJSON(credsJson, scopes...)
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: Unable to parse client secret file to config. More info => %w",
			cdlerrors.INPUT_ERROR,
			err,
		)
	}
	return config, nil
}

func ParseTokenJson(tokenJson []byte) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	err := json.Unmarshal(tokenJson, token)
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: Unable to parse token file. More info => %w",
			cdlerrors.INPUT_ERROR,
			err,
		)
	}
	return token, nil
}

func GetOAuthUrl(config *oauth2.Config) string {
	return config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
}

func ProcessAuthCode(ctx context.Context, authCode string, config *oauth2.Config) (*oauth2.Token, error) {
	if authCode == "" {
		return nil, fmt.Errorf(
			"gdrive error %d: No auth code provided",
			cdlerrors.INPUT_ERROR,
		)
	}

	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: Unable to retrieve token from web. More info => %w",
			cdlerrors.INPUT_ERROR,
			err,
		)
	}
	return token, nil
}
