package logutil

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func CreateJSONLogger() (logger *slog.Logger, err error) {
	var logDir string
	var logFilePath string
	var homeDir string
	var logFile *os.File

	homeDir, err = os.UserHomeDir()
	if err != nil {
		err = fmt.Errorf("failed to access home directory; %w\n", err)
		goto end
	}

	logDir = filepath.Join(homeDir, ".config", "gmover")
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		err = fmt.Errorf("failed to make log directory %s; %w\n", logDir, err)
		goto end
	}
	logFilePath = filepath.Join(logDir, "errors.log")
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		err = fmt.Errorf("failed to open log logFile %s; %w\n", logFilePath, err)
		goto end
	}
	defer mustClose(logFile)

	logger = slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Logger initialized", "log_file", logFilePath)

end:
	if err != nil {
		err = fmt.Errorf("failed to initialize logger; %w\n", err)
	}
	return logger, err
}
