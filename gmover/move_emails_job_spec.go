package gmover

import (
	"fmt"
	"strings"

	"github.com/mikeschinkel/gmail-mover/gmjobs"
)

func init() {
	gmjobs.RegisterJobSpec(&MoveEmailsJobSpec{})
}

// MoveEmailsJobSpec represents the specification for moving emails between accounts
type MoveEmailsJobSpec struct {
	SrcEmail        EmailAddress `json:"src_email"`
	DstEmail        EmailAddress `json:"dst_email"`
	SrcLabels       []LabelName  `json:"src_labels,omitempty"`
	DstLabel        LabelName    `json:"dst_label,omitempty"`
	SearchQuery     SearchQuery  `json:"search_query,omitempty"`
	MaxMessages     MaxMessages  `json:"max_messages,omitempty"`
	DryRun          bool         `json:"dry_run,omitempty"`
	DeleteAfterMove bool         `json:"delete_after_move,omitempty"`
}

// JobType returns the job type identifier
func (m *MoveEmailsJobSpec) JobType() string {
	return "move_emails"
}

// Name returns a descriptive name for this job
func (m *MoveEmailsJobSpec) Name() string {
	var srcLabels, dstLabel string

	if len(m.SrcLabels) > 0 {
		labelStrs := make([]string, len(m.SrcLabels))
		for i, label := range m.SrcLabels {
			labelStrs[i] = string(label)
		}
		srcLabels = fmt.Sprintf("[%s]", strings.Join(labelStrs, ","))
	}

	if !m.DstLabel.IsZero() {
		dstLabel = fmt.Sprintf("[%s]", string(m.DstLabel))
	}

	return fmt.Sprintf("Move emails from %s%s to %s%s",
		m.SrcEmail, srcLabels, m.DstEmail, dstLabel)
}

// NewMoveEmailsJobSpec creates a new job spec from a gmover config
func NewMoveEmailsJobSpec(config *Config) *MoveEmailsJobSpec {
	return &MoveEmailsJobSpec{
		SrcEmail:        config.SrcEmail,
		DstEmail:        config.DstEmail,
		SrcLabels:       config.SrcLabels,
		DstLabel:        config.DstLabel,
		SearchQuery:     config.SearchQuery,
		MaxMessages:     config.MaxMessages,
		DryRun:          config.DryRun,
		DeleteAfterMove: config.DeleteAfterMove,
	}
}

// ToConfig converts the job spec to gmover config - direct mapping since types match
func (m *MoveEmailsJobSpec) ToConfig() (config gmjobs.Config, err error) {
	config = &Config{
		SrcEmail:        m.SrcEmail,
		DstEmail:        m.DstEmail,
		SrcLabels:       m.SrcLabels,
		DstLabel:        m.DstLabel,
		SearchQuery:     m.SearchQuery,
		MaxMessages:     m.MaxMessages,
		DryRun:          m.DryRun,
		DeleteAfterMove: m.DeleteAfterMove,
		AutoConfirm:     false, // Jobs don't support interactive confirmation
	}

	return config, err
}
