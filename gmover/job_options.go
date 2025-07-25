package gmover

// JobOptions represents options for creating a Job from CLI flags
type JobOptions struct {
	Name            string
	SrcEmail        string
	SrcLabels       []string
	DstEmail        string
	DstLabel        string
	MaxMessages     int64
	DryRun          bool
	DeleteAfterMove bool
	SearchQuery     string
	FailOnError     bool
	LogLevel        string
}

// NewJobOptions creates a new Options with the provided parameters
func NewJobOptions(opts JobOptions) (options JobOptions) {
	if opts.LogLevel == "" {
		opts.LogLevel = "info"
	}
	return opts
}
