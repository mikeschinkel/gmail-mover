package gmover

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mikeschinkel/gmail-mover/gmutil"
)

const (
	ConfigDirName = "gmail-mover"
)

// Job represents a complete job configuration for moving Gmail messages
type Job struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	SrcAccount  SrcAccount `json:"src_account"`
	DstAccount  DstAccount `json:"dst_account"`
	Options     JobOptions `json:"options"`
}

// NewJob creates a new Job with the provided options
func NewJob(opts JobOptions) (job *Job, err error) {
	if opts.SrcEmail == "" || opts.DstEmail == "" {
		err = fmt.Errorf("source and dest email addresses are required")
		goto end
	}

	job = &Job{
		Name:    opts.Name,
		Options: NewJobOptions(opts),
		SrcAccount: NewSrcAccount(opts.SrcEmail, SrcAccountOpts{
			Labels:      []string{opts.SrcLabel},
			Query:       opts.SearchQuery,
			MaxMessages: opts.MaxMessages,
		}),
		DstAccount: NewDstAccount(opts.DstEmail, DstAccountOpts{
			ApplyLabel:           opts.DstLabel,
			CreateLabelIfMissing: true,
		}),
	}

end:
	return job, err
}

// Execute runs the job by moving messages from source to destination
func (job *Job) Execute() (err error) {
	return job.ExecuteWithApproval(nil)
}

// ExecuteWithApproval runs the job with an optional approval function
func (job *Job) ExecuteWithApproval(approvalFunc gmutil.ApprovalFunc) (err error) {
	var labelsToApply []string
	var api *gmutil.GMailAPI

	if job.DstAccount.ApplyLabel != "" {
		labelsToApply = []string{job.DstAccount.ApplyLabel}
	}

	opts := gmutil.TransferOpts{
		Labels:          job.SrcAccount.Labels,
		SearchQuery:     job.SrcAccount.Query,
		MaxMessages:     int(job.SrcAccount.MaxMessages),
		DryRun:          job.Options.DryRun,
		DeleteAfterMove: job.Options.DeleteAfterMove,
		LabelsToApply:   labelsToApply,
		FailOnError:     !job.Options.FailOnError, // Note: inverted logic
	}

	api = gmutil.NewGMailAPI(ConfigDirName)
	api.ApprovalFunc = approvalFunc
	return api.TransferMessagesWithOpts(job.SrcAccount.Email, job.DstAccount.Email, opts)
}

// LoadJob loads a job configuration from a JSON file
func LoadJob(filepath string) (job *Job, err error) {
	var data []byte

	data, err = os.ReadFile(filepath)
	if err != nil {
		goto end
	}

	job = &Job{}
	err = json.Unmarshal(data, job)
	if err != nil {
		goto end
	}

	// Set defaults
	if len(job.SrcAccount.Labels) == 0 {
		job.SrcAccount.Labels = []string{"INBOX"}
	}
	if job.SrcAccount.MaxMessages == 0 {
		job.SrcAccount.MaxMessages = 1000
	}
	if job.Options.LogLevel == "" {
		job.Options.LogLevel = "info"
	}

end:
	return job, err
}

func GetJob(config Config) (job *Job, err error) {

	// Load job configuration
	if config.JobFile() != "" {
		job, err = LoadJob(config.JobFile())
		goto end
	}

	// TODO: This is AI generated code smell; disinfect it
	job, err = NewJob(JobOptions{
		Name:            "CLI Job",
		SrcEmail:        config.SrcEmail(),
		SrcLabel:        config.SrcLabel(),
		DstEmail:        config.DstEmail(),
		DstLabel:        config.DstLabel(),
		MaxMessages:     config.MaxMessages(),
		DryRun:          config.DryRun(),
		DeleteAfterMove: config.DeleteAfterMove(),
		SearchQuery:     config.SearchQuery(),
		FailOnError:     false,
		LogLevel:        "info",
	})

end:
	return job, err
}
