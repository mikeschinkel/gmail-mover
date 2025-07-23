package gmover

// DstAccount represents the destination Gmail account configuration
type DstAccount struct {
	Email                string `json:"email"`
	ApplyLabel           string `json:"apply_label,omitempty"`
	CreateLabelIfMissing bool   `json:"create_label_if_missing"`
	PreserveLabels       bool   `json:"preserve_labels"`
}

type DstAccountOpts struct {
	Email                string `json:"email"`
	ApplyLabel           string `json:"apply_label,omitempty"`
	CreateLabelIfMissing bool   `json:"create_label_if_missing"`
	PreserveLabels       bool   `json:"preserve_labels"`
}

// NewDstAccount creates a new DstAccount with the provided parameters
func NewDstAccount(email string, opts DstAccountOpts) DstAccount {
	return DstAccount{
		Email:                email,
		ApplyLabel:           opts.ApplyLabel,
		CreateLabelIfMissing: opts.CreateLabelIfMissing,
		PreserveLabels:       opts.PreserveLabels,
	}
}
