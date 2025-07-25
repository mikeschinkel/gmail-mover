package gmover

import (
	"fmt"
	"io"
)

func fprintf(w io.Writer, format string, a ...any) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		logger.Error("Error attempting to output to writer",
			"writer", w,
			"output", fmt.Sprintf(format, a...),
			"error", err)
	}
}

func deRef[T any](ptr *T) (v T) {
	if ptr != nil {
		v = *ptr
	}
	return v
}

func toPtr[T any](v T) *T {
	return &v
}

//goland:noinspection GoUnusedParameter
func noop(...any) {}
