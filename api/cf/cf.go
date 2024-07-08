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
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

// TODO: log the panics

var (
	//go:embed requirements.txt
	requirementsTxtData []byte
	//go:embed main.py
	mainPyData []byte
	//go:embed cf_logic.py
	cfLogicPyData []byte
)

func getCfDirPath() string {
	return filepath.Join(iofuncs.APP_PATH, "kjhjason-cf-py")
}

func pipInstallRequirements(reqTxtFilePath string) {
	venvPath := filepath.Join(getCfDirPath(), "venv")
	// delete venv if it exists
	if iofuncs.PathExists(venvPath) {
		err := os.RemoveAll(venvPath)
		if err != nil {
			panic(err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python", "-m", "venv", venvPath)
	utils.PrepareCmdForBgTask(cmd)

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	installCtx, installCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer installCancel()
	cmd = exec.CommandContext(installCtx, filepath.Join(venvPath, "Scripts", "pip"), "install", "-r", reqTxtFilePath)
	utils.PrepareCmdForBgTask(cmd)

	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

func checkAndWriteFile(filePath string, embeddedData []byte, isReqTxt bool) {
	if iofuncs.PathExists(filePath) {
		localData, err := os.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		if string(localData) == string(embeddedData) {
			return
		}
	}

	err := os.WriteFile(filePath, embeddedData, constants.DEFAULT_PERMS)
	if err != nil {
		panic(err)
	}

	if isReqTxt {
		pipInstallRequirements(filePath)
	}
}

func init() {
	cfDirPath := getCfDirPath()
	os.MkdirAll(cfDirPath, constants.DEFAULT_PERMS)

	requirementsTxtPath := filepath.Join(cfDirPath, "requirements.txt")
	mainPyPath := filepath.Join(cfDirPath, "main.py")
	cfLogicPyPath := filepath.Join(cfDirPath, "cf_logic.py")

	checkAndWriteFile(requirementsTxtPath, requirementsTxtData, true)
	checkAndWriteFile(mainPyPath, mainPyData, false)
	checkAndWriteFile(cfLogicPyPath, cfLogicPyData, false)

	requirementsTxtData = nil
	mainPyData = nil
	cfLogicPyData = nil
}
