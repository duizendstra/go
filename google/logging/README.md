# Structured Logger

The **Structured Logger** is a custom logging implementation built on top of Go's `log/slog` package. It enables structured logging with additional trace context information, such as trace IDs and span IDs, and allows for logging at various levels like `DEBUG`, `INFO`, `WARNING`, `ERROR`, and custom levels like `NOTICE`, `CRITICAL`, `ALERT`, and `EMERGENCY`.

This package is designed to log messages in JSON format for better readability and analysis, especially in cloud environments like Google Cloud, where logs can include trace information.

## Features

- **Structured Logging**: Logs messages as structured JSON objects.
- **Trace Context**: Automatically includes trace IDs and span IDs from HTTP requests for better traceability across services.
- **Log Levels**: Supports different log levels (`DEBUG`, `INFO`, `NOTICE`, `WARNING`, `ERROR`, `CRITICAL`, `ALERT`, `EMERGENCY`).
- **Customizable**: Allows setting the minimum log level and adding extra attributes to each log entry.
- **Error Level Source Information**: Automatically captures and includes source location (file, line, function) for error and higher log levels.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [Creating a Logger](#creating-a-logger)
  - [Logging Messages](#logging-messages)
  - [Custom Log Levels](#custom-log-levels)
  - [Setting the Log Level](#setting-the-log-level)
- [Trace Context](#trace-context)
- [Testing](#testing)
- [License](#license)

## Installation

To install this package, you need to clone or download the package source.

You can then include it in your Go project:

```go
import "path/to/structured"
```

You also need to ensure that you have Go modules enabled and can install dependencies like the `log/slog` package.

## Usage

### Creating a Logger

You can create a new `StructuredLogger` using the `NewStructuredLogger` function. Optionally, you can provide an HTTP request to extract trace context information (e.g., from Google Cloud):

```go
logger := structured.NewStructuredLogger("my-project-id", "my-component", nil, nil)
```

You can also redirect logs to a custom writer, such as a file or a buffer:

```go
file, err := os.Create("logs.json")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

logger := structured.NewStructuredLogger("my-project-id", "my-component", nil, file)
```

### Logging Messages

The logger provides methods for logging messages at various levels:

- `LogDebug(ctx context.Context, msg string, args ...any)`
- `LogInfo(ctx context.Context, msg string, args ...any)`
- `LogNotice(ctx context.Context, msg string, args ...any)`
- `LogWarning(ctx context.Context, msg string, args ...any)`
- `LogError(ctx context.Context, msg string, args ...any)`
- `LogCritical(ctx context.Context, msg string, args ...any)`
- `LogAlert(ctx context.Context, msg string, args ...any)`
- `LogEmergency(ctx context.Context, msg string, args ...any)`

Example:

```go
ctx := context.Background()
logger.LogInfo(ctx, "Starting the service", "version", "1.0.0")
logger.LogError(ctx, "Failed to connect to database", "error", err)
```

Additional attributes can be passed as key-value pairs:

```go
logger.LogInfo(ctx, "User login", "userID", 12345, "role", "admin")
```

### Custom Log Levels

This package includes custom log levels:

- `LevelNotice`: Info level + 1
- `LevelCritical`: Error level + 1
- `LevelAlert`: Error level + 2
- `LevelEmergency`: Error level + 3

These levels can be used in the `LogNotice`, `LogCritical`, `LogAlert`, and `LogEmergency` methods.

### Setting the Log Level

You can control the minimum log level by calling `SetLogLevel`:

```go
logger.SetLogLevel("DEBUG")
```

This will ensure that only messages with a severity level of `DEBUG` and above will be logged. Supported levels are:

- `DEBUG`
- `INFO`
- `NOTICE`
- `WARNING`
- `ERROR`
- `CRITICAL`
- `ALERT`
- `EMERGENCY`

## Trace Context

The logger automatically extracts trace information from the `X-Cloud-Trace-Context` header of an HTTP request. This is useful in distributed systems where logs can be correlated across multiple services.

To enable trace logging, pass the HTTP request to the logger:

```go
req := httptest.NewRequest("GET", "http://example.com", nil)
req.Header.Set("X-Cloud-Trace-Context", "105445aa7843bc8bf206b120001000/1;o=1")

logger := structured.NewStructuredLogger("my-project-id", "my-component", req, nil)
```

The logger will include `traceID`, `spanID`, and `trace_sampled` in every log message.

## Testing

The package comes with unit tests. You can run the tests using:

```bash
go test -v
```

Example tests include:

- Creating a logger with or without trace context.
- Logging messages at various levels and checking the output.
- Setting different log levels and verifying which messages are logged.
- Adding additional attributes to log messages.

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.

---

For any issues or questions, feel free to contact [Jasper Duizendstra](mailto:jasper@duizendstra.com).