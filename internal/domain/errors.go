// Sentinel errors shared across domain, handlers, and HTTP response mapping.
package domain

import "errors"

var (
	ErrNotFound        = errors.New("banana not found")
	ErrInvalidID       = errors.New("invalid id")
	ErrInvalidContent  = errors.New("invalid content")
	ErrInvalidJSON     = errors.New("invalid json")
	ErrMethodNotAllowed = errors.New("method not allowed")
)
