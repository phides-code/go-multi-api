package handler_test

import (
	"encoding/json"
	"maps"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/platform"
)

func requireStatusAndEnvelope(t *testing.T, resp events.APIGatewayProxyResponse, wantStatus int) platform.APIResponse {
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

func assertAPIError(t *testing.T, envelope platform.APIResponse, wantMsg string) {
	t.Helper()
	if envelope.Data != nil {
		t.Fatalf("expected nil data, got %v", envelope.Data)
	}
	if envelope.Error == nil || *envelope.Error != wantMsg {
		t.Fatalf("error = %v, want %q", envelope.Error, wantMsg)
	}
}

func decodeBananaData(t *testing.T, envelope platform.APIResponse) domain.Banana {
	t.Helper()
	if envelope.Error != nil {
		t.Fatalf("unexpected error: %s", *envelope.Error)
	}
	data, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var banana domain.Banana
	if err := json.Unmarshal(data, &banana); err != nil {
		t.Fatalf("unmarshal banana: %v", err)
	}
	return banana
}

func decodePageData(t *testing.T, envelope platform.APIResponse) domain.BananaPage {
	t.Helper()
	if envelope.Error != nil {
		t.Fatalf("unexpected error: %s", *envelope.Error)
	}
	data, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var page domain.BananaPage
	if err := json.Unmarshal(data, &page); err != nil {
		t.Fatalf("unmarshal page: %v", err)
	}
	return page
}

func assertBananaDataKeys(t *testing.T, envelope platform.APIResponse) {
	t.Helper()

	raw, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}

	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatalf("unmarshal data keys: %v", err)
	}

	want := []string{"content", "createdOn", "id"}
	if len(keys) != len(want) {
		t.Fatalf("data has %d keys %v, want exactly %v", len(keys), maps.Keys(keys), want)
	}
	for _, k := range want {
		if _, ok := keys[k]; !ok {
			t.Fatalf("missing data key %q; got %v", k, maps.Keys(keys))
		}
	}
}
