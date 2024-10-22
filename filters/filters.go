package filters

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/logger"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils"
)

const (
	NO_MAX_FILESIZE = -1
)

type filtersDateInfo struct {
	hasStartDate bool
	hasEndDate   bool
}

// Note: ValidateArgs must be called
// after initialising the Filters struct.
type Filters struct {
	MinFileSize int64 // In bytes
	MaxFileSize int64 // In bytes

	FileExt []string

	StartDate time.Time
	EndDate   time.Time
	dateInfo  *filtersDateInfo

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

	now := time.Now()
	hasStartDate := f.StartDate.IsZero()
	hasEndDate := f.EndDate.IsZero()
	if hasStartDate && hasEndDate && f.StartDate.After(f.EndDate) {
		return errors.New("start date cannot be after end date")
	}
	if hasStartDate && f.StartDate.After(now) {
		return errors.New("start date cannot be newer than today's date")
	}
	if hasEndDate && f.EndDate.After(now) {
		return errors.New("end date cannot be newer than today's date")
	}
	f.dateInfo = &filtersDateInfo{
		hasStartDate: hasStartDate,
		hasEndDate:   hasEndDate,
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
	if f.dateInfo == nil {
		logger.MainLogger.Fatalf(
			"error %d: Filters struct was not initialised properly, did you forget to call ValidateArgs()?",
			cdlerrors.DEV_ERROR,
		)
	}

	if postDate.IsZero() {
		return true // if the fileDate is invalid, fallback to true/ignore the date filter
	}

	// No date given for filtering, return true for valid.
	if f.StartDate.IsZero() && f.EndDate.IsZero() {
		return true
	}

	// Check if the user has provided a starting date.
	// If provided, check if the post date is after the starting date
	if f.dateInfo.hasStartDate && !postDate.After(f.StartDate) {
		return false
	}

	// After checking the starting date, if the end date is not given, it is valid.
	// Otherwise, check if the post date is before the given end date.
	return !f.dateInfo.hasEndDate || postDate.Before(f.EndDate)
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
