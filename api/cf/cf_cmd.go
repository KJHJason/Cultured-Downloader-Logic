package cf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

func getPyVenvBinDirName() string {
	switch runtime.GOOS {
	case "windows":
		return "Scripts"
	default:
		return "bin"
	}
}

func TestScript() error {
	cfPyPath := getMainPyPath()
	venvPath := getVenvDirPath()
	if !iofuncs.PathExists(venvPath) {
		return cdlerrors.ErrVenvDoesNotExist
	}

	chromePath, err := utils.GetChromeExecPath()
	if err != nil {
		// chrome exec path check should have been done
		// at the start of the program. Hence, the panic here.
		panic(err)
	}

	cmd := exec.Command(
		filepath.Join(venvPath, getPyVenvBinDirName(), "python"), 
		cfPyPath, 
		"--test-connection", 
		"--headless=true",
		"--browser-path", chromePath, 
		"--log-path", logger.CdlCfLogFilePath,
	)
	utils.PrepareCmdForBgTask(cmd)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func CallScript(args *CfArgs) (Cookies, error) {
	if args == nil {
		return nil, fmt.Errorf("error %d: args is nil", cdlerrors.UNEXPECTED_ERROR)
	}

	cfPyPath := getMainPyPath()
	venvPath := getVenvDirPath()
	if !iofuncs.PathExists(venvPath) {
		return nil, cdlerrors.ErrVenvDoesNotExist
	}

	parsedCfArgs := args.ParseCmdArgs()
	cmdArgs := make([]string, 0, len(parsedCfArgs)+1)
	cmdArgs = append(cmdArgs, cfPyPath)
	cmdArgs = append(cmdArgs, parsedCfArgs...)

	cmd := exec.Command(filepath.Join(venvPath, getPyVenvBinDirName(), "python"), cmdArgs...)
	utils.PrepareCmdForBgTask(cmd)
	stdout, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Since the error should already be logged
			// by the Python script, we just return the error here.
			return nil, cdlerrors.ErrPyExitCode
		}
		return nil, err
	}

	cookieFilePath := strings.TrimPrefix(string(stdout), "cookies saved to ")
	cookieFilePath = strings.TrimSpace(cookieFilePath)
	cookieFilePath = filepath.Clean(cookieFilePath)
	if !iofuncs.PathExists(cookieFilePath) {
		return nil, fmt.Errorf("error %d: cookie file not found at %s", cdlerrors.UNEXPECTED_ERROR, cookieFilePath)
	}
	defer os.Remove(cookieFilePath)

	cookies, err := parseCookies(cookieFilePath)
	if err != nil {
		return nil, err
	}
	return cookies, nil
}
