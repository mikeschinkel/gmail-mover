package gmover

import (
	"fmt"

	"github.com/mikeschinkel/gmail-mover/gapi"
)

// ListLabels executes a list labels operation with the provided configuration
func ListLabels(config *Config) (err error) {
	var api *gapi.GMailAPI

	ensureLogger()

	// Validate configuration for list labels
	err = validateListLabelsConfig(config)
	if err != nil {
		goto end
	}

	api = gapi.NewGMailAPI(ConfigDirName)
	err = api.ListLabels(string(config.SrcEmail))
	if err != nil {
		logger.Error("Failed to list labels", "error", err)
	}

end:
	return err
}

// validateListLabelsConfig validates configuration for list labels mode
func validateListLabelsConfig(config *Config) (err error) {
	if config.SrcEmail.IsZero() {
		err = fmt.Errorf("source email address is required for listing labels (use -src flag)")
		goto end
	}
	// Note: src-label is NOT required for ListLabels mode

end:
	return err
}
