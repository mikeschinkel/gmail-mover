package cliutil

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// OutputWriter defines the interface for user-facing writer
type OutputWriter interface {
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Console writes to stdout/stderr for normal CLI usage
type outputWriter struct {
	stdout io.Writer
	stderr io.Writer
}

// NewOutputWriter creates a console writer writer
func NewOutputWriter() OutputWriter {
	return &outputWriter{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Printf writes formatted writer to stdout
func (c *outputWriter) Printf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(c.stdout, format, args...)
}

// Errorf writes formatted error writer to stderr
func (c *outputWriter) Errorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(c.stderr, format, args...)
}

// Writer holds the structured Writer instance for the golang package
var writer OutputWriter

// Package-level output instance
var printMu sync.RWMutex
var errorMu sync.RWMutex

// SetWriter sets the global writer writer (primarily for testing)
func SetWriter(w OutputWriter) {
	printMu.Lock()
	defer printMu.Unlock()
	writer = w
	ensureWriter()
}

// GetWriter returns the current writer writer
func GetWriter() OutputWriter {
	printMu.RLock()
	defer printMu.RUnlock()
	return writer
}

// Package-level convenience functions

// Printf writes formatted writer
func Printf(format string, args ...interface{}) {
	printMu.RLock()
	defer printMu.RUnlock()
	writer.Printf(format, args...)
}

// Errorf writes formatted error writer
func Errorf(format string, args ...interface{}) {
	errorMu.RLock()
	defer errorMu.RUnlock()
	writer.Errorf(format, args...)
}

// ensureWriter panics if no Writer has been set, preventing uninitialized usage
func ensureWriter() {
	if writer == nil {
		panic("Must set Writer with golang.SetWriter() before using golang package")
	}
}

// init registers the Writer initialization function
func init() {
	RegisterInitializerFunc(func(args InitializerArgs) error {
		SetWriter(args.Writer)
		return nil
	})
}
