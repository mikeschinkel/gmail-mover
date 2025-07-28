package gmjobs

import (
	"errors"
	"fmt"
	"os"
)

// ensureFileDoesNotExist returns nil if file doesn't exist, error if file exists or other problems
// Returns nil if file doesn't exist (safe to create)
// Returns error if file exists or there was a problem checking (permissions, etc.)
func ensureFileDoesNotExist(filename string) (err error) {
	var info os.FileInfo
	info, err = os.Lstat(filename)
	if err == nil {
		// File exists - return error
		err = fmt.Errorf("file already exists: %s", filename)
		goto end
	}

	if errors.Is(err, os.ErrNotExist) {
		// File doesn't exist - that's what we want
		err = nil
		goto end
	}

	if info.IsDir() {
		err = fmt.Errorf("%s is a directory", filename)
	}
	// Some other error occurred (permissions, etc.) - return it

end:
	return err
}
