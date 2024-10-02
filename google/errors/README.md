# Google API Error Handling Package

This package provides structured error handling and logging for Google API requests. It defines custom error types for API responses and a flexible error handling function that can be integrated with any logger that implements the required interface.

## Features

- **Custom Google API Error (`GoogleAPIError`)**: 
  A structured error type that captures the HTTP status code, response body, error code, and error message from Google API responses.
  
- **Error Handling (`HandleError`)**: 
  A centralized function for handling and logging errors. It differentiates between custom Google API errors and generic Go errors, sending appropriate HTTP responses to the client.

- **Logger Interface**: 
  The package uses a simple logger interface that allows any logging library to be used as long as it implements the `LogError` method.

## Usage

### 1. Custom Google API Error

The `GoogleAPIError` struct is used to represent an error response from Google APIs. It captures the following fields:

- `StatusCode`: The HTTP status code returned by the API.
- `Body`: The body of the error response.
- `ErrorCode`: A custom error code (optional).
- `ErrorMessage`: A custom error message (optional).

Example of how to create and use the `GoogleAPIError`:

```go
err := &GoogleAPIError{
    StatusCode: http.StatusNotFound,
    Body:       "Resource not found",
    ErrorCode:  "404",
    ErrorMessage: "The requested resource could not be found.",
}
fmt.Println(err.Error()) // Output: API request failed with status 404: Resource not found
```

### 2. Error Handling with `HandleError`

The `HandleError` function logs errors and sends the appropriate HTTP response based on the error type. It accepts the following arguments:

- `logger`: An instance of a logger that implements the `LogError` method.
- `w`: The `http.ResponseWriter` to send the HTTP response.
- `err`: The error to be handled (either `GoogleAPIError` or any standard Go error).

Example usage:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    // Create a mock logger
    logger := &MockLogger{}

    // Simulate an error
    err := &GoogleAPIError{
        StatusCode: http.StatusInternalServerError,
        Body:       "Internal server error",
    }

    // Handle the error
    HandleError(logger, w, err)
}
```

### 3. Logger Interface

The logger used in `HandleError` must implement the following interface:

```go
type Logger interface {
    LogError(message string)
}
```

This allows you to use any logging library (e.g., `log`, `zap`, `slog`) as long as it supports this interface. For testing, you can mock this interface with a custom struct.

### Example of Mock Logger:

```go
type MockLogger struct {
    Messages []string
}

func (ml *MockLogger) LogError(message string) {
    ml.Messages = append(ml.Messages, message)
}
```

## Tests

The package includes unit tests for both the error handling and the custom error struct. The tests are written using the `testify/assert` package.

### Running Tests

To run the tests, execute:

```bash
go test ./...
```

The tests cover:

- Handling `GoogleAPIError` with the correct status code and body.
- Handling generic Go errors.
- Verifying that errors are properly logged by the mock logger.

### Example Test Case

```go
func TestHandleError(t *testing.T) {
    logger := &MockLogger{}
    recorder := httptest.NewRecorder()

    err := &GoogleAPIError{
        StatusCode: http.StatusNotFound,
        Body:       "Not Found",
    }

    HandleError(logger, recorder, err)

    result := recorder.Result()
    assert.Equal(t, http.StatusNotFound, result.StatusCode)
    assert.Contains(t, logger.Messages, "API request failed with status 404: Not Found")
}
```

## Installation

Install the package using `go get`:

```bash
go get github.com/duizendstra/go/google/errors
```

Then import the package in your project:

```go
import "github.com/duizendstra/go/google/errors"
```

## License

This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for details.

---

For any issues or questions, feel free to contact [Jasper Duizendstra](mailto:jasper@duizendstra.com).