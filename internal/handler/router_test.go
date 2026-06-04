// Unit tests for request routing, resource dispatch, and X-CF-Token gate.
package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/handler"
	"github.com/phides-code/go-multi-api/internal/platform"
)

const testCFTToken = "test-token"

func cfTokenHeaders(token string) map[string]string {
	return map[string]string{"X-CF-Token": token}
}

func TestRouterDispatchesBananas(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	repo := &mockBananaRepository{
		getFn: func(_ context.Context, gotID string) (domain.Banana, error) {
			return domain.Banana{ID: gotID, Content: "found"}, nil
		},
		listFn:   func(_ context.Context, _ domain.ListOptions) (domain.Page, error) { return domain.Page{}, nil },
		createFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) { return banana, nil },
		updateFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) { return banana, nil },
		deleteFn: func(_ context.Context, _ string) (domain.Banana, error) { return domain.Banana{}, nil },
	}

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testCFTToken)
	router.Register("bananas", handler.NewBananaHandler(repo, platform.NewLogger()))

	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Path:           "/bananas/" + id,
		PathParameters: map[string]string{"id": id},
		Headers:        cfTokenHeaders(testCFTToken),
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

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testCFTToken)
	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/apples",
		Headers:    cfTokenHeaders(testCFTToken),
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

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testCFTToken)
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

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testCFTToken)
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

func TestRouterAllowsOptionsWithoutCFTToken(t *testing.T) {
	t.Parallel()

	router := handler.NewRouterWithCFTToken(platform.NewLogger(), testCFTToken)
	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "OPTIONS",
		Path:       "/bananas",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatal("expected OPTIONS preflight to bypass token check")
	}
}
