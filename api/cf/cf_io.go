package cf

import (
	"bytes"
	"embed"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
)

func removeOldCfDir() {
	entries, err := os.ReadDir(iofuncs.APP_PATH)
	if err != nil {
		panicHandler(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		entryName := entry.Name()
		if strings.HasPrefix(entryName, CF_DIR_PREFIX) && entryName != CF_DIR_NAME {
			os.RemoveAll(filepath.Join(iofuncs.APP_PATH, entryName))
		}
	}
}

type FileInfo struct {
	data []byte
	path string
}

// Note the issue https://github.com/golang/go/issues/45230,
// filepath.Join on Windows uses '\' instead of '/'
// which will cause embed.FS to return file does not exist error!
func readFsDir(fs embed.FS, dirName string) ([]*FileInfo, error) {
	dir, err := fs.ReadDir(dirName)
	if err != nil {
		return nil, err
	}

	var files []*FileInfo
	for _, file := range dir {
		fileOrDirPath := path.Join(dirName, file.Name())
		if file.IsDir() {
			subFiles, err := readFsDir(fs, fileOrDirPath)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else {
			data, err := fs.ReadFile(fileOrDirPath)
			if err != nil {
				return nil, err
			}
			files = append(files, &FileInfo{
				data: data,
				path: strings.TrimPrefix(fileOrDirPath, EMBEDDED_DIR_NAME+"/"),
			})
		}
	}
	return files, nil
}

func checkAndWriteFile(filePath string, embeddedData []byte) {
	if iofuncs.PathExists(filePath) {
		localData, err := os.ReadFile(filePath)
		if err != nil {
			panicHandler(err)
		}

		if bytes.Equal(embeddedData, localData) {
			return
		}
	}

	os.MkdirAll(filepath.Dir(filePath), constants.DEFAULT_PERMS)
	err := os.WriteFile(filePath, embeddedData, constants.DEFAULT_PERMS)
	if err != nil {
		panicHandler(err)
	}

	if filepath.Base(filePath) == "requirements.txt" {
		pipInstallRequirements(filePath)
	}
}
