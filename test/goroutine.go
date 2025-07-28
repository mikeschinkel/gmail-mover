package test

import (
	"runtime"
	"strconv"
	"strings"
)

// GoID extracts the current goroutine ID using runtime.Stack()
// This is test-only code where performance is less critical
func GoID() int64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	line := strings.Fields(string(buf[:n]))[1] // e.g., "goroutine 42 [running]"
	id, _ := strconv.ParseInt(line, 10, 64)
	return id
}
