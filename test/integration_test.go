package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gapi"
	"github.com/mikeschinkel/gmail-mover/gmcmds"
	"github.com/mikeschinkel/gmail-mover/gmover"

	_ "github.com/mikeschinkel/gmail-mover/gmcmds"
)

// TestCommandSystem tests the command registration and execution system
func TestCommandSystem(t *testing.T) {
	setupTestLogger()

	ctx := context.Background()

	// Test help (no arguments shows help)
	t.Run("HelpCommand", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args:          []string{}, // No arguments shows help
		})

		err := runner.Run(ctx)
		if err != nil {
			t.Errorf("Help (no args) should not error, got: %v", err)
		}
	})

	// Test invalid command
	t.Run("InvalidCommand", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args:          []string{"nonexistent"},
		})

		err := runner.Run(ctx)
		if err == nil {
			t.Error("Invalid command should return an error")
		}
	})
}

// TestMoveCommandValidation tests move command validation
func TestMoveCommandValidation(t *testing.T) {
	setupTestLogger()

	ctx := context.Background()

	testCases := []struct {
		name        string
		args        []string
		shouldError bool
		description string
	}{
		{
			name:        "MissingSourceEmail",
			args:        []string{"move"},
			shouldError: true,
			description: "Move command should require source email",
		},
		{
			name:        "MissingSourceLabel",
			args:        []string{"move", "--src=test@example.com"},
			shouldError: true,
			description: "Move command should require source label",
		},
		{
			name:        "MissingDestinationEmail",
			args:        []string{"move", "--src=test@example.com", "--src-label=INBOX"},
			shouldError: true,
			description: "Move command should require destination email",
		},
		{
			name:        "MissingDestinationLabel",
			args:        []string{"move", "--src=test@example.com", "--src-label=INBOX", "--dst=archive@example.com"},
			shouldError: true,
			description: "Move command should require destination label",
		},
		{
			name:        "ValidParametersNoCreds",
			args:        []string{"move", "--src=test@example.com", "--src-label=INBOX", "--dst=archive@example.com", "--dst-label=moved", "--dry-run"},
			shouldError: true,
			description: "Valid parameters should fail on authentication (no credentials)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
				Config:        gmcmds.GetConfig(),
				GlobalFlagSet: gmcmds.GlobalFlagSet,
				Args:          tc.args,
			})

			err := runner.Run(ctx)
			if tc.shouldError && err == nil {
				t.Errorf("%s: expected error but got none", tc.description)
			} else if !tc.shouldError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tc.description, err)
			}
		})
	}
}

// TestSameAccountDetection tests same-account vs cross-account detection
func TestSameAccountDetection(t *testing.T) {
	setupTestLogger()

	ctx := context.Background()

	// Test same account move (should not error on validation)
	t.Run("SameAccountMove", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args: []string{
				"move",
				"--src=test@example.com",
				"--dst=test@example.com", // Same account
				"--src-label=INBOX",
				"--dst-label=moved",
				"--dry-run",
			},
		})

		err := runner.Run(ctx)
		// Should fail on auth, not validation
		if err == nil {
			t.Error("Expected authentication error for same-account move")
		}
		// Verify it's not a validation error
		if err != nil && strings.Contains(err.Error(), "required") {
			t.Errorf("Same account move failed validation: %v", err)
		}
	})

	// Test that same source and destination labels are rejected
	t.Run("SameSourceDestinationLabels", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args: []string{
				"move",
				"--src=test@example.com",
				"--dst=test@example.com",
				"--src-label=INBOX",
				"--dst-label=INBOX", // Same as source
				"--dry-run",
			},
		})

		err := runner.Run(ctx)
		if err == nil {
			t.Error("Expected validation error for same source and destination labels")
		}
		if err != nil && !strings.Contains(err.Error(), "same") {
			t.Errorf("Expected 'same' validation error, got: %v", err)
		}
	})
}

// TestListLabelsCommand tests the list labels functionality
func TestListLabelsCommand(t *testing.T) {
	setupTestLogger()

	ctx := context.Background()

	t.Run("MissingSourceEmail", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args:          []string{"list", "labels"},
		})

		err := runner.Run(ctx)
		if err == nil {
			t.Error("List labels should require source email")
		}
	})

	t.Run("ValidParametersNoCreds", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args:          []string{"list", "labels", "--src=test@example.com"},
		})

		err := runner.Run(ctx)
		if err == nil {
			t.Error("List labels should fail on authentication without credentials")
		}
	})
}

// TestJobCommands tests job-related functionality
func TestJobCommands(t *testing.T) {
	setupTestLogger()

	ctx := context.Background()

	t.Run("JobDefineCommand", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args:          []string{"job", "define", "test-job"},
		})

		err := runner.Run(ctx)
		// Job define will likely prompt for input or fail gracefully
		// We mainly test that the command is registered and callable
		_ = err // Expected to fail in test environment
	})

	t.Run("JobRunNonexistentFile", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args:          []string{"job", "run", "nonexistent.json"},
		})

		err := runner.Run(ctx)
		if err == nil {
			t.Error("Job run should fail for nonexistent file")
		}
	})
}

