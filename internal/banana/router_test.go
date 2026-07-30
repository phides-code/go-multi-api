// Gateway integration tests for the bananas resource.
package banana_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/gateway"
	"github.com/phides-code/go-multi-api/internal/platform"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func cfTokenHeaders(token string) map[string]string {
	return map[string]string{platform.CFTTokenHeader: token}
}

func registeredBananaGateway(repo banana.Repository) *gateway.Gateway {
	g := gateway.NewGatewayWithCFTToken(platform.NewLogger(), testutil.TestCFTToken)
	g.Register(banana.PathPrefix, banana.NewHandler(repo, platform.NewLogger()))
	return g
}

func TestGatewayRoutesBananas(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	g := registeredBananaGateway(dispatchBananaRepo())

	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodGet,
		Path:           "/" + banana.PathPrefix + "/" + id,
		PathParameters: map[string]string{"id": id},
		Headers:        cfTokenHeaders(testutil.TestCFTToken),
	})
	testutil.RequireHandle(t, resp, err, http.StatusOK)
}

func TestGatewaySkipsCFTTokenUnderSAMLocal(t *testing.T) {
	t.Setenv("AWS_SAM_LOCAL", "true")

	id := uuid.NewString()
	g := registeredBananaGateway(dispatchBananaRepo())

	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     http.MethodGet,
		Path:           "/" + banana.PathPrefix + "/" + id,
		PathParameters: map[string]string{"id": id},
	})
	testutil.RequireHandle(t, resp, err, http.StatusOK)
}

func TestGatewayAllowsOptionsWithoutCFTToken(t *testing.T) {
	t.Parallel()

	g := registeredBananaGateway(emptyBananaRepo())

	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodOptions,
		Path:       "/" + banana.PathPrefix,
	})
	envelope := testutil.RequireHandle(t, resp, err, http.StatusMethodNotAllowed)
	testutil.AssertAPIError(t, envelope, domain.ErrMethodNotAllowed.Error())
}
