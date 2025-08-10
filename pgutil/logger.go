package pgutil

import (
	"log/slog"
)

const (
	ErrorLogArg = "error"
)

var logger = slog.Default().With("package", "pgutil")
