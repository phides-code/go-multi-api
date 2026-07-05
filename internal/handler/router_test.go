// Unit tests for request routing, resource dispatch, and X-CF-Token gate.
package handler_test

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/handler"
	"github.com/phides-code/go-multi-api/internal/platform"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func cfTokenHeaders(token string) map[string]string {
	return map[string]string{"X-CF-Token": token}
}

type stubResourceHandler struct{}

func (stubResourceHandler) Handle(_ context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return platform.SuccessResponse(http.StatusOK, map[string]bool{"dispatched": true})
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

func TestRouterDispatchesBananas(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	repo := dispatchBananaRepo()
	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	router.Register("bananas", handler.NewBananaHandler(repo, platform.NewLogger()))

	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Path:           "/bananas/" + id,
		PathParameters: map[string]string{"id": id},
		Headers:        cfTokenHeaders(testutil.TestCFTToken),
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRouterUnknownResource(t *testing.T) {
	t.Parallel()

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
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

func TestRouterEmptyPath(t *testing.T) {
	t.Parallel()

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
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

func TestRouterRejectsMissingCFTToken(t *testing.T) {
	t.Parallel()

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
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

func TestRouterRejectsInvalidCFTToken(t *testing.T) {
	t.Parallel()

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
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

func TestRouterSkipsCFTTokenUnderSAMLocal(t *testing.T) {
	t.Setenv("AWS_SAM_LOCAL", "true")

	id := uuid.NewString()
	repo := dispatchBananaRepo()
	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	router.Register("bananas", handler.NewBananaHandler(repo, platform.NewLogger()))

	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Path:           "/bananas/" + id,
		PathParameters: map[string]string{"id": id},
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRouterAllowsOptionsWithoutCFTToken(t *testing.T) {
	t.Parallel()

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	router.Register("bananas", handler.NewBananaHandler(emptyBananaRepo(), platform.NewLogger()))

	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "OPTIONS",
		Path:       "/bananas",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}

	var envelope platform.APIResponse
	if err := json.Unmarshal([]byte(resp.Body), &envelope); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if envelope.Data != nil {
		t.Fatalf("expected nil data, got %v", envelope.Data)
	}
	if envelope.Error == nil || *envelope.Error != "method not allowed" {
		t.Fatalf("error = %v, want %q", envelope.Error, "method not allowed")
	}
}

func TestRouterDispatchesRegisteredPrefix(t *testing.T) {
	t.Parallel()

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	router.Register("apples", stubResourceHandler{})

	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
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
	if dispatched, _ := data["dispatched"].(bool); !dispatched {
		t.Fatalf("data[dispatched] = %v, want true", data["dispatched"])
	}
}

func TestRouterResponseEnvelopeShape(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		router := handler.NewRouterWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
		router.Register("apples", stubResourceHandler{})

		resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
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

		router := handler.NewRouterWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)

		resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
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
