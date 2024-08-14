package startup

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/KJHJason/Cultured-Downloader-Logic/api/cdlsolvers/cdldocker"
	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

func checkDockerDaemonIsRunning() bool {
	cmd := exec.Command("docker", "version")
	utils.PrepareCmdForBgTask(cmd)
	return cmd.Run() == nil
}

func checkDockerRequirements(ctx context.Context, infoHandler func(msg string)) {
	msgPrefix := fmt.Sprintf(
		"Features like the Captcha Solver will not work properly due to error %d",
		cdlerrors.STARTUP_ERROR,
	)
	if _, err := exec.LookPath("docker"); err != nil {
		infoHandler(
			fmt.Sprintf(
				"%s: Docker executable not found, please install Docker and ensure it's in the PATH environment variable",
				msgPrefix,
			),
		)
		return
	}
	if !checkDockerDaemonIsRunning() {
		infoHandler(
			fmt.Sprintf(
				"%s: Docker is not running, please start the Docker daemon",
				msgPrefix,
			),
		)
		return
	}

	if err := cdldocker.PullDockerImage(ctx); err != nil {
		infoHandler(
			fmt.Sprintf(
				"%s: failed to pull Docker image -> %v",
				msgPrefix,
				err,
			),
		)
		return
	}
}

func CheckPrerequisites(ctx context.Context, infoHandler func(msg string), panicHandler func(msg string)) {
	if _, err := utils.GetChromeExecPath(); err != nil {
		panicHandler(
			fmt.Sprintf(
				"error %d: Google Chrome executable not found, please install Google Chrome or set the CHROME_EXECUTABLE environment variable",
				cdlerrors.STARTUP_ERROR,
			),
		)
	}
	checkDockerRequirements(ctx, infoHandler)
}
