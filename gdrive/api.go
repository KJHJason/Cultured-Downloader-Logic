package gdrive

import (
	"fmt"
	"strconv"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

// censor the key=... part of the URL to <REDACTED>.
// This is to prevent the API key from being leaked in the logs.
func censorApiKeyFromStr(str string) string {
	return API_KEY_PARAM_REGEX.ReplaceAllString(str, "key=<REDACTED>")
}

// Gets the error message for a failed GDrive API call
func getFailedApiCallErr(res *http.Response) error {
	requestUrl := res.Request.URL.String()
	return fmt.Errorf(
		"error while fetching from GDrive...\n" +
			"GDrive URL (May not be accurate): https://drive.google.com/file/d/%s/view?usp=sharing\n" +
				"Status Code: %s\nURL: %s",
		httpfuncs.GetLastPartOfUrl(requestUrl),
		res.Status,
		censorApiKeyFromStr(requestUrl),
	)
}

// Returns the contents of the given GDrive folder using Google's GDrive package
func (gdrive *GDrive) getFolderContentsWithClient(folderId, logPath string, config *configs.Config) ([]*GdriveFileToDl, error) {
	var pageToken string
	var gdriveFiles []*GdriveFileToDl
	for {
		action := gdrive.client.Files.List().Q(fmt.Sprintf("'%s' in parents", folderId)).Fields(GDRIVE_FOLDER_FIELDS)
		if pageToken != "" {
			action = action.PageToken(pageToken)
		}
		files, err := action.Do()
		if err != nil {
			return nil, fmt.Errorf(
				"gdrive error %d: failed to get folder contents with ID of %s, more info => %v",
				constants.CONNECTION_ERROR,
				folderId,
				err,
			)
		}

		for _, file := range files.Files {
			gdriveFiles = append(gdriveFiles, &GdriveFileToDl{
				Id:          file.Id,
				Name:        file.Name,
				Size:        strconv.FormatInt(file.Size, 10),
				MimeType:    file.MimeType,
				Md5Checksum: file.Md5Checksum,
				FilePath:    "",
			})
		}

		if files.NextPageToken == "" {
			break
		} else {
			pageToken = files.NextPageToken
		}
	}
	return gdriveFiles, nil
}

// Returns the contents of the given GDrive folder using API calls to GDrive API v3
func (gdrive *GDrive) getFolderContentsWithApi(folderId, logPath string, config *configs.Config) ([]*GdriveFileToDl, error) {
	params := map[string]string{
		"key":    gdrive.apiKey,
		"q":      fmt.Sprintf("'%s' in parents", folderId),
		"fields": GDRIVE_FOLDER_FIELDS,
	}
	var files []*GdriveFileToDl
	pageToken := ""
	for {
		if pageToken != "" {
			params["pageToken"] = pageToken
		} else {
			delete(params, "pageToken")
		}
		res, err := httpfuncs.CallRequest(
			&httpfuncs.RequestArgs{
				Url:       gdrive.apiUrl,
				Method:    "GET",
				Timeout:   gdrive.timeout,
				Params:    params,
				UserAgent: config.UserAgent,
				Http2:     !HTTP3_SUPPORTED,
				Http3:     HTTP3_SUPPORTED,
			},
		)
		if err != nil {
			return nil, fmt.Errorf(
				"gdrive error %d: failed to get folder contents with ID of %s, more info => %v",
				constants.CONNECTION_ERROR,
				folderId,
				err,
			)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return nil, fmt.Errorf(
				"gdrive error %d: failed to get folder contents with ID of %s, more info => %s",
				constants.RESPONSE_ERROR,
				folderId,
				res.Status,
			)
		}

		var gdriveFolder GDriveFolder
		if err := httpfuncs.LoadJsonFromResponse(res, &gdriveFolder); err != nil {
			return nil, err
		}

		for _, file := range gdriveFolder.Files {
			files = append(files, &GdriveFileToDl{
				Id:          file.Id,
				Name:        file.Name,
				Size:        file.Size,
				MimeType:    file.MimeType,
				Md5Checksum: file.Md5Checksum,
				FilePath:    "",
			})
		}

		if gdriveFolder.NextPageToken == "" {
			break
		} else {
			pageToken = gdriveFolder.NextPageToken
		}
	}
	return files, nil
}

// Returns the contents of the given GDrive folder
func (gdrive *GDrive) GetFolderContents(folderId, logPath string, config *configs.Config) ([]*GdriveFileToDl, error) {
	if gdrive.client != nil {
		return gdrive.getFolderContentsWithClient(folderId, logPath, config)
	}
	return gdrive.getFolderContentsWithApi(folderId, logPath, config)
}

// Retrieves the content of a GDrive folder and its subfolders recursively using GDrive API v3
func (gdrive *GDrive) GetNestedFolderContents(folderId, logPath string, config *configs.Config) ([]*GdriveFileToDl, error) {
	var files []*GdriveFileToDl
	folderContents, err := gdrive.GetFolderContents(folderId, logPath, config)
	if err != nil {
		return nil, err
	}

	for _, file := range folderContents {
		if file.MimeType == "application/vnd.google-apps.folder" {
			subFolderFiles, err := gdrive.GetNestedFolderContents(file.Id, logPath, config)
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

// Retrieves the file details of the given GDrive file by making a HTTP request to GDrive API v3
func (gdrive *GDrive) getFileDetailsWithAPI(gdriveInfo *GDriveToDl, config *configs.Config) (*GdriveFileToDl, error) {
	params := map[string]string{
		"key":    gdrive.apiKey,
		"fields": GDRIVE_FILE_FIELDS,
	}
	url := fmt.Sprintf("%s/%s", gdrive.apiUrl, gdriveInfo.Id)
	res, err := httpfuncs.CallRequest(
		&httpfuncs.RequestArgs{
			Url:       url,
			Method:    "GET",
			Timeout:   gdrive.timeout,
			Params:    params,
			UserAgent: config.UserAgent,
			Http2:     !HTTP3_SUPPORTED,
			Http3:     HTTP3_SUPPORTED,
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: failed to get file details with ID of %s, more info => %v",
			constants.CONNECTION_ERROR,
			gdriveInfo.Id,
			err,
		)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, getFailedApiCallErr(res)
	}

	var gdriveFile GDriveFile
	if err := httpfuncs.LoadJsonFromResponse(res, &gdriveFile); err != nil {
		return nil, err
	}

	return &GdriveFileToDl{
		Id:          gdriveFile.Id,
		Name:        gdriveFile.Name,
		Size:        gdriveFile.Size,
		MimeType:    gdriveFile.MimeType,
		Md5Checksum: gdriveFile.Md5Checksum,
		FilePath:    gdriveInfo.FilePath,
	}, nil
}

// Retrieves the file details of the given GDrive file using Google's GDrive package
func (gdrive *GDrive) getFileDetailsWithClient(gdriveInfo *GDriveToDl, config *configs.Config) (*GdriveFileToDl, error) {
	file, err := gdrive.client.Files.Get(gdriveInfo.Id).Fields(GDRIVE_FILE_FIELDS).Do()
	if err != nil {
		return nil, fmt.Errorf(
			"gdrive error %d: failed to get file details with ID of %s, more info => %v",
			constants.CONNECTION_ERROR,
			gdriveInfo.Id,
			err,
		)
	}
	return &GdriveFileToDl{
		Id:          file.Id,
		Name:        file.Name,
		Size:        strconv.FormatInt(file.Size, 10),
		MimeType:    file.MimeType,
		Md5Checksum: file.Md5Checksum,
		FilePath:    gdriveInfo.FilePath,
	}, nil
}

// Retrieves the file details of the given GDrive file using GDrive API v3
func (gdrive *GDrive) GetFileDetails(gdriveInfo *GDriveToDl, config *configs.Config) (*GdriveFileToDl, error) {
	if gdrive.client != nil {
		return gdrive.getFileDetailsWithClient(gdriveInfo, config)
	} 
	return gdrive.getFileDetailsWithAPI(gdriveInfo, config)
}
