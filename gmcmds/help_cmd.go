package gmcmds

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/gmail-mover/cliutil"
)

// HelpCmd handles showing help information
type HelpCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&HelpCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "help",
			Usage:       "help [command]",
			Description: "Show help information",
		}),
	})
}

// Handle executes the help command
func (c *HelpCmd) Handle(_ context.Context, _ cliutil.Config, args []string) (err error) {

	if len(args) == 0 {
		c.showMainHelp()
		goto end
	}

	err = cliutil.ShowCmdHelp(args[0])

end:
	return err
}

// showMainHelp displays the main help screen
func (c *HelpCmd) showMainHelp() {
	fmt.Printf(`Gmail Mover - Move emails between Gmail accounts and labels

USAGE:
    gmover <command> [subcommand] [options]

COMMANDS:
`)

	// Show all top-level commands
	topCmds := cliutil.GetTopLevelCmds()
	for _, cmd := range topCmds {
		subCmds := cliutil.GetSubCmds(cmd.Name())
		subCmdText := ""
		if len(subCmds) > 0 {
			subCmdText = fmt.Sprintf(" [%s]", subCmds[0].Name()) // Show first subcommand as example
		}
		fmt.Printf("    %-20s %s\n", cmd.Name()+subCmdText, cmd.Description())
	}

	fmt.Printf(`
EXAMPLES:
    # Show help for a specific command
    gmover help list
    gmover help move

    # List available labels
    gmover list --src=user@example.com

    # Move emails  
    gmover move --src=user@example.com --dst=archive@example.com --src-label="INBOX" --dst-label="archived"

    # Job operations
    gmover job define daily-archive.json --src=user@example.com --dst=archive@example.com
    gmover job run daily-archive.json --auto-confirm

For more information, visit: https://github.com/mikeschinkel/gmail-mover
`)
}
