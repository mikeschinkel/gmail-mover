package gmcmds

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/gmail-mover/cliutil"
)

var _ cliutil.Command = (*RunJobCmd)(nil)

// RunJobCmd handles running job files
type RunJobCmd struct {
	*cliutil.CmdBase
}

func init() {
	cmd := cliutil.NewCmdBase(cliutil.CmdArgs{
		Name:        "run",
		Usage:       "run FILE",
		Description: "Execute a job file",
	})
	cliutil.RegisterCommand(&RunJobCmd{CmdBase: cmd}, &JobCmd{})
}

// Handle executes the job run command
func (c *RunJobCmd) Handle(ctx context.Context, config cliutil.Config, args []string) (err error) {
	noop(ctx, config, args)
	// TODO: Need to implement job run logic without directly importing gmover
	return fmt.Errorf("command '%v' not yet implemented in new architecture", c.FullNames())
}
