package gapi

import (
	"log"
)

var writer OutputWriter

func GetWriter() OutputWriter {
	return writer
}
func SetWriter(w OutputWriter) {
	writer = w
}
func ensureOutput() {
	if writer == nil {
		log.Fatal("OutputWriter is not set; call cliutil.SetOutputWriter() first.")
	}
}
