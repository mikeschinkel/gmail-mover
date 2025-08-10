package gmcmds

import (
	"github.com/mikeschinkel/gmover/cliutil"
)

// ListCmd handles listing Gmail resources
type ListCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&ListCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "list",
			Usage:       "list [subcommand]",
			Description: "List Gmail resources (default: labels)",
			DelegateTo:  (*ListLabelsCmd)(nil),
		}),
	})
}

// ListCmd implements Command only (delegates to subcommands)
var _ cliutil.Command = (*ListCmd)(nil)
