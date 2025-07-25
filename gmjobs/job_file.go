package gmjobs

import (
	"fmt"
	"strings"
)

// JobFile represents a validated job file path
type JobFile string

// ParseJobFile validates and creates a JobFile
func ParseJobFile(path string) (jobFile JobFile, err error) {
	if strings.TrimSpace(path) == "" {
		err = fmt.Errorf("job file path cannot be empty")
		goto end
	}

	// TODO: Add more validation as needed (file extension, path format, etc.)
	jobFile = JobFile(path)

end:
	return jobFile, err
}

// IsZero returns true if the JobFile is empty
func (j JobFile) IsZero() bool {
	return string(j) == ""
}
