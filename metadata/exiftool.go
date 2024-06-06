package metadata

import (
	"context"
	"fmt"
	//"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

var (
	exiftoolPath      string
	ACCEPTED_FILE_EXT = [...]string{
		".jpeg",
		".jpg",
		".png",
		".tif",
		".tiff",
		// .gif does not have the metadata for creation date
	}
)

func init() {
	_, err := exec.LookPath("exiftool")
	if err == nil {
		exiftoolPath = "exiftool"
	}
}

// Note: This is not thread-safe
func ChangeExiftoolPath(newPath string) {
	exiftoolPath = newPath
}

func isAcceptedFileExt(fileExt string) bool {
	if fileExt == "" {
		return false
	}

	for _, ext := range ACCEPTED_FILE_EXT {
		if fileExt == ext {
			return true
		}
	}
	return false
}

func GetExifDataFromImage(ctx context.Context, filePath string) (string, error) {
	if !isAcceptedFileExt(filePath) {
		return "", nil
	}

	cmd := exec.CommandContext(ctx, exiftoolPath, filePath)
	utils.PrepareCmdForBgTask(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// https://exiftool.org/TagNames/
func getCreationDateArg(fileExt string) string {
	switch fileExt {
	case ".jpeg", ".jpg", ".tif", ".tiff":
		return "-CreateDate"
	case ".png":
		return "-PNG:CreationTime"
	case ".webp":
		return "-DateTimeOriginal"
	default:
		panic("unexpected file extension in getCreationDateArg")
	}
}

func SetExifDataToImage(ctx context.Context, filePath string, exifData Metadata) error {
	if filePath == "" {
		return nil
	}

	fileExt := strings.ToLower(filepath.Ext(filePath))
	if !isAcceptedFileExt(fileExt) {
		return nil
	}

	// YYYY:MM:DD HH:MM:SS
	formattedDate := exifData.CreationDate.Format("2006:01:02 15:04:05")

	args := []string{
		fmt.Sprintf("%s=%q", getCreationDateArg(fileExt), formattedDate),
		"-overwrite_original", // to prevent creating a backup of the original file
		filePath,
	}

	cmd := exec.CommandContext(ctx, exiftoolPath, args...)
	utils.PrepareCmdForBgTask(cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
