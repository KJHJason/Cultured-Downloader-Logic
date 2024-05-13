package gdrive

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/progress"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func getConfigAndProgInfo() (*configs.Config, *progress.ProgressBarInfo) {
	config := &configs.Config{
		DownloadPath:   "./",
		FfmpegPath:     "",
		OverwriteFiles: false,
		LogUrls: 	    false,
		UserAgent:      httpfuncs.DEFAULT_USER_AGENT,
	}
	prog := &progress.ProgressBarInfo{
		MainProgressBar:      &progress.DummyProgBar{},
		DownloadProgressBars: nil,
	}
	return config, prog
}

func loadDotEnv(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		t.Fatal("Error loading .env file")
	}
}

func initTestGDrive(t *testing.T) (*GDrive, context.CancelFunc, *configs.Config, *progress.ProgressBarInfo) {
	loadDotEnv(t)

	apiKey := os.Getenv("GDRIVE_API_KEY")
	if apiKey == "" {
		t.Fatal("GDRIVE_API_KEY is empty")
	}

	ctx, cancel := context.WithCancel(context.Background())
	creds := &CredsInputs{
		ApiKey: apiKey,
	}
	gdriveClient, err := GetNewGDrive(ctx, creds, 2)
	if err != nil {
		t.Fatalf("Error creating GDrive client: %v", err)
	}

	config, prog := getConfigAndProgInfo()
	return gdriveClient, cancel, config, prog
}

func TestGDriveFileDownload(t *testing.T) {
	gdriveClient, cancel, config, progInfo := initTestGDrive(t)
	defer cancel()

	url := "https://drive.google.com/file/d/1xnDYjiH866KOlAGnZ3mDJuqpPq3mRF1F/view?usp=sharing"
	dirPath := "test-dir"
	defer os.RemoveAll(dirPath)

	toDlInfo := &httpfuncs.ToDownload{
		Url:      url,
		FilePath: dirPath,
	}

	errSlice := gdriveClient.DownloadGdriveUrls(
		[]*httpfuncs.ToDownload{toDlInfo},
		config,
		progInfo,
	)
	if len(errSlice) > 0 {
		t.Logf("Errors downloading file in %s", dirPath)
		for _, err := range errSlice {
			t.Error(err)
		}
		t.Fail()
	} else {
		t.Logf("Downloaded file in %s", dirPath)
	}
}

func TestGDriveFolderDownload(t *testing.T) {
	gdriveClient, cancel, config, progInfo := initTestGDrive(t)
	defer cancel()

	url := "https://drive.google.com/drive/folders/1zhP5ZzpxFX654-KSNP8d4nA2-zqLa-qq?usp=sharing"
	dirPath := "test-dir"
	defer os.RemoveAll(dirPath)

	toDlInfo := &httpfuncs.ToDownload{
		Url:      url,
		FilePath: dirPath,
	}

	errSlice := gdriveClient.DownloadGdriveUrls(
		[]*httpfuncs.ToDownload{toDlInfo},
		config,
		progInfo,
	)
	if len(errSlice) > 0 {
		t.Logf("Errors downloading folder at %s", dirPath)
		for _, err := range errSlice {
			t.Error(err)
		}
		t.Fail()
	} else {
		t.Logf("Downloaded folder at %s", dirPath)
	}
}

