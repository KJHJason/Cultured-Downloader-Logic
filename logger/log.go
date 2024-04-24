package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
	"github.com/KJHJason/Cultured-Downloader-Logic/iofuncs"
)

const LogSuffix = "\n\n"
var (
	MainLogger *Logger
	logFolder = filepath.Join(iofuncs.APP_PATH, "logs")
	logFilePath = filepath.Join(
		logFolder,
		fmt.Sprintf(
			"cultured_downloader-logic_v%s_%s.log", 
			constants.VERSION, 
			time.Now().Format("2006-01-02"),
		),
	)
)

func init() {
	// create the logs directory if it does not exist
	os.MkdirAll(logFolder, 0755)

	// will be opened througout the program's runtime
	// hence, there is no need to call f.Close() at the end of this function
	f, fileErr := os.OpenFile(
		logFilePath, 
		os.O_WRONLY|os.O_CREATE|os.O_APPEND, 
		0666,
	)
	if fileErr != nil {
		fileErr = fmt.Errorf(
			"error opening log file: %v\nlog file path: %s", 
			fileErr, 
			logFilePath,
		)
		panic(fileErr)
	}
	MainLogger = NewLogger(f)
}

// Delete all empty log files and log files
// older than 30 days except for the current day's log file.
func DeleteEmptyAndOldLogs() error {
	return filepath.Walk(logFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || path == logFilePath {
			return nil
		}

		if info.Size() == 0 || info.ModTime().Before(time.Now().AddDate(0, 0, -30)) {
			return os.Remove(path)
		}

		return nil
	})
}

// Thread-safe logging function that logs to "cultured_downloader.log" in the logs directory
func LogError(err error, exit bool, level int) {
	if err == nil {
		return
	}

	MainLogger.LogBasedOnLvl(level, err.Error() + LogSuffix)
	if exit {
		os.Exit(1)
	}
}

// Uses the thread-safe LogError() function to log multiple errors
//
// Also returns if any errors were due to context.Canceled which is caused by Ctrl + C.
func LogErrors(exit bool, level int, errs ...error) bool {
	var hasCanceled bool
	for _, err := range errs {
		if err == context.Canceled {
			if !hasCanceled {
				hasCanceled = true
			}
			continue
		}
		LogError(err, exit, level)
	}
	return hasCanceled
}

// Uses the thread-safe LogError() function to log a channel of errors
//
// Also returns if any errors were due to context.Canceled which is caused by Ctrl + C.
func LogChanErrors(exit bool, level int, errChan chan error) bool {
	var hasCanceled bool
	for err := range errChan {
		if err == context.Canceled {
			if !hasCanceled {
				hasCanceled = true
			}
			continue
		}
		LogError(err, exit, level)
	}
	return hasCanceled
}

var logToPathMux sync.Mutex

// Thread-safe logging function that logs to the provided file path
func LogMessageToPath(message, filePath string, level int) {
	logToPathMux.Lock()
	defer logToPathMux.Unlock()

	os.MkdirAll(filepath.Dir(filePath), 0755)
	if iofuncs.PathExists(filePath) {
		logFileContents, err := os.ReadFile(filePath)
		if err != nil {
			err = fmt.Errorf(
				"error %d: failed to read log file, more info => %w\nfile path: %s\noriginal message: %s",
				errs.OS_ERROR,
				err,
				filePath,
				message,
			)
			LogError(err, false, ERROR)
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
		0666,
	)
	if err != nil {
		err = fmt.Errorf(
			"error %d: failed to open log file, more info => %w\nfile path: %s\noriginal message: %s",
			errs.OS_ERROR,
			err,
			filePath,
			message,
		)
		LogError(err, false, ERROR)
		return
	}
	defer logFile.Close()

	pathLogger := NewLogger(logFile)
	pathLogger.LogBasedOnLvl(level, message)
}
