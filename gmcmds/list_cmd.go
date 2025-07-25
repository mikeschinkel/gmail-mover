package gmcmds

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/gmail-mover/cliutil"
)

// ListCmd handles listing Gmail resources
type ListCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&ListCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "list",
			Usage:       "list",
			Description: "List Gmail resources (default: labels)",
		}),
	})
}

// Handle executes the list command
func (c *ListCmd) Handle(ctx context.Context, config cliutil.Config, args []string) (err error) {
	noop(ctx, config, args)
	// TODO: Need to implement list logic without directly importing gmover
	return fmt.Errorf("command '%v' not yet implemented in new architecture", c.FullNames())

}
