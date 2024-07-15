package cf

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

const (
	EMBEDDED_DIR_NAME = "python_scripts"
)

var (
	//go:embed python_scripts/LICENSE
	//go:embed python_scripts/README.md
	//go:embed python_scripts/requirements.txt
	//go:embed python_scripts/main.py
	//go:embed python_scripts/_logger/*.py
	//go:embed python_scripts/_types/*.py
	//go:embed python_scripts/constants/*.py
	//go:embed python_scripts/errors/*.py
	//go:embed python_scripts/extensions/**
	//go:embed python_scripts/logic/*.py
	//go:embed python_scripts/parser/*.py
	//go:embed python_scripts/test/*.py
	//go:embed python_scripts/utils/*.py
	pythonScripts embed.FS

	panicHandler = func(err error) {
		logger.LogError(err, logger.FATAL)
	}
)

func getCfDirPath() string {
	return filepath.Join(iofuncs.APP_PATH, "kjhjason-cf-py")
}

func getMainPyPath() string {
	return filepath.Join(getCfDirPath(), "main.py")
}

func getVenvDirPath() string {
	return filepath.Join(getCfDirPath(), "venv")
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

func InitFiles() {
	cfDirPath := getCfDirPath()
	os.MkdirAll(cfDirPath, constants.DEFAULT_PERMS)

	files, err := readFsDir(pythonScripts, EMBEDDED_DIR_NAME)
	if err != nil {
		panicHandler(
			fmt.Errorf(
				"error %d: failed to read embedded files -> %w",
				cdlerrors.UNEXPECTED_ERROR, err,
			),
		)
	}
	for _, file := range files {
		checkAndWriteFile(filepath.Join(cfDirPath, file.path), file.data)
	}

	if err := TestScript(); err != nil {
		panicHandler(err)
	}
}

func pipInstallRequirements(reqTxtFilePath string) {
	venvPath := getVenvDirPath()
	if iofuncs.PathExists(venvPath) {
		// delete venv if it exists
		err := os.RemoveAll(venvPath)
		if err != nil {
			panicHandler(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python", "-m", "venv", venvPath)
	utils.PrepareCmdForBgTask(cmd)

	err := cmd.Run()
	if err != nil {
		panicHandler(err)
	}

	installCtx, installCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer installCancel()
	pipPath := filepath.Join(venvPath, getPyVenvBinDirName(), "pip")
	cmd = exec.CommandContext(installCtx, pipPath, "install", "-r", reqTxtFilePath)
	utils.PrepareCmdForBgTask(cmd)

	err = cmd.Run()
	if err != nil {
		panicHandler(err)
	}
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
