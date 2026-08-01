// Composition smoke tests: built gateway serves registered routes without panicking.
package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/gateway"
	"github.com/phides-code/go-multi-api/internal/platform"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func TestWiringSmokeGETBananas(t *testing.T) {
	assertWiringSmokeGET(t, testGateway(t), "/"+banana.PathPrefix)
}

func testGateway(t *testing.T) *gateway.Gateway {
	t.Helper()
	t.Setenv(platform.CFTTokenEnvVar, testutil.TestCFTToken)
	return buildGateway(platform.NewLogger(), stubBananaRepo{})
}

func assertWiringSmokeGET(t *testing.T, g *gateway.Gateway, path string) {
	t.Helper()

	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       path,
		Headers:    testutil.CFTokenHeaders(testutil.TestCFTToken),
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		t.Fatalf("status = %d, want < 500", resp.StatusCode)
	}
}
