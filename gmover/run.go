package gmover

import (
	"github.com/mikeschinkel/gmail-mover/gmutil"
)

type RunMode string

const (
	ListLabels RunMode = "list_labels"
	MoveEmails RunMode = "move_emails"
)

var validModes = []RunMode{ListLabels, MoveEmails}

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

// Run executes Gmail Mover with the provided configuration
func Run(config *Config) (err error) {
	var job *Job

	ensureLogger()

	if config.RunMode() == "" {
		config.SetRunMode(ListLabels)
	}

	// Handle list labels command
	switch config.RunMode() {
	case ListLabels:
		err = gmutil.ListLabels(config.SrcEmail())
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

		err = job.Execute()
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
