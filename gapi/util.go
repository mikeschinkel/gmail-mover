package gapi

import (
	"io"
	"log"
)

func mustCloseOrLog(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}
