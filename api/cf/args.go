package cf

import (
	"strconv"

	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

type CfArgs struct {
	Version     bool
	Attempts    int
	BrowserPath string
	Headless    bool
	TargetUrl   string
	UserAgent   string
}

func (args CfArgs) ParseCmdArgs() []string {
	cmdArgs := []string{
		"--log-path", logger.CfPyLogFilePath,
	}

	if args.Version {
		cmdArgs = append(cmdArgs, "-v")
	}
	if args.Attempts > 0 {
		cmdArgs = append(cmdArgs, "--attempts", strconv.Itoa(args.Attempts))
	}
	if args.BrowserPath != "" {
		cmdArgs = append(cmdArgs, "--browser-path", args.BrowserPath)
	}
	if args.Headless {
		cmdArgs = append(cmdArgs, "--headless")
	}
	if args.TargetUrl != "" {
		cmdArgs = append(cmdArgs, "--target-url", args.TargetUrl)
	}
	if args.UserAgent != "" {
		cmdArgs = append(cmdArgs, "-ua", args.UserAgent)
	}
	return cmdArgs
}

func NewCfArgs(url string) *CfArgs {
	cfArgs := CfArgs{}
	cfArgs.Default()
	cfArgs.SetTargetUrl(url)
	return &cfArgs
}

func (args *CfArgs) Default() {
	args.Version = false
	args.Attempts = 4
	args.Headless = true
	args.UserAgent = httpfuncs.DEFAULT_USER_AGENT

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
