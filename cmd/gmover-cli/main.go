package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gmcmds"
	"github.com/mikeschinkel/gmail-mover/gmover"

	// Import all commands to trigger their init() functions
	_ "github.com/mikeschinkel/gmail-mover/gmcmds"
)

func main() {
	// Initialize CLI-friendly slog logger first
	handler := NewCLIHandler()
	logger := slog.New(handler)

	// Create context with cancellation for Ctrl-C handling
	ctx, cancel := cancelContext(logger)
	defer cancel()

	// Initialize gmover package
	err := gmover.Initialize(&gmover.Opts{
		Logger: logger,
		Output: cliutil.GetOutput(),
	})
	if err != nil {
		logger.Error("Failed to initialize", "error", err)
		os.Exit(1)
	}

	// Cannot do this in gmover.Initialize() because of import cycles.
	// Meed to find a better way
	gmcmds.SetLogger(logger)

	runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
		Config:        gmcmds.GetConfig(),
		GlobalFlagSet: gmcmds.GlobalFlagSet,
		Args:          os.Args[1:],
	})

	// Execute command using new command system with context and config
	err = runner.Run(ctx)
	if err != nil {
		// Handle context cancellation (Ctrl-C) gracefully
		if errors.Is(err, context.Canceled) {
			logger.Info("Operation cancelled by user")
			os.Exit(0)
		}
		logger.Error("Command failed", "error", err)
		os.Exit(1)
	}
}

// CLAUDE: Would it not make sense for this to be in  gmover.Initialize()?
func cancelContext(logger *slog.Logger) (ctx context.Context, cancel context.CancelFunc) {
	// Create context with cancellation for Ctrl-C handling
	ctx, cancel = context.WithCancel(context.Background())

	// Set up signal handling for graceful shutdown
	// CLAUDE: Would it not make sense for this to be in  gmover.Initialize()?
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received interrupt signal, shutting down...")
		cancel()
	}()
	return ctx, cancel
}
