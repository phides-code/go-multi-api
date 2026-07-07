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

func cfTokenHeaders(token string) map[string]string {
	return map[string]string{platform.CFTTokenHeader: token}
}

type stubResourceHandler struct{}

func (stubResourceHandler) Handle(_ context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return platform.SuccessResponse(http.StatusOK, map[string]bool{"routed": true})
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

	g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/apples",
		Headers:    cfTokenHeaders(testutil.TestCFTToken),
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestGatewayEmptyPath(t *testing.T) {
	t.Parallel()

	g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/",
		Headers:    cfTokenHeaders(testutil.TestCFTToken),
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestGatewayRejectsMissingCFTToken(t *testing.T) {
	t.Parallel()

	g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/bananas",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestGatewayRejectsInvalidCFTToken(t *testing.T) {
	t.Parallel()

	g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/bananas",
		Headers:    cfTokenHeaders("wrong"),
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestGatewayRoutesRegisteredPrefix(t *testing.T) {
	t.Parallel()

	g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	g.Register("apples", stubResourceHandler{})

	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/apples",
		Headers:    cfTokenHeaders(testutil.TestCFTToken),
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var envelope platform.APIResponse
	if err := json.Unmarshal([]byte(resp.Body), &envelope); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
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

		g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
		g.Register("apples", stubResourceHandler{})

		resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod: http.MethodGet,
			Path:       "/apples",
			Headers:    cfTokenHeaders(testutil.TestCFTToken),
		})
		if err != nil {
			t.Fatalf("handle: %v", err)
		}

		assertEnvelopeShape(t, resp.Body)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)

		resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod: http.MethodGet,
			Path:       "/apples",
			Headers:    cfTokenHeaders(testutil.TestCFTToken),
		})
		if err != nil {
			t.Fatalf("handle: %v", err)
		}

		assertEnvelopeShape(t, resp.Body)
	})
}
