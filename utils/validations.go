package utils

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
)

func ValidateDlDirPath(dlDirPath, targetDirName string) (validatedDirPath string, err error) {
	if dlDirPath == "" {
		return filepath.Join(iofuncs.DOWNLOAD_PATH, targetDirName), nil
	}

	if !iofuncs.DirPathExists(dlDirPath) {
		return "", fmt.Errorf(
			"error %d, download path does not exist or is not a directory, please create the directory and try again",
			cdlerrors.INPUT_ERROR,
		)
	}

	if filepath.Base(dlDirPath) != targetDirName {
		return filepath.Join(dlDirPath, targetDirName), nil
	}
	return dlDirPath, nil
}

// check page nums if they are in the correct format.
//
// E.g. "1-10" is valid, but "0-9" is not valid because "0" is not accepted
func ValidatePageNumInput(baseSliceLen int, pageNums []string, errMsgs []string) error {
	pageNumsLen := len(pageNums)
	if baseSliceLen != pageNumsLen {
		var msgBody error
		if len(errMsgs) > 0 {
			msgBody = errors.New(strings.Join(errMsgs, "\n"))
		} else {
			msgBody = fmt.Errorf(
				"error %d: %d URLS provided, but %d page numbers provided\nPlease provide the same number of page numbers as the number of URLs",
				cdlerrors.INPUT_ERROR,
				baseSliceLen,
				pageNumsLen,
			)
		}
		return msgBody
	}

	valid, outlier := SliceMatchesRegex(constants.PAGE_NUM_REGEX, pageNums, false)
	if !valid {
		return fmt.Errorf(
			"error %d: invalid page number format: %q\nPlease follow the format, \"1-10\", as an example.\nNote that \"0\" are not accepted! E.g. \"0-9\" is invalid",
			cdlerrors.INPUT_ERROR,
			outlier,
		)
	}
	return nil
}

type SliceTypes interface {
	~string | ~int
}

// Checks if the given target is in the given arr and returns a boolean
func SliceContains[T SliceTypes](arr []T, target T) bool {
	for _, el := range arr {
		if el == target {
			return true
		}
	}
	return false
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
		"%w\nExpecting one of the following: %s",
		msgBody,
		strings.TrimSpace(strings.Join(slice, ", ")),
	)
}

// Validates if the slice of strings contains only numbers
// Otherwise, os.Exit(1) is called after printing error messages for the user to read
func ValidateIds(args []string) error {
	for _, id := range args {
		err := ValidateId(id)
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateId(arg string) error {
	if !constants.NUMBER_REGEX.MatchString(arg) {
		return fmt.Errorf(
			"error %d: invalid ID, %q, must be a number",
			cdlerrors.INPUT_ERROR,
			arg,
		)
	}
	return nil
}
