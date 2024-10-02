// structured.go

// [License Header Omitted for Brevity]

package structured

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
)

type StructuredLogger struct {
	logger       *slog.Logger
	component    string
	traceID      string
	spanID       string
	traceSampled bool
	writer       io.Writer
}

// NewStructuredLogger creates a new StructuredLogger instance with optional trace information.
func NewStructuredLogger(projectID, component string, r *http.Request, writer io.Writer) *StructuredLogger {
	if writer == nil {
		writer = os.Stderr
	}

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		AddSource: false, // We'll add source manually for error levels
	})

	logger := slog.New(handler)

	sl := &StructuredLogger{
		logger:    logger,
		component: component,
		writer:    writer,
	}

	if r != nil {
		traceID, spanID, traceSampled := extractTraceContext(projectID, r)
		sl.traceID = traceID
		sl.spanID = spanID
		sl.traceSampled = traceSampled
	}

	return sl
}

// extractTraceContext extracts trace information from the request headers.
func extractTraceContext(projectID string, r *http.Request) (string, string, bool) {
	get := r.Header.Get("X-Cloud-Trace-Context")
	traceID, spanID, traceSampled := deconstructXCloudTraceContext(get)
	if traceID != "" {
		traceID = fmt.Sprintf("projects/%s/traces/%s", projectID, traceID)
	}
	return traceID, spanID, traceSampled
}

// Log logs a message with the specified level and message.
func (sl *StructuredLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	attrs := []slog.Attr{
		slog.String("component", sl.component),
	}

	if sl.traceID != "" {
		attrs = append(attrs, slog.String("logging.googleapis.com/trace", sl.traceID))
	}

	if sl.spanID != "" {
		attrs = append(attrs, slog.String("logging.googleapis.com/spanId", sl.spanID))
	}

	if sl.traceSampled {
		attrs = append(attrs, slog.Bool("logging.googleapis.com/trace_sampled", true))
	}

	if level >= slog.LevelError {
		// Add source location
		pc, file, line, ok := runtime.Caller(2) // Adjust skip level as needed
		if ok {
			fn := runtime.FuncForPC(pc).Name()
			attrs = append(attrs, slog.Group("logging.googleapis.com/sourceLocation",
				slog.String("file", file),
				slog.Int("line", line),
				slog.String("function", fn),
			))
		}
	}

	// Process additional args as attributes
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if !ok {
				continue // Key must be a string
			}
			attrs = append(attrs, slog.Any(key, args[i+1]))
		}
	}

	// Use LogAttrs to pass slog.Attr
	sl.logger.LogAttrs(ctx, level, msg, attrs...)
}

// LogDebug logs a debug message.
func (sl *StructuredLogger) LogDebug(ctx context.Context, msg string, args ...any) {
	sl.Log(ctx, slog.LevelDebug, msg, args...)
}

// LogInfo logs an info message.
func (sl *StructuredLogger) LogInfo(ctx context.Context, msg string, args ...any) {
	sl.Log(ctx, slog.LevelInfo, msg, args...)
}

// LogNotice logs a notice message (mapped to LevelNotice).
func (sl *StructuredLogger) LogNotice(ctx context.Context, msg string, args ...any) {
	sl.Log(ctx, LevelNotice, msg, args...)
}

// LogWarning logs a warning message.
func (sl *StructuredLogger) LogWarning(ctx context.Context, msg string, args ...any) {
	sl.Log(ctx, slog.LevelWarn, msg, args...)
}

// LogError logs an error message.
func (sl *StructuredLogger) LogError(ctx context.Context, msg string, args ...any) {
	sl.Log(ctx, slog.LevelError, msg, args...)
}

// LogCritical logs a critical message (custom level).
func (sl *StructuredLogger) LogCritical(ctx context.Context, msg string, args ...any) {
	sl.Log(ctx, LevelCritical, msg, args...)
}

// LogAlert logs an alert message (custom level).
func (sl *StructuredLogger) LogAlert(ctx context.Context, msg string, args ...any) {
	sl.Log(ctx, LevelAlert, msg, args...)
}

// LogEmergency logs an emergency message (custom level).
func (sl *StructuredLogger) LogEmergency(ctx context.Context, msg string, args ...any) {
	sl.Log(ctx, LevelEmergency, msg, args...)
}

// SetLogLevel sets the minimum level of logs to output.
func (sl *StructuredLogger) SetLogLevel(level string) {
	var slogLevel slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		slogLevel = slog.LevelDebug
	case "INFO":
		slogLevel = slog.LevelInfo
	case "NOTICE":
		slogLevel = LevelNotice
	case "WARNING":
		slogLevel = slog.LevelWarn
	case "ERROR":
		slogLevel = slog.LevelError
	case "CRITICAL":
		slogLevel = LevelCritical
	case "ALERT":
		slogLevel = LevelAlert
	case "EMERGENCY":
		slogLevel = LevelEmergency
	default:
		slogLevel = slog.LevelInfo
	}

	// Update the handler options to set the log level
	handler := slog.NewJSONHandler(sl.writer, &slog.HandlerOptions{
		Level:     slogLevel,
		AddSource: false,
	})
	sl.logger = slog.New(handler)
}

// Custom log levels beyond the standard slog levels
const (
	LevelNotice    = slog.LevelInfo + 1
	LevelCritical  = slog.LevelError + 1
	LevelAlert     = slog.LevelError + 2
	LevelEmergency = slog.LevelError + 3
)

// deconstructXCloudTraceContext parses the X-Cloud-Trace-Context header.
var reCloudTraceContext = regexp.MustCompile(
	`^([a-f\d]+)/([a-f\d]+);o=(\d+)$`,
)

func deconstructXCloudTraceContext(s string) (traceID, spanID string, traceSampled bool) {
	matches := reCloudTraceContext.FindStringSubmatch(s)
	if len(matches) == 4 {
		traceID = matches[1]
		spanID = matches[2]
		traceSampled = matches[3] == "1"
	}
	return
}
