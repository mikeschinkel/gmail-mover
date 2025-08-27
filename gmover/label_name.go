package gmover

import (
	"fmt"
	"strings"
)

// LabelName represents a validated Gmail label name
type LabelName string

// ParseLabelName validates and creates a LabelName
func ParseLabelName(label string) (LabelName, error) {
	var labelName LabelName
	var err error

	if label == "" {
		err = fmt.Errorf("label name cannot be empty")
		goto end
	}

	label = strings.TrimSpace(label)

	// Basic validation - Gmail labels can't contain backslashes (forward slashes are allowed for nesting)
	if strings.Contains(label, "\\") {
		err = fmt.Errorf("label name '%s' contains invalid characters (\\)", label)
		goto end
	}

	labelName = LabelName(label)

end:
	return labelName, err
}

// IsZero returns true if the label name is empty
func (l LabelName) IsZero() bool {
	return string(l) == ""
}

// String returns the string representation of the label name
func (l LabelName) String() string {
	return string(l)
}
