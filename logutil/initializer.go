package logutil

import (
	"errors"
	"log/slog"
)

type InitializerArgs struct {
	AppName string // Name of the application using logutil for logging
	Logger  *slog.Logger
}

type InitializerFunc func(InitializerArgs) error

var initializerFuncs []InitializerFunc

func RegisterInitializerFunc(f InitializerFunc) {
	initializerFuncs = append(initializerFuncs, f)
}

func CallInitializerFuncs(args InitializerArgs) (err error) {
	var errs []error
	for _, f := range initializerFuncs {
		errs = append(errs, f(args))
	}
	return errors.Join(errs...)
}

// logger holds the structured logger instance for the golang package
var logger *slog.Logger

// SetLogger sets the logger instance for the golang package and ensures it's valid
func SetLogger(l *slog.Logger) {
	logger = l
	ensureLogger()
}

// ensureLogger panics if no logger has been set, preventing uninitialized usage
func ensureLogger() {
	if logger == nil {
		panic("Must set logger with golang.SetLogger() before using golang package")
	}
}

// init registers the logger initialization function
func init() {
	RegisterInitializerFunc(func(args InitializerArgs) error {
		SetLogger(args.Logger)
		return nil
	})
}
