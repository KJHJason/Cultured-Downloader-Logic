package cf

import (
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
)

type CfArgs struct {
	Version     bool
	BrowserPath string
	Headless    bool
	TargetUrl   string
	UserAgent   string
}

func (args CfArgs) ParseCmdArgs() []string {
	cmdArgs := make([]string, 0, 6)
	cmdArgs = append(cmdArgs, "--log-path", logger.CfPyLogFilePath)

	if args.Version {
		cmdArgs = append(cmdArgs, "-v")
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
