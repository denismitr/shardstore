package logger

import (
	"fmt"
	"log"
	"os"
)

type Env string

const (
	Dev   Env = "dev"
	Prod  Env = "prod"
	Local Env = "local"
)

// Logger - provides small interface for logging errors and debuging
type Logger interface {
	Error(err error)
	Debugf(format string, args ...interface{})
}

// NewStdoutLogger creates StdLogger that uses stderr and stdout for logging
func NewStdoutLogger(env Env, appName string) *StdLogger {
	debugLogger := log.New(os.Stdout, " : "+appName+" : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	errLogger := log.New(os.Stderr, " : "+appName+" : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	return &StdLogger{
		debugLogger: debugLogger,
		errLogger:   errLogger,
		env:         env,
	}
}

// StdLogger uses stderr and stdout for logging
type StdLogger struct {
	errLogger   *log.Logger
	debugLogger *log.Logger
	env         Env
}

// Error - logs error as a string message
func (l *StdLogger) Error(err error) {
	if err != nil {
		_ = l.errLogger.Output(2, fmt.Sprintf("\nERROR in ["+string(l.env)+"] env: %s", err.Error()))
	}
}

// Debugf - prints debug message with params
func (l *StdLogger) Debugf(format string, args ...interface{}) {
	if l.env == Prod {
		return
	}

	_ = l.debugLogger.Output(2, fmt.Sprintf("\nDEBUG in ["+string(l.env)+"] "+format, args...))
}
