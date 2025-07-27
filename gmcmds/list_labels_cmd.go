package gmcmds

import (
	"context"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gmover"
)

var ListLabelsFlagSet = &cliutil.FlagSet{
	Name: "list_labels",
	FlagDefs: []cliutil.FlagDef{
		{Name: "src", Default: "", Usage: "Source Gmail address", Required: true, String: cfg.SrcEmail},
	},
}

// ListLabelsCmd handles listing Gmail labels
type ListLabelsCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&ListLabelsCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "labels",
			Usage:       "list labels --src=EMAIL",
			Description: "List available labels for a Gmail account",
			FlagSets: []*cliutil.FlagSet{
				ListLabelsFlagSet,
			},
		}),
	}, &ListCmd{})
}

// ListLabelsCmd implements CommandHandler (executes logic)
var _ cliutil.CommandHandler = (*ListLabelsCmd)(nil)

// Handle executes the list labels command
func (c *ListLabelsCmd) Handle(_ context.Context, config cliutil.Config, _ []string) (err error) {
	var gmCfg *gmover.Config

	ensureLogger()

	gmCfg, err = ConvertConfig(config)
	if err != nil {
		goto end
	}

	err = gmover.ListLabels(gmCfg)

end:
	return err
}
