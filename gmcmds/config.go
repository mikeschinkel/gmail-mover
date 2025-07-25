package gmcmds

import (
	"github.com/mikeschinkel/gmail-mover/cliutil"
)

// Singleton instance for CLI command configuration
var config *Config

// GetConfig returns the singleton config instance
func GetConfig() *Config {
	if config == nil {
		config = &Config{
			JobFile:         new(string),
			SrcEmail:        new(string),
			SrcLabel:        new(string),
			DstEmail:        new(string),
			DstLabel:        new(string),
			Search:          new(string),
			MaxMessages:     new(int64),
			DryRun:          new(bool),
			DeleteAfterMove: new(bool),
			AutoConfirm:     new(bool),
		}
	}
	return config
}

var _ cliutil.Config = (*Config)(nil)

// Config represents the parsed configuration for Gmail Mover operations
type Config struct {
	JobFile         *string
	SrcEmail        *string
	SrcLabel        *string
	DstEmail        *string
	DstLabel        *string
	Search          *string
	MaxMessages     *int64
	DryRun          *bool
	DeleteAfterMove *bool
	AutoConfirm     *bool
}

func (c *Config) SetValues(values map[string]any) {
	noop(values)
	//TODO implement me
	panic("implement me")
}

func (c *Config) Config() {}
