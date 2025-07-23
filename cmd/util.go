package main

import (
	"fmt"
	"io"
	"log"
)

func fprintf(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		log.Printf("Error attempting to output to writer: %v [writer=%v,message=%s]",
			err,
			w,
			fmt.Sprintf(format, a...),
		)
	}
}

func deRef[T any](ptr *T) (v T) {
	if ptr != nil {
		v = *ptr
	}
	return v
}
