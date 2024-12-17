# Logger

[![Go](https://github.com/smarter-day/logger/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/smarter-day/logger/actions/workflows/go.yml)

 Logger is a Go-based logging library designed to provide structured and efficient logging capabilities. It leverages popular libraries like Logrus and OpenTelemetry to offer advanced logging features and tracing support.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Features

- Structured logging with Logrus
- OpenTelemetry tracing support
- Easy integration with Go applications
- Configurable logging levels

## Installation

To install Logger, ensure you have Go installed and run the following command:

```bash
go get github.com/smarter-day/logger
```

## Usage

Here's a basic example of how to use Logger in your Go application:

```go
package main

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/smarter-day/logger"
)

func main() {
	log := logger.Log(context.Background()).SetLevel(logrus.TraceLevel)
	log.WithValues("key", "value").Info("This is an info message")
	log.WithError(errors.New("some error")).Error("This is an error message")
}
```

For more detailed usage and configuration options, please refer to the documentation within the codebase.

## Contributing

We welcome contributions to Logger!

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
