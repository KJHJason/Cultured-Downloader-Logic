package cdldocker

type dockerCfArgs struct {
	common     CdlSolversArgs
	targetUrl  string
	cookiePath string
}

func (args dockerCfArgs) parseCmdArgs() []string {
	cmdArgs := args.common.parseCmdArgs()
	if args.cookiePath != "" {
		cmdArgs = append(cmdArgs, "--cookie-path", args.cookiePath)
	}
	if args.targetUrl != "" {
		cmdArgs = append(cmdArgs, "--target-url", args.targetUrl)
	}
	return cmdArgs
}

func newCfArgs(targetUrl, cookieFilePath, userAgent, logFilePath string) dockerCfArgs {
	cfArgs := dockerCfArgs{
		common: CdlSolversArgs{
			attempts:  4,
			logPath:   logFilePath,
			subcmd:    Cf,
			userAgent: userAgent,
		},
		targetUrl:  targetUrl,
		cookiePath: cookieFilePath,
	}
	return cfArgs
}
