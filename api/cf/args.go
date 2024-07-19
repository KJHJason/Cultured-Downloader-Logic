package cf

import (
	"runtime"
	"strconv"

	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

type dockerCfArgs struct {
	attempts   int
	targetUrl  string
	cookiePath string
	logPath    string
}

func (args dockerCfArgs) parseCmdArgs() []string {
	cmdArgs := []string{
		"--user-agent", httpfuncs.DEFAULT_USER_AGENT,
		"--os-name", runtime.GOOS,
		"--virtual-display",
		// yes, it is hardcoded mainly to make the docker
		// image harder to run for people without dev knowledge
		"--app-key", "fzN9Hvkb9s+mwPGCDd5YFnLiqKx8WhZfWoZE5nZC",
	}

	if args.cookiePath != "" {
		cmdArgs = append(cmdArgs, "--cookie-path", args.cookiePath)
	}
	if args.logPath != "" {
		cmdArgs = append(cmdArgs, "--log-path", args.logPath)
	}
	if args.attempts > 0 {
		cmdArgs = append(cmdArgs, "--attempts", strconv.Itoa(args.attempts))
	}
	if args.targetUrl != "" {
		cmdArgs = append(cmdArgs, "--target-url", args.targetUrl)
	}
	return cmdArgs
}

func newCfArgs(url, cookieFilePath, logFilePath string) dockerCfArgs {
	cfArgs := dockerCfArgs{
		attempts:   4,
		targetUrl:  url,
		cookiePath: cookieFilePath,
		logPath:    logFilePath,
	}
	return cfArgs
}
