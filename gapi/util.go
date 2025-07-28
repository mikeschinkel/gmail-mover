package gapi

import (
	"io"
	"log"
	"time"
)

func mustCloseOrLog(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// parseEmailDate parses email Date headers using common RFC2822 formats
// Returns the parsed date and whether parsing was successful
func parseEmailDate(dateStr string) (parsedDate time.Time, success bool) {
	// Common RFC2822 email date formats to try in order
	dateFormats := []string{
		time.RFC1123Z,                           // "Mon, 02 Jan 2006 15:04:05 -0700"
		"Mon, 2 Jan 2006 15:04:05 -0700",        // Common variant without leading zero
		"Mon, 02 Jan 2006 15:04:05 MST",         // With timezone name
		"Mon, 2 Jan 2006 15:04:05 MST",          // With timezone name, no leading zero
		"2 Jan 2006 15:04:05 -0700",             // Without day of week
		"02 Jan 2006 15:04:05 -0700",            // Without day of week, with leading zero
		"Mon, 02 Jan 2006 15:04:05 -0700 (MST)", // With timezone in parentheses
		"Mon, 2 Jan 2006 15:04:05 -0700 (MST)",  // With timezone in parentheses, no leading zero
		time.RFC822Z,                            // "02 Jan 06 15:04 -0700"
		time.RFC822,                             // "02 Jan 06 15:04 MST"
	}

	for _, format := range dateFormats {
		if parsed, err := time.Parse(format, dateStr); err == nil {
			parsedDate = parsed
			success = true
			goto end
		}
	}

	// If all parsing attempts failed, use current time and log warning
	parsedDate = time.Now()
	success = false
	logger.Warn("Failed to parse email date header", "date_header", dateStr)

end:
	return parsedDate, success
}
