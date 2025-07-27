package gmcmds

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/gmail-mover/cliutil"
)

var _ cliutil.Command = (*JobCmd)(nil)

// JobCmd handles job operations (parent command for run/define)
type JobCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&JobCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "job",
			Usage:       "job [run|define] [options]",
			Description: "Job operations",
		}),
	})
}

/*
// Handle executes the job command
// Default subcommand routing is now handled declaratively by the framework
func (c *JobCmd) Handle(ctx context.Context, config cliutil.Config, args []string) (err error) {
	noop(ctx, config, args)
	// This should not be reached in normal operation since the framework
	// should route to the default subcommand. If we get here, show help.
	err = fmt.Errorf("usage: %s", c.Usage())
	return err
}
*/
