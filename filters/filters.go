package filters

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Filters struct {
	MinFileSize int64
	MaxFileSize int64

	FileExt []string

	StartDate time.Time
	EndDate   time.Time

	FileNameFilter *regexp.Regexp
}

func (f *Filters) ValidateArgs() error {
	if f.MinFileSize < 0 || f.MaxFileSize < 0 {
		return errors.New("min and max file size cannot be negative")
	}

	if f.MinFileSize > f.MaxFileSize {
		return errors.New("min file size cannot be greater than max file size")
	}

	if f.StartDate.After(f.EndDate) {
		return errors.New("start date cannot be after end date")
	}

	for idx, ext := range f.FileExt {
		if !strings.HasSuffix(ext, ".") {
			return errors.New("file extension must start with a period")
		}

		ext = strings.TrimSpace(ext)
		if ext == "" {
			return errors.New("file extension cannot be empty")
		}
		f.FileExt[idx] = ext
	}

	return nil
}

func (f *Filters) IsFileSizeInRange(fileSize int64) bool {
	if f.MinFileSize == 0 && f.MaxFileSize == 0 {
		return true
	}
	return fileSize >= f.MinFileSize && fileSize <= f.MaxFileSize
}

func (f *Filters) IsFileExtValid(fileExt string) bool {
	if len(f.FileExt) == 0 {
		return true
	}
	for _, ext := range f.FileExt {
		if ext == fileExt {
			return true
		}
	}
	return false
}

func (f *Filters) IsFilePathExtValid(filePath string) bool {
	return f.IsFileExtValid(filepath.Ext(filePath))
}

func (f *Filters) IsPostDateValid(postDate time.Time) bool {
	if postDate.IsZero() {
		return true // if the fileDate is invalid, fallback to true/ignore the date filter
	}
	if f.StartDate.IsZero() && f.EndDate.IsZero() {
		return true
	}
	return postDate.After(f.StartDate) && postDate.Before(f.EndDate)
}

func (f *Filters) IsFileNameValid(fileName string) bool {
	if f.FileNameFilter == nil {
		return true
	}
	return f.FileNameFilter.MatchString(fileName)
}

func (f *Filters) IsFilePathFileNameValid(filePath string) bool {
	return f.IsFileNameValid(filepath.Base(filePath))
}
