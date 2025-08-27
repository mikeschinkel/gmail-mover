package gmover

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// JobFile represents a validated job file path
type JobFile string

// ParseJobFile validates and creates a JobFile
func ParseJobFile(path string, mustExist bool) (JobFile, error) {
	var jobFile JobFile
	var err error

	if path == "" {
		err = fmt.Errorf("job file path cannot be empty")
		goto end
	}

	path = strings.TrimSpace(path)

	// Convert to absolute path
	path, err = filepath.Abs(path)
	if err != nil {
		err = fmt.Errorf("invalid job file path '%s': %w", path, err)
		goto end
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(path), ".json") {
		err = fmt.Errorf("job file must have .json extension: %s", path)
		goto end
	}

	if mustExist {
		// Check if file exists
		_, err = os.Stat(path)
		if err != nil {
			err = fmt.Errorf("job file does not exist '%s': %w", path, err)
			goto end
		}
	}

	jobFile = JobFile(path)

end:
	return jobFile, err
}

// IsZero returns true if the job file path is empty
func (j JobFile) IsZero() bool {
	return string(j) == ""
}
