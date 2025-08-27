package gmover

// OutputWriter defines the interface for user-facing output
type OutputWriter interface {
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Package-level output writer
var outputWriter OutputWriter

// SetWriter sets the package-level output writer
func SetWriter(writer OutputWriter) {
	outputWriter = writer
}

// GetWriter returns the current output writer
func GetWriter() OutputWriter {
	return outputWriter
}

// Printf writes formatted output using the configured output writer
func Printf(format string, args ...interface{}) {
	if outputWriter != nil {
		outputWriter.Printf(format, args...)
	}
}

// Errorf writes formatted error output using the configured output writer
func Errorf(format string, args ...interface{}) {
	if outputWriter != nil {
		outputWriter.Errorf(format, args...)
	}
}
