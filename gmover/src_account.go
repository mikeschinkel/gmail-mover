package gmover

// SrcAccount represents the source Gmail account configuration
type SrcAccount struct {
	Email       string     `json:"email"`
	Labels      []string   `json:"labels"`
	Query       string     `json:"query,omitempty"`
	MaxMessages int64      `json:"max_messages"`
	DateFilter  DateFilter `json:"date_filter,omitempty"`
}

type SrcAccountOpts struct {
	Labels      []string   `json:"labels"`
	Query       string     `json:"query,omitempty"`
	MaxMessages int64      `json:"max_messages"`
	DateFilter  DateFilter `json:"date_filter,omitempty"`
}

// NewSrcAccount creates a new Source with the provided parameters
func NewSrcAccount(email string, opts SrcAccountOpts) (source SrcAccount) {
	if len(opts.Labels) == 0 {
		opts.Labels = []string{"INBOX"}
	}
	source = SrcAccount{
		Email:       email,
		Labels:      opts.Labels,
		Query:       opts.Query,
		MaxMessages: opts.MaxMessages,
		DateFilter:  opts.DateFilter,
	}
	return source
}
