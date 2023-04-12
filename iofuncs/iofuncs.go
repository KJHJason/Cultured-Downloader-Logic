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

// Returns the file size based on the provided file path
//
// If the file does not exist or
// there was an error opening the file at the given file path string, -1 is returned
func GetFileSize(filePath string) (int64, error) {
	if !PathExists(filePath) {
		return -1, os.ErrNotExist
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return -1, err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return -1, err
	}
	return fileInfo.Size(), nil
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

// Used in CleanPathName to remove illegal characters in a path name
func removeIllegalRuneInPath(r rune) rune {
	if strings.ContainsRune("<>:\"/\\|?*\n\r\t", r) {
		return '-'
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
