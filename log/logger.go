package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// logWithPC captures the caller's PC and sends the record directly to the handler,
// bypassing slog.Logger's internal PC capture which would point to this package.
func logWithPC(level slog.Level, msg string) {
	ctx := context.Background()
	if !logger.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip: runtime.Callers, logWithPC, public wrapper
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	_ = logger.Handler().Handle(ctx, r)
}

// Fatal logs a fatal message and exits the program.
func Fatal(msg string) {
	logWithPC(slog.LevelError, msg)
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits the program.
func Fatalf(format string, args ...interface{}) {
	logWithPC(slog.LevelError, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Error logs an error message.
func Error(msg string) {
	logWithPC(slog.LevelError, msg)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...interface{}) {
	logWithPC(slog.LevelError, fmt.Sprintf(format, args...))
}

// Errorln logs an error message with a newline character.
func Errorln(args ...interface{}) {
	logWithPC(slog.LevelError, fmt.Sprint(args...))
}

// Warn logs a warning message.
func Warn(msg string) {
	logWithPC(slog.LevelWarn, msg)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	logWithPC(slog.LevelWarn, fmt.Sprintf(format, args...))
}

// Info logs an info message.
func Info(msg string) {
	logWithPC(slog.LevelInfo, msg)
}

// Infof logs a formatted info message.
func Infof(format string, args ...interface{}) {
	logWithPC(slog.LevelInfo, fmt.Sprintf(format, args...))
}

// Debug logs a debug message.
func Debug(msg string) {
	logWithPC(slog.LevelDebug, msg)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...interface{}) {
	logWithPC(slog.LevelDebug, fmt.Sprintf(format, args...))
}

// Println logs a message with info level.
func Println(args ...interface{}) {
	logWithPC(slog.LevelInfo, fmt.Sprint(args...))
}

// Printf logs a formatted message with info level.
func Printf(format string, args ...interface{}) {
	logWithPC(slog.LevelInfo, fmt.Sprintf(format, args...))
}

// Trace logs a trace message.
func Trace(msg string) {
	logWithPC(LevelTrace, msg)
}

// Tracef logs a formatted trace message.
func Tracef(format string, args ...interface{}) {
	logWithPC(LevelTrace, fmt.Sprintf(format, args...))
}