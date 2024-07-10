package cdlerrors

import (
	"errors"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

var (
	ErrRecaptcha = errors.New(constants.ERR_RECAPTCHA_STR)
	ErrSkipLine  = errors.New("skip line")

	ErrPyExitCode     = errors.New("python script exited with non-zero exit code")
	ErrChromeNotFound = errors.New("could not find google chrome executable/binary path")
)
