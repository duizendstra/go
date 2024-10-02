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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"time"
)

const (
	LevelDebug     = 100
	LevelInfo      = 200
	LevelNotice    = 300
	LevelWarning   = 400
	LevelError     = 500
	LevelCritical  = 600
	LevelAlert     = 700
	LevelEmergency = 800
)

type StructuredLogger struct {
	SpanID       string `json:"logging.googleapis.com/spanId,omitempty"`
	TraceID      string `json:"logging.googleapis.com/trace,omitempty"`
	TraceSampled bool   `json:"logging.googleapis.com/traceSampled,omitempty"`
	Component    string
	logLevel     int
}

func NewStructuredLogger(projectID, component string, r *http.Request) *StructuredLogger {
	level := LevelInfo // default log level

	structuredLogger := &StructuredLogger{
		Component: component,
		logLevel:  level,
	}

	if r != nil {
		get := r.Header.Get("X-Cloud-Trace-Context")
		traceID, spanID, traceSampled := deconstructXCloudTraceContext(get)
		if traceID != "" {
			traceID = fmt.Sprintf("projects/%s/traces/%s", projectID, traceID)
			structuredLogger.TraceID = traceID
			structuredLogger.SpanID = spanID
			structuredLogger.TraceSampled = traceSampled
		}
	}

	return structuredLogger
}

type logEntry struct {
	Severity       string            `json:"severity"`
	Message        string            `json:"message"`
	Timestamp      string            `json:"timestamp"`
	Labels         map[string]string `json:"logging.googleapis.com/labels,omitempty"`
	Operation      *operation        `json:"logging.googleapis.com/operation,omitempty"`
	SourceLocation *sourceLocation   `json:"logging.googleapis.com/sourceLocation,omitempty"`
	SpanID         string            `json:"logging.googleapis.com/spanId,omitempty"`
	TraceID        string            `json:"logging.googleapis.com/trace,omitempty"`
	TraceSampled   bool              `json:"logging.googleapis.com/trace_sampled,omitempty"`
	Component      string            `json:"component,omitempty"`
}

type operation struct {
	Id       string `json:"id,omitempty"`
	Producer string `json:"producer,omitempty"`
	First    string `json:"first,omitempty"`
	Last     string `json:"last,omitempty"`
}

type sourceLocation struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
}

func getSourceLocation(skip int) *sourceLocation {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	function := runtime.FuncForPC(pc).Name()

	return &sourceLocation{
		File:     file,
		Line:     line,
		Function: function,
	}
}

func (cl *StructuredLogger) logWithEntry(entry logEntry) {
	entry.Timestamp = time.Now().Format(time.RFC3339)
	entry.Component = cl.Component
	entry.TraceID = cl.TraceID
	entry.SpanID = cl.SpanID
	entry.TraceSampled = cl.TraceSampled

	if entry.Severity == "ERROR" || entry.Severity == "CRITICAL" || entry.Severity == "ALERT" || entry.Severity == "EMERGENCY" {
		entry.SourceLocation = getSourceLocation(3)
	}

	logEntryJSON, err := json.Marshal(entry)
	if err != nil {
		log.Printf(`{"severity": "ERROR", "message": "Error marshaling log entry: %v"}`, err)
		return
	}

	log.Printf("%s", string(logEntryJSON))
}

func (cl *StructuredLogger) Log(severity int, message string) {
	if severity >= cl.logLevel {
		entry := logEntry{
			Severity: mapSeverity(severity),
			Message:  message,
		}

		if severity >= LevelError {
			entry.SourceLocation = getSourceLocation(3)
		}

		entry.Timestamp = time.Now().Format(time.RFC3339)
		entry.Component = cl.Component
		entry.TraceID = cl.TraceID
		entry.SpanID = cl.SpanID
		entry.TraceSampled = cl.TraceSampled

		logEntryJSON, err := json.Marshal(entry)
		if err != nil {
			log.Printf(`{"severity": "ERROR", "message": "Error marshaling log entry: %v"}`, err)
			return
		}

		log.Printf("%s", string(logEntryJSON))
	}
}

func (cl *StructuredLogger) LogDebug(message string) {
	cl.Log(LevelDebug, message)
}

func (cl *StructuredLogger) LogInfo(message string) {
	cl.Log(LevelInfo, message)
}

func (cl *StructuredLogger) LogNotice(message string) {
	cl.Log(LevelNotice, message)
}

func (cl *StructuredLogger) LogWarning(message string) {
	cl.Log(LevelWarning, message)
}

func (cl *StructuredLogger) LogError(message string) {
	cl.Log(LevelError, message)
}

func (cl *StructuredLogger) LogCritical(message string) {
	cl.Log(LevelCritical, message)
}

func (cl *StructuredLogger) LogAlert(message string) {
	cl.Log(LevelAlert, message)
}

func (cl *StructuredLogger) LogEmergency(message string) {
	cl.Log(LevelEmergency, message)
}

func (cl *StructuredLogger) LogLevel(level string) {
	switch level {
	case "DEBUG":
		cl.logLevel = LevelDebug
	case "INFO":
		cl.logLevel = LevelInfo
	case "NOTICE":
		cl.logLevel = LevelNotice
	case "WARNING":
		cl.logLevel = LevelWarning
	case "ERROR":
		cl.logLevel = LevelError
	case "CRITICAL":
		cl.logLevel = LevelCritical
	case "ALERT":
		cl.logLevel = LevelAlert
	case "EMERGENCY":
		cl.logLevel = LevelEmergency
	default:
		cl.logLevel = LevelInfo // default log level
	}
}

func mapSeverity(level int) string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelNotice:
		return "NOTICE"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelCritical:
		return "CRITICAL"
	case LevelAlert:
		return "ALERT"
	case LevelEmergency:
		return "EMERGENCY"
	default:
		return "DEFAULT"
	}
}

var reCloudTraceContext = regexp.MustCompile(
	`([a-f\d]+)?` +
		`(?:/([a-f\d]+))?` +
		`(?:;o=(\d))?`)

func deconstructXCloudTraceContext(s string) (traceID, spanID string, traceSampled bool) {
	matches := reCloudTraceContext.FindStringSubmatch(s)
	if len(matches) >= 4 {
		traceID, spanID, traceSampled = matches[1], matches[2], matches[3] == "1"
		if spanID == "0" {
			spanID = ""
		}
	}
	return
}
