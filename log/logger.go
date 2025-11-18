package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var logger *slog.Logger
var location = time.Local
var config Config
var logLevel = new(slog.LevelVar)

// Config holds the logging configuration.
type Config struct {
	save          string
	level         string
	timezone      string
	directory     string
	retentionDays int
}

func init() {
	config.save = os.Getenv("LOG_SAVE")
	config.level = os.Getenv("LOG_LEVEL")
	config.timezone = os.Getenv("LOG_TIMEZONE")
	config.directory = os.Getenv("LOG_DIRECTORY")
	if config.directory == "" {
		config.directory = "data/logs" // Default value
	}

	// Parse retention days with default of 30 days
	config.retentionDays = 30
	if retentionStr := os.Getenv("LOG_RETENTION_DAYS"); retentionStr != "" {
		if days, err := strconv.Atoi(retentionStr); err == nil && days > 0 {
			config.retentionDays = days
		} else {
			fmt.Fprintf(os.Stderr, "Invalid LOG_RETENTION_DAYS value: %s. Using default: %d days\n", retentionStr, config.retentionDays)
		}
	}

	// Set log level
	setLogLevel()

	// Set timezone for log timestamps
	if config.timezone != "" {
		loc, err := time.LoadLocation(config.timezone)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid LOG_TIMEZONE: %s. Falling back to local time.\n", err)
		} else {
			location = loc
		}
	}

	// Create handlers
	consoleHandler := newConsoleHandler(os.Stdout)

	if config.save == "true" {
		fileHandler := newFileHandler(config.directory)
		multiHandler := newMultiHandler(consoleHandler, fileHandler)
		logger = slog.New(multiHandler)
	} else {
		logger = slog.New(consoleHandler)
	}
}

// setLogLevel sets the logging level based on the LOG_LEVEL environment variable.
func setLogLevel() {
	switch config.level {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelDebug) // Default to debug (includes trace)
	}
}

// ConsoleHandler is a custom slog handler that outputs colorful logs to the console.
type ConsoleHandler struct {
	w     io.Writer
	level slog.Leveler
	mu    sync.Mutex
}

func newConsoleHandler(w io.Writer) *ConsoleHandler {
	return &ConsoleHandler{
		w:     w,
		level: logLevel,
	}
}

func (h *ConsoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *ConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	var levelColor string
	levelText := strings.ToUpper(r.Level.String())

	// Set color for each log level
	switch r.Level {
	case slog.LevelDebug:
		levelColor = "" // No color for debug
	case slog.LevelWarn:
		levelColor = "\033[33m" // Yellow
	case slog.LevelError:
		levelColor = "\033[31m" // Red
	default:
		levelColor = "\033[36m" // Cyan (for info)
	}

	// Adjust level text length to 5 characters
	levelText = fmt.Sprintf("%-5s", levelText)
	if levelColor != "" {
		levelText = fmt.Sprintf("%s%s\x1b[0m", levelColor, levelText)
	}

	timestamp := r.Time.In(location).Format("02.01.2006 15:04:05.000")
	message := fmt.Sprintf("%s | %s | %s\n", timestamp, levelText, r.Message)

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write([]byte(message))
	return err
}

func (h *ConsoleHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *ConsoleHandler) WithGroup(_ string) slog.Handler {
	return h
}

// FileHandler is a custom slog handler that writes logs to daily files without colors.
type FileHandler struct {
	basePath string
	file     *os.File
	level    slog.Leveler
	mu       sync.Mutex
}

func newFileHandler(basePath string) *FileHandler {
	h := &FileHandler{
		basePath: basePath,
		level:    logLevel,
	}
	h.ensureLogFile()
	return h
}

func (h *FileHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *FileHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.ensureLogFile()

	levelText := strings.ToUpper(r.Level.String())
	levelText = fmt.Sprintf("%-5s", levelText)
	timestamp := r.Time.In(location).Format("02.01.2006 15:04:05.000")
	message := fmt.Sprintf("%s | %s | %s\n", timestamp, levelText, r.Message)

	if h.file != nil {
		_, err := h.file.Write([]byte(message))
		return err
	}
	return nil
}

func (h *FileHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *FileHandler) WithGroup(_ string) slog.Handler {
	return h
}

// ensureLogFile ensures that the log file for the current day is open.
func (h *FileHandler) ensureLogFile() {
	now := time.Now().In(location)
	fileName := filepath.Join(h.basePath, now.Format("2006-01-02")+".log")

	// Check if the file is already open and is current
	if h.file != nil {
		stat, err := h.file.Stat()
		if err == nil && stat.Name() == filepath.Base(fileName) {
			return
		}
		h.file.Close()
	}

	// Ensure the log directory exists
	if err := os.MkdirAll(h.basePath, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory %s: %v\n", h.basePath, err)
		return
	}

	// Clean up old log files
	h.cleanOldLogs()

	// Open the file for writing
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file %s: %v\n", fileName, err)
		return
	}

	h.file = file
}

// cleanOldLogs removes log files older than the configured retention period.
func (h *FileHandler) cleanOldLogs() {
	if config.retentionDays <= 0 {
		return // Retention disabled
	}

	entries, err := os.ReadDir(h.basePath)
	if err != nil {
		return
	}

	cutoffDate := time.Now().In(location).AddDate(0, 0, -config.retentionDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if filename matches the log file pattern (YYYY-MM-DD.log)
		name := entry.Name()
		if !strings.HasSuffix(name, ".log") {
			continue
		}

		// Extract date from filename
		dateStr := strings.TrimSuffix(name, ".log")
		fileDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue // Skip files that don't match the date pattern
		}

		// Remove file if it's older than retention period
		if fileDate.Before(cutoffDate) {
			filePath := filepath.Join(h.basePath, name)
			os.Remove(filePath)
		}
	}
}

// MultiHandler combines multiple handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}

// removeColorCodes removes ANSI color codes from a log line.
func removeColorCodes(line []byte) []byte {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAll(line, []byte(""))
}

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
