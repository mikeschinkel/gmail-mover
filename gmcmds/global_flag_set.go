package gmcmds

import (
	"github.com/mikeschinkel/gmover/cliutil"
	"github.com/mikeschinkel/gmover/gmjobs"
	"github.com/mikeschinkel/gmover/gmover"
)

var cfg = GetConfig()

var (
	GlobalFlagSet = &cliutil.FlagSet{
		Name: "global",
		FlagDefs: []cliutil.FlagDef{
			{Name: "auto-confirm", Default: false, Usage: "Skip interactive confirmation prompts", Bool: cfg.AutoConfirm},
			{Name: "dry-run", Default: false, Usage: "Show what would happen without executing", Bool: cfg.DryRun},
		},
	}
)

// OverrideGlobals applies global flags to job config functionally
func OverrideGlobals(jobConfig gmjobs.Config, globals cliutil.Config) *gmover.Config {
	var config gmover.Config
	var gc *Config

	config = *jobConfig.(*gmover.Config) // copy
	gc = globals.(*Config)

	if gc.DryRun != nil {
		config.DryRun = *gc.DryRun
	}
	if gc.AutoConfirm != nil {
		config.AutoConfirm = *gc.AutoConfirm
	}

	return &config
}
