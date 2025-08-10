package gmcmds

import (
	"github.com/mikeschinkel/gmover/cliutil"
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
