package gapi

import "time"

// MessageInfo contains details about a message for approval decisions
type MessageInfo struct {
	Subject string
	From    string
	To      string
	Date    time.Time
	ID      string
}

// ApprovalFunc is called before moving each message
// Returns: approved (move this message), approveAll (auto-approve remaining), error
type ApprovalFunc func(msg MessageInfo) (approved bool, approveAll bool, err error)

// GMailAPI provides Gmail API operations for a specific app configuration
type GMailAPI struct {
	appConfigDir string
	ApprovalFunc ApprovalFunc // Optional - if nil, auto-approve all messages
}

// NewGMailAPI creates a new GMailAPI instance for the specified app config directory
func NewGMailAPI(appConfigDir string) *GMailAPI {
	return &GMailAPI{appConfigDir: appConfigDir}
}
