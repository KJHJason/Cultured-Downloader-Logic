package iofuncs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

type ConfigFile struct {
	DownloadDir string `json:"download_directory"`
	Language    string `json:"language"`
}

// Returns the download path from the config file
func GetDefaultDownloadPath() string {
	configFilePath := filepath.Join(APP_PATH, "config.json")
	if !PathExists(configFilePath) {
		return ""
	}

	configFile, err := os.ReadFile(configFilePath)
	if err != nil {
		os.Remove(configFilePath)
		return ""
	}

	var config ConfigFile
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		os.Remove(configFilePath)
		return ""
	}

	if !PathExists(config.DownloadDir) {
		return ""
	}
	return config.DownloadDir
}

// saves the new download path to the config file if it does not exist
func saveConfig(newDownloadPath, configFilePath string) error {
	config := ConfigFile{
		DownloadDir: newDownloadPath,
		Language:    "en",
	}
	configFile, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf(
			"error %d: failed to marshal config file, more info => %w",
			errs.JSON_ERROR,
			err,
		)
	}

	err = os.WriteFile(configFilePath, configFile, 0666)
	if err != nil {
		return fmt.Errorf(
			"error %d: failed to write config file, more info => %w",
			errs.OS_ERROR,
			err,
		)
	}
	return nil
}

// saves the new download path to the config file and overwrites the old one
func overwriteConfig(newDownloadPath, configFilePath string) error {
	// read the file
	configFile, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf(
			"error %d: failed to read config file, more info => %w",
			errs.OS_ERROR,
			err,
		)
	}

	var config ConfigFile
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return fmt.Errorf(
			"error %d: failed to unmarshal config file, more info => %w",
			errs.JSON_ERROR,
			err,
		)
	}

	// update the file if the download directory is different
	if config.DownloadDir == newDownloadPath {
		return nil
	}

	config.DownloadDir = newDownloadPath
	configFile, err = json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf(
			"error %d: failed to marshal config file, more info => %w",
			errs.JSON_ERROR,
			err,
		)
	}

	err = os.WriteFile(configFilePath, configFile, 0666)
	if err != nil {
		return fmt.Errorf(
			"error %d: failed to write config file, more info => %w",
			errs.OS_ERROR,
			err,
		)
	}
	return nil
}

// Configure and saves the config file with updated download path
func SetDefaultDownloadPath(newDownloadPath string) error {
	if !PathExists(newDownloadPath) {
		return fmt.Errorf(
			"error %d: download path does not exist, please create the directory and try again", 
			errs.INPUT_ERROR,
		)
	}

	os.MkdirAll(APP_PATH, 0755)
	configFilePath := filepath.Join(APP_PATH, "config.json")
	if !PathExists(configFilePath) {
		return saveConfig(newDownloadPath, configFilePath)
	}
	return overwriteConfig(newDownloadPath, configFilePath)
}
