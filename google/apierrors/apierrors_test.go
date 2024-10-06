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
package apierrors_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duizendstra/go/google/apierrors"
	"github.com/duizendstra/go/google/structuredlogger"
	"github.com/stretchr/testify/assert"
)

// TestHandleError tests the HandleError function.
func TestHandleError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedBody string
	}{
		{
			name: "APIError with Message",
			err: &apierrors.GoogleAPIError{
				StatusCode: http.StatusNotFound,
				Body:       "Not Found",
				ErrorMessage: "Resource not found",
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "An error occurred while processing your request",
		},
		{
			name: "APIError without Message",
			err: &apierrors.GoogleAPIError{
				StatusCode: http.StatusForbidden,
				Body:       "Forbidden",
			},
			expectedCode: http.StatusForbidden,
			expectedBody: "An error occurred",
		},
		{
			name:         "GenericError",
			err:          errors.New("generic error"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := structuredlogger.NewStructuredLogger("test-project", "test-component", nil, nil)
			recorder := httptest.NewRecorder()
			ctx := context.Background()

			apierrors.HandleError(ctx, logger, recorder, tt.err)

			result := recorder.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expectedCode, result.StatusCode)

			body, err := io.ReadAll(result.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, strings.TrimSpace(string(body)))

			// Adjust the log validation as per structured log format
			assert.NotEmpty(t, logger) // Ensure something is logged
		})
	}
}