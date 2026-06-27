// Maps domain errors to HTTP status codes and client-facing error messages.
package platform

import (
	"errors"
	"net/http"

	"github.com/phides-code/go-multi-api/internal/domain"
)

func HTTPStatusForError(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidID),
		errors.Is(err, domain.ErrValidationFailed),
		errors.Is(err, domain.ErrInvalidJSON),
		errors.Is(err, domain.ErrInvalidCursor):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, domain.ErrMethodNotAllowed):
		return http.StatusMethodNotAllowed
	default:
		return http.StatusInternalServerError
	}
}

func ClientErrorMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrInvalidID):
		return "invalid id"
	case errors.Is(err, domain.ErrValidationFailed):
		return "validation failed"
	case errors.Is(err, domain.ErrInvalidJSON):
		return "invalid json"
	case errors.Is(err, domain.ErrNotFound):
		return "not found"
	case errors.Is(err, domain.ErrAlreadyExists):
		return "already exists"
	case errors.Is(err, domain.ErrMethodNotAllowed):
		return "method not allowed"
	case errors.Is(err, domain.ErrInvalidCursor):
		return "invalid cursor"
	default:
		return "internal server error"
	}
}

func IsClientError(err error) bool {
	status := HTTPStatusForError(err)
	return status >= 400 && status < 500
}
