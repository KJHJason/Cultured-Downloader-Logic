package startup

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cf"
	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

func checkDockerDaemonIsRunning() bool {
	cmd := exec.Command("docker", "version")
	utils.PrepareCmdForBgTask(cmd)
	return cmd.Run() == nil
}

func checkDockerRequirements(ctx context.Context, panicHandler func(msg string)) {
	if _, err := exec.LookPath("docker"); err != nil {
		panicHandler(
			fmt.Sprintf(
				"error %d: Docker executable not found, please install Docker and ensure it's in the PATH environment variable",
				cdlerrors.STARTUP_ERROR,
			),
		)
	}
	if !checkDockerDaemonIsRunning() {
		panicHandler(
			fmt.Sprintf(
				"error %d: Docker is not running, please start Docker daemon",
				cdlerrors.STARTUP_ERROR,
			),
		)
	}

	if err := cf.PullCfDockerImage(ctx); err != nil {
		panicHandler(
			fmt.Sprintf(
				"error %d: failed to pull Docker image -> %v",
				cdlerrors.STARTUP_ERROR,
				err,
			),
		)
	}
}

func CheckPrerequisites(ctx context.Context, panicHandler func(msg string)) {
	if _, err := utils.GetChromeExecPath(); err != nil {
		panicHandler(
			fmt.Sprintf(
				"error %d: Google Chrome executable not found, please install Google Chrome or set the CHROME_EXECUTABLE environment variable",
				cdlerrors.STARTUP_ERROR,
			),
		)
	}
	checkDockerRequirements(ctx, panicHandler)
}
