package test

import (
	"bytes"
	"strings"
	"sync"

	"github.com/mikeschinkel/gmail-mover/cliutil"
)

// TestOutput captures output for testing
type TestOutput struct {
	buffer *bytes.Buffer
	mu     sync.Mutex
}

// NewTestOutput creates a test output writer
func NewTestOutput() *TestOutput {
	return &TestOutput{
		buffer: &bytes.Buffer{},
	}
}

// Printf captures formatted output
func (t *TestOutput) Printf(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fprintf(t.buffer, format, args...)
}

// Errorf captures formatted error output
func (t *TestOutput) Errorf(format string, args ...interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fprintf(t.buffer, format, args...)
}

// GetOutput returns captured output
func (t *TestOutput) GetOutput() string {
	return t.buffer.String()
}

// ClearOutput clears captured output
func (t *TestOutput) ClearOutput() {
	t.buffer.Reset()
}

// ContainsOutput checks if output contains specific text
func (t *TestOutput) ContainsOutput(text string) bool {
	return strings.Contains(t.buffer.String(), text)
}

// Global test output instance
var testOutput *TestOutput

// SetupTestOutput creates and sets a test output writer
func SetupTestOutput() *TestOutput {
	testOutput = NewTestOutput()
	cliutil.SetOutput(testOutput)
	return testOutput
}

// GetTestOutput returns the global test output instance
func GetTestOutput() *TestOutput {
	if testOutput == nil {
		panic("Test output not initialized - call SetupTestOutput() first")
	}
	return testOutput
}
