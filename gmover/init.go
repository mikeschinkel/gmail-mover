package gmover

import (
	"log/slog"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gapi"
)

// Opts holds configuration options for initializing the gmover package
type Opts struct {
	Logger *slog.Logger
	// Add other fields as needed for future extensibility
}

// Initialize sets up the gmover package with the provided options
func Initialize(opts *Opts) (err error) {

	if opts == nil {
		goto end
	}

	if opts.Logger != nil {
		SetLogger(opts.Logger)
		gapi.SetLogger(opts.Logger)
		cliutil.SetLogger(opts.Logger)
	}

	// Build the command tree after all init() functions have completed
	err = cliutil.Initialize()

end:
	return err
}
