// Sentinel errors shared across domain, handlers, and HTTP response mapping.
// Error() strings are the client-facing API messages (see platform.ClientErrorMessage).
package domain

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidID        = errors.New("invalid id")
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidJSON      = errors.New("invalid json")
	ErrMethodNotAllowed = errors.New("method not allowed")
	ErrAlreadyExists    = errors.New("already exists")
)
