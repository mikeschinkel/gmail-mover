package gmcmds

import (
	"context"

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
		ArgDefs: []*cliutil.ArgDef{
			{Name: "filename", Usage: "Job filename", Required: true, String: cfg.JobFile},
		},
	})
	cliutil.RegisterCommand(&RunJobCmd{CmdBase: cmd}, &JobCmd{})
}

// RunJobCmd implements CommandHandler (executes logic)
var _ cliutil.CommandHandler = (*RunJobCmd)(nil)

// Handle executes the job run command
func (c *RunJobCmd) Handle(ctx context.Context, config cliutil.Config, args []string) (err error) {
	var job *gmjobs.Job
	var jobCfg gmjobs.Config
	var cfg *Config

	cfg = config.(*Config)

	job, err = gmjobs.LoadJobFile(gmjobs.JobFile(*cfg.JobFile))
	if err != nil {
		goto end
	}

	jobCfg, err = job.ToConfig()
	if err != nil {
		goto end
	}

	// Apply global overrides and run move
	err = gmover.MoveEmails(ctx,
		OverrideGlobals(jobCfg, config),
		gmover.MoveEmailOpts{
			ApprovalFunc: EmailMoverApprover,
		},
	)

end:
	return err
}
