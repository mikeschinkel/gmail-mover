package gmdb

import (
	"time"

	"github.com/mikeschinkel/gmover/gmover"
)

// GmailMessage represents a Gmail message in the database
type GmailMessage struct {
	ID           string
	Subject      string
	From         gmover.EmailAddress
	To           []gmover.EmailAddress
	Labels       []gmover.LabelName
	InternalDate time.Time
	ThreadID     string
	Snippet      string
	Body         string
	Headers      map[string]string
	Attachments  []GmailAttachment
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GmailAttachment represents a Gmail message attachment
type GmailAttachment struct {
	ID           string
	MessageID    string
	Filename     string
	MimeType     string
	Size         int64
	AttachmentID string
	CreatedAt    time.Time
}

// GmailThread represents a Gmail thread
type GmailThread struct {
	ID           string
	Subject      string
	MessageCount int
	Labels       []gmover.LabelName
	LastMessage  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GmailLabel represents a Gmail label
type GmailLabel struct {
	ID                    string
	Name                  gmover.LabelName
	Type                  string // "system" or "user"
	MessageListVisibility string
	LabelListVisibility   string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// SyncState represents the sync state for a Gmail account
type SyncState struct {
	EmailAddress gmover.EmailAddress
	LastSyncTime time.Time
	HistoryID    string
	Status       string
	ErrorCount   int
	LastError    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
