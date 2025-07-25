package test

import (
	"log/slog"
	"os"

	"github.com/mikeschinkel/gmail-mover/gapi"
	"github.com/mikeschinkel/gmail-mover/gmover"
)

// setupTestLogger creates a basic slog logger for tests
func setupTestLogger() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	gmover.SetLogger(logger)
	gapi.SetLogger(logger)
}
