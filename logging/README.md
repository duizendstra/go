
# logging

## Overview
The `logging` package provides structured logging capabilities for Go applications. It supports various log levels and integrates with Google Cloud Trace for enhanced traceability.

## Features
- Structured logging with JSON format
- Different log levels: Debug, Info, Notice, Warning, Error, Critical, Alert, Emergency
- Trace context integration for Google Cloud

## Installation
To install the package, run:
```sh
go get github.com/duizendstra/go/logging
```

## Usage

### Initialization
Create a new structured logger with trace context:
```go
package main

import (
    "net/http"
    "github.com/duizendstra/go/logging/structured"
)

func main() {
    req, _ := http.NewRequest("GET", "http://example.com", nil)
    logger := structured.NewStructuredLogger("project-id", "component", req)
    logger.LogInfo("This is an info message")
}
```

### Log Levels
Log messages with different severity levels:
```go
logger.LogDebug("Debug message")
logger.LogInfo("Info message")
logger.LogNotice("Notice message")
logger.LogWarning("Warning message")
logger.LogError("Error message")
logger.LogCritical("Critical message")
logger.LogAlert("Alert message")
logger.LogEmergency("Emergency message")
```

## Middleware Example
Use the logger in HTTP middleware:
```go
func newMiddleware(projectID string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logger := structured.NewStructuredLogger(projectID, "component", r)
            logger.LogInfo("Request received")
            ctx := context.WithValue(r.Context(), "logger", logger)
            r = r.WithContext(ctx)
            next.ServeHTTP(w, r)
            logger.LogInfo("Response sent")
        })
    }
}
```

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
