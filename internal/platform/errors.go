// Maps domain errors to HTTP status codes and client-facing error messages.
package platform

import (
	"errors"
	"net/http"

	"github.com/phides-code/go-multi-api/internal/domain"
)

// InternalServerErrorMessage is the client-facing text for unexpected failures (500).
const InternalServerErrorMessage = "internal server error"

type clientErrorMapping struct {
	sentinel error
	status   int
}

// clientErrorMappings ties each domain sentinel to its HTTP status.
// ClientErrorMessage uses sentinel.Error() so strings live only in domain/errors.go.
var clientErrorMappings = []clientErrorMapping{
	{domain.ErrInvalidID, http.StatusBadRequest},
	{domain.ErrValidationFailed, http.StatusBadRequest},
	{domain.ErrInvalidJSON, http.StatusBadRequest},
	{domain.ErrInvalidCursor, http.StatusBadRequest},
	{domain.ErrNotFound, http.StatusNotFound},
	{domain.ErrAlreadyExists, http.StatusConflict},
	{domain.ErrMethodNotAllowed, http.StatusMethodNotAllowed},
}

func HTTPStatusForError(err error) int {
	for _, m := range clientErrorMappings {
		if errors.Is(err, m.sentinel) {
			return m.status
		}
	}
	return http.StatusInternalServerError
}

func ClientErrorMessage(err error) string {
	for _, m := range clientErrorMappings {
		if errors.Is(err, m.sentinel) {
			return m.sentinel.Error()
		}
	}
	return InternalServerErrorMessage
}

func IsClientError(err error) bool {
	status := HTTPStatusForError(err)
	return status >= 400 && status < 500
}
