package gmcmds

import (
	"log/slog"

	"github.com/mikeschinkel/gmover/logutil"
)

var logger *slog.Logger

// SetLogger sets the slog.Logger to use
func SetLogger(l *slog.Logger) {
	logger = l
}

// ensureLogger panics if logger is not set
func ensureLogger() {
	if logger == nil {
		panic("Must set logger with gmcmds.SetLogger() before using gmcmds package")
	}
}

func init() {
	logutil.RegisterInitializerFunc(func(args logutil.InitializerArgs) error {
		SetLogger(args.Logger)
		return nil
	})
}
