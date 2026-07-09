// Composition smoke tests: verify the built gateway handles banana routes without panicking.
package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/platform"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

type stubBananaRepo struct{}

func (stubBananaRepo) Create(_ context.Context, _ banana.Banana) (banana.Banana, error) {
	return banana.Banana{}, nil
}
func (stubBananaRepo) GetByID(_ context.Context, _ string) (banana.Banana, error) {
	return banana.Banana{}, nil
}
func (stubBananaRepo) List(_ context.Context) ([]banana.Banana, error) {
	return nil, nil
}
func (stubBananaRepo) Update(_ context.Context, _ banana.Banana) (banana.Banana, error) {
	return banana.Banana{}, nil
}
func (stubBananaRepo) Delete(_ context.Context, _ string) (banana.Banana, error) {
	return banana.Banana{}, nil
}

func TestWiringSmokeGETBananas(t *testing.T) {
	t.Setenv("AWS_CF_TOKEN", testutil.TestCFTToken)

	g := buildGateway(platform.NewLogger(), stubBananaRepo{})

	resp, err := g.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/bananas",
		Headers:    map[string]string{platform.CFTTokenHeader: testutil.TestCFTToken},
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		t.Fatalf("status = %d, want < 500", resp.StatusCode)
	}
}
