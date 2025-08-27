package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mikeschinkel/gmover/cliutil"
	"github.com/mikeschinkel/gmover/gmcmds"
	"github.com/mikeschinkel/gmover/gmover"
	"github.com/mikeschinkel/gmover/logutil"

	// Import all commands to trigger their init() functions
	_ "github.com/mikeschinkel/gmover/gmcmds"
)

func main() {
	logger, err := logutil.CreateJSONLogger()
	if err != nil {
		err = fmt.Errorf("failed to initialize logger: %v\n", err)
		goto end
	}
	err = gmover.Run(context.Background(), gmover.RunArgs{
		Args:           os.Args,
		Logger:         logger,
		CLIWriter:      cliutil.NewOutputWriter(),
		ConfigProvider: gmcmds.NewConfigProvider(),
	})
end:
	if err != nil {
		os.Exit(1)
	}
}