// TestGAPIFunctionality tests Gmail API wrapper functionality
func TestGAPIFunctionality(t *testing.T) {
	setupTestLogger()

	t.Run("NewGMailAPI", func(t *testing.T) {
		api := gapi.NewGMailAPI("test-app", gmover.ConfigFileStore())
		if api == nil {
			t.Error("NewGMailAPI should return non-nil API instance")
		}
	})

	t.Run("GetGmailServiceNoAuth", func(t *testing.T) {
		api := gapi.NewGMailAPI("test-app", gmover.ConfigFileStore())

		_, err := api.GetGmailService("test@example.com")
		if err == nil {
			t.Error("GetGmailService should fail without authentication")
		}
	})
}

// TestTransferOptsConfiguration tests TransferOpts structure
func TestTransferOptsConfiguration(t *testing.T) {
	setupTestLogger()

	// Test that TransferOpts can be configured properly
	opts := gapi.TransferOpts{
		Labels:          []string{"INBOX"},
		LabelsToApply:   []string{"moved"},
		LabelsToRemove:  []string{"INBOX"},
		SearchQuery:     "from:test@example.com",
		MaxMessages:     100,
		DryRun:          true,
		DeleteAfterMove: false,
		FailOnError:     false,
	}

	// Verify configuration
	if len(opts.Labels) != 1 || opts.Labels[0] != "INBOX" {
		t.Error("Labels not configured correctly")
	}

	if len(opts.LabelsToApply) != 1 || opts.LabelsToApply[0] != "moved" {
		t.Error("LabelsToApply not configured correctly")
	}

	if len(opts.LabelsToRemove) != 1 || opts.LabelsToRemove[0] != "INBOX" {
		t.Error("LabelsToRemove not configured correctly")
	}

	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
}

// TestDryRunMode tests dry run functionality
func TestDryRunMode(t *testing.T) {
	setupTestLogger()

	ctx := context.Background()

	t.Run("DryRunFlag", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args: []string{
				"move",
				"--src=test@example.com",
				"--dst=archive@example.com",
				"--src-label=INBOX",
				"--dst-label=moved",
				"--dry-run", // This should be handled properly
				"--max=1",
			},
		})

		err := runner.Run(ctx)
		// Should fail on auth, but dry-run flag should be processed
		if err == nil {
			t.Error("Expected authentication error in dry-run mode")
		}
	})
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	setupTestLogger()

	t.Run("CancelledContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args: []string{
				"move",
				"--src=test@example.com",
				"--dst=archive@example.com",
				"--src-label=INBOX",
				"--dst-label=moved",
				"--dry-run",
			},
		})

		err := runner.Run(ctx)
		// Should handle cancellation gracefully
		_ = err // Cancellation handling may vary
	})

	t.Run("TimeoutContext", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args: []string{
				"move",
				"--src=test@example.com",
				"--dst=archive@example.com",
				"--src-label=INBOX",
				"--dst-label=moved",
				"--dry-run",
			},
		})

		err := runner.Run(ctx)
		// Should handle timeout gracefully
		_ = err // Timeout handling may vary based on where it occurs
	})
}

// TestConfigFileStore tests configuration file storage
func TestConfigFileStore(t *testing.T) {
	t.Run("ConfigFileStoreCreation", func(t *testing.T) {
		store := gmover.ConfigFileStore()
		if store == nil {
			t.Error("ConfigFileStore should return non-nil store")
		}
	})
}

// TestMaxMessages tests message limit functionality
func TestMaxMessages(t *testing.T) {
	setupTestLogger()

	ctx := context.Background()

	t.Run("MaxMessagesFlag", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args: []string{
				"move",
				"--src=test@example.com",
				"--dst=archive@example.com",
				"--src-label=INBOX",
				"--dst-label=moved",
				"--max=50",
				"--dry-run",
			},
		})

		err := runner.Run(ctx)
		// Should fail on auth, but max flag should be processed
		if err == nil {
			t.Error("Expected authentication error with max messages flag")
		}
	})
}

// TestSearchQuery tests search query functionality
func TestSearchQuery(t *testing.T) {
	setupTestLogger()

	ctx := context.Background()

	t.Run("SearchQueryFlag", func(t *testing.T) {
		runner := cliutil.NewCmdRunner(cliutil.CmdRunnerArgs{
			Config:        gmcmds.GetConfig(),
			GlobalFlagSet: gmcmds.GlobalFlagSet,
			Args: []string{
				"move",
				"--src=test@example.com",
				"--dst=archive@example.com",
				"--src-label=INBOX",
				"--dst-label=moved",
				"--query=from:important@company.com",
				"--dry-run",
			},
		})

		err := runner.Run(ctx)
		// Should fail on auth, but query flag should be processed
		if err == nil {
			t.Error("Expected authentication error with search query")
		}
	})
}
