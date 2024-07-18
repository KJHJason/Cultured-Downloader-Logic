package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

type MetadataTypes interface {
	FantiaPost | FantiaProduct | KemonoPost | PixivFanboxPost
}

// Marshal the metadata into a JSON file
// and write it to the given file path.
//
// If the file already exists, do nothing.
func WriteMetadata[T MetadataTypes](metadata T, dirPath string) error {
	filePath := filepath.Join(dirPath, "post_metadata.json")
	if iofuncs.PathExists(filePath) {
		return nil
	}

	os.MkdirAll(dirPath, constants.DEFAULT_PERMS)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, constants.DEFAULT_PERMS)
	if err != nil {
		logger.MainLogger.Errorf(
			"error %d: error opening file for writing metadata %q => %v",
			cdlerrors.OS_ERROR, filePath, err,
		)
		return err
	}
	defer file.Close()

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		logger.MainLogger.Errorf(
			"error %d: error marshalling metadata json => %v",
			cdlerrors.JSON_ERROR, err,
		)
		return err
	}

	_, err = file.Write(metadataBytes)
	if err != nil {
		logger.MainLogger.Errorf(
			"error %d: error writing metadata to file %q => %v",
			cdlerrors.OS_ERROR, filePath, err,
		)
		return err
	}
	logger.MainLogger.Infof("Metadata written to %q", filePath)
	return nil
}
