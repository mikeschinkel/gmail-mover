package gmcmds

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gmjobs"
	"github.com/mikeschinkel/gmail-mover/gmover"
)

var _ cliutil.Command = (*DefineJobCmd)(nil)

// DefineMoveJobCmd handles creating email move job files from CLI options
type DefineMoveJobCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&DefineMoveJobCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name: "move",
			// CLAUDE: Should we have to define USAGE? Shouldn't we be able to declare the metadata
			//         and then cliutil be able to generate it?
			Usage:       "define move FILE --src=EMAIL --dst=EMAIL [options]",
			Description: "Create an email move job file from command line options",
			ArgDefs: []*cliutil.ArgDef{
				{Name: "filename", Usage: "Job filename", Required: true, String: cfg.JobFile},
			},
			FlagSets: []*cliutil.FlagSet{
				MoveEmailFlagSet,
			},
			// Remove GetFlagSets for now - will be handled by FlagSet() method
		}),
	}, &DefineJobCmd{})
}

// DefineMoveJobCmd implements CommandHandler (executes logic)
var _ cliutil.CommandHandler = (*DefineMoveJobCmd)(nil)

// Handle executes the job define command
func (c *DefineMoveJobCmd) Handle(_ context.Context, config cliutil.Config, _ []string) (err error) {
	var gmCfg *gmover.Config
	var jobSpec *gmover.MoveEmailsJobSpec
	var jobFile gmjobs.JobFile

	gmCfg, err = ConvertConfig(config)
	if err != nil {
		goto end
	}

	jobFile = gmjobs.JobFile(gmCfg.JobFile)
	jobSpec = gmover.NewMoveEmailsJobSpec(gmCfg, jobFile)

	err = gmjobs.Save(jobFile, jobSpec)
	if err != nil {
		goto end
	}

	fmt.Printf("Job file created: %s\n", gmCfg.JobFile)

end:
	return err
}
