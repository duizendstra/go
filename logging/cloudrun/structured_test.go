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

package structured

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewStructuredLogger(t *testing.T) {
	projectID := "test-project"
	component := "test-component"

	// Test case 1: http.Request is nil
	logger := NewStructuredLogger(projectID, component, nil)
	if logger.TraceID != "" || logger.SpanID != "" || logger.TraceSampled {
		t.Errorf("Expected empty trace details for nil request, got TraceID: %s, SpanID: %s, TraceSampled: %v", logger.TraceID, logger.SpanID, logger.TraceSampled)
	}
	if logger.Component != component {
		t.Errorf("Expected component %s, got %s", component, logger.Component)
	}

	// Test case 2: http.Request with valid X-Cloud-Trace-Context header
	traceHeader := "105445aa7843bc8bf206b120001000/1;o=1"
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Cloud-Trace-Context", traceHeader)

	logger = NewStructuredLogger(projectID, component, req)
	expectedTraceID := "projects/test-project/traces/105445aa7843bc8bf206b120001000"
	expectedSpanID := "1"
	expectedTraceSampled := true

	if logger.TraceID != expectedTraceID {
		t.Errorf("Expected TraceID %s, got %s", expectedTraceID, logger.TraceID)
	}
	if logger.SpanID != expectedSpanID {
		t.Errorf("Expected SpanID %s, got %s", expectedSpanID, logger.SpanID)
	}
	if logger.TraceSampled != expectedTraceSampled {
		t.Errorf("Expected TraceSampled %v, got %v", expectedTraceSampled, logger.TraceSampled)
	}
	if logger.Component != component {
		t.Errorf("Expected component %s, got %s", component, logger.Component)
	}

	// Test case 3: http.Request with invalid X-Cloud-Trace-Context header
	traceHeader = "invalid-header"
	req = httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Cloud-Trace-Context", traceHeader)

	logger = NewStructuredLogger(projectID, component, req)
	if logger.TraceID != "" || logger.SpanID != "" || logger.TraceSampled {
		t.Errorf("Expected empty trace details for invalid header, got TraceID: %s, SpanID: %s, TraceSampled: %v", logger.TraceID, logger.SpanID, logger.TraceSampled)
	}
	if logger.Component != component {
		t.Errorf("Expected component %s, got %s", component, logger.Component)
	}
}

func TestLogWithEntry(t *testing.T) {
	component := "test-component"
	traceID := "projects/test-project/traces/105445aa7843bc8bf206b120001000"
	spanID := "1"
	traceSampled := true

	logger := &StructuredLogger{
		TraceID:      traceID,
		SpanID:       spanID,
		TraceSampled: traceSampled,
		Component:    component,
	}

	var buf bytes.Buffer
	log.SetFlags(0)
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(nil)
	}()

	entry := logEntry{
		Severity: "INFO",
		Message:  "Test message",
	}

	logger.logWithEntry(entry)

	logOutput := buf.String()
	expectedTimestamp := time.Now().Format(time.RFC3339)[:19] // Only match up to seconds for consistency

	var loggedEntry logEntry
	err := json.Unmarshal([]byte(logOutput), &loggedEntry)
	if err != nil {
		t.Fatalf("Error unmarshaling log output: %v", err)
	}

	if loggedEntry.Timestamp[:19] != expectedTimestamp {
		t.Errorf("Expected timestamp %s, got %s", expectedTimestamp, loggedEntry.Timestamp[:19])
	}
	if loggedEntry.Severity != entry.Severity {
		t.Errorf("Expected severity %s, got %s", entry.Severity, loggedEntry.Severity)
	}
	if loggedEntry.Component != component {
		t.Errorf("Expected component %s, got %s", component, loggedEntry.Component)
	}
	if loggedEntry.Message != entry.Message {
		t.Errorf("Expected message %s, got %s", entry.Message, loggedEntry.Message)
	}
	if loggedEntry.TraceID != traceID {
		t.Errorf("Expected TraceID %s, got %s", traceID, loggedEntry.TraceID)
	}
	if loggedEntry.SpanID != spanID {
		t.Errorf("Expected SpanID %s, got %s", spanID, loggedEntry.SpanID)
	}
	if loggedEntry.TraceSampled != traceSampled {
		t.Errorf("Expected TraceSampled %v, got %v", traceSampled, loggedEntry.TraceSampled)
	}
}