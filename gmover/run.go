package gmover

import (
	"fmt"
	"os"

	"github.com/mikeschinkel/gmail-mover/gmutil"
)

type RunMode string

const (
	ShowHelp   RunMode = "show_help"
	ListLabels RunMode = "list_labels"
	MoveEmails RunMode = "move_emails"
)

var validModes = []RunMode{ShowHelp, ListLabels, MoveEmails}

// Config represents the parsed configuration for Gmail Mover operations
type Config struct {
	runMode         RunMode
	jobFile         *string
	srcEmail        *string
	srcLabel        *string
	dstEmail        *string
	dstLabel        *string
	maxMessages     *int64
	dryRun          *bool
	deleteAfterMove *bool
	searchQuery     *string
	autoConfirm     *bool
}

func NewConfig(runMode RunMode) Config {
	return Config{runMode: runMode}
}

// Getter methods
func (c Config) RunMode() RunMode {
	return c.runMode
}

func (c Config) JobFile() string {
	return deRef(c.jobFile)
}

func (c Config) SrcEmail() string {
	return deRef(c.srcEmail)
}

func (c Config) SrcLabel() string {
	return deRef(c.srcLabel)
}

func (c Config) DstEmail() string {
	return deRef(c.dstEmail)
}

func (c Config) DstLabel() string {
	return deRef(c.dstLabel)
}

func (c Config) MaxMessages() int64 {
	return deRef(c.maxMessages)
}

func (c Config) DryRun() bool {
	return deRef(c.dryRun)
}

func (c Config) DeleteAfterMove() bool {
	return deRef(c.deleteAfterMove)
}

func (c Config) SearchQuery() string {
	return deRef(c.searchQuery)
}

func (c Config) AutoConfirm() bool {
	return deRef(c.autoConfirm)
}

// Setter methods
func (c *Config) SetRunMode(mode RunMode) {
	c.runMode = mode
}

func (c *Config) SetJobFile(file string) {
	c.jobFile = toPtr(file)
}

func (c *Config) SetSrcEmail(email string) {
	c.srcEmail = toPtr(email)
}

func (c *Config) SetSrcLabel(label string) {
	c.srcLabel = toPtr(label)
}

func (c *Config) SetDstEmail(email string) {
	c.dstEmail = toPtr(email)
}

func (c *Config) SetDstLabel(label string) {
	c.dstLabel = toPtr(label)
}

func (c *Config) SetMaxMessages(max int64) {
	c.maxMessages = toPtr(max)
}

func (c *Config) SetDryRun(dryRun bool) {
	c.dryRun = toPtr(dryRun)
}

func (c *Config) SetDeleteAfterMove(delete bool) {
	c.deleteAfterMove = toPtr(delete)
}

func (c *Config) SetSearchQuery(query string) {
	c.searchQuery = toPtr(query)
}

func (c *Config) SetAutoConfirm(autoConfirm bool) {
	c.autoConfirm = toPtr(autoConfirm)
}

// Run executes Gmail Mover with the provided configuration
func Run(config *Config) (err error) {
	return RunWithApproval(config, nil)
}

// RunWithApproval executes Gmail Mover with the provided configuration and approval function
func RunWithApproval(config *Config, approvalFunc gmutil.ApprovalFunc) (err error) {
	var job *Job

	ensureLogger()

	// Validate configuration for the specified run mode
	err = validateConfig(config)
	if err != nil {
		goto end
	}

	// Handle different run modes
	switch config.RunMode() {
	case ShowHelp:
		showUsage()
	case ListLabels:
		api := gmutil.NewGMailAPI(ConfigDirName)
		err = api.ListLabels(config.SrcEmail())
		if err != nil {
			logger.Error("Failed to list labels", "error", err)
		}
	case MoveEmails:
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

		// Execute the job with approval function
		err = job.ExecuteWithApproval(approvalFunc)
		if err != nil {
			logger.Error("Job execution failed", "error", err)
		}
	default:
		logger.Error("Not a valid mode",
			"mode_specified", config.RunMode(),
			"valid_modes", validModes,
		)
	}
end:
	return err
}

// validateConfig validates the configuration for the specified run mode
func validateConfig(config *Config) (err error) {
	switch config.RunMode() {
	case ShowHelp:
		// No validation needed for help mode
	case ListLabels:
		err = validateListLabelsConfig(config)
	case MoveEmails:
		err = validateMoveEmailsConfig(config)
	default:
		err = fmt.Errorf("invalid run mode: %s", config.RunMode())
	}
	return err
}

// validateListLabelsConfig validates configuration for list labels mode
func validateListLabelsConfig(config *Config) (err error) {
	if config.SrcEmail() == "" {
		err = fmt.Errorf("source email address is required for listing labels (use -src flag)")
		goto end
	}
	// Note: src-label is NOT required for ListLabels mode

end:
	return err
}

// validateMoveEmailsConfig validates configuration for move emails mode
func validateMoveEmailsConfig(config *Config) (err error) {
	// If using job file, skip individual field validation
	if config.JobFile() != "" {
		goto end
	}

	if config.SrcEmail() == "" {
		err = fmt.Errorf("source email address is required (use -src flag)")
		goto end
	}

	if config.SrcLabel() == "" {
		err = fmt.Errorf("source label is required to prevent accidental mass operations (use -src-label flag, or '*' for all messages)")
		goto end
	}

	if config.DstEmail() == "" {
		err = fmt.Errorf("destination email address is required (use -dst flag)")
		goto end
	}

	if config.DstLabel() == "" {
		err = fmt.Errorf("destination label is required for organizing moved messages (use -dst-label flag)")
		goto end
	}

	if config.SrcEmail() == config.DstEmail() && config.SrcLabel() == config.DstLabel() {
		err = fmt.Errorf("source and destination cannot be the same (same email and same label)")
		goto end
	}

end:
	return err
}

// showUsage displays help information for the Gmail Mover application
func showUsage() {
	fmt.Fprintf(os.Stderr, `Gmail Mover - Move emails between Gmail accounts and labels

USAGE:
    gmail-mover [OPTIONS]

MODES:
    By default, this help is shown. Use one of these flags to specify the mode:

    -list-labels                List available labels for source email address
    -job FILE                   Execute move operation using job configuration file
    -dst EMAIL                  Move emails to destination (requires -src)

COMMON OPTIONS:
    -src EMAIL                  Source Gmail address (required for most operations)
    -src-label LABEL            Source Gmail label (default: "INBOX")
    -dst-label LABEL            Label to apply to moved messages
    -max N                      Maximum messages to process (default: 10000)
    -dry-run                    Show what would be moved without moving (default: false)
    -delete                     Delete from source after move (default: true)
    -query QUERY                Gmail search query to filter messages
    -auto-confirm               Skip interactive confirmation prompts (default: false)

EXAMPLES:
    # Show available labels for an account
    gmail-mover -list-labels -src user@example.com

    # Move emails from INBOX to archive account
    gmail-mover -src user@example.com -dst archive@example.com -dst-label "archived"

    # Dry run with custom query and label
    gmail-mover -src user@example.com -dst backup@example.com -query "from:newsletter" -dry-run

    # Execute from job file
    gmail-mover -job backup-job.json

For more information, visit: https://github.com/mikeschinkel/gmail-mover
`)
}
