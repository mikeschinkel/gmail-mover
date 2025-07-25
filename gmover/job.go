package gmover

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mikeschinkel/gmail-mover/gapi"
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
			Labels:      opts.SrcLabels,
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
func (job *Job) ExecuteWithApproval(approvalFunc gapi.ApprovalFunc) (err error) {
	return job.ExecuteWithApprovalAndContext(context.Background(), approvalFunc)
}

// ExecuteWithApprovalAndContext runs the job with context and optional approval function
func (job *Job) ExecuteWithApprovalAndContext(ctx context.Context, approvalFunc gapi.ApprovalFunc) (err error) {
	var labelsToApply []string
	var api *gapi.GMailAPI

	if job.DstAccount.ApplyLabel != "" {
		labelsToApply = []string{job.DstAccount.ApplyLabel}
	}

	opts := gapi.TransferOpts{
		Labels:          job.SrcAccount.Labels,
		SearchQuery:     job.SrcAccount.Query,
		MaxMessages:     int(job.SrcAccount.MaxMessages),
		DryRun:          job.Options.DryRun,
		DeleteAfterMove: job.Options.DeleteAfterMove,
		LabelsToApply:   labelsToApply,
		FailOnError:     !job.Options.FailOnError, // Note: inverted logic
	}

	api = gapi.NewGMailAPI(ConfigDirName)
	// CLAUDE: I am wondering if approveFunc should be in api or opts?
	//         Argue pros and cons of each then let me decide.
	api.ApprovalFunc = approvalFunc
	// TODO: Create TransferMessagesWithOptsAndContext that accepts context
	noop(ctx) // Context will be used when gapi supports it
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
	if !config.JobFile.IsZero() {
		job, err = LoadJob(string(config.JobFile))
		goto end
	}

	job, err = NewJob(JobOptions{
		Name:            "CLI Job",
		SrcEmail:        string(config.SrcEmail),
		SrcLabels:       stringSlice(config.SrcLabels),
		DstEmail:        string(config.DstEmail),
		DstLabel:        string(config.DstLabel),
		MaxMessages:     int64(config.MaxMessages),
		DryRun:          config.DryRun,
		DeleteAfterMove: config.DeleteAfterMove,
		SearchQuery:     string(config.SearchQuery),
		FailOnError:     false,
		LogLevel:        "info",
	})

end:
	return job, err
}

func stringSlice[T ~string](tt []T) []string {
	ss := make([]string, len(tt))
	for i, t := range tt {
		ss[i] = string(t)
	}
	return ss
}
