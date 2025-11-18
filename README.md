# Extended Log Go

A lightweight, high-performance logging package for Go projects, built on top of the native `log/slog` package. Zero external dependencies - pure Go standard library!

## Features

- 🚀 **Zero Dependencies** - Uses only Go standard library (`log/slog`)
- 🎨 **Colored Console Output** - Enhanced readability with color-coded log levels
- 📁 **Daily Log Rotation** - Automatic file rotation with configurable retention
- ⚙️ **Environment Configuration** - Easy setup via environment variables
- 📊 **Fprintf API** - Full support for `fmt.Fprintf`-like logging with byte count
- 🌍 **Timezone Support** - Configure timezone for log timestamps
- 🔄 **Multiple Handlers** - Console and file logging simultaneously

## Installation

Install the package:

```shell
go get github.com/tsisar/extended-log-go
```

## Usage

### Basic Logging Example

```go
package main

import (
	"github.com/tsisar/extended-log-go/log"
)

func main() {
	log.Info("Application started")
	log.Warn("This is a warning")
	log.Error("An error occurred")
	log.Debugf("User %s logged in", "Alice")
}
```

### Fprintf API with Byte Count

```go
package main

import (
	"github.com/tsisar/extended-log-go/log"
)

func main() {
	// Using InfoFprintf - returns byte count and error
	name, age := "Alice", 30
	n, err := log.InfoFprintf("User %s is %d years old", name, age)
	if err != nil {
		log.Errorf("Logging error: %v", err)
	}
	log.Infof("%d bytes written", n)

	// Using Writer API with standard fmt.Fprintf
	count, err := log.Fprintf(log.WarnWriter(), "Server load is at %d%%", 85)
	if err != nil {
		log.Errorf("Write error: %v", err)
	}
	log.Infof("Written %d bytes to warn log", count)
}
```

### Available Fprintf Functions

- `InfoFprintf(format, args...)` - Info level with byte count
- `WarnFprintf(format, args...)` - Warn level with byte count
- `ErrorFprintf(format, args...)` - Error level with byte count
- `DebugFprintf(format, args...)` - Debug level with byte count

### Writer API

Get `io.Writer` for any log level:

```go
infoWriter := log.InfoWriter()
warnWriter := log.WarnWriter()
errorWriter := log.ErrorWriter()
debugWriter := log.DebugWriter()

// Use with fmt.Fprintf
fmt.Fprintf(infoWriter, "This goes to info log")
```

## Configuration

Configure logging via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_SAVE` | Enable saving logs to files (`true`/`false`) | `false` |
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `debug` |
| `LOG_TIMEZONE` | Timezone for timestamps (e.g., `UTC`, `Europe/Kiev`) | System local |
| `LOG_DIRECTORY` | Directory for log files | `data/logs` |
| `LOG_RETENTION_DAYS` | Days to keep old log files | `30` |

### Example

```bash
export LOG_SAVE=true
export LOG_LEVEL=info
export LOG_TIMEZONE=UTC
export LOG_DIRECTORY=/var/log/myapp
export LOG_RETENTION_DAYS=7

./myapp
```

## Log Levels

- `Trace` / `Tracef` - Most verbose, for tracing execution
- `Debug` / `Debugf` - Detailed information for debugging
- `Info` / `Infof` - General informational messages
- `Warn` / `Warnf` - Warning messages
- `Error` / `Errorf` - Error messages
- `Fatal` / `Fatalf` - Fatal errors (exits program)

## Log Format

Console output (with colors):
```
18.11.2025 11:04:17.250 | INFO  | Application started
18.11.2025 11:04:17.251 | WARN  | Database connection slow
18.11.2025 11:04:17.252 | ERROR | Failed to connect
```

File output (without colors):
```
18.11.2025 11:04:17.250 | INFO  | Application started
18.11.2025 11:04:17.251 | WARN  | Database connection slow
18.11.2025 11:04:17.252 | ERROR | Failed to connect
```

## Log Rotation

When `LOG_SAVE=true`, logs are automatically:
- Saved to daily files (format: `YYYY-MM-DD.log`)
- Rotated at midnight (based on configured timezone)
- Cleaned up after retention period expires

## Dependencies

**None!** 🎉

This package uses only Go standard library:
- `log/slog` - Structured logging
- `fmt`, `os`, `time`, etc. - Standard utilities

## Project Structure

```
log/
├── config.go      - Configuration and initialization
├── handlers.go    - Console, File, and Multi handlers
└── logger.go      - Public API functions
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

Developed with ❤️ by [Tsisar](https://github.com/tsisar)
