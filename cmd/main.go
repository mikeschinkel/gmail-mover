package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/mikeschinkel/gmail-mover/gmover"
	"github.com/mikeschinkel/gmail-mover/gmutil"
)

func main() {
	var err error
	var handler *CLIHandler
	var logger *slog.Logger
	var config gmover.Config

	// Initialize CLI-friendly slog logger
	handler = NewCLIHandler()
	logger = slog.New(handler)

	// Initialize gmover package
	err = gmover.Initialize(&gmover.Opts{
		Logger: logger,
	})
	if err != nil {
		logger.Error("Failed to initialize", "error", err)
		os.Exit(1)
	}

	config = parseFlags()
	
	// Create approval function based on auto-confirm setting
	approvalFunc := createApprovalFunc(config.AutoConfirm())
	
	err = gmover.RunWithApproval(&config, approvalFunc)
	if err != nil {
		logger.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

// parseFlags handles all command-line flag parsing and returns a Config
func parseFlags() (config gmover.Config) {
	config = gmover.NewConfig(gmover.ShowHelp)

	listLabels := flag.Bool("list-labels", false, "List available labels for source email address")

	jobFile := flag.String("job", "", "Path to job configuration file")
	srcEmail := flag.String("src", "", "Source Gmail address")
	srcLabel := flag.String("src-label", "INBOX", "Source Gmail label")
	dstEmail := flag.String("dst", "", "Destination Gmail address")
	dstLabel := flag.String("dst-label", "", "Label to apply to moved messages")
	maxMessages := flag.Int64("max", 10000, "Maximum messages to process")
	dryRun := flag.Bool("dry-run", false, "Show what would be moved without moving")
	deleteAfterMove := flag.Bool("delete", true, "Delete from source after move")
	searchQuery := flag.String("query", "", "Gmail search query")
	autoConfirm := flag.Bool("auto-confirm", false, "Skip interactive confirmation prompts")

	flag.Parse()

	// Set config values using setter methods
	config.SetJobFile(*jobFile)
	config.SetSrcEmail(*srcEmail)
	config.SetSrcLabel(*srcLabel)
	config.SetDstEmail(*dstEmail)
	config.SetDstLabel(*dstLabel)
	config.SetMaxMessages(*maxMessages)
	config.SetDryRun(*dryRun)
	config.SetDeleteAfterMove(*deleteAfterMove)
	config.SetSearchQuery(*searchQuery)
	config.SetAutoConfirm(*autoConfirm)

	// Determine run mode based on provided flags
	if *listLabels {
		config.SetRunMode(gmover.ListLabels)
	} else if *jobFile != "" {
		// Job file explicitly indicates move operation
		config.SetRunMode(gmover.MoveEmails)
	} else if *dstEmail != "" {
		// Destination email explicitly indicates move operation
		config.SetRunMode(gmover.MoveEmails)
	}
	// Otherwise, stay in ShowHelp mode (default)
	
	return config
}

// createApprovalFunc creates an approval function based on autoConfirm setting
func createApprovalFunc(autoConfirm bool) gmutil.ApprovalFunc {
	if autoConfirm {
		return nil // Auto-approve everything
	}
	
	return func(msg gmutil.MessageInfo) (approved bool, approveAll bool, err error) {
		var input string
		var scanner *bufio.Scanner
		
		// Display message details
		fmt.Printf("\n--- Message Details ---\n")
		fmt.Printf("Subject: %s\n", msg.Subject)
		fmt.Printf("From: %s\n", msg.From)
		fmt.Printf("To: %s\n", msg.To)
		fmt.Printf("Date: %s\n", msg.Date.Format("2006-01-02 15:04:05"))
		fmt.Printf("ID: %s\n", msg.ID)
		
		// Prompt for approval
		for {
			fmt.Print("Move this message? [Y/n/a]: ")
			scanner = bufio.NewScanner(os.Stdin)
			if !scanner.Scan() {
				err = scanner.Err()
				goto end
			}
			
			input = strings.ToLower(strings.TrimSpace(scanner.Text()))
			
			switch input {
			case "", "y", "yes":
				approved = true
				goto end
			case "n", "no":
				approved = false
				goto end
			case "a", "all":
				approved = true
				approveAll = true
				goto end
			default:
				fmt.Println("Please enter 'y' (yes), 'n' (no), or 'a' (all)")
				continue
			}
		}
		
	end:
		return approved, approveAll, err
	}
}
