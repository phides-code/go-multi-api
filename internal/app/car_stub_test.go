// No-op car.Repository for composition smoke tests.
package app

import (
	"context"

	"github.com/phides-code/go-multi-api/internal/car"
)

type stubCarRepo struct{}

func (stubCarRepo) Create(_ context.Context, _ car.Car) (car.Car, error) {
	return car.Car{}, nil
}
func (stubCarRepo) GetByID(_ context.Context, _ string) (car.Car, error) {
	return car.Car{}, nil
}
func (stubCarRepo) List(_ context.Context) ([]car.Car, error) {
	return nil, nil
}
func (stubCarRepo) Update(_ context.Context, _ car.Car) (car.Car, error) {
	return car.Car{}, nil
}
func (stubCarRepo) Delete(_ context.Context, _ string) (car.Car, error) {
	return car.Car{}, nil
}
