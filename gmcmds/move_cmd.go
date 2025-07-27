package gmcmds

import (
	"context"
	"fmt"
	"os"

	"github.com/mikeschinkel/gmail-mover/cliutil"
	"github.com/mikeschinkel/gmail-mover/gapi"
	"github.com/mikeschinkel/gmail-mover/gmover"
	"golang.org/x/term"
)

var MoveEmailFlagSet = &cliutil.FlagSet{
	Name: "move_email",
	FlagDefs: []cliutil.FlagDef{
		{Name: "src", Default: "", Usage: "Source Gmail address", Required: true, String: cfg.SrcEmail},
		{Name: "dst", Default: "", Usage: "Destination Gmail address", Required: true, String: cfg.DstEmail},
		{Name: "src-label", Default: "INBOX", Usage: "Source Gmail label", Required: true, String: cfg.SrcLabel},
		{Name: "dst-label", Usage: "Destination label", Required: true, String: cfg.DstLabel},
		{Name: "search", Usage: "Gmail search query", String: cfg.Search},
		{Name: "max", Default: int64(10000), Usage: "Maximum messages to process", Int64: cfg.MaxMessages},
		{Name: "delete", Default: true, Usage: "Delete from source after move", Bool: cfg.DeleteAfterMove},
	},
}

// MoveCmd handles moving emails between accounts/labels
type MoveCmd struct {
	*cliutil.CmdBase
}

func init() {
	cliutil.RegisterCommand(&MoveCmd{
		CmdBase: cliutil.NewCmdBase(cliutil.CmdArgs{
			Name:        "move",
			Usage:       "move --src=EMAIL --dst=EMAIL [options]",
			Description: "Move emails between accounts/labels",
			FlagSets: []*cliutil.FlagSet{
				MoveEmailFlagSet,
			},
		}),
	})
}

// MoveCmd implements CommandHandler (executes logic)
var _ cliutil.CommandHandler = (*MoveCmd)(nil)

// Handle executes the move command
func (c *MoveCmd) Handle(ctx context.Context, config cliutil.Config, _ []string) (err error) {
	var gmCfg *gmover.Config

	ensureLogger()

	gmCfg, err = ConvertConfig(config)
	if err != nil {
		goto end
	}

	err = gmover.MoveEmails(ctx, gmCfg,
		gmover.MoveEmailOpts{
			ApprovalFunc: EmailMoverApprover,
		},
	)

end:
	return err
}

func EmailMoverApprover(ctx context.Context, prompt string) (approved bool, approveAll bool, err error) {
	var char byte
	var buffer [1]byte
	var stdinFd int
	var oldState *term.State

	ensureLogger()

	fmt.Print(prompt + "\nMove Message? [ Y(es) / N(o) / A(ll) / C(ancel) ]: ")

	// Try raw mode for single character input
	stdinFd = int(os.Stdin.Fd())
	oldState, err = term.MakeRaw(stdinFd)
	if err != nil {
		// Fallback for non-TTY environments (like GoLand console)
		if gapi.IsTerminalError(err) {
			fmt.Print("\n(Running in non-TTY environment; type choice and press Enter): ")
			approved, approveAll, err = fallbackLineInput(ctx)
			goto end
		}
		goto end
	}

	// Read single character
	_, err = os.Stdin.Read(buffer[:])
	if err != nil {
		goto end
	}
	// Restore terminal immediately after reading
	defer must(term.Restore(stdinFd, oldState))
	_ = term.Restore(stdinFd, oldState)

	char = buffer[0]

	// Echo the character and add newline (now in normal mode)
	if char != 3 { // Don't echo Ctrl-C
		fmt.Printf("%c\n", char)
	} else {
		fmt.Println("^C")
	}

	// Process the input
	switch char {
	case 'y', 'Y':
		approved = true
	case 'a', 'A':
		approved = true
		approveAll = true
	case 'n', 'N':
		approved = false
	case 'c', 'C', 3: // 3 is Ctrl-C in raw mode
		err = context.Canceled
	default:
		err = fmt.Errorf("invalid input: %c (expected Y/N/A/C or Ctrl-C)", char)
	}

end:
	return approved, approveAll, err
}

func fallbackLineInput(ctx context.Context) (approved bool, approveAll bool, err error) {
	var input string
	var inputChan chan string
	var errChan chan error

	inputChan = make(chan string, 1)
	errChan = make(chan error, 1)

	// Read input in a goroutine
	go func() {
		var scanInput string
		_, scanErr := fmt.Scanln(&scanInput)
		if scanErr != nil {
			errChan <- scanErr
			return
		}
		inputChan <- scanInput
	}()

	// Wait for either input or cancellation
	select {
	case <-ctx.Done():
		fmt.Println() // Add newline for clean output
		err = ctx.Err()
		goto end
	case err = <-errChan:
		goto end
	case input = <-inputChan:
		// Continue with processing
	}

	// Process the input (accept first character or full words)
	if len(input) > 0 {
		switch input[0] {
		case 'y', 'Y':
			approved = true
		case 'a', 'A':
			approved = true
			approveAll = true
		case 'n', 'N':
			approved = false
		case 'c', 'C':
			err = context.Canceled
		default:
			err = fmt.Errorf("invalid input: %s (expected Y/N/A/C)", input)
		}
	}

end:
	return approved, approveAll, err
}
