// Unit tests for gateway routing and X-CF-Token auth gate.
package gateway_test

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/gateway"
	"github.com/phides-code/go-multi-api/internal/platform"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

// Stub resource used only inside this package (not a real app resource).
const (
	stubPrefix = "apples"
	stubPath   = "/" + stubPrefix
)

type stubResourceHandler struct{}

func (stubResourceHandler) Handle(_ context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return platform.SuccessResponse(http.StatusOK, map[string]bool{"routed": true})
}

func newTestGateway() *gateway.Gateway {
	return gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
}

func getRequest(path string, headers map[string]string) events.APIGatewayProxyRequest {
	return events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       path,
		Headers:    headers,
	}
}

func assertEnvelopeShape(t *testing.T, body string) {
	t.Helper()

	var keys map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &keys); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	for _, k := range []string{"data", "error"} {
		if _, ok := keys[k]; !ok {
			t.Fatalf("missing top-level key %q; got %v", k, maps.Keys(keys))
		}
	}
	if len(keys) != 2 {
		t.Fatalf("body has %d top-level keys %v, want exactly data and error", len(keys), maps.Keys(keys))
	}
}

func TestGatewayUnknownResource(t *testing.T) {
	t.Parallel()

	resp, err := newTestGateway().Handle(context.Background(), getRequest(stubPath, testutil.CFTokenHeaders(testutil.TestCFTToken)))
	testutil.RequireHandle(t, resp, err, http.StatusNotFound)
}

func TestGatewayEmptyPath(t *testing.T) {
	t.Parallel()

	resp, err := newTestGateway().Handle(context.Background(), getRequest("/", testutil.CFTokenHeaders(testutil.TestCFTToken)))
	testutil.RequireHandle(t, resp, err, http.StatusNotFound)
}

func TestGatewayRejectsMissingCFTToken(t *testing.T) {
	t.Parallel()

	resp, err := newTestGateway().Handle(context.Background(), getRequest(stubPath, nil))
	testutil.RequireHandle(t, resp, err, http.StatusUnauthorized)
}

func TestGatewayRejectsInvalidCFTToken(t *testing.T) {
	t.Parallel()

	resp, err := newTestGateway().Handle(context.Background(), getRequest(stubPath, testutil.CFTokenHeaders("wrong")))
	testutil.RequireHandle(t, resp, err, http.StatusUnauthorized)
}

func TestGatewayRoutesRegisteredPrefix(t *testing.T) {
	t.Parallel()

	g := newTestGateway()
	g.Register(stubPrefix, stubResourceHandler{})

	resp, err := g.Handle(context.Background(), getRequest(stubPath, testutil.CFTokenHeaders(testutil.TestCFTToken)))
	envelope := testutil.RequireHandle(t, resp, err, http.StatusOK)
	if envelope.Error != nil {
		t.Fatalf("unexpected error: %v", envelope.Error)
	}

	data, ok := envelope.Data.(map[string]any)
	if !ok {
		t.Fatalf("data type = %T, want map[string]any", envelope.Data)
	}
	if routed, _ := data["routed"].(bool); !routed {
		t.Fatalf("data[routed] = %v, want true", data["routed"])
	}
}

func TestGatewayResponseEnvelopeShape(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		g := newTestGateway()
		g.Register(stubPrefix, stubResourceHandler{})

		resp, err := g.Handle(context.Background(), getRequest(stubPath, testutil.CFTokenHeaders(testutil.TestCFTToken)))
		testutil.RequireHandle(t, resp, err, http.StatusOK)
		assertEnvelopeShape(t, resp.Body)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		resp, err := newTestGateway().Handle(context.Background(), getRequest(stubPath, testutil.CFTokenHeaders(testutil.TestCFTToken)))
		testutil.RequireHandle(t, resp, err, http.StatusNotFound)
		assertEnvelopeShape(t, resp.Body)
	})
}
