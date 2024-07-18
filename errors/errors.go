package cdlerrors

import (
	"errors"
)

var (
	ErrSkipLine = errors.New("skip line")

	ErrPyExitCode       = errors.New("python script exited with non-zero exit code")
	ErrChromeNotFound   = errors.New("could not find google chrome executable/binary path")
	ErrVenvDoesNotExist = errors.New("venv does not exist")
)
