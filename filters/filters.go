package filters

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

const (
	NO_MAX_FILESIZE = -1
)

type Filters struct {
	MinFileSize int64 // In bytes
	MaxFileSize int64 // In bytes

	FileExt []string

	StartDate time.Time
	EndDate   time.Time

	FileNameFilter *regexp.Regexp
}

func (f *Filters) RemoveDuplicateFileExt() {
	f.FileExt = utils.RemoveDuplicatesFromSlice(f.FileExt)
}

func (f *Filters) ConvertFileSizeFromMB() {
	if f.MinFileSize == 0 {
		return
	}

	const mb = 1024 * 1024
	f.MinFileSize = f.MinFileSize * mb

	if f.MaxFileSize != NO_MAX_FILESIZE {
		f.MaxFileSize = f.MaxFileSize * mb
	}
}

func (f *Filters) Copy() *Filters {
	return &Filters{
		MinFileSize:    f.MinFileSize,
		MaxFileSize:    f.MaxFileSize,
		FileExt:        append([]string{}, f.FileExt...),
		StartDate:      f.StartDate,
		EndDate:        f.EndDate,
		FileNameFilter: f.FileNameFilter,
	}
}

func (f *Filters) ValidateArgs() error {
	hasMaxFileSize := f.MaxFileSize != NO_MAX_FILESIZE
	if f.MinFileSize < 0 || (hasMaxFileSize && f.MaxFileSize < 0) {
		return errors.New("min and max file size cannot be negative")
	}

	if hasMaxFileSize && f.MinFileSize > f.MaxFileSize {
		return errors.New("min file size cannot be greater than max file size")
	}

	noStartDate := f.StartDate.IsZero()
	noEndDate := f.EndDate.IsZero()
	if !noStartDate && !noEndDate && f.StartDate.After(f.EndDate) {
		return errors.New("start date cannot be after end date")
	} else if noStartDate != noEndDate { // same as XOR
		return errors.New("both start and end date must be set or unset")
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

// Note: In bytes
func (f *Filters) IsFileSizeInRange(fileSize int64) bool {
	if f.MinFileSize == 0 && f.MaxFileSize == NO_MAX_FILESIZE {
		return true
	}
	return fileSize >= f.MinFileSize && fileSize <= f.MaxFileSize
}

// Note: fileExt should start with a period/dot
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
