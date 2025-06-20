# Extended Log Go

A lightweight logging package for Go projects, built on top of `logrus` for advanced structured logging. Includes support for colored logs for enhanced readability.

## Features

- Built-in integration with [logrus](https://github.com/sirupsen/logrus) for structured logging.
- Supports multiple log levels: INFO, WARN, ERROR, DEBUG, etc.
- Colored log output with [ginkgo](https://github.com/onsi/ginkgo).

## Installation

Install the package and its dependencies:

```shell
go get github.com/tsisar/extended-log-go@v1.0.0
```

## Usage

Basic Logging Example

```go
package main

import (
	"github.com/tsisar/extended-log-go/log"
)

func main() {
	logger := log.NewLogger()

	logger.Info("Application started")
	logger.Warn("This is a warning")
	logger.Error("An error occurred")
}
```

## Dependencies

This package relies on the following external libraries:

- [logrus](https://github.com/sirupsen/logrus): For structured logging.
- [ginkgo](https://github.com/onsi/ginkgo): For color support in logs.
- tzdata in target image: For timezone data.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

Developed with ❤️ by [Tsisar](https://github.com/Tsisar).