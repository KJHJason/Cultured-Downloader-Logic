package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/cdlerrors"
	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
	"github.com/KJHJason/Cultured-Downloader-Logic/utils/threadsafe"
)

const (
	LOG_SUFFIX    = "\n\n"
	LOG_PERMS     = 0644 // rw-r--r--
	LOG_THRESHOLD = 15 * 24 * time.Hour
)

var (
	MainLogger Logger

	logFolder        = filepath.Join(iofuncs.APP_PATH, "logs")
	logFilePath      = filepath.Join(logFolder, getLogFileName())
	CdlCfLogFilePath = filepath.Join(logFolder, getCdlCfLogFileName())
)

func getFilenameDateFmt() string {
	return time.Now().Format("2006-01-02")
}

func getLogFileName() string {
	return fmt.Sprintf(
		"cultured-downloader-logic-v%s_%s.log",
		constants.VERSION,
		getFilenameDateFmt(),
	)
}

func getCdlCfLogFileName() string {
	return fmt.Sprintf(
		"kjhjason-cdl-cf_%s.log",
		getFilenameDateFmt(),
	)
}

func init() {
	// create the logs directory if it does not exist
	os.MkdirAll(logFolder, LOG_PERMS)

	// will be opened throughout the program's runtime
	// hence, there is no need to call f.Close() at the end of this function
	logFlags := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	f, fileErr := os.OpenFile(
		logFilePath,
		logFlags,
		LOG_PERMS,
	)
	if fileErr == nil {
		MainLogger = NewLogger(f)
	} else {
		fileErr = fmt.Errorf(
			"error opening log file: %w\nlog file path: %s",
			fileErr,
			logFilePath,
		)

		// fallback to cwd if the logs directory cannot be created
		var fallbackFileErr error
		f, fallbackFileErr = os.OpenFile(
			getLogFileName(),
			logFlags,
			LOG_PERMS,
		)
		if fallbackFileErr != nil {
			panic(fileErr)
		}

		MainLogger = NewLogger(f)
		LogError(fileErr, ERROR)
	}
	DeleteEmptyAndOldLogs()
}

func DeleteLogsOnCond(condToSkip func(os.FileInfo) bool) error {
	return filepath.Walk(logFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || path == logFilePath || (condToSkip != nil && condToSkip(info)) {
			return nil
		}

		return os.Remove(path)
	})
}

// Delete all empty log files and log files
// older than the log threshold except for the current day's log file.
func DeleteEmptyAndOldLogs() error {
	return DeleteLogsOnCond(func(info os.FileInfo) bool {
		isNewerThan7Days := time.Since(info.ModTime()) < LOG_THRESHOLD
		return isNewerThan7Days
	})
}

// Thread-safe logging function that logs to "cultured_downloader.log" in the logs directory
func LogError(err error, level int) {
	if err == nil {
		return
	}

	MainLogger.LogBasedOnLvl(level, err.Error()+LOG_SUFFIX)
}

// Uses the thread-safe LogError() function to log multiple errors
//
// Also returns if any errors were due to context.Canceled which is caused by Ctrl + C.
func LogErrors(level int, errs ...error) bool {
	var hasCanceled bool
	for _, err := range errs {
		if errors.Is(err, context.Canceled) {
			if !hasCanceled {
				hasCanceled = true
			}
			continue
		}
		LogError(err, level)
	}
	return hasCanceled
}

// Uses the thread-safe LogError() function to log a slice of errors
//
// Note that the thread-safe slice will be cleared after logging using `ClearUnsafe`.
//
// Also returns if any errors were due to context.Canceled which is caused by Ctrl + C.
func LogSliceErrors(level int, tsErrSlice *threadsafe.Slice[error]) (bool, []error) {
	var hasCanceled bool

	errSlice := make([]error, 0, tsErrSlice.Len())
	it := tsErrSlice.NewIter()
	for it.Next() {
		err := it.Item()
		if errors.Is(err, context.Canceled) {
			if !hasCanceled {
				hasCanceled = true
			}
			continue
		}
		LogError(err, level)
		errSlice = append(errSlice, err)
	}
	tsErrSlice.ClearUnsafe()
	return hasCanceled, errSlice
}

var logToPathMux sync.Mutex

// Thread-safe logging function that logs to the provided file path
func LogMessageToPath(message, filePath string, level int) {
	logToPathMux.Lock()
	defer logToPathMux.Unlock()

	os.MkdirAll(filepath.Dir(filePath), LOG_PERMS)
	if iofuncs.PathExists(filePath) {
		logFileContents, err := os.ReadFile(filePath)
		if err != nil {
			err = fmt.Errorf(
				"error %d: failed to read log file, more info => %w\nfile path: %s\noriginal message: %s",
				cdlerrors.OS_ERROR,
				err,
				filePath,
				message,
			)
			LogError(err, ERROR)
			return
		}

		// check if the same message has already been logged
		if strings.Contains(string(logFileContents), message) {
			return
		}
	}

	logFile, err := os.OpenFile(
		filePath,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		LOG_PERMS,
	)
	if err != nil {
		err = fmt.Errorf(
			"error %d: failed to open log file, more info => %w\nfile path: %s\noriginal message: %s",
			cdlerrors.OS_ERROR,
			err,
			filePath,
			message,
		)
		LogError(err, ERROR)
		return
	}
	defer logFile.Close()

	pathLogger := NewLogger(logFile)
	pathLogger.LogBasedOnLvl(level, message)
}
