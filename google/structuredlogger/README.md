Here is a revised version of the README for the **Structured Logger**:

---

# Structured Logger

The **Structured Logger** is a flexible logging package built on top of Go's `log/slog`. It is designed to support structured logging in JSON format, with advanced features for cloud-based environments, like trace context and detailed logging levels.

## Features

- **Structured JSON Logging**: Outputs logs in JSON format for easy parsing and analysis in cloud logging platforms.
- **Trace Context**: Automatically includes trace and span IDs from incoming HTTP requests to support distributed tracing.
- **Detailed Log Levels**: Supports standard log levels like `DEBUG`, `INFO`, `WARNING`, and `ERROR`, as well as custom levels like `NOTICE`, `CRITICAL`, `ALERT`, and `EMERGENCY`.
- **Customizable Log Levels**: Set the minimum log level to filter out unnecessary logs.
- **Source Location for Errors**: Automatically includes file, line number, and function for error-level logs and higher.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [Creating a Logger](#creating-a-logger)
  - [Logging Examples](#logging-examples)
  - [Custom Log Levels](#custom-log-levels)
  - [Setting the Log Level](#setting-the-log-level)
- [Trace Context](#trace-context)
- [Testing](#testing)
- [License](#license)

## Installation

To use the **Structured Logger**, install the package using:

```bash
go get github.com/your-repo/structured
```

Then, import it in your Go project:

```go
import "github.com/your-repo/structured"
```

## Usage

### Creating a Logger

Create a new `StructuredLogger` using the `NewStructuredLogger` function. Optionally, pass an HTTP request to extract trace information:

```go
logger := structured.NewStructuredLogger("my-project-id", "my-component", nil, nil)
```

You can also direct logs to a custom output like a file:

```go
file, err := os.Create("logs.json")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

logger := structured.NewStructuredLogger("my-project-id", "my-component", nil, file)
```

### Logging Examples

Log messages with various log levels:

```go
ctx := context.Background()

logger.LogInfo(ctx, "Service started", "version", "1.0.0")
logger.LogWarning(ctx, "Disk usage is high", "threshold", 85)
logger.LogError(ctx, "Failed to connect to database", "error", err)
```

Each log entry can include additional attributes as key-value pairs:

```go
logger.LogDebug(ctx, "Debugging request", "requestID", "abc123", "userID", 42)
```

### Custom Log Levels

The package supports custom log levels beyond the standard ones:

- `LevelNotice`: Information that requires special attention.
- `LevelCritical`: Critical issues that require immediate resolution.
- `LevelAlert`: Issues requiring prompt action.
- `LevelEmergency`: Severe issues that need immediate attention.

Use them like this:

```go
logger.LogNotice(ctx, "System maintenance starting soon")
logger.LogCritical(ctx, "Database corruption detected")
logger.LogEmergency(ctx, "Service unavailable, escalating")
```

### Setting the Log Level

Control the verbosity of logs by setting the minimum log level:

```go
logger.SetLogLevel("DEBUG")
```

Supported levels are:

- `DEBUG`
- `INFO`
- `NOTICE`
- `WARNING`
- `ERROR`
- `CRITICAL`
- `ALERT`
- `EMERGENCY`

Only logs at or above the specified level will be captured.

## Trace Context

The logger can automatically extract trace information (such as trace IDs and span IDs) from HTTP requests via the `X-Cloud-Trace-Context` header. This is particularly useful for cloud environments like Google Cloud, where logs from multiple services need to be correlated.

To enable trace logging, pass an HTTP request to the logger:

```go
req := httptest.NewRequest("GET", "http://example.com", nil)
req.Header.Set("X-Cloud-Trace-Context", "105445aa7843bc8bf206b120001000/1;o=1")

logger := structured.NewStructuredLogger("my-project-id", "my-component", req, nil)
```

The logger will then include trace-related information in all logs:

```json
{
  "traceID": "projects/my-project-id/traces/105445aa7843bc8bf206b120001000",
  "spanID": "1",
  "traceSampled": true,
  "msg": "Service started"
}
```

## Testing

To run the tests:

```bash
go test -v
```

Tests cover the following scenarios:

- Logger creation with and without trace context.
- Logging at various levels with proper output.
- Verifying which messages are logged based on the set log level.
- Testing for correct trace and error information in logs.

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for more details.

---

For any questions or issues, feel free to contact [Jasper Duizendstra](mailto:jasper@duizendstra.com).
