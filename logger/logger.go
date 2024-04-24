package logger

import (
	"fmt"
	"log"
	"io"
	"os"

	"github.com/KJHJason/Cultured-Downloader-Logic/constants"
	"github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

const (
	// Log levels
	INFO = iota
	ERROR
	DEBUG
)

type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

var loggerPrefix = fmt.Sprintf("Cultured Downloader Logic V%s ", constants.VERSION)
func NewLogger(out io.Writer) *Logger {
	if out == nil {
		out = os.Stdout
	}

	return &Logger{
		infoLogger:  log.New(out, loggerPrefix + "[INFO]: ", log.Ldate|log.Ltime),
		errorLogger: log.New(out, loggerPrefix + "[ERROR]: ", log.Ldate|log.Ltime),
		debugLogger: log.New(out, loggerPrefix + "[DEBUG]: ", log.Ldate|log.Ltime),
	}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.infoLogger.SetOutput(w)
	l.errorLogger.SetOutput(w)
	l.debugLogger.SetOutput(w)
}

// LogBasedOnLvlf logs a message based on the log level passed in
//
// You can use this function to log a message with a format string
//
// However, please ensure that the 
// lvl passed in is valid (i.e. INFO, ERROR, or DEBUG), otherwise this function will panic
func (l *Logger) LogBasedOnLvlf(lvl int, format string, args ...any) {
	switch lvl {
	case INFO:
		l.Infof(format, args...)
	case ERROR:
		l.Errorf(format, args...)
	case DEBUG:
		l.Debugf(format, args...)
	default:
		panic(
			fmt.Sprintf(
				"error %d: invalid log level %d passed to LogBasedOnLvl()",
				errs.DEV_ERROR,
				lvl,
			),
		)
	}
}

// LogBasedOnLvl is a wrapper for LogBasedOnLvlf() that takes a string instead of a format string
//
// However, please ensure that the 
// lvl passed in is valid (i.e. INFO, ERROR, or DEBUG), otherwise this function will panic
func (l *Logger) LogBasedOnLvl(lvl int, msg string) {
	l.LogBasedOnLvlf(lvl, msg)
}

func (l *Logger) Debug(args ...any) {
	l.debugLogger.Println(args...)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.debugLogger.Printf(format, args...)
}

func (l *Logger) Info(args ...any) {
	l.infoLogger.Println(args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.infoLogger.Printf(format, args...)
}

func (l *Logger) Error(args ...any) {
	l.errorLogger.Println(args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.errorLogger.Printf(format, args...)
}
