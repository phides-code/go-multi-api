// Unit tests for HTTP status mapping and JSON response envelopes.
package platform_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/platform"
)

func TestHTTPStatusForError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err    error
		status int
	}{
		{err: domain.ErrInvalidID, status: http.StatusBadRequest},
		{err: domain.ErrNotFound, status: http.StatusNotFound},
		{err: domain.ErrMethodNotAllowed, status: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		if status := platform.HTTPStatusForError(tt.err); status != tt.status {
			t.Fatalf("status for %v = %d, want %d", tt.err, status, tt.status)
		}
	}
}

func TestSuccessResponseEnvelope(t *testing.T) {
	t.Parallel()

	resp, err := platform.SuccessResponse(http.StatusOK, domain.Banana{ID: "id", Content: "c"})
	if err != nil {
		t.Fatalf("success response: %v", err)
	}
	if resp.Headers["Content-Type"] != "application/json" {
		t.Fatalf("content type = %q", resp.Headers["Content-Type"])
	}

	var envelope platform.APIResponse
	if err := json.Unmarshal([]byte(resp.Body), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Error != nil {
		t.Fatalf("expected nil error, got %v", envelope.Error)
	}
	if envelope.Data == nil {
		t.Fatal("expected data")
	}
}

func TestErrorResponseEnvelope(t *testing.T) {
	t.Parallel()

	resp, err := platform.ErrorResponse(http.StatusBadRequest, "invalid id")
	if err != nil {
		t.Fatalf("error response: %v", err)
	}

	var envelope platform.APIResponse
	if err := json.Unmarshal([]byte(resp.Body), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Data != nil {
		t.Fatalf("expected nil data, got %v", envelope.Data)
	}
	if envelope.Error == nil || *envelope.Error != "invalid id" {
		t.Fatalf("unexpected error field: %v", envelope.Error)
	}
}
