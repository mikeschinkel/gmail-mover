package test

import (
	"bytes"
	"strings"
	"sync"

	"github.com/mikeschinkel/gmover/cliutil"
	"github.com/mikeschinkel/gmover/gapi"
)

// TestOutput captures output for testing with goroutine isolation
type TestOutput struct {
	buffers sync.Map // map[int64]*bytes.Buffer
	mu      sync.Mutex
}

// InitTestOutput creates a test output writer
func InitTestOutput() *TestOutput {
	output := &TestOutput{}
	cliutil.SetOutput(output)
	gapi.SetOutput(output)
	return output
}

// getBuffer returns the buffer for the current goroutine
func (t *TestOutput) getBuffer() *bytes.Buffer {
	id := GoID()
	if buf, ok := t.buffers.Load(id); ok {
		return buf.(*bytes.Buffer)
	}

	// Create new buffer for this goroutine
	newBuf := &bytes.Buffer{}
	t.buffers.Store(id, newBuf)
	return newBuf
}

// Printf captures formatted output
func (t *TestOutput) Printf(format string, args ...interface{}) {
	buf := t.getBuffer()
	t.mu.Lock()
	defer t.mu.Unlock()
	fprintf(buf, format, args...)
}

// Errorf captures formatted error output
func (t *TestOutput) Errorf(format string, args ...interface{}) {
	buf := t.getBuffer()
	t.mu.Lock()
	defer t.mu.Unlock()
	fprintf(buf, format, args...)
}

// GetOutput returns captured output for the current goroutine
func (t *TestOutput) GetOutput() string {
	buf := t.getBuffer()
	return buf.String()
}

// ClearOutput clears captured output for the current goroutine
func (t *TestOutput) ClearOutput() {
	buf := t.getBuffer()
	buf.Reset()
}

// ContainsOutput checks if output contains specific text for the current goroutine
func (t *TestOutput) ContainsOutput(text string) bool {
	buf := t.getBuffer()
	return strings.Contains(buf.String(), text)
}
