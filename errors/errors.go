package cdlerrors

import (
	"errors"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
)

var (
	ErrRecaptcha = errors.New(constants.ERR_RECAPTCHA_STR)
	ErrSkipLine  = errors.New("skip line")
)
