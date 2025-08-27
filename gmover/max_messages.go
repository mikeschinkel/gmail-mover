package gmover

import (
	"fmt"
)

// MaxMessages represents a validated maximum message count
type MaxMessages int64

// ParseMaxMessages validates and creates a MaxMessages
func ParseMaxMessages(count int64) (MaxMessages, error) {
	var maxMessages MaxMessages
	var err error

	if count < 0 {
		err = fmt.Errorf("max messages cannot be negative: %d", count)
		goto end
	}

	if count == 0 {
		// Default to reasonable limit if 0 is provided
		count = 10000
	}

	maxMessages = MaxMessages(count)

end:
	return maxMessages, err
}

// IsZero returns true if the max messages is 0
func (m MaxMessages) IsZero() bool {
	return int64(m) == 0
}
