package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/mikeschinkel/gmail-mover/gmover"
	"github.com/mikeschinkel/gmail-mover/gmutil"
)

func main() {
	// Initialize CLI-friendly slog logger
	handler := NewCLIHandler()
	logger := slog.New(handler)
	gmover.SetLogger(logger)
	gmutil.SetLogger(logger)

	config := parseFlags()
	err := gmover.Run(&config)
	if err != nil {
		logger.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

// parseFlags handles all command-line flag parsing and returns a Config
func parseFlags() (config gmover.Config) {
	config = gmover.NewConfig(gmover.MoveEmails)

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

	if *listLabels {
		config.SetRunMode(gmover.ListLabels)
	}
	return config
}
