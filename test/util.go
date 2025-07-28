package test

import (
	"fmt"
	"io"
	"log"
)

// Utility functions that should be in the standard library but aren't

//goland:noinspection GoUnusedFunction
func fprintf(w io.Writer, format string, a ...any) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		log.Printf("Error attempting to output to writer: %v [writer=%v,message=%s]",
			err,
			w,
			fmt.Sprintf(format, a...),
		)
	}
}
