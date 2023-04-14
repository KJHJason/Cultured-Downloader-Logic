package gdrive

import (
	"fmt"
	"net/http"

	"github.com/KJHJason/Cultured-Downloader-Logic/configs"
	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
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

// Returns the contents of the given GDrive folder
func (gdrive *GDrive) GetFolderContents(folderId, logPath string, config *configs.Config) ([]*GdriveFileToDl, error) {
	params := map[string]string{
		"key":    gdrive.apiKey,
		"q":      fmt.Sprintf("'%s' in parents", folderId),
		"fields": fmt.Sprintf("nextPageToken,files(%s)", GDRIVE_FILE_FIELDS),
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

// Retrieves the file details of the given GDrive file using GDrive API v3
func (gdrive *GDrive) GetFileDetails(gdriveInfo *GDriveToDl, config *configs.Config) (*GdriveFileToDl, error) {
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
