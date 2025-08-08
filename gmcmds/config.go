package gmcmds

import (
	"errors"
	"fmt"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gapi"
	"github.com/mikeschinkel/gmail-mover/gmover"
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

func (c *Config) Config() {}

type GMoverConfigArgs struct {
	JobFileMustExist bool
}

// GMoverConfig converts the gmcmds.Config to a gmover.Config with domain types
func (c *Config) GMoverConfig(args GMoverConfigArgs) (*gmover.Config, error) {
	var config *gmover.Config
	var errs []error
	var err error

	config = gmover.NewConfig()

	// Parse all fields, appending any errors (errors.Join filters out nils)
	if c.JobFile != nil && *c.JobFile != "" {
		config.JobFile, err = gmover.ParseJobFile(*c.JobFile, args.JobFileMustExist)
		errs = append(errs, err)
	}

	if c.SrcEmail != nil && *c.SrcEmail != "" {
		config.SrcEmail, err = gapi.ParseEmailAddress(*c.SrcEmail)
		errs = append(errs, err)
	}

	if c.DstEmail != nil && *c.DstEmail != "" {
		config.DstEmail, err = gapi.ParseEmailAddress(*c.DstEmail)
		errs = append(errs, err)
	}

	if c.SrcLabel != nil && *c.SrcLabel != "" {
		var srcLabel gmover.LabelName
		srcLabel, err = gmover.ParseLabelName(*c.SrcLabel)
		config.SrcLabels = []gmover.LabelName{srcLabel}
		errs = append(errs, err)
	}

	if c.DstLabel != nil && *c.DstLabel != "" {
		var dstLabel gmover.LabelName
		dstLabel, err = gmover.ParseLabelName(*c.DstLabel)
		config.DstLabels = []gmover.LabelName{dstLabel}
		errs = append(errs, err)
	}

	if c.Search != nil {
		config.SearchQuery, err = gmover.ParseSearchQuery(*c.Search)
		errs = append(errs, err)
	}

	if c.MaxMessages != nil {
		config.MaxMessages, err = gmover.ParseMaxMessages(*c.MaxMessages)
		errs = append(errs, err)
	}

	// Set boolean flags
	if c.DryRun != nil {
		config.DryRun = *c.DryRun
	}
	if c.DeleteAfterMove != nil {
		config.DeleteAfterMove = *c.DeleteAfterMove
	}
	if c.AutoConfirm != nil {
		config.AutoConfirm = *c.AutoConfirm
	}

	// If there were errors, return nil config and joined error
	err = errors.Join(errs...)
	if err != nil {
		config = nil
	}

	return config, err
}

// ConvertConfig converts a cliutil.Config to a gmover.Config
func ConvertConfig(config cliutil.Config) (gmc *gmover.Config, err error) {
	cfg, ok := config.(*Config)
	if !ok {
		err = fmt.Errorf("invalid config type")
		goto end
	}
	gmc, err = cfg.GMoverConfig(GMoverConfigArgs{
		JobFileMustExist: false,
	})
end:
	return gmc, err
}
