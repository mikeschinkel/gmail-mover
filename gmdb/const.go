package gmdb

const (
	// Gmail API constants
	MaxBatchSize   = 100
	MaxQueryLength = 1000

	// Sync status constants
	SyncStatusPending    = "pending"
	SyncStatusInProgress = "in_progress"
	SyncStatusCompleted  = "completed"
	SyncStatusFailed     = "failed"

	// Label types
	LabelTypeSystem = "system"
	LabelTypeUser   = "user"

	// Message list visibility
	VisibilityShow = "show"
	VisibilityHide = "hide"

	// Label list visibility
	LabelVisibilityLabelShow         = "labelShow"
	LabelVisibilityLabelHide         = "labelHide"
	LabelVisibilityLabelShowIfUnread = "labelShowIfUnread"
)
