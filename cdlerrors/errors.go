package cdlerrors

import (
	"errors"
)

var (
	ErrSkipLine          = errors.New("skip line")
	ErrCaptchaPrevFailed = errors.New("captcha has failed previously, skipping attempt of trying to bypass captcha again")
)
