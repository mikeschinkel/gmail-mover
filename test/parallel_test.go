package test

import (
	"context"
	"strings"
	"testing"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gmcmds"
)

// TestParallelOutput tests that goroutine isolation works in parallel tests
func TestParallelOutput(t *testing.T) {
	setupTest()

	// Set up ONE shared TestOutput instance before parallel tests
	output := InitTestOutput()

	t.Run("ParallelTest1", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 100; i++ {
			output.ClearOutput()

			ctx := context.Background()
			runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
				Config:        gmcmds.GetConfig(),
				GlobalFlagSet: gmcmds.GlobalFlagSet,
				Args:          []string{}, // Shows help
			})

			err := runner.Run(ctx)
			if err != nil {
				t.Errorf("Iteration %d: Help command should not error, got: %v", i, err)
			}

			result := output.GetOutput()
			if !strings.Contains(result, "Gmail Mover") {
				t.Errorf("Iteration %d: Expected output to contain 'Gmail Mover', got: %q", i, result)
			}

			// Check for cross-contamination from other goroutines
			if strings.Contains(result, "TEST2_MARKER") {
				t.Errorf("Iteration %d: Found Test2 output in Test1, goroutine isolation failed!", i)
			}
		}
	})

	t.Run("ParallelTest2", func(t *testing.T) {
		t.Parallel()

		for i := 0; i < 100; i++ {
			output.ClearOutput()

			// Write a unique marker to this goroutine's output
			output.Printf("TEST2_MARKER")

			ctx := context.Background()
			runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
				Config:        gmcmds.GetConfig(),
				GlobalFlagSet: gmcmds.GlobalFlagSet,
				Args:          []string{}, // Shows help
			})

			err := runner.Run(ctx)
			if err != nil {
				t.Errorf("Iteration %d: Help command should not error, got: %v", i, err)
			}

			result := output.GetOutput()
			if !strings.Contains(result, "Gmail Mover") {
				t.Errorf("Iteration %d: Expected output to contain 'Gmail Mover', got: %q", i, result)
			}

			// Verify our marker is there (proving this is our buffer)
			if !strings.Contains(result, "TEST2_MARKER") {
				t.Errorf("Iteration %d: Expected to find TEST2_MARKER in our own output", i)
			}
		}
	})
}
