package main

import "github.com/tsisar/extended-log-go/log"

// main demonstrates various logging functions and levels available in the extended-log-go package.
// It showcases trace, debug, info, warn, and error level logging with both simple and formatted messages.
func main() {
	// Basic logging examples
	log.Println("=== Basic Logging Examples ===")
	log.Println("Simple println message")
	log.Printf("Formatted message: %s = %d", "answer", 42)

	// Trace level - most verbose, used for tracing program flow
	log.Println("\n=== Trace Level ===")
	log.Trace("Entering function calculateSum()")
	log.Tracef("Function parameters: a=%d, b=%d", 10, 20)

	// Debug level - detailed information for debugging
	log.Println("\n=== Debug Level ===")
	log.Debug("Processing user request")
	log.Debugf("Request ID: %s, User ID: %d", "req-12345", 1001)

	// Info level - general informational messages
	log.Println("\n=== Info Level ===")
	log.Info("Application started successfully")
	log.Infof("Server listening on port %d", 8080)

	// Warn level - warning messages that might require attention
	log.Println("\n=== Warn Level ===")
	log.Warn("Database connection pool is running low")
	log.Warnf("Retry attempt %d of %d failed", 2, 3)

	// Error level - error messages for failures
	log.Println("\n=== Error Level ===")
	log.Error("Failed to connect to database")
	log.Errorf("Invalid configuration: %s is missing", "api_key")
	log.Errorln("Connection", "timeout", "occurred")

	// Configuration examples
	log.Println("\n=== Configuration Info ===")
	log.Info("To configure logging, use these environment variables:")
	log.Info("  LOG_SAVE=true          - enable saving logs to files")
	log.Info("  LOG_LEVEL=debug        - set log level (trace, debug, info, warn, error, fatal, panic)")
	log.Info("  LOG_TIMEZONE=UTC       - set timezone for timestamps")
	log.Info("  LOG_DIRECTORY=logs     - set custom directory for log files (default: data/logs)")
	log.Info("  LOG_RETENTION_DAYS=30  - number of days to keep log files (default: 30)")
	log.Info("  LOG_SHOW_CALLER=true   - show file:line where log was called")

	log.Println("\n=== End of Examples ===")
}
