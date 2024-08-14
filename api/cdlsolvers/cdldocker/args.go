package cdldocker

import (
	"strconv"

	"github.com/KJHJason/Cultured-Downloader-Logic/httpfuncs"
)

type SubCommand int

const (
	Cf SubCommand = iota
	Fantia
)

type CdlSolversArgs struct {
	attempts  int
	logPath   string
	subcmd    SubCommand
	userAgent string
}

func (args CdlSolversArgs) parseCmdArgs() []string {
	userAgent := args.userAgent
	if userAgent == "" {
		userAgent = httpfuncs.DEFAULT_USER_AGENT
	}

	cmdArgs := []string{
		"--headless", "0",
		"--virtual-display",
		"--user-agent", userAgent,
		"--adv-feature",
	}
	switch args.subcmd {
	case Cf:
		cmdArgs = append(cmdArgs, "1HrEnmTKajwzUBwPkeJXtTezuX8jsUhWQKasNb9Z9ShNB6nBJScvMGhG6ujtbfhC")
	case Fantia:
		cmdArgs = append(cmdArgs, "dNf9fgd6GabVZdFGERMZUzHGt8QkurCmdZ0G5KsvnCaBAn3PmAWakgaFE7VDAxgs")
	default:
		panic("Invalid subcommand")
	}

	if args.logPath != "" {
		cmdArgs = append(cmdArgs, "--log-path", args.logPath)
	}
	if args.attempts > 0 {
		cmdArgs = append(cmdArgs, "--attempts", strconv.Itoa(args.attempts))
	}
	return cmdArgs
}
