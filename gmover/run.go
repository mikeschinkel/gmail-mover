package gmover

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/mikeschinkel/gmover/cliutil"
)

// ConfigProvider provides access to CLI commands and configuration
type ConfigProvider interface {
	GetConfig() cliutil.Config
	GlobalFlagSet() *cliutil.FlagSet
}

type RunArgs struct {
	Args           []string
	CLIWriter      cliutil.OutputWriter
	ConfigProvider ConfigProvider
	Logger         *slog.Logger
}

func Run(ctx context.Context, ra RunArgs) (err error) {
	var runner *cliutil.CmdRunner
	var cancel context.CancelFunc

	// Initialize Scout
	err = Initialize(&Opts{
		Logger:    ra.Logger,
		CLIWriter: ra.CLIWriter,
	})
	if err != nil {
		logger.Error("Failed to initialize GMover", "error", err)
		goto end
	}

	// Set up signal handling for the context
	ctx, cancel = setupSignalHandling(ctx, logger)
	defer cancel()

	// Set up command runner
	runner = cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
		Config:        ra.ConfigProvider.GetConfig(),
		GlobalFlagSet: ra.ConfigProvider.GlobalFlagSet(),
		Args:          ra.Args[1:], // Skip program name
	})

	// Execute command
	err = runner.Run(ctx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			cliutil.Printf("Command failed: %v", err)
			logger.Error("Run aborted", "error", err)
			os.Exit(1)
		}
		cliutil.Printf("Operation cancelled by user")
	}

end:
	return err
}

func setupSignalHandling(ctx context.Context, logger *slog.Logger) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	// Create context with cancellation for Ctrl-C handling
	ctx, cancel = context.WithCancel(ctx)

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
