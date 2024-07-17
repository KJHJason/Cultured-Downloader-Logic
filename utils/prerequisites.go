package utils

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

var UseDockerForCf bool

const dockerCfEnvKey = "CDL_CF_USE_DOCKER"

func init() {
	if runtime.GOOS != "linux" {
		UseDockerForCf = true
		return
	}

	useDockerArg := os.Getenv(dockerCfEnvKey)
	UseDockerForCf = useDockerArg == "1" || useDockerArg == "true"
}

func CheckIsArm() bool {
	return runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" ||
		runtime.GOARCH == "arm64be" || runtime.GOARCH == "armbe"
}

func isPythonVersionAtLeast(output string, major int, minor int, patch int) (bool, error) {
	ver := strings.TrimPrefix(output, "Python ")
	verParts := strings.Split(ver, ".")

	if len(verParts) < 3 {
		return false, fmt.Errorf(
			"error %d: python version %q is not in the format of 'major.minor.patch'",
			cdlerrors.UNEXPECTED_ERROR,
			output,
		)
	}

	majorVer, err := strconv.Atoi(verParts[0])
	if err != nil {
		panic(err) // shouldn't happen
	}
	if majorVer < major {
		return false, nil
	}

	minorVer, err := strconv.Atoi(verParts[1])
	if err != nil {
		panic(err) // shouldn't happen
	}
	if minorVer < minor {
		return false, nil
	}

	patchVer, err := strconv.Atoi(verParts[2])
	if err != nil {
		panic(err) // shouldn't happen
	}
	return patchVer >= patch, nil
}

func checkPythonVersion(major int, minor int, patch int) (bool, error) {
	cmd := exec.Command("python", "--version")
	PrepareCmdForBgTask(cmd)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return isPythonVersionAtLeast(string(output), major, minor, patch)
}

func checkPythonRequirements(panicHandler func(msg string)) {
	if _, err := exec.LookPath("python"); err != nil {
		var errMsg string
		if runtime.GOOS == "linux" {
			errMsg = fmt.Sprintf(
				"error %d: Python executable not found, please install Python and ensure that it can ran by calling 'python' in the terminal (not 'python3')",
				cdlerrors.STARTUP_ERROR,
			)
		} else {
			errMsg = fmt.Sprintf(
				"error %d: Python executable not found, please install Python and ensure it's in the PATH environment variable",
				cdlerrors.STARTUP_ERROR,
			)
		}
		panicHandler(errMsg)
	}

	const (
		major = 3
		minor = 10
		patch = 0
	)
	if ok, err := checkPythonVersion(major, minor, patch); err != nil {
		panicHandler(
			fmt.Sprintf(
				"error %d: could not check Python version, more info => %v",
				cdlerrors.STARTUP_ERROR,
				err,
			),
		)
	} else if !ok {
		pythonVer := fmt.Sprintf("%d.%d.%d", major, minor, patch)
		panicHandler(
			fmt.Sprintf(
				"error %d: Python version %s or higher is required, please install Python %s or higher",
				cdlerrors.STARTUP_ERROR,
				pythonVer, pythonVer,
			),
		)
	}
}

func checkXvfbExec(dockerEnvKey string, panicHandler func(msg string)) {
	_, err := exec.LookPath("Xvfb")
	if err != nil {
		panicHandler(
			fmt.Sprintf(
				"error %d: Xvfb executable not found, please install Xvfb or set the %s environment variable to true to use a Docker image instead",
				cdlerrors.STARTUP_ERROR,
				dockerEnvKey,
			),
		)
	}
}

func checkDockerDaemonIsRunning() bool {
	cmd := exec.Command("docker", "version")
	PrepareCmdForBgTask(cmd)
	return cmd.Run() == nil
}

func checkDockerRequirements(panicHandler func(msg string)) {
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
}

func CheckPrerequisites(panicHandler func(msg string)) {
	if _, err := GetChromeExecPath(); err != nil {
		panicHandler(
			fmt.Sprintf(
				"error %d: Google Chrome executable not found, please install Google Chrome or set the CHROME_EXECUTABLE environment variable",
				cdlerrors.STARTUP_ERROR,
			),
		)
	}

	if !UseDockerForCf {
		checkXvfbExec(dockerCfEnvKey, panicHandler)
		checkPythonRequirements(panicHandler)
	} else {
		checkDockerRequirements(panicHandler)
	}
}
