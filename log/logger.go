package log

import (
	"fmt"
	"os"
)

// Fatal logs a fatal message and exits the program.
func Fatal(msg string) {
	logger.Error(msg)
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits the program.
func Fatalf(format string, args ...interface{}) {
	logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Error logs an error message.
func Error(msg string) {
	logger.Error(msg)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...interface{}) {
	logger.Error(fmt.Sprintf(format, args...))
}

// Errorln logs an error message with a newline character.
func Errorln(args ...interface{}) {
	logger.Error(fmt.Sprint(args...))
}

// Warn logs a warning message.
func Warn(msg string) {
	logger.Warn(msg)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	logger.Warn(fmt.Sprintf(format, args...))
}

// Info logs an info message.
func Info(msg string) {
	logger.Info(msg)
}

// Infof logs a formatted info message.
func Infof(format string, args ...interface{}) {
	logger.Info(fmt.Sprintf(format, args...))
}

// Debug logs a debug message.
func Debug(msg string) {
	logger.Debug(msg)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...interface{}) {
	logger.Debug(fmt.Sprintf(format, args...))
}

// Println logs a message with info level.
func Println(args ...interface{}) {
	logger.Info(fmt.Sprint(args...))
}

// Printf logs a formatted message with info level.
func Printf(format string, args ...interface{}) {
	logger.Info(fmt.Sprintf(format, args...))
}

// Trace logs a trace message (using debug level in slog).
func Trace(msg string) {
	logger.Debug(msg)
}

// Tracef logs a formatted trace message (using debug level in slog).
func Tracef(format string, args ...interface{}) {
	logger.Debug(fmt.Sprintf(format, args...))
}
