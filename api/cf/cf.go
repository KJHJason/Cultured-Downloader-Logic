package cf

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

const (
	cfPyFilename            = "cf.py"
	cfLogicPyFilename       = "cf_logic.py"
	requirementsTxtFilename = "requirements.txt"
	licenseFilename         = "LICENSE"
	readmeFilename          = "README.md"
)

var (
	//go:embed python_scripts/cf.py
	cfPyData []byte
	//go:embed python_scripts/cf_logic.py
	cfLogicPyData []byte
	//go:embed python_scripts/requirements.txt
	requirementsTxtData []byte
	//go:embed python_scripts/LICENSE
	licenseData []byte
	//go:embed python_scripts/README.md
	readmeData []byte

	panicHandler = func(err error) {
		logger.LogError(err, logger.FATAL)
	}
)

func getCfDirPath() string {
	return filepath.Join(iofuncs.APP_PATH, "kjhjason-cf-py")
}

func getCfPyPath() string {
	return filepath.Join(getCfDirPath(), cfPyFilename)
}

func getVenvDirPath() string {
	return filepath.Join(getCfDirPath(), "venv")
}

func InitFiles() {
	cfDirPath := getCfDirPath()
	os.MkdirAll(cfDirPath, constants.DEFAULT_PERMS)

	// Get the local paths for the files
	cfPyPath := filepath.Join(cfDirPath, cfPyFilename)
	cfLogicPyPath := filepath.Join(cfDirPath, cfLogicPyFilename)
	requirementsTxtPath := filepath.Join(cfDirPath, requirementsTxtFilename)
	licensePath := filepath.Join(cfDirPath, licenseFilename)
	readmePath := filepath.Join(cfDirPath, readmeFilename)

	checkAndWriteFile(cfPyPath, cfPyData)
	checkAndWriteFile(cfLogicPyPath, cfLogicPyData)
	checkAndWriteFile(requirementsTxtPath, requirementsTxtData)
	checkAndWriteFile(licensePath, licenseData)
	checkAndWriteFile(readmePath, readmeData)

	// free up memory after writing the files
	cfPyData = nil
	cfLogicPyData = nil
	requirementsTxtData = nil
	licenseData = nil
	readmeData = nil

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

		if string(localData) == string(embeddedData) {
			return
		}
	}

	err := os.WriteFile(filePath, embeddedData, constants.DEFAULT_PERMS)
	if err != nil {
		panicHandler(err)
	}

	if filepath.Base(filePath) == requirementsTxtFilename {
		pipInstallRequirements(filePath)
	}
}
