package gmcmds

import (
	"context"

	"github.com/mikeschinkel/gmover/cliutil"
	"github.com/mikeschinkel/gmover/gmover"
)

var SyncFlagSet = &cliutil.FlagSet{
	Name: "sync",
	FlagDefs: []cliutil.FlagDef{
		{Name: "account", Usage: "Gmail account to sync (required)", Required: true, String: cfg.SrcEmail},
		{Name: "db", Default: "", Usage: "Database name from config (default: use config default)", String: cfg.DstEmail}, // Reusing DstEmail field temporarily
		{Name: "label", Usage: "Specific Gmail label to sync (optional, default: full account)", String: cfg.SrcLabel},
		{Name: "query", Usage: "Gmail search query to filter sync (optional)", String: cfg.Search},
		{Name: "force", Default: false, Usage: "Force full resync, ignoring previous state", Bool: cfg.DeleteAfterMove}, // Reusing DeleteAfterMove temporarily
		{Name: "dry-run", Default: false, Usage: "Preview sync without writing to database", Bool: cfg.DryRun},
	},
}

// SyncCmd handles syncing Gmail account to SQLite database
type SyncCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&SyncCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "sync",
			Usage:       "sync --account=EMAIL [options]",
			Description: "Synchronize Gmail account to local SQLite database archive",
			FlagSets: []*cliutil.FlagSet{
				SyncFlagSet,
			},
		}),
	})
}

// SyncCmd implements CommandHandler (executes logic)
var _ cliutil.CommandHandler = (*SyncCmd)(nil)

// Handle executes the sync command
func (c *SyncCmd) Handle(ctx context.Context, config cliutil.Config, _ []string) (err error) {
	var gmCfg *gmover.Config
	var syncOpts gmover.SyncOptions

	ensureLogger()

	gmCfg, err = ConvertConfig(config)
	if err != nil {
		goto end
	}

	// Execute sync operation
	syncOpts = gmover.SyncOptions{
		Account:   gmCfg.SrcEmail,
		DBName:    gmCfg.DstEmail.String(), // Temporarily using DstEmail field for DB name
		Query:     gmCfg.SearchQuery.String(),
		Force:     gmCfg.DeleteAfterMove, // Temporarily reusing this field for force flag
		DryRun:    gmCfg.DryRun,
		BatchSize: 100, // Default batch size
	}

	// Handle labels - use first label if available, empty for full account sync
	if len(gmCfg.SrcLabels) > 0 {
		syncOpts.Label = gmCfg.SrcLabels[0].String()
	} else {
		syncOpts.Label = ""
	}

	// Handle empty DB name (use default)
	if syncOpts.DBName == "" {
		syncOpts.DBName = "default"
	}

	_, err = gmover.RunSync(ctx, syncOpts)

end:
	return err
}
