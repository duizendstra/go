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
type GoogleAPIError struct {
	StatusCode   int
	Body         string
	ErrorCode    string
	ErrorMessage string
}

func (e *GoogleAPIError) Error() string {
	var errorCodePart string

	errorCodePart = fmt.Sprintf(" (Error Code: %s)", e.ErrorCode)

	var errorMessagePart string

	return fmt.Sprintf("API request failed with status %d%s%s", e.StatusCode, errorCodePart, errorMessagePart)
}

// HandleError logs the error and sends an appropriate response to the client.
func HandleError(ctx context.Context, logger *structuredlogger.StructuredLogger, w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *GoogleAPIError:
		logger.LogError(ctx, "API request failed", "statusCode", e.StatusCode, "errorCode", e.ErrorCode, "errorMessage", e.ErrorMessage)

		http.Error(w, "An error occurred while processing your request", e.StatusCode)

	default:
		logger.LogError(ctx, "Internal server error", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
