package gmover

import (
	"github.com/mikeschinkel/gmail-mover/gmjobs"
)

// Config represents the parsed configuration for Gmail Mover operations
type Config struct {
	JobFile         gmjobs.JobFile
	SrcEmail        EmailAddress
	SrcLabels       []LabelName
	DstEmail        EmailAddress
	DstLabel        LabelName
	MaxMessages     MaxMessages
	DryRun          bool
	DeleteAfterMove bool
	SearchQuery     SearchQuery
	AutoConfirm     bool
}

func (c *Config) Config() {}

func NewConfig() *Config {
	return &Config{}
}

// Singleton instance for CLI command configuration
var globalConfig *Config

// GetConfig returns the singleton config instance
//
//goland:noinspection GoUnusedExportedFunction
func GetConfig() *Config {
	if globalConfig == nil {
		globalConfig = NewConfig()
	}
	return globalConfig
}

// ResetConfig resets the singleton for testing
//
//goland:noinspection GoUnusedExportedFunction
func ResetConfig() {
	globalConfig = nil
}
