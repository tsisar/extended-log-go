package log

import (
	"bufio"
	"log/slog"
	"os"
	"strconv"
	"strings"
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
	showCaller    bool
}

// loadEnv loads environment variables from .env file if it exists.
// This function parses simple KEY=VALUE pairs and ignores comments.
func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		// .env file is optional, so we don't return error if it doesn't exist
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip invalid lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Set environment variable only if not already set
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func init() {
	// Try to load .env file from current directory
	_ = loadEnv(".env")
	config.save = os.Getenv("LOG_SAVE")
	config.level = os.Getenv("LOG_LEVEL")
	config.timezone = os.Getenv("LOG_TIMEZONE")
	config.directory = os.Getenv("LOG_DIRECTORY")
	if config.directory == "" {
		config.directory = "data/logs" // Default value
	}

	// Show caller information (file:line)
	config.showCaller = os.Getenv("LOG_SHOW_CALLER") == "true"

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
