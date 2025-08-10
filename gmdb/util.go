package gmdb

import (
	"fmt"
	"time"
)

func panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

func throttle(d time.Duration) {
	time.Sleep(d)
}
