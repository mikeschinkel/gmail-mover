package gmover

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/gmail-mover/gapi"
)

type MoveEmailOpts struct {
	ApprovalFunc gapi.ApprovalFunc
}

// MoveEmails executes a move emails operation with approval function
func MoveEmails(ctx context.Context, config *Config, opts MoveEmailOpts) (err error) {
	var api *gapi.GMailAPI

	ensureLogger()

	// Validate configuration - function-specific business logic validation
	err = validateMoveEmailsConfig(config)
	if err != nil {
		goto end
	}

	// Log execution info
	logger.Info("Executing move emails operation")
	if config.DryRun {
		logger.Info("DRY RUN MODE - No messages will be moved")
	}

	// Execute the transfer
	api = gapi.NewGMailAPI(ConfigDirName)
	// TODO: Create TransferMessagesWithOptsAndContext that accepts context
	noop(ctx) // Context will be used when gapi supports it
	err = api.TransferMessages(string(config.SrcEmail), string(config.DstEmail), gapi.TransferOpts{
		Labels:          StringSlice(config.SrcLabels),
		LabelsToApply:   StringSlice(config.DstLabels),
		SearchQuery:     string(config.SearchQuery),
		MaxMessages:     int(config.MaxMessages),
		ApprovalFunc:    opts.ApprovalFunc,
		DeleteAfterMove: config.DeleteAfterMove,
		DryRun:          config.DryRun,
		FailOnError:     false, // Continue on individual message errors
	})
	if err != nil {
		logger.Error("Move operation failed", "error", err)
	}

end:
	return err
}

// validateMoveEmailsConfig validates configuration for move emails operation
// This validates business logic that can't be enforced by the type system
func validateMoveEmailsConfig(config *Config) (err error) {
	// If using job file, skip individual field validation
	if !config.JobFile.IsZero() {
		goto end
	}

	if config.SrcEmail.IsZero() {
		err = fmt.Errorf("source email address is required (use -src flag)")
		goto end
	}

	if len(config.SrcLabels) == 0 || config.SrcLabels[0].IsZero() {
		err = fmt.Errorf("source label is required to prevent accidental mass operations (use -src-label flag, or '*' for all messages)")
		goto end
	}

	if config.DstEmail.IsZero() {
		err = fmt.Errorf("destination email address is required (use -dst flag)")
		goto end
	}

	if len(config.DstLabels) == 0 || config.DstLabels[0].IsZero() {
		err = fmt.Errorf("destination label is required for organizing moved messages (use -dst-label flag)")
		goto end
	}

	if string(config.SrcEmail) == string(config.DstEmail) && SlicesIntersect(config.DstLabels, config.DstLabels) {
		err = fmt.Errorf("source and destination cannot be the same (same email and same label)")
		goto end
	}

end:
	return err
}
