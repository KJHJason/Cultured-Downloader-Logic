package metadata

import (
	"context"
	"fmt"
	"os"
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
		".webp",
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
		return "CreateDate"
	case ".png":
		return "PNG:CreationTime"
	case ".webp":
		return "DateTimeOriginal"
	default:
		panic("unexpected file extension in getCreationDateArg")
	}
}

func getCreationDateArgWithFilePath(filePath string) string {
	return getCreationDateArg(strings.ToLower(filepath.Ext(filePath)))
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
		fmt.Sprintf("-%s=%q", getCreationDateArg(fileExt), formattedDate),
		"-overwrite_original", // to prevent creating a backup of the original file
		filePath,
	}
	cmd := exec.CommandContext(ctx, exiftoolPath, args...)
	utils.PrepareCmdForBgTask(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set exif data to image: %w\n%s", err, string(out))
	}

	return nil
}

func GetFileSizeWithoutExifData(ctx context.Context, filePath string) (int64, error) {
	if filePath == "" {
		return 0, nil
	}

	tempFile, err := os.CreateTemp("", "cdl-file-wo-metadata-")
	if err != nil {
		return 0, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFile.Close()

	tempPath := tempFile.Name()
	os.Remove(tempPath)

	// remove again as exiftool will create another file
	// without the metadata for getting the file size
	defer os.Remove(tempPath)

	args := []string{
		fmt.Sprintf("-%s=", getCreationDateArgWithFilePath(filePath)),
		"-o", tempPath,
		filePath,
	}
	cmd := exec.CommandContext(ctx, exiftoolPath, args...)
	utils.PrepareCmdForBgTask(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("failed to produce a file without exif data: %w\n%s", err, string(out))
	}

	stat, err := os.Stat(tempPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size without exif data: %w", err)
	}
	return stat.Size(), nil
}
