package gmcmds

import (
	"context"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gmover"
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

// MoveCmd implements CommandHandler (executes logic)
var _ cliutil.CommandHandler = (*MoveCmd)(nil)

// Handle executes the move command
func (c *MoveCmd) Handle(ctx context.Context, config cliutil.Config, _ []string) (err error) {
	var gmCfg *gmover.Config

	gmCfg, err = ConvertConfig(config)
	if err != nil {
		goto end
	}

	err = gmover.MoveEmails(ctx, gmCfg,
		gmover.MoveEmailOpts{
			ApprovalFunc: EmailMoverApprover,
		},
	)

end:
	return err
}

func EmailMoverApprover(prompt string) (approved bool, approveAll bool, err error) {
	panic("IMPLEMENT ME")
	return approved, approveAll, err
}
