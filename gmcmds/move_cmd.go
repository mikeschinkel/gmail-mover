package gmcmds

import (
	"context"

	"github.com/mikeschinkel/gmail-mover/cliutil"
)

var MoveEmailFlagSet = &cliutil.FlagSet{
	Name: "move_email",
	FlagDefs: []cliutil.FlagDef{
		{Name: "src", Default: "", Usage: "Source Gmail address", Required: true, String: cfg.SrcEmail},
		{Name: "dst", Default: "", Usage: "Destination Gmail address", Required: true, String: cfg.DstEmail},
		{Name: "src-label", Default: "INBOX", Usage: "Source Gmail label", Required: true, String: cfg.SrcLabel},
		{Name: "dst-label", Usage: "Destination label", Required: true, String: cfg.DstLabel},
		{Name: "search", Usage: "Gmail search query", String: cfg.Search},
		{Name: "max", Default: int64(10000), Usage: "Maximum messages to process", Int64: cfg.MaxMessages},
		{Name: "delete", Default: true, Usage: "Delete from source after move", Bool: cfg.DeleteAfterMove},
	},
}

// MoveCmd handles moving emails between accounts/labels
type MoveCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&MoveCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "move",
			Usage:       "move --src=EMAIL --dst=EMAIL [options]",
			Description: "Move emails between accounts/labels",
			FlagSets: []*cliutil.FlagSet{
				MoveEmailFlagSet,
			},
		}),
	})
}

// Handle executes the move command
func (c *MoveCmd) Handle(ctx context.Context, config cliutil.Config, args []string) (err error) {
	noop(ctx, config, args)
	// TODO: Need to implement move logic without directly importing gmover
	// This will require refactoring the business logic to accept config as parameter
	return nil
}
