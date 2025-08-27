package gmover

import (
	"encoding/json"
	"time"

	"github.com/mikeschinkel/gmover/gapi"
	"github.com/mikeschinkel/gmover/gmcfg"
)

// MoveLogger handles JSON lines logging for email move operations
type MoveLogger struct {
	fileStore *gmcfg.FileStore
}

// NewMoveLogger creates a new move logger
func NewMoveLogger() *MoveLogger {
	return &MoveLogger{
		fileStore: gmcfg.NewFileStore(AppName),
	}
}

// LogMove writes a move operation entry to the JSON lines log
func (ml *MoveLogger) LogMove(entry gapi.MoveLogEntry) (err error) {
	var data []byte

	entry.Timestamp = time.Now()

	data, err = json.Marshal(entry)
	if err != nil {
		goto end
	}

	// Append with newline for JSON lines format
	data = append(data, '\n')
	err = ml.fileStore.Append("moves.jsonl", data)

end:
	return err
}
