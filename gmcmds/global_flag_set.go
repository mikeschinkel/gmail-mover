package gmcmds

import (
	"github.com/mikeschinkel/gmail-mover/cliutil"
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
