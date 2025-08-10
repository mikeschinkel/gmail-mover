package gmdb

import "errors"

var (
	ErrMessageNotFound   = errors.New("message not found")
	ErrThreadNotFound    = errors.New("thread not found")
	ErrLabelNotFound     = errors.New("label not found")
	ErrSyncStateNotFound = errors.New("sync state not found")
	ErrInvalidEmail      = errors.New("invalid email address")
	ErrInvalidLabel      = errors.New("invalid label name")
	ErrSyncInProgress    = errors.New("sync already in progress")
	ErrSyncFailed        = errors.New("sync failed")
)
