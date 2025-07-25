package gmover

import (
	"context"
	"fmt"
	"slices"

	"github.com/mikeschinkel/gmail-mover/gapi"
)

// MoveEmails executes a move emails operation with the provided configuration
func MoveEmails(ctx context.Context, config *Config) (err error) {
	return MoveEmailsWithApproval(ctx, config, nil)
}

// MoveEmailsWithApproval executes a move emails operation with approval function
func MoveEmailsWithApproval(ctx context.Context, config *Config, approvalFunc gapi.ApprovalFunc) (err error) {
	var job *Job

	ensureLogger()

	// Validate configuration - function-specific business logic validation
	err = validateMoveEmailsConfig(config)
	if err != nil {
		goto end
	}

	job, err = GetJob(*config)
	if err != nil {
		logger.Error("Failed to get job", "error", err)
		goto end
	}

	// Execute the job
	logger.Info("Executing job", "job_name", job.Name)
	if job.Options.DryRun {
		logger.Info("DRY RUN MODE - No messages will be moved")
	}

	// Execute the job with approval function and context
	err = job.ExecuteWithApprovalAndContext(ctx, approvalFunc)
	if err != nil {
		logger.Error("Job execution failed", "error", err)
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

	if config.DstLabel.IsZero() {
		err = fmt.Errorf("destination label is required for organizing moved messages (use -dst-label flag)")
		goto end
	}

	if string(config.SrcEmail) == string(config.DstEmail) && slices.Contains(config.SrcLabels, config.DstLabel) {
		err = fmt.Errorf("source and destination cannot be the same (same email and same label)")
		goto end
	}

end:
	return err
}
