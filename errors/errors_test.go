package errors

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockLogger struct {
	Messages []string
}

func (ml *MockLogger) LogError(message string) {
	ml.Messages = append(ml.Messages, message)
}

func TestHandleError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedBody string
	}{
		{
			name: "APIError",
			err: &APIError{
				StatusCode: http.StatusNotFound,
				Body:       "Not Found",
			},
			expectedCode: http.StatusNotFound,
			expectedBody: "Not Found",
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
			logger := &MockLogger{}
			recorder := httptest.NewRecorder()

			HandleError(logger, recorder, tt.err)

			result := recorder.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.expectedCode, result.StatusCode)

			body, err := io.ReadAll(result.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, strings.TrimSpace(string(body)))

			assert.Contains(t, logger.Messages, tt.err.Error())
		})
	}
}
