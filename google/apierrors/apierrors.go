// Copyright 2024 Jasper Duizendstra
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package apierrors

import (
	"context"
	"fmt"
	"net/http"

	"github.com/duizendstra/go/google/structuredlogger"
)

// GoogleAPIError represents an error response from an API request.
// It contains details such as the HTTP status code, body, error code, and message.
type GoogleAPIError struct {
	StatusCode   int    // HTTP status code returned by the API
	Body         string // Response body returned by the API
	ErrorCode    string // Specific error code provided by the API
	ErrorMessage string // Human-readable error message
}

// Error implements the error interface for GoogleAPIError.
// It returns a formatted string describing the error.
func (e *GoogleAPIError) Error() string {
	// Construct the error code part only if an error code is present.
	errorCodePart := ""
	if e.ErrorCode != "" {
		errorCodePart = fmt.Sprintf(" (Error Code: %s)", e.ErrorCode)
	}
	// Return the complete error message including the status code and error code part if available.
	return fmt.Sprintf("API request failed with status %d%s", e.StatusCode, errorCodePart)
}

// HandleError logs the error and sends an appropriate response to the client.
// It handles different error types and logs them accordingly.
func HandleError(ctx context.Context, logger *structuredlogger.StructuredLogger, w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *GoogleAPIError:
		// Log detailed information about the Google API error.
		if e.ErrorMessage != "" {
			logger.LogError(ctx, "API request failed", "statusCode", e.StatusCode, "errorCode", e.ErrorCode, "errorMessage", e.ErrorMessage)
		} else {
			logger.LogError(ctx, "API request failed", "statusCode", e.StatusCode, "errorCode", e.ErrorCode)
		}
		// Log the response body if it's available for debugging purposes.
		if e.Body != "" {
			logger.LogDebug(ctx, "Detailed error response body", "body", e.Body)
		}
		// Send a generic error response to the client to avoid exposing sensitive information.
		http.Error(w, "An error occurred while processing your request", e.StatusCode)
	default:
		// Handle unknown error types by logging the error and returning a generic server error.
		logger.LogError(ctx, "Unknown error type", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}