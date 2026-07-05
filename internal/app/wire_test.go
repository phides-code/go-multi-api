package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/platform"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

type stubBananaRepo struct{}

func (stubBananaRepo) Create(_ context.Context, _ domain.Banana) (domain.Banana, error) {
	return domain.Banana{}, nil
}
func (stubBananaRepo) GetByID(_ context.Context, _ string) (domain.Banana, error) {
	return domain.Banana{}, nil
}
func (stubBananaRepo) List(_ context.Context, _ domain.ListOptions) (domain.BananaPage, error) {
	return domain.BananaPage{}, nil
}
func (stubBananaRepo) Update(_ context.Context, _ domain.Banana) (domain.Banana, error) {
	return domain.Banana{}, nil
}
func (stubBananaRepo) Delete(_ context.Context, _ string) (domain.Banana, error) {
	return domain.Banana{}, nil
}

func TestWiringSmokeGETBananas(t *testing.T) {
	t.Setenv("AWS_CF_TOKEN", testutil.TestCFTToken)

	router := newRouter(platform.NewLogger(), stubBananaRepo{})

	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/bananas",
		Headers:    map[string]string{"X-CF-Token": testutil.TestCFTToken},
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode >= http.StatusInternalServerError {
		t.Fatalf("status = %d, want < 500", resp.StatusCode)
	}
}
