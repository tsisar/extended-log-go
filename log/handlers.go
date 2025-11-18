package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// getCaller returns the file and line number of the caller.
// skip is the number of stack frames to skip (typically 4 for our logger functions).
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	// Get only the filename, not full path
	file = filepath.Base(file)
	return fmt.Sprintf("%s:%d", file, line)
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

	var message string
	if config.showCaller {
		caller := getCaller(6) // Skip: getCaller -> Handle -> slog -> public func -> user code
		message = fmt.Sprintf("%s | %s | [%s] %s\n", timestamp, levelText, caller, r.Message)
	} else {
		message = fmt.Sprintf("%s | %s | %s\n", timestamp, levelText, r.Message)
	}

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

	var message string
	if config.showCaller {
		caller := getCaller(6) // Skip: getCaller -> Handle -> slog -> public func -> user code
		message = fmt.Sprintf("%s | %s | [%s] %s\n", timestamp, levelText, caller, r.Message)
	} else {
		message = fmt.Sprintf("%s | %s | %s\n", timestamp, levelText, r.Message)
	}

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
		if err := h.file.Close(); err != nil {
			return
		}
	}

	// Ensure the log directory exists
	if err := os.MkdirAll(h.basePath, os.ModePerm); err != nil {
		fprintf(os.Stderr, "Failed to create log directory %s: %v\n", h.basePath, err)
		return
	}

	// Clean up old log files
	h.cleanOldLogs()

	// Open the file for writing
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fprintf(os.Stderr, "Failed to open log file %s: %v\n", fileName, err)
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
			if err := os.Remove(filePath); err != nil {
				fprintf(os.Stderr, "Failed to remove old log file %s: %v\n", filePath, err)
			}
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
