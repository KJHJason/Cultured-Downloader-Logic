package cf

import (
	"runtime"
	"strconv"

	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

type CfArgs struct {
	Attempts       int
	BrowserPath    string
	Headless       bool
	VirtualDisplay bool
	TargetUrl      string
}

func (args CfArgs) ParseCmdArgs() []string {
	cmdArgs := []string{
		"--user-agent", httpfuncs.DEFAULT_USER_AGENT,
	}

	if args.Attempts > 0 {
		cmdArgs = append(cmdArgs, "--attempts", strconv.Itoa(args.Attempts))
	}

	if args.BrowserPath != "" {
		cmdArgs = append(cmdArgs, "--browser-path", args.BrowserPath)
	}

	if args.Headless {
		cmdArgs = append(cmdArgs, "--headless", "1")
	} else {
		cmdArgs = append(cmdArgs, "--headless", "0")
	}

	if args.VirtualDisplay {
		cmdArgs = append(cmdArgs, "--virtual-display")
	}

	if args.TargetUrl != "" {
		cmdArgs = append(cmdArgs, "--target-url", args.TargetUrl)
	}
	return cmdArgs
}

func AddDefaultLogPath(cmdArgs []string) []string {
	return append(cmdArgs, "--log-path", logger.CdlCfLogFilePath)
}

func NewCfArgs(url string) CfArgs {
	cfArgs := CfArgs{}
	cfArgs.Default()
	cfArgs.SetTargetUrl(url)
	return cfArgs
}

func (args *CfArgs) Default() {
	args.Attempts = 4
	args.VirtualDisplay = runtime.GOOS == "linux"
	args.Headless = !args.VirtualDisplay

	browserPath, err := utils.GetChromeExecPath()
	if err != nil {
		// shouldn't happen as the error check
		// should be done at the start of the program.
		panic(err)
	}
	args.BrowserPath = browserPath
}

func (args *CfArgs) SetTargetUrl(url string) {
	args.TargetUrl = url
}
