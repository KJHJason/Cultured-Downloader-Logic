package logger

import (
	"fmt"
	"io"
	"log"
	"os"

	cdlerrors "github.com/KJHJason/Cultured-Downloader-Logic/errors"
)

const (
	// Log levels
	TRACE = iota
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
)

type Logger struct {
	traceLogger   *log.Logger
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	fatalLogger   *log.Logger
}

func NewLogger(out io.Writer) Logger {
	if out == nil {
		out = os.Stdout
	}

	return Logger{
		traceLogger:   log.New(out, "[TRACE]: ", log.Ldate|log.Ltime),
		debugLogger:   log.New(out, "[DEBUG]: ", log.Ldate|log.Ltime),
		infoLogger:    log.New(out, "[INFO]: ", log.Ldate|log.Ltime),
		warningLogger: log.New(out, "[WARNING]: ", log.Ldate|log.Ltime),
		errorLogger:   log.New(out, "[ERROR]: ", log.Ldate|log.Ltime),
		fatalLogger:   log.New(out, "[FATAL]: ", log.Ldate|log.Ltime),
	}
}

func (l Logger) SetOutput(w io.Writer) {
	l.traceLogger.SetOutput(w)
	l.debugLogger.SetOutput(w)
	l.infoLogger.SetOutput(w)
	l.warningLogger.SetOutput(w)
	l.errorLogger.SetOutput(w)
	l.fatalLogger.SetOutput(w)
}

// LogBasedOnLvlf logs a message based on the log level passed in
//
// # You can use this function to log a message with a format string
//
// However, please ensure that the
// lvl passed in is valid (i.e. TRACE, DEBUG, etc.), otherwise this function will panic
func (l Logger) LogBasedOnLvlf(lvl int, format string, args ...any) {
	switch lvl {
	case TRACE:
		l.Tracef(format, args...)
	case DEBUG:
		l.Debugf(format, args...)
	case INFO:
		l.Infof(format, args...)
	case WARNING:
		l.Warningf(format, args...)
	case ERROR:
		l.Errorf(format, args...)
	case FATAL:
		l.Fatalf(format, args...)
	default:
		panic(
			fmt.Sprintf(
				"error %d: invalid log level %d passed to LogBasedOnLvl()",
				cdlerrors.DEV_ERROR,
				lvl,
			),
		)
	}
}

// LogBasedOnLvl is a wrapper for LogBasedOnLvlf() that takes a string instead of a format string
//
// However, please ensure that the
// lvl passed in is valid (i.e. INFO, ERROR, or DEBUG), otherwise this function will panic
func (l Logger) LogBasedOnLvl(lvl int, msg string) {
	l.LogBasedOnLvlf(lvl, msg)
}

var printLogger = log.New(os.Stdout, "[PRINT]: ", log.Ldate|log.Ltime)
func (l Logger) Print(message string) {
	printLogger.Println(message)
}

func (l Logger) Printf(format string, args ...any) {
	printLogger.Printf(format, args...)
}

func (l Logger) Trace(message string) {
	l.traceLogger.Println(message)
}

func (l Logger) Tracef(format string, args ...any) {
	l.traceLogger.Printf(format, args...)
}

func (l Logger) Debug(message string) {
	l.debugLogger.Println(message)
}

func (l Logger) Debugf(format string, args ...any) {
	l.debugLogger.Printf(format, args...)
}

func (l Logger) Info(message string) {
	l.infoLogger.Println(message)
}

func (l Logger) Infof(format string, args ...any) {
	l.infoLogger.Printf(format, args...)
}

func (l Logger) Warning(message string) {
	l.warningLogger.Println(message)
}

func (l Logger) Warningf(format string, args ...any) {
	l.warningLogger.Printf(format, args...)
}

func (l Logger) Error(message string) {
	l.errorLogger.Println(message)
}

func (l Logger) Errorf(format string, args ...any) {
	l.errorLogger.Printf(format, args...)
}

func (l Logger) Fatal(message string) {
	l.fatalLogger.Fatal(message)
}

func (l Logger) Fatalf(format string, args ...any) {
	l.fatalLogger.Fatalf(format, args...)
}
