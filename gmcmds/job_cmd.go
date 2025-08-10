package gmcmds

import (
	"github.com/mikeschinkel/gmover/cliutil"
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
