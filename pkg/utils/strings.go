package utils

import "net/mail"

const (
	ellipses = "..."
)

func Truncate(input string, size int) string {
	if len(input) <= size {
		return input
	}
	if len(input) <= len(ellipses) {
		return ellipses
	}

	return input[:size-len(ellipses)] + ellipses
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
