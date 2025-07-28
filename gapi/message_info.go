package gapi

import (
	"fmt"
	"time"
)

// MessageInfo contains details about a message for approval decisions
type MessageInfo struct {
	Subject    string
	From       string
	To         string
	Date       time.Time
	Id         string
	Moved      bool
	Deleted    bool
	Labeled    bool
	DateParsed bool
}

func (info MessageInfo) String() string {
	return fmt.Sprintf("Id: %s, From: %s, To: %s, Date: %s, Subject: %s",
		info.Id,
		info.From,
		info.To,
		info.Date.Format(time.DateOnly),
		info.Subject,
	)
}
