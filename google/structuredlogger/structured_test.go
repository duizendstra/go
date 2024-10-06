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

package structuredlogger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"testing"
	"github.com/duizendstra/go/google/structuredlogger"
)

func TestNewStructuredLogger(t *testing.T) {
	projectID := "test-project"
	component := "test-component"

	// Test case 1: http.Request is nil
	logger := structuredlogger.NewStructuredLogger(projectID, component, nil, nil)
	if logger.GetTraceID() != "" || logger.GetSpanID() != "" || logger.IsTraceSampled() {
		t.Errorf("Expected empty trace details for nil request, got TraceID: %s, SpanID: %s, TraceSampled: %v", logger.GetTraceID(), logger.GetSpanID(), logger.IsTraceSampled())
	}
	if logger.GetComponent() != component {
		t.Errorf("Expected component %s, got %s", component, logger.GetComponent())
	}

	// Test case 2: http.Request with valid X-Cloud-Trace-Context header
	traceHeader := "105445aa7843bc8bf206b120001000/1;o=1"
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Cloud-Trace-Context", traceHeader)

	logger = structuredlogger.NewStructuredLogger(projectID, component, req, nil)
	expectedTraceID := "projects/test-project/traces/105445aa7843bc8bf206b120001000"
	expectedSpanID := "1"
	expectedTraceSampled := true

	if logger.GetTraceID() != expectedTraceID {
		t.Errorf("Expected TraceID %s, got %s", expectedTraceID, logger.GetTraceID())
	}
	if logger.GetSpanID() != expectedSpanID {
		t.Errorf("Expected SpanID %s, got %s", expectedSpanID, logger.GetSpanID())
	}
	if logger.IsTraceSampled() != expectedTraceSampled {
		t.Errorf("Expected TraceSampled %v, got %v", expectedTraceSampled, logger.IsTraceSampled())
	}
	if logger.GetComponent() != component {
		t.Errorf("Expected component %s, got %s", component, logger.GetComponent())
	}

	// Test case 3: http.Request with invalid X-Cloud-Trace-Context header
	traceHeader = "invalid-header"
	req = httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Cloud-Trace-Context", traceHeader)

	logger = structuredlogger.NewStructuredLogger(projectID, component, req, nil)
	if logger.GetTraceID() != "" || logger.GetSpanID() != "" || logger.IsTraceSampled() {
		t.Errorf("Expected empty trace details for invalid header, got TraceID: %s, SpanID: %s, TraceSampled: %v", logger.GetTraceID(), logger.GetSpanID(), logger.IsTraceSampled())
	}
	if logger.GetComponent() != component {
		t.Errorf("Expected component %s, got %s", component, logger.GetComponent())
	}
}

