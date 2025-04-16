package log

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/onsi/ginkgo/reporters/stenographer/support/go-colorable"
	"github.com/sirupsen/logrus"
)

var l = logrus.New()
var location *time.Location = time.Local

var config Config

type Config struct {
	save     string
	level    string
	timezone string
}

func init() {
	config.save = os.Getenv("LOG_SAVE")
	config.level = os.Getenv("LOG_LEVEL")
	config.timezone = os.Getenv("LOG_TIMEZONE")

	// Colorful output for console
	l.Out = colorable.NewColorableStdout()
	setLogLevel()
	l.SetFormatter(&customFormatter{
		TextFormatter: logrus.TextFormatter{
			FullTimestamp:          true,
			TimestampFormat:        "02.01.2006 15:04:05.000",
			ForceColors:            true,
			DisableLevelTruncation: true,
		},
	})

	// Add hook for saving logs to daily files without colors
	if config.save == "true" {
		l.AddHook(newDailyFileHook())
	}

	// Set timezone for log timestamps
	if config.timezone != "" {
		loc, err := time.LoadLocation(config.timezone)
		if err != nil {
			l.Warnf("Invalid LOG_TIMEZONE: %s. Falling back to local time.", err)
		} else {
			location = loc
		}
	}

	markHealthy()
}

// setLogLevel sets the logging level based on the LOG_LEVEL environment variable.
func setLogLevel() {
	switch config.level {
	case "debug":
		l.SetLevel(logrus.DebugLevel)
	case "info":
		l.SetLevel(logrus.InfoLevel)
	case "warn":
		l.SetLevel(logrus.WarnLevel)
	case "error":
		l.SetLevel(logrus.ErrorLevel)
	case "fatal":
		l.SetLevel(logrus.FatalLevel)
	case "panic":
		l.SetLevel(logrus.PanicLevel)
	default:
		l.SetLevel(logrus.DebugLevel)
	}
}

// customFormatter is a custom logrus formatter that adds colors to the log level.
type customFormatter struct {
	logrus.TextFormatter
}

// Format adds color to the log level and formats the log entry.
func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor string
	levelText := strings.ToUpper(entry.Level.String())

	// Set color for each log level
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = "" // No color for debug and trace
	case logrus.WarnLevel:
		levelColor = "\033[33m" // Yellow
		levelText = "WARN"
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = "\033[31m" // Red
	default:
		levelColor = "\033[36m" // Cyan
	}

	// Adjust level text length to 5 characters
	levelText = fmt.Sprintf("%-5s", levelText)
	if levelColor != "" {
		levelText = fmt.Sprintf("%s%s\x1b[0m", levelColor, levelText)
	}

	formattedMessage := fmt.Sprintf("%s | %s | %s\n",
		entry.Time.In(location).Format(f.TimestampFormat),
		levelText,
		entry.Message)
	return []byte(formattedMessage), nil
}

// newDailyFileHook creates a new logrus hook that writes logs to a daily file.
func newDailyFileHook() *dailyFileHook {
	hook := &dailyFileHook{
		basePath: "data/logs", // Directory for log files
		formatter: &logrus.TextFormatter{ // Formatter without colors
			FullTimestamp:          true,
			TimestampFormat:        "02.01.2006 15:04:05.000",
			ForceColors:            false,
			DisableLevelTruncation: true,
		},
	}
	hook.ensureLogFile()
	return hook
}

// dailyFileHook is a logrus hook that writes logs to a daily file.
type dailyFileHook struct {
	basePath  string
	file      *os.File
	formatter logrus.Formatter
}

func (hook *dailyFileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *dailyFileHook) Fire(entry *logrus.Entry) error {
	hook.ensureLogFile()

	line, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}

	// Remove color codes from the log line before writing to the file
	line = removeColorCodes(line)

	_, err = hook.file.Write(line)
	return err
}

// ensureLogFile ensures that the log file for the current day is open.
func (hook *dailyFileHook) ensureLogFile() {
	now := time.Now().In(location)
	fileName := filepath.Join(hook.basePath, now.Format("2006-01-02")+".log")

	// Check if the file is already open and is current
	if hook.file != nil {
		stat, err := hook.file.Stat()
		if err == nil && stat.Name() == filepath.Base(fileName) {
			return
		}
		hook.file.Close()
	}

	// Ensure the log directory exists
	os.MkdirAll(hook.basePath, os.ModePerm)

	// Open the file for writing
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		l.Errorf("Failed to open log file %s: %v", fileName, err)
		return
	}

	hook.file = file
}

// removeColorCodes removes ANSI color codes from a log line.
func removeColorCodes(line []byte) []byte {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAll(line, []byte(""))
}

// Fatal logs a fatal message and exits the program.
func Fatal(msg string) {
	markUnhealthy()
	l.Fatal(msg)
}

// Fatalf logs a formatted fatal message and exits the program.
func Fatalf(format string, args ...interface{}) {
	markUnhealthy()
	l.Fatalf(format, args...)
}

// Error logs an error message.
func Error(msg string) {
	l.Error(msg)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...interface{}) {
	l.Errorf(format, args...)
}

// Errorln logs an error message with a newline character.
func Errorln(args ...interface{}) {
	l.Errorln(args...)
}

// Warn logs a warning message.
func Warn(msg string) {
	l.Warn(msg)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	l.Warnf(format, args...)
}

// Info logs an info message.
func Info(msg string) {
	l.Info(msg)
}

// Infof logs a formatted info message.
func Infof(format string, args ...interface{}) {
	l.Infof(format, args...)
}

// Debug logs a debug message.
func Debug(msg string) {
	l.Debug(msg)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...interface{}) {
	l.Debugf(format, args...)
}

// markHealthy creates a /tmp/healthy file to indicate that the application is healthy.
func markHealthy() {
	_, err := os.Create("/tmp/healthy")
	if err != nil {
		Errorf("Error creating /tmp/healthy file: %s", err)
	} else {
		Info("The /tmp/healthy file was successfully created.")
	}
}

// markUnhealthy removes the /tmp/healthy file to indicate that the application is unhealthy.
func markUnhealthy() {
	err := os.Remove("/tmp/healthy")
	if err != nil {
		Errorf("Error removing /tmp/healthy file: %s", err)
	} else {
		Info("The /tmp/healthy file was successfully removed.")
	}
}
