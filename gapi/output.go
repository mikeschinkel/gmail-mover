package gapi

import (
	"log"
)

var output OutputWriter

func GetOutput() OutputWriter {
	return output
}
func SetOutput(writer OutputWriter) {
	output = writer
}
func ensureOutput() {
	if output == nil {
		log.Fatal("OutputWriter is not set; call cliutil.SetOutputWriter() first.")
	}
}
