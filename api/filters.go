package api

import (
	"errors"
	"regexp"
	"time"
)

type Filters struct {
	MinFileSize int64
	MaxFileSize int64

	FileExt []string

	MinDate time.Time
	MaxDate time.Time

	FileNameFilter *regexp.Regexp
}

func (f *Filters) ValidateArgs() error {
	if f.MinFileSize < 0 || f.MaxFileSize < 0 {
		return errors.New("min and max file size must be greater than or equal to 0")
	}

	if f.MinFileSize > f.MaxFileSize {
		return errors.New("min file size cannot be greater than max file size")
	}

	if f.MinDate.After(f.MaxDate) {
		return errors.New("min date cannot be after max date")
	}

	return nil
}
