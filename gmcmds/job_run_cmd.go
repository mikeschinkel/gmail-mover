package gmcmds

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gmjobs"
	"github.com/mikeschinkel/gmail-mover/gmover"
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

// RunJobCmd implements CommandHandler (executes logic)
var _ cliutil.CommandHandler = (*RunJobCmd)(nil)

// Handle executes the job run command
func (c *RunJobCmd) Handle(ctx context.Context, config cliutil.Config, args []string) (err error) {
	var filename string
	var job *gmjobs.Job
	var gmCfg gmjobs.Config

	if len(args) < 1 {
		err = fmt.Errorf("job filename required")
		goto end
	}
	filename = args[0]

	job, err = gmjobs.LoadJobFile(filename)
	if err != nil {
		goto end
	}

	gmCfg, err = job.ToConfig()
	if err != nil {
		goto end
	}

	// Apply global overrides and run move
	err = gmover.MoveEmails(ctx, OverrideGlobals(gmCfg, config))

end:
	return err
}
