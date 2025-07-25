package gmcmds

import (
	"context"

	"github.com/mikeschinkel/gmail-mover/cliutil"
)

var _ cliutil.Command = (*DefineJobCmd)(nil)

// DefineJobCmd handles creating job files from CLI options
type DefineJobCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&DefineJobCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name: "define",
			// CLAUDE: Should we have to define USAGE? Shouldn't we be able to declare the metdadata
			//         and then cliutil be able to generate it?
			Usage:       "define [move] [args] [options]",
			Description: "Create a job file from command line options",
			// Remove GetFlagSets for now - will be handled by FlagSet() method
		}),
	}, &JobCmd{})
}

// Handle executes the job define command
func (c *DefineJobCmd) Handle(ctx context.Context, config cliutil.Config, args []string) (err error) {
	noop(ctx, config, args)
	// TODO: Need to implement job define logic without directly importing gmover
	// This will require refactoring to work with the cliutil.Config interface
	panic("Implement show help on commands not meant to be handled")
}
