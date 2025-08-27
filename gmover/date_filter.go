package gmover

import (
	"time"
)

// DateFilter represents date-based filtering options
type DateFilter struct {
	Before *time.Time `json:"before,omitempty"`
	After  *time.Time `json:"after,omitempty"`
}

// NewDateFilter creates a new DateFilter with the provided parameters
func NewDateFilter(before, after *time.Time) (dateFilter DateFilter) {
	dateFilter = DateFilter{
		Before: before,
		After:  after,
	}
	return dateFilter
}
