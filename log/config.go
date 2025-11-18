package log

import (
	"log/slog"
	"os"
	"strconv"
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
			fprintf(os.Stderr, "Invalid LOG_RETENTION_DAYS value: %s. Using default: %d days\n", retentionStr, config.retentionDays)
		}
	}

	// Set log level
	setLogLevel()

	// Set timezone for log timestamps
	if config.timezone != "" {
		loc, err := time.LoadLocation(config.timezone)
		if err != nil {
			fprintf(os.Stderr, "Invalid LOG_TIMEZONE: %s. Falling back to local time.\n", err)
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
