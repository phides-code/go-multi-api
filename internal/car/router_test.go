// Gateway integration tests for the cars resource.
package car_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/car"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/gateway"
	"github.com/phides-code/go-multi-api/internal/platform"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func cfTokenHeaders(token string) map[string]string {
	return map[string]string{platform.CFTTokenHeader: token}
}

func TestGatewayRoutesCars(t *testing.T) {
	t.Parallel()
	id := uuid.NewString()
	g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	g.Register("cars", car.NewHandler(dispatchCarRepo(), platform.NewLogger()))

	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Path:           "/cars/" + id,
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

func TestGatewaySkipsCFTTokenUnderSAMLocal(t *testing.T) {
	t.Setenv("AWS_SAM_LOCAL", "true")
	id := uuid.NewString()
	g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	g.Register("cars", car.NewHandler(dispatchCarRepo(), platform.NewLogger()))

	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Path:           "/cars/" + id,
		PathParameters: map[string]string{"id": id},
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestGatewayAllowsOptionsWithoutCFTToken(t *testing.T) {
	t.Parallel()
	g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	g.Register("cars", car.NewHandler(emptyCarRepo(), platform.NewLogger()))

	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "OPTIONS",
		Path:       "/cars",
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
	wantErr := domain.ErrMethodNotAllowed.Error()
	if envelope.Error == nil || *envelope.Error != wantErr {
		t.Fatalf("error = %v, want %q", envelope.Error, wantErr)
	}
}
