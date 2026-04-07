package log

import (
	"bytes"
	"log/slog"
	"regexp"
	"strings"
	"testing"
)

// setupTestLogger replaces the global logger with one that writes to buf
// and returns a cleanup function.
func setupTestLogger(buf *bytes.Buffer) func() {
	origLogger := logger
	origShowCaller := config.showCaller
	origLevel := logLevel.Level()

	logLevel.Set(LevelTrace)
	config.showCaller = true
	h := newConsoleHandler(buf)
	logger = slog.New(h)

	return func() {
		logger = origLogger
		config.showCaller = origShowCaller
		logLevel.Set(origLevel)
	}
}

// extractCaller returns the caller string from a log line like "[file.go:42]".
func extractCaller(output string) string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	clean := ansi.ReplaceAllString(output, "")
	re := regexp.MustCompile(`\[([^\]]+)\]`)
	m := re.FindStringSubmatch(clean)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func TestInfo_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Info("test info") // this line is the expected caller

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestInfof_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Infof("hello %s", "world")

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
	if !strings.Contains(buf.String(), "hello world") {
		t.Errorf("expected formatted message, got %s", buf.String())
	}
}

func TestError_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Error("test error")

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestErrorf_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Errorf("err %d", 42)

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestErrorln_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Errorln("error", "line")

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestWarn_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Warn("test warn")

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestWarnf_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Warnf("warn %v", true)

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestDebug_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Debug("test debug")

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestDebugf_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Debugf("debug %s", "msg")

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestTrace_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Trace("test trace")

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestTracef_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Tracef("trace %d", 1)

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestPrintln_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Println("test println")

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestPrintf_CallerPointsHere(t *testing.T) {
	var buf bytes.Buffer
	cleanup := setupTestLogger(&buf)
	defer cleanup()

	Printf("printf %s", "test")

	caller := extractCaller(buf.String())
	if !strings.HasPrefix(caller, "logger_test.go:") {
		t.Errorf("expected caller logger_test.go:*, got %s", caller)
	}
}

func TestCallerThroughMultiHandler(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	origLogger := logger
	origShowCaller := config.showCaller
	origLevel := logLevel.Level()
	defer func() {
		logger = origLogger
		config.showCaller = origShowCaller
		logLevel.Set(origLevel)
	}()

	logLevel.Set(slog.LevelInfo)
	config.showCaller = true
	h1 := newConsoleHandler(&buf1)
	h2 := newConsoleHandler(&buf2)
	multi := newMultiHandler(h1, h2)
	logger = slog.New(multi)

	Info("multi handler test")

	caller1 := extractCaller(buf1.String())
	caller2 := extractCaller(buf2.String())

	if caller1 != caller2 {
		t.Errorf("callers differ between handlers: %s vs %s", caller1, caller2)
	}
	if !strings.HasPrefix(caller1, "logger_test.go:") {
		t.Errorf("expected logger_test.go caller, got %s", caller1)
	}
}