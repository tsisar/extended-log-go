package log

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

// testRecord creates a slog.Record with PC pointing to the caller of testRecord.
func testRecord(level slog.Level, msg string) slog.Record {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip runtime.Callers and testRecord
	return slog.NewRecord(time.Date(2026, 4, 7, 12, 0, 0, 0, time.UTC), level, msg, pcs[0])
}

func TestCallerFromRecord(t *testing.T) {
	r := testRecord(slog.LevelInfo, "test")
	caller := callerFromRecord(r)

	if !strings.HasPrefix(caller, "handlers_test.go:") {
		t.Errorf("expected caller to start with handlers_test.go:, got %s", caller)
	}
}

func TestCallerFromRecord_ZeroPC(t *testing.T) {
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	caller := callerFromRecord(r)

	if caller != "unknown:0" {
		t.Errorf("expected unknown:0, got %s", caller)
	}
}

func TestConsoleHandler_Enabled(t *testing.T) {
	logLevel.Set(slog.LevelInfo)
	h := newConsoleHandler(&bytes.Buffer{})

	tests := []struct {
		level    slog.Level
		expected bool
	}{
		{LevelTrace, false},
		{slog.LevelDebug, false},
		{slog.LevelInfo, true},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}

	for _, tt := range tests {
		if got := h.Enabled(context.Background(), tt.level); got != tt.expected {
			t.Errorf("Enabled(%v) = %v, want %v", tt.level, got, tt.expected)
		}
	}
}

func TestConsoleHandler_LevelText(t *testing.T) {
	logLevel.Set(LevelTrace)
	var buf bytes.Buffer
	h := newConsoleHandler(&buf)
	origShowCaller := config.showCaller
	config.showCaller = false
	defer func() { config.showCaller = origShowCaller }()

	tests := []struct {
		level    slog.Level
		expected string
	}{
		{LevelTrace, "TRACE"},
		{slog.LevelDebug, "DEBUG"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelWarn, "WARN"},
		{slog.LevelError, "ERROR"},
	}

	for _, tt := range tests {
		buf.Reset()
		r := slog.NewRecord(time.Date(2026, 4, 7, 12, 0, 0, 0, time.UTC), tt.level, "test message", 0)
		if err := h.Handle(context.Background(), r); err != nil {
			t.Fatalf("Handle() error: %v", err)
		}
		// Strip ANSI codes for comparison
		ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
		clean := ansi.ReplaceAllString(buf.String(), "")
		if !strings.Contains(clean, tt.expected) {
			t.Errorf("level=%v: expected output to contain %q, got %q", tt.level, tt.expected, clean)
		}
	}
}

func TestConsoleHandler_WithCaller(t *testing.T) {
	logLevel.Set(slog.LevelInfo)
	var buf bytes.Buffer
	h := newConsoleHandler(&buf)
	origShowCaller := config.showCaller
	config.showCaller = true
	defer func() { config.showCaller = origShowCaller }()

	r := testRecord(slog.LevelInfo, "hello caller")
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[handlers_test.go:") {
		t.Errorf("expected caller info [handlers_test.go:...], got %s", output)
	}
	if !strings.Contains(output, "hello caller") {
		t.Errorf("expected message in output, got %s", output)
	}
}

func TestConsoleHandler_WithoutCaller(t *testing.T) {
	logLevel.Set(slog.LevelInfo)
	var buf bytes.Buffer
	h := newConsoleHandler(&buf)
	origShowCaller := config.showCaller
	config.showCaller = false
	defer func() { config.showCaller = origShowCaller }()

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "no caller", 0)
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	clean := ansi.ReplaceAllString(buf.String(), "")
	if strings.Contains(clean, "[") {
		t.Errorf("expected no caller brackets in output, got %s", clean)
	}
}

func TestMultiHandler_Enabled(t *testing.T) {
	logLevel.Set(slog.LevelWarn)
	warnHandler := newConsoleHandler(&bytes.Buffer{})

	logLevel.Set(slog.LevelDebug)
	debugHandler := newConsoleHandler(&bytes.Buffer{})

	// Reset to a higher level so only debugHandler enables Debug
	logLevel.Set(slog.LevelWarn)

	// MultiHandler with mixed levels - we create them with different level snapshots
	// Since both share logLevel, we test that at least one is enabled
	multi := newMultiHandler(warnHandler, debugHandler)

	if !multi.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("MultiHandler should be enabled for Warn")
	}
}

