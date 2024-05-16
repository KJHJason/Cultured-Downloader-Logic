package gdrive

import (
	"fmt"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

type GDriveToDl struct {
	Id       string
	Type     string
	FilePath string
}

type GdriveFileToDl struct {
	Id          string
	Name        string
	Size        int64
	MimeType    string
	Md5Checksum string
	FilePath    string
}

func (g GdriveFileToDl) GetUrl() string {
	return fmt.Sprintf("%s/%s", constants.GDRIVE_FILE_API_URL, g.Id)
}

type GdriveError struct {
	Err      error
	FilePath string
}
