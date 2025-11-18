package log

import (
	"fmt"
	"io"
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

// Writer returns an io.Writer that writes to the logger at the specified level.
type logWriter struct {
	logFunc func(string)
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.logFunc(string(p))
	return len(p), nil
}

// ErrorWriter returns an io.Writer that writes to the logger at error level.
func ErrorWriter() io.Writer {
	return &logWriter{logFunc: Error}
}

// WarnWriter returns an io.Writer that writes to the logger at warn level.
func WarnWriter() io.Writer {
	return &logWriter{logFunc: Warn}
}

// InfoWriter returns an io.Writer that writes to the logger at info level.
func InfoWriter() io.Writer {
	return &logWriter{logFunc: Info}
}

// DebugWriter returns an io.Writer that writes to the logger at debug level.
func DebugWriter() io.Writer {
	return &logWriter{logFunc: Debug}
}

// Fprintf formats according to a format specifier and writes to the logger at info level.
// It returns the number of bytes written and any write error encountered.
func Fprintf(w io.Writer, format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(w, format, args...)
}

// ErrorFprintf formats according to a format specifier and writes to the logger at error level.
// It returns the number of bytes written and any write error encountered.
func ErrorFprintf(format string, args ...interface{}) (n int, err error) {
	msg := fmt.Sprintf(format, args...)
	Error(msg)
	return len(msg), nil
}

// WarnFprintf formats according to a format specifier and writes to the logger at warn level.
// It returns the number of bytes written and any write error encountered.
func WarnFprintf(format string, args ...interface{}) (n int, err error) {
	msg := fmt.Sprintf(format, args...)
	Warn(msg)
	return len(msg), nil
}

// InfoFprintf formats according to a format specifier and writes to the logger at info level.
// It returns the number of bytes written and any write error encountered.
func InfoFprintf(format string, args ...interface{}) (n int, err error) {
	msg := fmt.Sprintf(format, args...)
	Info(msg)
	return len(msg), nil
}

// DebugFprintf formats according to a format specifier and writes to the logger at debug level.
// It returns the number of bytes written and any write error encountered.
func DebugFprintf(format string, args ...interface{}) (n int, err error) {
	msg := fmt.Sprintf(format, args...)
	Debug(msg)
	return len(msg), nil
}
