# Go Packages by Duizendstra

## Overview
This repository contains public Go packages for various utilities.

## Packages

### logging/structured
A structured logging package with support for different log levels and trace context.

## Installation
```sh
go get github.com/duizendstra/go/logging/structured
```

## usage
```go
package main

import (
    "github.com/duizendstra/go/logging/structured"
    "net/http"
)

func main() {
    req, _ := http.NewRequest("GET", "http://example.com", nil)
    logger := structured.NewStructuredLogger("project-id", "component", req)
    logger.LogInfo("This is an info message")
}
```