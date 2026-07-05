// Unit tests for HTTP status mapping and JSON response envelopes.
package platform_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/phides-code/go-multi-api/internal/platform"
)

func TestSuccessResponseEnvelope(t *testing.T) {
	t.Parallel()

	resp, err := platform.SuccessResponse(http.StatusOK, map[string]string{"key": "value"})
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
