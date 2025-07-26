package gapi

import (
	"fmt"
	"time"
)

// MessageInfo contains details about a message for approval decisions
type MessageInfo struct {
	Subject string
	From    string
	To      string
	Date    time.Time
	Id      string
}

func (info MessageInfo) String() string {
	return fmt.Sprintf("Id: %s, Subject: %s, From: %s", info.Id, info.Subject, info.From)
}

// ApprovalFunc is called before moving each message
// Returns: approved (move this message), approveAll (auto-approve remaining), error
type ApprovalFunc func(prompt string) (approved bool, approveAll bool, err error)

// GMailAPI provides Gmail API operations for a specific app configuration
type GMailAPI struct {
	appConfigDir string
}

// NewGMailAPI creates a new GMailAPI instance for the specified app config directory
func NewGMailAPI(appConfigDir string) *GMailAPI {
	return &GMailAPI{appConfigDir: appConfigDir}
}
