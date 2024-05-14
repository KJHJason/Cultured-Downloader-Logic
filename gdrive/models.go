package gdrive

import (
	"fmt"
	"strconv"

	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

type GDriveToDl struct {
	Id       string
	Type     string
	FilePath string
}

type GdriveFileToDl struct {
	Id          string
	Name        string
	Size        string
	MimeType    string
	Md5Checksum string
	FilePath    string
}

// Convert the size of the file to int64 and return it
func (f GdriveFileToDl) GetIntSize() int64 {
	size, err := strconv.ParseInt(f.Size, 10, 64)
	if err != nil {
		// shouldn't happen
		panic(
			fmt.Errorf(
				"gdrive error %d: failed to convert the size of the file to int64, more info => %w",
				cdlerrors.UNEXPECTED_ERROR,
				err,
			),
		)
	}
	return size
}

type GdriveError struct {
	Err      error
	FilePath string
}
