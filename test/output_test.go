package test

import (
	"context"
	"strings"
	"testing"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gmcmds"
)

// TestOutputCapture tests that our output capture mechanism works
func TestOutputCapture(t *testing.T) {
	setupTest()

	// Setup test output capture

	ctx := context.Background()

	// Test that help command output is captured
	t.Run("HelpOutputCapture", func(t *testing.T) {
		output := InitTestOutput()
		defer output.ClearOutput()

		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args:          []string{}, // No arguments shows help

		})

		err := runner.Run(ctx)
		if err != nil {
			t.Errorf("Help command should not error, got: %v", err)
		}

		// Check if we captured the output
		result := output.GetOutput()
		if !strings.Contains(result, "Gmail Mover") {
			t.Errorf("Expected help output to contain 'Gmail Mover', got: %q", result)
		}

		if !strings.Contains(result, "USAGE:") {
			t.Errorf("Expected help output to contain 'USAGE:', got: %q", result)
		}
	})
}
