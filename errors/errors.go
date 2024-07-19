package cdlerrors

import (
	"errors"
)

var (
	ErrSkipLine          = errors.New("skip line")
	ErrChromeNotFound    = errors.New("could not find google chrome executable/binary path")
	ErrCaptchaPrevFailed = errors.New("captcha has failed previously, skipping attempt of trying to bypass captcha again")
)