func TestGDriveServiceAcc(t *testing.T) {
	gdriveJsonPath := "../test-gdrive-service-acc.json"
	if _, err := os.Stat(gdriveJsonPath); os.IsNotExist(err) {
		t.Fatalf("gdrive-service-acc.json not found at %s", gdriveJsonPath)
	} 

	credJson, err := os.ReadFile(gdriveJsonPath)
	if err != nil {
		t.Fatalf("Error reading gdrive-service-acc.json: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	creds := &CredsInputs{
		SrvAccJson: credJson,
	}
	gdriveClient, err := GetNewGDrive(ctx, creds, 2)
	if err != nil {
		t.Fatalf("Error setting with service account: %v", err)
	}

	url := "https://drive.google.com/file/d/1ZjhOns-rZeSWS0EQPMziqsZINxWge468/view?usp=sharing"
	dirPath := "test-dir"
	defer os.RemoveAll(dirPath)

	toDlInfo := &httpfuncs.ToDownload{
		Url:      url,
		FilePath: dirPath,
	}

	config, progInfo := getConfigAndProgInfo()
	errSlice := gdriveClient.DownloadGdriveUrls(
		[]*httpfuncs.ToDownload{toDlInfo},
		config,
		progInfo,
	)
	if len(errSlice) > 0 {
		t.Logf("Errors downloading file in %s", dirPath)
		for _, err := range errSlice {
			t.Error(err)
		}
		t.Fail()
	} else {
		t.Logf("Downloaded file in %s", dirPath)
	}
}

func getGDriveUserClientSecret(t *testing.T) (*oauth2.Config, []byte) {
	gdriveJsonPath := "../test-gcp-client-secret.json"
	if _, err := os.Stat(gdriveJsonPath); os.IsNotExist(err) {
		t.Fatalf("test-gcp-client-secret.json not found at %s", gdriveJsonPath)
	}

	credJson, err := os.ReadFile(gdriveJsonPath)
	if err != nil {
		t.Fatalf("Error reading test-gcp-client-secret.json: %v", err)
	}

	oauthConfig, err := ParseConfigFromClientJson(credJson)
	if err != nil {
		t.Fatalf("Error parsing client secret file: %v", err)
	}
	return oauthConfig, credJson
}

func TestGDriveOauthProcess(t *testing.T) {
	oauthConfig, _ := getGDriveUserClientSecret(t)

	url := GetOAuthUrl(oauthConfig)
	t.Logf("Visit the URL for the auth dialog: %v", url)
}

func TestGDriveTokenExchange(t *testing.T) {
	loadDotEnv(t)
	oauthConfig, _ := getGDriveUserClientSecret(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gdriveAccessCode := os.Getenv("GDRIVE_ACCESS_CODE")
	if gdriveAccessCode == "" {
		t.Fatal("GDRIVE_ACCESS_CODE is empty")
	}

	token, err := ProcessAuthCode(ctx, gdriveAccessCode, oauthConfig)
	if err != nil {
		t.Fatalf("Error processing auth code: %v", err)
	}

	tokenJson, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("Error marshalling token: %v", err)
	}
	t.Logf("Token JSON: %s", tokenJson)
}

func TestGDriveOauthDownload(t *testing.T) {
	loadDotEnv(t)
	_, credJson := getGDriveUserClientSecret(t)

	tokenJsonPath := "../test-gcp-token.json"
	if _, err := os.Stat(tokenJsonPath); os.IsNotExist(err) {
		t.Fatalf("test-gcp-token.json not found at %s", tokenJsonPath)
	}

	tokenJson, err := os.ReadFile(tokenJsonPath)
	if err != nil {
		t.Fatalf("Error reading test-gcp-token.json: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	creds := &CredsInputs{
		ClientSecretJson: credJson,
		UserOauthTokenJson: tokenJson,
	}
	gdriveClient, err := GetNewGDrive(ctx, creds, 2)
	if err != nil {
		t.Fatalf("Error setting with oauth2 credentials: %v", err)
	}

	url := "https://drive.google.com/file/d/1ZjhOns-rZeSWS0EQPMziqsZINxWge468/view?usp=sharing"
	dirPath := "test-dir"
	defer os.RemoveAll(dirPath)

	toDlInfo := &httpfuncs.ToDownload{
		Url:      url,
		FilePath: dirPath,
	}

	config, progInfo := getConfigAndProgInfo()
	errSlice := gdriveClient.DownloadGdriveUrls(
		[]*httpfuncs.ToDownload{toDlInfo},
		config,
		progInfo,
	)
	if len(errSlice) > 0 {
		t.Logf("Errors downloading file in %s", dirPath)
		for _, err := range errSlice {
			t.Error(err)
		}
	} else {
		t.Logf("Downloaded file in %s", dirPath)
	}
}
