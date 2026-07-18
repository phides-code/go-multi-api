// Shared validation helpers for create/update payloads across resources.
package domain

import (
	"strings"
	"unicode/utf8"
)

// Default rune length bounds for required strings unless a resource opts out.
const (
	DefaultMinStringLength = 1
	DefaultMaxStringLength = 100
)

// ValidateRequiredString rejects blank values (after trim) and enforces rune length bounds.
func ValidateRequiredString(s string, minLen, maxLen int) error {
	if strings.TrimSpace(s) == "" {
		return ErrValidationFailed
	}
	length := utf8.RuneCountInString(s)
	if length < minLen || length > maxLen {
		return ErrValidationFailed
	}
	return nil
}
