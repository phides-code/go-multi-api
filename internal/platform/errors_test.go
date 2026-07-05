package platform_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/platform"
)

func TestErrorMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
		wantClient bool
	}{
		{
			name:       "invalid id",
			err:        domain.ErrInvalidID,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "invalid id",
			wantClient: true,
		},
		{
			name:       "validation failed",
			err:        domain.ErrValidationFailed,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "validation failed",
			wantClient: true,
		},
		{
			name:       "invalid json",
			err:        domain.ErrInvalidJSON,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "invalid json",
			wantClient: true,
		},
		{
			name:       "not found",
			err:        domain.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantMsg:    "not found",
			wantClient: true,
		},
		{
			name:       "method not allowed",
			err:        domain.ErrMethodNotAllowed,
			wantStatus: http.StatusMethodNotAllowed,
			wantMsg:    "method not allowed",
			wantClient: true,
		},
		{
			name:       "unknown error",
			err:        errors.New("something broke"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "internal server error",
			wantClient: false,
		},
		{
			name:       "wrapped not found",
			err:        fmt.Errorf("get item: %w", domain.ErrNotFound),
			wantStatus: http.StatusNotFound,
			wantMsg:    "not found",
			wantClient: true,
		},
		{
			name:       "invalid cursor",
			err:        domain.ErrInvalidCursor,
			wantStatus: http.StatusBadRequest,
			wantMsg:    "invalid cursor",
			wantClient: true,
		},
		{
			name:       "already exists",
			err:        domain.ErrAlreadyExists,
			wantStatus: http.StatusConflict,
			wantMsg:    "already exists",
			wantClient: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := platform.HTTPStatusForError(tt.err); got != tt.wantStatus {
				t.Errorf("HTTPStatusForError() = %d, want %d", got, tt.wantStatus)
			}
			if got := platform.ClientErrorMessage(tt.err); got != tt.wantMsg {
				t.Errorf("ClientErrorMessage() = %q, want %q", got, tt.wantMsg)
			}
			if got := platform.IsClientError(tt.err); got != tt.wantClient {
				t.Errorf("IsClientError() = %v, want %v", got, tt.wantClient)
			}
		})
	}
}

func TestClientErrorMessageUsesDomainSentinelText(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		domain.ErrNotFound,
		domain.ErrInvalidID,
		domain.ErrValidationFailed,
		domain.ErrInvalidJSON,
		domain.ErrMethodNotAllowed,
		domain.ErrAlreadyExists,
		domain.ErrInvalidCursor,
	}

	for _, sentinel := range sentinels {
		t.Run(sentinel.Error(), func(t *testing.T) {
			t.Parallel()
			if got := platform.ClientErrorMessage(sentinel); got != sentinel.Error() {
				t.Errorf("ClientErrorMessage() = %q, want sentinel Error() = %q", got, sentinel.Error())
			}
		})
	}
}
