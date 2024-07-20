package gdrive

import (
	"fmt"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

// censor the key=... part of the URL to <REDACTED>.
// This is to prevent the API key from being leaked in the logs.
func censorApiKeyFromStr(str string) string {
	return constants.GDRIVE_API_KEY_PARAM_REGEX.ReplaceAllString(str, "key=<REDACTED>")
}

// Gets the error message for a failed GDrive API call
func getFailedApiCallErr(res *http.Response) error {
	requestUrl := res.Request.URL.String()
	return fmt.Errorf(
		"error while fetching from GDrive...\n"+
			"GDrive URL (May not be accurate): https://drive.google.com/file/d/%s/view?usp=sharing\n"+
			"Status Code: %s\nURL: %s",
		httpfuncs.GetLastPartOfUrl(requestUrl),
		res.Status,
		censorApiKeyFromStr(requestUrl),
	)
}

// Returns the contents of the given GDrive folder using Google's GDrive package
func (gdrive *GDrive) getFolderContentsWithClient(folderId string) ([]*GdriveFileToDl, error) {
	var pageToken string
	var gdriveFiles []*GdriveFileToDl
	query := fmt.Sprintf("'%s' in parents", folderId)

	for {
		action := gdrive.client.Files.List().Q(query).Fields(constants.GDRIVE_FOLDER_FIELDS)
		if pageToken != "" {
			action = action.PageToken(pageToken)
		}
		files, err := action.Context(gdrive.ctx).Do()
		if err != nil {
			return nil, fmt.Errorf(
				"gdrive error %d: failed to get folder contents with ID of %s, more info => %w",
				cdlerrors.CONNECTION_ERROR,
				folderId,
				err,
			)
		}

		for _, file := range files.Files {
			gdriveFiles = append(gdriveFiles, &GdriveFileToDl{
				Id:          file.Id,
				Name:        file.Name,
				Size:        file.Size,
				MimeType:    file.MimeType,
				Md5Checksum: file.Md5Checksum,
				FilePath:    "",
			})
		}

		if files.NextPageToken == "" {
			break
		}
		pageToken = files.NextPageToken
	}
	return gdriveFiles, nil
}

// Returns the contents of the given GDrive folder
func (gdrive *GDrive) GetFolderContents(folderId, logPath string) ([]*GdriveFileToDl, error) {
	return gdrive.getFolderContentsWithClient(folderId)
}

// Retrieves the content of a GDrive folder and its subfolders recursively using GDrive API v3
func (gdrive *GDrive) GetNestedFolderContents(folderId, logPath string) ([]*GdriveFileToDl, error) {
	var files []*GdriveFileToDl
	folderContents, err := gdrive.GetFolderContents(folderId, logPath)
	if err != nil {
		return nil, err
	}

	for _, file := range folderContents {
		if file.MimeType == "application/vnd.google-apps.folder" {
			subFolderFiles, err := gdrive.GetNestedFolderContents(file.Id, logPath)
			if err != nil {
				return nil, err
			}
			files = append(files, subFolderFiles...)
		} else {
			files = append(files, file)
		}
	}
	return files, nil
}

// Retrieves the file details of the given GDrive file using Google's GDrive package
func (gdrive *GDrive) getFileDetailsWithClient(gdriveInfo *GDriveToDl) (*GdriveFileToDl, error) {
	file, err := gdrive.client.Files.Get(gdriveInfo.Id).Fields(constants.GDRIVE_FILE_FIELDS).Context(gdrive.ctx).Do()
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: failed to get file details with ID of %s, more info => %w",
			cdlerrors.CONNECTION_ERROR,
			gdriveInfo.Id,
			err,
		)
	}
	return &GdriveFileToDl{
		Id:          file.Id,
		Name:        file.Name,
		Size:        file.Size,
		MimeType:    file.MimeType,
		Md5Checksum: file.Md5Checksum,
		FilePath:    gdriveInfo.FilePath,
	}, nil
}

// Retrieves the file details of the given GDrive file using GDrive API v3
func (gdrive *GDrive) GetFileDetails(gdriveInfo *GDriveToDl) (*GdriveFileToDl, error) {
	return gdrive.getFileDetailsWithClient(gdriveInfo)
}