func TestLoggingMethods(t *testing.T) {
	component := "test-component"
	traceID := "projects/test-project/traces/105445aa7843bc8bf206b120001000"
	spanID := "1"
	traceSampled := true

	// Capture the output
	var buf bytes.Buffer

	sl := structuredlogger.NewStructuredLogger("", component, nil, &buf)
	sl.SetTrace(traceID, spanID, traceSampled)

	// Set log level to DEBUG to ensure all messages are captured
	sl.SetLogLevel("DEBUG")

	ctx := context.Background()

	// Test different log levels
	logMethods := []struct {
		name   string
		method func(ctx context.Context, msg string, args ...any)
		level  slog.Level
	}{
		{"LogDebug", sl.LogDebug, slog.LevelDebug},
		{"LogInfo", sl.LogInfo, slog.LevelInfo},
		{"LogNotice", sl.LogNotice, structuredlogger.LevelNotice},
		{"LogWarning", sl.LogWarning, slog.LevelWarn},
		{"LogError", sl.LogError, slog.LevelError},
		{"LogCritical", sl.LogCritical, structuredlogger.LevelCritical},
		{"LogAlert", sl.LogAlert, structuredlogger.LevelAlert},
		{"LogEmergency", sl.LogEmergency, structuredlogger.LevelEmergency},
	}

	for _, lm := range logMethods {
		buf.Reset()
		msg := "Test message for " + lm.name
		lm.method(ctx, msg, "extraKey", "extraValue")

		var loggedEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &loggedEntry)
		if err != nil {
			t.Fatalf("Error unmarshaling log output: %v", err)
		}

		// Check the basic fields
		if loggedEntry["message"] != msg {
			t.Errorf("Expected message '%s', got '%s'", msg, loggedEntry["message"])
		}
		if loggedEntry["component"] != component {
			t.Errorf("Expected component '%s', got '%s'", component, loggedEntry["component"])
		}
		if loggedEntry["logging.googleapis.com/trace"] != traceID {
			t.Errorf("Expected TraceID '%s', got '%s'", traceID, loggedEntry["logging.googleapis.com/trace"])
		}
		if loggedEntry["logging.googleapis.com/spanId"] != spanID {
			t.Errorf("Expected SpanID '%s', got '%s'", spanID, loggedEntry["logging.googleapis.com/spanId"])
		}
		if loggedEntry["logging.googleapis.com/trace_sampled"] != traceSampled {
			t.Errorf("Expected TraceSampled '%v', got '%v'", traceSampled, loggedEntry["logging.googleapis.com/trace_sampled"])
		}

		// Check the level
		if loggedEntry["level"] != lm.level.String() {
			t.Errorf("Expected level '%s', got '%s'", lm.level.String(), loggedEntry["level"])
		}

		// Check extra arguments
		if loggedEntry["extraKey"] != "extraValue" {
			t.Errorf("Expected extraKey 'extraValue', got '%v'", loggedEntry["extraKey"])
		}

		// Check source location for error levels and above
		if lm.level >= slog.LevelError {
			sourceLocation, ok := loggedEntry["logging.googleapis.com/sourceLocation"].(map[string]interface{})
			if !ok {
				t.Errorf("Expected sourceLocation to be present for level '%s'", lm.level.String())
			} else {
				if sourceLocation["file"] == "" || sourceLocation["line"] == nil || sourceLocation["function"] == "" {
					t.Errorf("Incomplete sourceLocation information")
				}
			}
		} else {
			if _, exists := loggedEntry["logging.googleapis.com/sourceLocation"]; exists {
				t.Errorf("Did not expect sourceLocation for level '%s'", lm.level.String())
			}
		}
	}
}

func TestSetLogLevel(t *testing.T) {
	component := "test-component"

	// Capture the output
	var buf bytes.Buffer

	sl := structuredlogger.NewStructuredLogger("", component, nil, &buf)

	ctx := context.Background()

	// Set log level to WARNING
	sl.SetLogLevel("WARNING")

	// Log an INFO message (should not be logged)
	sl.LogInfo(ctx, "This is an info message")

	if buf.Len() != 0 {
		t.Errorf("Expected no output for INFO level when log level is WARNING")
	}

	// Log a WARNING message (should be logged)
	sl.LogWarning(ctx, "This is a warning message")

	if buf.Len() == 0 {
		t.Errorf("Expected output for WARNING level when log level is WARNING")
	}

	// Reset buffer and set log level to DEBUG
	buf.Reset()
	sl.SetLogLevel("DEBUG")

	// Log an INFO message (should be logged)
	sl.LogInfo(ctx, "This is an info message")

	if buf.Len() == 0 {
		t.Errorf("Expected output for INFO level when log level is DEBUG")
	}
}

func TestAdditionalAttributes(t *testing.T) {
	component := "test-component"

	// Capture the output
	var buf bytes.Buffer

	sl := structuredlogger.NewStructuredLogger("", component, nil, &buf)

	// Set log level to DEBUG to capture all messages
	sl.SetLogLevel("DEBUG")

	ctx := context.Background()

	// Log with additional attributes
	sl.LogInfo(ctx, "Test message", "userID", 12345, "role", "admin")

	var loggedEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &loggedEntry)
	if err != nil {
		t.Fatalf("Error unmarshaling log output: %v", err)
	}

	if loggedEntry["userID"] != float64(12345) { // JSON numbers are float64
		t.Errorf("Expected userID '12345', got '%v'", loggedEntry["userID"])
	}

	if loggedEntry["role"] != "admin" {
		t.Errorf("Expected role 'admin', got '%v'", loggedEntry["role"])
	}
}
