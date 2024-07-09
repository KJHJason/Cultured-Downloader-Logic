package cf

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

var (
	//go:embed requirements.txt
	requirementsTxtData []byte
	//go:embed cf.py
	cfPyData []byte
	//go:embed cf_logic.py
	cfLogicPyData []byte

	panicHandler = func(err error) {
		logger.LogError(err, logger.FATAL)
	}
)

func getCfDirPath() string {
	return filepath.Join(iofuncs.APP_PATH, "kjhjason-cf-py")
}

func getVenvDirPath() string {
	return filepath.Join(getCfDirPath(), "venv")
}

func pipInstallRequirements(reqTxtFilePath string) {
	venvPath := getVenvDirPath()
	// delete venv if it exists
	if iofuncs.PathExists(venvPath) {
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
	cmd = exec.CommandContext(installCtx, filepath.Join(venvPath, "Scripts", "pip"), "install", "-r", reqTxtFilePath)
	utils.PrepareCmdForBgTask(cmd)

	err = cmd.Run()
	if err != nil {
		panicHandler(err)
	}
}

func checkAndWriteFile(filePath string, embeddedData []byte, isReqTxt bool) {
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

	if isReqTxt {
		pipInstallRequirements(filePath)
	}
}

func InitFiles() {
	cfDirPath := getCfDirPath()
	os.MkdirAll(cfDirPath, constants.DEFAULT_PERMS)

	requirementsTxtPath := filepath.Join(cfDirPath, "requirements.txt")
	mainPyPath := filepath.Join(cfDirPath, "cf.py")
	cfLogicPyPath := filepath.Join(cfDirPath, "cf_logic.py")

	checkAndWriteFile(requirementsTxtPath, requirementsTxtData, true)
	checkAndWriteFile(mainPyPath, cfPyData, false)
	checkAndWriteFile(cfLogicPyPath, cfLogicPyData, false)

	requirementsTxtData = nil
	cfPyData = nil
	cfLogicPyData = nil
}

var (
	ErrPyExitCode       = errors.New("python script exited with non-zero exit code")
	ErrVenvDoesNotExist = fmt.Errorf("venv does not exist at %s", getVenvDirPath())
)

func CallScript(ctx context.Context, timeout float64, args CfArgs) (string, error) {
	cfDirPath := getCfDirPath()
	mainPyPath := filepath.Join(cfDirPath, "cf.py")

	venvPath := getVenvDirPath()
	if !iofuncs.PathExists(venvPath) {
		return "", ErrVenvDoesNotExist
	}

	cmd := exec.CommandContext(ctx, filepath.Join(venvPath, "Scripts", "python"), mainPyPath)
	cmd.Args = append(cmd.Args, args.ParseCmdArgs()...)
	utils.PrepareCmdForBgTask(cmd)

	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Since the error should already be logged
			// by the Python script, we just return the error here.
			return "", ErrPyExitCode
		}
		if errors.Is(err, context.DeadlineExceeded) {
			if sigErr := utils.InterruptProcess(cmd); sigErr != nil {
				logger.LogError(
					fmt.Errorf(
						"error %d: failed to send SIGINT to process, more info => %w", 
						cdlerrors.OS_ERROR, 
						sigErr,
					), 
					logger.ERROR,
				)
			}
		}
		return "", err
	}

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	cookieFilePath := strings.TrimPrefix(string(stdout), "cookies saved to ")
	return cookieFilePath, nil
}
