package api

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

// Returns a readable format of the website name for the user
//
// Will panic if the site string doesn't match one of its cases.
func GetReadableSiteStr(site string) string {
	switch site {
	case constants.FANTIA:
		return constants.FANTIA_TITLE
	case constants.PIXIV_FANBOX:
		return constants.PIXIV_FANBOX_TITLE
	case constants.PIXIV:
		return constants.PIXIV_TITLE
	case constants.KEMONO:
		return constants.KEMONO_TITLE
	default:
		// panic since this is a dev error
		panic(
			fmt.Errorf(
				"error %d: invalid website, %q, in GetReadableSiteStr",
				errs.DEV_ERROR,
				site,
			),
		)
	}
}

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

// check page nums if they are in the correct format.
//
// E.g. "1-10" is valid, but "0-9" is not valid because "0" is not accepted
// If the page nums are not in the correct format, os.Exit(1) is called
func ValidatePageNumInput(baseSliceLen int, pageNums []string, errMsgs []string) error {
	pageNumsLen := len(pageNums)
	if baseSliceLen != pageNumsLen {
		var msgBody error
		if len(errMsgs) > 0 {
			msgBody = errors.New(strings.Join(errMsgs, "\n"))
		} else {
			msgBody = fmt.Errorf(
				"error %d: %d URLS provided, but %d page numbers provided\nPlease provide the same number of page numbers as the number of URLs",
				errs.INPUT_ERROR,
				baseSliceLen,
				pageNumsLen,
			)
		}
		return msgBody
	}

	valid, outlier := SliceMatchesRegex(constants.PAGE_NUM_REGEX, pageNums)
	if !valid {
		return fmt.Errorf(
			"error %d: invalid page number format: %q\nPlease follow the format, \"1-10\", as an example.\nNote that \"0\" are not accepted! E.g. \"0-9\" is invalid",
			errs.INPUT_ERROR,
			outlier,
		)
	}
	return nil
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
				errs.UNEXPECTED_ERROR,
				nums[0],
			)
		}

		max, err = strconv.Atoi(nums[1])
		if err != nil {
			return -1, -1, false, fmt.Errorf(
				"error %d: failed to convert max page number, %q, to int",
				errs.UNEXPECTED_ERROR,
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
				errs.UNEXPECTED_ERROR,
				numStr,
			)
		}
		max = min
	}
	return min, max, true, nil
}

// Checks if the given str is in the given arr and returns a boolean
func SliceContains(arr []string, str string) bool {
	for _, el := range arr {
		if el == str {
			return true
		}
	}
	return false
}

type SliceTypes interface {
	~string | ~int
}

// Removes duplicates from the given slice.
func RemoveSliceDuplicates[T SliceTypes](s []T) []T {
	var result []T
	seen := make(map[T]struct{})
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Used for removing duplicate IDs with its corresponding page number from the given slices.
//
// Returns the the new idSlice and pageSlice with the duplicates removed.
func RemoveDuplicateIdAndPageNum[T SliceTypes](idSlice, pageSlice []T) ([]T, []T) {
	var idResult, pageResult []T
	seen := make(map[T]struct{})
	for idx, v := range idSlice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			idResult = append(idResult, v)
			pageResult = append(pageResult, pageSlice[idx])
		}
	}
	return idResult, pageResult
}

// Checks if the slice of string contains the target str. Otherwise, returns an error.
func ValidateStrArgs(str string, slice, errMsgs []string) (string, error) {
	if SliceContains(slice, str) {
		return str, nil
	}

	var msgBody error
	if len(errMsgs) > 0 {
		msgBody = errors.New(strings.Join(errMsgs, "\n"))
	} else {
		msgBody = fmt.Errorf("input error, got: %s", str)
	}
	return "", fmt.Errorf(
		"%v\nExpecting one of the following: %s", 
		msgBody,
		strings.TrimSpace(strings.Join(slice, ", ")),
	)
}

// Validates if the slice of strings contains only numbers
// Otherwise, os.Exit(1) is called after printing error messages for the user to read
func ValidateIds(args []string) error {
	for _, id := range args {
		if !constants.NUMBER_REGEX.MatchString(id) {
			return fmt.Errorf(
				"error %d: invalid ID, %q, must be a number",
				errs.INPUT_ERROR,
				id,
			)
		}
	}
	return nil
}

// Same as strings.Join([]string, "\n")
func CombineStringsWithNewline(strs ...string) string {
	return strings.Join(strs, "\n")
}

// Checks if the slice of string all matches the given regex pattern
//
// Returns true if all matches, false otherwise with the outlier string
func SliceMatchesRegex(regex *regexp.Regexp, slice []string) (bool, string) {
	for _, str := range slice {
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
