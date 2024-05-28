package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

// Convert the page number to the offset as one page might have x posts.
//
// Usually for paginated results like Pixiv's mobile API (60 per page), checkPixivMax should be set to true.
func ConvertPageNumToOffset(minPageNum, maxPageNum, perPage int) (int, int) {
	// Check for negative page numbers
	if minPageNum < 0 {
		minPageNum = 1
	}
	if maxPageNum < 0 {
		maxPageNum = 1
	}

	// Swap the page numbers if the min is greater than the max
	if minPageNum > maxPageNum {
		minPageNum, maxPageNum = maxPageNum, minPageNum
	}

	minOffset := perPage * (minPageNum - 1)
	maxOffset := perPage * (maxPageNum - minPageNum + 1)
	return minOffset, maxOffset
}

// Returns the min, max, hasMaxNum, and error from the given string of "num" or "min-max"
//
// E.g.
//
//	"1-10" => 1, 10, true, nil
//	"1" => 1, 1, true, nil
//	"" => 1, 1, false, nil (defaults to min = 1, max = inf)
func GetMinMaxFromStr(numStr string) (int, int, bool, error) {
	if numStr == "" {
		// defaults to min = 1, max = inf
		return 1, 1, false, nil
	}

	var err error
	var min, max int
	if strings.Contains(numStr, "-") {
		nums := strings.SplitN(numStr, "-", 2)
		min, err = strconv.Atoi(nums[0])
		if err != nil {
			return -1, -1, false, fmt.Errorf(
				"error %d: failed to convert min page number, %q, to int",
				cdlerrors.UNEXPECTED_ERROR,
				nums[0],
			)
		}

		max, err = strconv.Atoi(nums[1])
		if err != nil {
			return -1, -1, false, fmt.Errorf(
				"error %d: failed to convert max page number, %q, to int",
				cdlerrors.UNEXPECTED_ERROR,
				nums[1],
			)
		}

		if min > max {
			min, max = max, min
		}
	} else {
		min, err = strconv.Atoi(numStr)
		if err != nil {
			return -1, -1, false, fmt.Errorf(
				"error %d: failed to convert page number, %q, to int",
				cdlerrors.UNEXPECTED_ERROR,
				numStr,
			)
		}
		max = min
	}
	return min, max, true, nil
}

// Same as strings.Join([]string, "\n")
func CombineStringsWithNewline(strs ...string) string {
	return strings.Join(strs, "\n")
}

// Checks if the slice of string all matches the given regex pattern
// If strict is true, then all string must match the regex pattern. Otherwise, empty strings are allowed.
//
// Returns true if all matches, false otherwise with the outlier string
func SliceMatchesRegex(regex *regexp.Regexp, slice []string, strict bool) (bool, string) {
	for _, str := range slice {
		if str == "" && !strict {
			continue
		}
		if !regex.MatchString(str) {
			return false, str
		}
	}
	return true, ""
}

// Detects if the given string contains any passwords
func DetectPasswordInText(text string) bool {
	for _, passwordText := range constants.PASSWORD_TEXTS {
		if strings.Contains(text, passwordText) {
			return true
		}
	}

	for _, passwordRegex := range constants.PASSWORD_REGEXES {
		if passwordRegex.MatchString(text) {
			return true
		}
	}
	return false
}

// Detects if the given string contains any GDrive links and logs it if detected
func DetectGDriveLinks(text, postFolderPath string, isUrl, logUrls bool) bool {
	gdriveFilepath := filepath.Join(postFolderPath, constants.GDRIVE_FILENAME)
	containsGDriveLink := false
	if isUrl && strings.HasPrefix(text, constants.GDRIVE_URL) {
		containsGDriveLink = true
	} else if strings.Contains(text, constants.GDRIVE_URL) {
		containsGDriveLink = true
	}

	if !containsGDriveLink {
		return false
	}

	if isUrl {
		gdriveText := fmt.Sprintf(
			"Google Drive link detected: %s\n\n",
			text,
		)
		logger.LogMessageToPath(gdriveText, gdriveFilepath, logger.INFO)
	}
	return true
}

// Detects if the given string contains any other external file hosting providers links and logs it if detected
func DetectOtherExtDLLink(text, postFolderPath string) bool {
	otherExtFilepath := filepath.Join(postFolderPath, constants.OTHER_LINKS_FILENAME)
	for _, extDownloadProvider := range constants.EXTERNAL_DOWNLOAD_PLATFORMS {
		if strings.Contains(text, extDownloadProvider) {
			otherExtText := fmt.Sprintf(
				"Detected a link that points to an external file hosting in post's description:\n%s\n\n",
				text,
			)
			logger.LogMessageToPath(otherExtText, otherExtFilepath, logger.INFO)
			return true
		}
	}
	return false
}
