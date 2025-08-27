package gmover

import (
	"github.com/mikeschinkel/gmover/gmcfg"
)

var AppName = "gmover"

// Config represents the parsed configuration for Gmail Mover operations
type Config struct {
	JobFile         JobFile
	SrcEmail        EmailAddress
	SrcLabels       []LabelName
	DstEmail        EmailAddress
	DstLabels       []LabelName
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

func ConfigFileStore() *gmcfg.FileStore {
	return gmcfg.NewFileStore(AppName)
}
