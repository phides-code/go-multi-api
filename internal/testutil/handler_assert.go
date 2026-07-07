// Shared handler test helpers for any resource (envelope parsing and API errors).
package testutil

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/platform"
)

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

func AssertAPIError(t *testing.T, envelope platform.APIResponse, wantMsg string) {
	t.Helper()
	if envelope.Data != nil {
		t.Fatalf("expected nil data, got %v", envelope.Data)
	}
	if envelope.Error == nil || *envelope.Error != wantMsg {
		t.Fatalf("error = %v, want %q", envelope.Error, wantMsg)
	}
}
