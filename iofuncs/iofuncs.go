package iofuncs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// checks if a file or directory exists
func PathExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

// similar to PathExists but checks if the path exists and is a directory
func DirPathExists(dirPath string) bool {
	stat, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

// Returns the file size based on the provided file path
//
// If the file does not exist or
// there was an error opening the file at the given file path string, -1 is returned
func GetFileSize(filePath string) (int64, error) {
	fileStat, err := os.Stat(filePath)
	if err != nil {
		return -1, err
	}

	// check if it's a directory
	if fileStat.IsDir() {
		return -1, fmt.Errorf("file %s is a directory", filePath)
	}
	return fileStat.Size(), nil
}

// Uses bufio.Reader to read a line from a file and returns it as a byte slice
//
// Mostly thanks to https://devmarkpro.com/working-big-files-golang
func ReadLine(reader *bufio.Reader) ([]byte, error) {
	var err error
	var isPrefix = true
	var totalLine, line []byte

	// Read until isPrefix is false as
	// that means the line has been fully read
	for isPrefix && err == nil {
		line, isPrefix, err = reader.ReadLine()
		totalLine = append(totalLine, line...)
	}
	return totalLine, err
}

// Returns the path without the file extension
func RemoveExtFromFilename(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}

// Used in CleanPathName to remove illegal characters in a path name
func removeIllegalRuneInPath(r rune) rune {
	if strings.ContainsRune("<>:\"/\\|?*\n\r\t", r) {
		return '-'
	} else if r == '.' {
		return ','
	}
	return r
}

// Removes any illegal characters in a path name
// to prevent any error with file I/O using the path name
func CleanPathName(pathName string) string {
	pathName = strings.TrimSpace(pathName)
	if len(pathName) > 255 {
		pathName = pathName[:255]
	}
	return strings.Map(removeIllegalRuneInPath, pathName)
}

// Returns a directory path for a post, artwork, etc.
// based on the user's saved download path and the provided arguments
func GetPostFolder(downloadPath, creatorName, postId, postTitle string) string {
	creatorName = CleanPathName(creatorName)
	postTitle = CleanPathName(postTitle)

	postFolderPath := filepath.Join(
		downloadPath,
		creatorName,
		fmt.Sprintf("[%s] %s", postId, postTitle),
	)
	return postFolderPath
}
