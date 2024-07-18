package cf

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
)

const (
	CF_DIR_PREFIX     = "kjhjason-cdl-cf"
	CF_DIR_NAME       = CF_DIR_PREFIX + "-" + VERSION
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
	//go:embed python_scripts/extensions/**
	//go:embed python_scripts/logic/*.py
	//go:embed python_scripts/parser/*.py
	//go:embed python_scripts/test/*.py
	//go:embed python_scripts/utils/*.py
	pythonScripts embed.FS
)

func getCfDirPath() string {
	return filepath.Join(iofuncs.APP_PATH, CF_DIR_NAME)
}

func getMainPyPath() string {
	return filepath.Join(getCfDirPath(), "main.py")
}

func getVenvDirPath() string {
	return filepath.Join(getCfDirPath(), "venv")
}

func InitPyFiles() error {
	cfDirPath := getCfDirPath()
	os.MkdirAll(cfDirPath, constants.DEFAULT_PERMS)

	if err := removeOldCfDir(); err != nil {
		return err
	}

	files, err := readFsDir(pythonScripts, EMBEDDED_DIR_NAME)
	if err != nil {
		return fmt.Errorf(
			"error %d: failed to read embedded files -> %w",
			cdlerrors.UNEXPECTED_ERROR, err,
		)
	}
	for _, file := range files {
		err := checkAndWriteFile(filepath.Join(cfDirPath, file.path), file.data)
		if err != nil {
			return err
		}
	}

	if err := TestScript(); err != nil {
		return err
	}
	return nil
}
