package test

import (
	"log/slog"
	"os"

	"github.com/mikeschinkel/gmail-mover/gmover"
	"github.com/mikeschinkel/gmail-mover/gmutil"
)

func toPtr[T any](v T) *T {
	return &v
}

// setupTestLogger creates a basic slog logger for tests
func setupTestLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	gmover.SetLogger(logger)
	gmutil.SetLogger(logger)
}
