package gmover

import (
	"fmt"
	"net/mail"
	"strings"
)

// EmailAddress represents a validated email address
type EmailAddress string

// ParseEmailAddress validates and creates an EmailAddress
func ParseEmailAddress(email string) (EmailAddress, error) {
	var addr EmailAddress
	var err error

	if email == "" {
		err = fmt.Errorf("email address cannot be empty")
		goto end
	}

	email = strings.TrimSpace(email)

	// Use Go's standard mail package for validation
	_, err = mail.ParseAddress(email)
	if err != nil {
		err = fmt.Errorf("invalid email address '%s': %w", email, err)
		goto end
	}

	addr = EmailAddress(email)

end:
	return addr, err
}

// IsZero returns true if the email address is empty
func (e EmailAddress) IsZero() bool {
	return string(e) == ""
}
