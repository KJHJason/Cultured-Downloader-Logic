package cdldocker

type dockerFantiaArgs struct {
	common     CdlSolversArgs
	cookiePath string
}

func (args dockerFantiaArgs) parseCmdArgs() []string {
	cmdArgs := args.common.parseCmdArgs()
	if args.cookiePath != "" {
		cmdArgs = append(cmdArgs, "--cookie-path", args.cookiePath)
	}
	return cmdArgs
}

func newFantiaArgs(userAgent, logFilePath, cookiePath string) dockerFantiaArgs {
	fantiaArgs := dockerFantiaArgs{
		common: CdlSolversArgs{
			attempts:  4,
			logPath:   logFilePath,
			subcmd:    Fantia,
			userAgent: userAgent,
		},
		cookiePath: cookiePath,
	}
	return fantiaArgs
}
