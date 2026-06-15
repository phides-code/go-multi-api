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
		errors.Is(err, domain.ErrInvalidContent),
		errors.Is(err, domain.ErrInvalidJSON):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
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
	case errors.Is(err, domain.ErrInvalidContent):
		return "invalid content"
	case errors.Is(err, domain.ErrInvalidJSON):
		return "invalid json"
	case errors.Is(err, domain.ErrNotFound):
		return "not found"
	case errors.Is(err, domain.ErrMethodNotAllowed):
		return "method not allowed"
	default:
		return "internal server error"
	}
}

func IsClientError(err error) bool {
	status := HTTPStatusForError(err)
	return status >= 400 && status < 500
}
