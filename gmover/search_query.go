package gmover

import (
	"strings"
)

// SearchQuery represents a Gmail search query
type SearchQuery string

// ParseSearchQuery validates and creates a SearchQuery
func ParseSearchQuery(query string) (SearchQuery, error) {
	// Trim whitespace
	query = strings.TrimSpace(query)

	// Empty queries are valid (means no filtering)
	return SearchQuery(query), nil
}

// IsZero returns true if the search query is empty
func (s SearchQuery) IsZero() bool {
	return string(s) == ""
}
