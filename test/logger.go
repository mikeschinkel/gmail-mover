package test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/mikeschinkel/gmover/gapi"
	"github.com/mikeschinkel/gmover/gmcmds"
	"github.com/mikeschinkel/gmover/gmover"
)

// TestLogger captures log output for testing
type TestLogger struct {
	buffer *bytes.Buffer
	mu     sync.RWMutex
}

// NewTestLogHandler creates a new test logger
func NewTestLogHandler() *TestLogger {
	return &TestLogger{
		buffer: &bytes.Buffer{},
	}
}

// Enabled implements slog.Handler interface
func (t *TestLogger) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle processes log records
func (t *TestLogger) Handle(_ context.Context, r slog.Record) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Format: LEVEL: message [key=value key=value]
	var attrs strings.Builder
	r.Attrs(func(a slog.Attr) bool {
		if attrs.Len() > 0 {
			attrs.WriteString(" ")
		}
		attrs.WriteString(fmt.Sprintf("%s=%v", a.Key, a.Value))
		return true
	})

	var line string
	if attrs.Len() > 0 {
		line = fmt.Sprintf("%s: %s [%s]\n", r.Level.String(), r.Message, attrs.String())
	} else {
		line = fmt.Sprintf("%s: %s\n", r.Level.String(), r.Message)
	}

	t.buffer.WriteString(line)
	return nil
}

// WithAttrs returns a new handler with additional attributes
func (t *TestLogger) WithAttrs(_ []slog.Attr) slog.Handler {
	return t
}

// WithGroup returns a new handler with a group name
func (t *TestLogger) WithGroup(_ string) slog.Handler {
	return t
}

// GetLogOutput returns all captured log output
func (t *TestLogger) GetLogOutput() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.buffer.String()
}

// ClearLogs clears the captured log output
func (t *TestLogger) ClearLogs() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buffer.Reset()
}

// ContainsLog checks if the log output contains a specific message
func (t *TestLogger) ContainsLog(message string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return strings.Contains(t.buffer.String(), message)
}

// ContainsLogLevel checks if the log output contains a message at a specific level
func (t *TestLogger) ContainsLogLevel(level slog.Level, message string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	expectedPrefix := fmt.Sprintf("%s: %s", level.String(), message)
	return strings.Contains(t.buffer.String(), expectedPrefix)
}

// CountLogLevel returns the number of log entries at a specific level
func (t *TestLogger) CountLogLevel(level slog.Level) int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	levelPrefix := level.String() + ":"
	return strings.Count(t.buffer.String(), levelPrefix)
}

// Global test logger instance
var testLogHandler *TestLogger
var testOutput *TestOutput

// setupTest creates a test logger and initializes gmover
func setupTest() {
	testLogHandler = NewTestLogHandler()
	logger := slog.New(testLogHandler)
	testOutput = InitTestOutput()

	// Initialize gmover package
	err := gmover.Initialize(&gmover.Opts{
		Logger: logger,
		Output: testOutput,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize gmover: %v", err))
	}

	// Set logger for other packages
	gapi.SetLogger(logger)
	gmcmds.SetLogger(logger)
}
