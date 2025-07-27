package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// CLIHandler provides user-friendly CLI output without timestamps
type CLIHandler struct {
	out io.Writer
	err io.Writer
}

// NewCLIHandler creates a new CLI-friendly slog handler
func NewCLIHandler() *CLIHandler {
	return &CLIHandler{
		out: os.Stdout,
		err: os.Stderr,
	}
}

// Enabled always returns true for all levels
func (h *CLIHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// Handle formats and outputs log records in a CLI-friendly way
func (h *CLIHandler) Handle(_ context.Context, r slog.Record) error {
	var output string
	var writer io.Writer

	// Choose message format and output stream based on level
	switch r.Level {
	case slog.LevelError:
		output = "Error: " + r.Message
		writer = h.err
	default:
		output = r.Message
		writer = h.out
	}

	// Add attributes in a CLI-friendly format
	r.Attrs(func(a slog.Attr) bool {
		output += fmt.Sprintf(" [%s=%v]", a.Key, a.Value)
		return true
	})

	fprintf(writer, "%s\n", output)
	return nil
}

// WithAttrs returns a new handler with additional attributes
func (h *CLIHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// For simplicity, return the same handler since we handle attrs in Handle()
	return h
}

// WithGroup returns a new handler with a group name
func (h *CLIHandler) WithGroup(_ string) slog.Handler {
	// For CLI output, we don't need grouping
	return h
}
