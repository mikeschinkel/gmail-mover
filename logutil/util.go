package logutil

import (
	"io"
	"log"
)

// mustClose closes an io.Closer and terminates the program on error.
func mustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}
