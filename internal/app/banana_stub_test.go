// No-op banana.Repository for composition smoke tests.
package app

import (
	"context"

	"github.com/phides-code/go-multi-api/internal/banana"
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