func TestMultiHandler_Handle(t *testing.T) {
	logLevel.Set(slog.LevelInfo)
	var buf1, buf2 bytes.Buffer
	h1 := newConsoleHandler(&buf1)
	h2 := newConsoleHandler(&buf2)
	multi := newMultiHandler(h1, h2)

	origShowCaller := config.showCaller
	config.showCaller = false
	defer func() { config.showCaller = origShowCaller }()

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "multi test", 0)
	if err := multi.Handle(context.Background(), r); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	if !strings.Contains(buf1.String(), "multi test") {
		t.Error("handler 1 did not receive the message")
	}
	if !strings.Contains(buf2.String(), "multi test") {
		t.Error("handler 2 did not receive the message")
	}
}

func TestMultiHandler_CallerConsistency(t *testing.T) {
	logLevel.Set(slog.LevelInfo)
	var buf1, buf2 bytes.Buffer
	h1 := newConsoleHandler(&buf1)
	h2 := newConsoleHandler(&buf2)
	multi := newMultiHandler(h1, h2)

	origShowCaller := config.showCaller
	config.showCaller = true
	defer func() { config.showCaller = origShowCaller }()

	r := testRecord(slog.LevelInfo, "caller check")
	if err := multi.Handle(context.Background(), r); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	// Both handlers should show the same caller info
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	clean1 := ansi.ReplaceAllString(buf1.String(), "")
	clean2 := ansi.ReplaceAllString(buf2.String(), "")

	// Extract caller from both outputs
	re := regexp.MustCompile(`\[([^\]]+)\]`)
	m1 := re.FindStringSubmatch(clean1)
	m2 := re.FindStringSubmatch(clean2)

	if len(m1) < 2 || len(m2) < 2 {
		t.Fatalf("could not extract caller from outputs:\n  1: %s\n  2: %s", clean1, clean2)
	}
	if m1[1] != m2[1] {
		t.Errorf("caller mismatch between handlers: %s vs %s", m1[1], m2[1])
	}
	if !strings.HasPrefix(m1[1], "handlers_test.go:") {
		t.Errorf("expected handlers_test.go caller, got %s", m1[1])
	}
}

func TestFileHandler_Handle(t *testing.T) {
	logLevel.Set(slog.LevelInfo)
	dir := t.TempDir()
	h := newFileHandler(dir)

	origShowCaller := config.showCaller
	config.showCaller = false
	defer func() { config.showCaller = origShowCaller }()

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "file test", 0)
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	// Read the log file
	today := time.Now().In(location).Format("2006-01-02") + ".log"
	data, err := os.ReadFile(filepath.Join(dir, today))
	if err != nil {
		t.Fatalf("could not read log file: %v", err)
	}
	if !strings.Contains(string(data), "file test") {
		t.Errorf("log file does not contain message, got: %s", string(data))
	}
}

func TestFileHandler_WithCaller(t *testing.T) {
	logLevel.Set(slog.LevelInfo)
	dir := t.TempDir()
	h := newFileHandler(dir)

	origShowCaller := config.showCaller
	config.showCaller = true
	defer func() { config.showCaller = origShowCaller }()

	r := testRecord(slog.LevelError, "error with caller")
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	today := time.Now().In(location).Format("2006-01-02") + ".log"
	data, err := os.ReadFile(filepath.Join(dir, today))
	if err != nil {
		t.Fatalf("could not read log file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "[handlers_test.go:") {
		t.Errorf("expected caller info in file output, got: %s", content)
	}
	if !strings.Contains(content, "ERROR") {
		t.Errorf("expected ERROR level in file output, got: %s", content)
	}
}

func TestCleanOldLogs(t *testing.T) {
	dir := t.TempDir()
	origRetention := config.retentionDays
	config.retentionDays = 7
	defer func() { config.retentionDays = origRetention }()

	// Create old and fresh log files
	oldDate := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	freshDate := time.Now().Format("2006-01-02")

	os.WriteFile(filepath.Join(dir, oldDate+".log"), []byte("old"), 0666)
	os.WriteFile(filepath.Join(dir, freshDate+".log"), []byte("fresh"), 0666)
	os.WriteFile(filepath.Join(dir, "not-a-date.log"), []byte("other"), 0666)

	h := &FileHandler{basePath: dir}
	h.cleanOldLogs()

	// Old file should be removed
	if _, err := os.Stat(filepath.Join(dir, oldDate+".log")); !os.IsNotExist(err) {
		t.Error("old log file should have been removed")
	}
	// Fresh file should remain
	if _, err := os.Stat(filepath.Join(dir, freshDate+".log")); err != nil {
		t.Error("fresh log file should still exist")
	}
	// Non-date file should remain
	if _, err := os.Stat(filepath.Join(dir, "not-a-date.log")); err != nil {
		t.Error("non-date log file should still exist")
	}
}