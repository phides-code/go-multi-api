// Shared handler test helpers for any resource (envelope parsing and API errors).
package testutil

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/platform"
)

// RequireStatusAndEnvelope fails unless status matches, then returns the parsed envelope.
func RequireStatusAndEnvelope(t *testing.T, resp events.APIGatewayProxyResponse, wantStatus int) platform.APIResponse {
	t.Helper()
	if resp.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d", resp.StatusCode, wantStatus)
	}
	var envelope platform.APIResponse
	if err := json.Unmarshal([]byte(resp.Body), &envelope); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	return envelope
}

// RequireHandle fails if Handle returned an error, then asserts status and returns the envelope.
func RequireHandle(t *testing.T, resp events.APIGatewayProxyResponse, err error, wantStatus int) platform.APIResponse {
	t.Helper()
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	return RequireStatusAndEnvelope(t, resp, wantStatus)
}

// AssertAPIError fails unless data is nil and error equals wantMsg.
func AssertAPIError(t *testing.T, envelope platform.APIResponse, wantMsg string) {
	t.Helper()
	if envelope.Data != nil {
		t.Fatalf("expected nil data, got %v", envelope.Data)
	}
	if envelope.Error == nil || *envelope.Error != wantMsg {
		t.Fatalf("error = %v, want %q", envelope.Error, wantMsg)
	}
}
