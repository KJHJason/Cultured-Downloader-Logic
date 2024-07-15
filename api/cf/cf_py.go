package cf

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

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
