package cf

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

var (
	ErrVenvDoesNotExist = fmt.Errorf("venv does not exist at %s", getVenvDirPath())
)

func CallScript(args *CfArgs) (Cookies, error) {
	if args == nil {
		return nil, fmt.Errorf("error %d: args is nil", cdlerrors.UNEXPECTED_ERROR)
	}

	cfDirPath := getCfDirPath()
	mainPyPath := filepath.Join(cfDirPath, "cf.py")

	venvPath := getVenvDirPath()
	if !iofuncs.PathExists(venvPath) {
		return nil, ErrVenvDoesNotExist
	}

	parsedCfArgs := args.ParseCmdArgs()
	cmdArgs := make([]string, 0, len(parsedCfArgs)+1)
	cmdArgs = append(cmdArgs, mainPyPath)
	cmdArgs = append(cmdArgs, parsedCfArgs...)

	cmd := exec.Command(filepath.Join(venvPath, "Scripts", "python"), cmdArgs...)
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
