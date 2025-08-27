package gmover

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/mikeschinkel/gmover/cliutil"
	"github.com/mikeschinkel/gmover/gapi"
	"github.com/mikeschinkel/gmover/gmcfg"
	"github.com/mikeschinkel/gmover/logutil"
)

// Opts holds configuration options for initializing the gmover package
type Opts struct {
	AppName   string
	Logger    *slog.Logger
	CLIWriter OutputWriter
	// Add other fields as needed for future extensibility
}

// Initialize sets up the gmover package with the provided options
func Initialize(opts *Opts) (err error) {

	if opts == nil {
		opts = &Opts{}
	}

	if opts.Logger == nil {
		err = errors.New("gmover.Initialize: Logger not set")
		goto end
	}
	err = logutil.CallInitializerFuncs(logutil.InitializerArgs{
		AppName: opts.AppName,
		Logger:  opts.Logger,
	})
	if err != nil {
		err = fmt.Errorf("gmover.Initialize: %w", err)
		goto end
	}
	initializeLoggers(opts.Logger)

	if opts.CLIWriter == nil {
		err = errors.New("gmover.Initialize: CLIWriter not set")
		goto end
	}
	err = cliutil.Initialize(opts.CLIWriter)
	if opts.CLIWriter == nil {
		err = fmt.Errorf("gmover.Initialize: clituil.Initialize() failed; %w", err)
		goto end
	}

	initializeWriters(opts.CLIWriter)
	err = cliutil.CallInitializerFuncs(cliutil.InitializerArgs{
		Writer: opts.CLIWriter,
	})

end:
	return err
}

func initializeWriters(w OutputWriter) {
	SetWriter(w)
	gapi.SetWriter(w)

	// This is redundant as it was already done in cliutil.Initialize(), but leaving
	// it here for symmetry.
	cliutil.SetWriter(w)
}

func initializeLoggers(logger *slog.Logger) {
	//TODO These should be registered by the package, not hard-coded here.
	// OR if possible resolved via reflection
	SetLogger(logger)
	gapi.SetLogger(logger)
	cliutil.SetLogger(logger)
	gmcfg.SetLogger(logger)
}
