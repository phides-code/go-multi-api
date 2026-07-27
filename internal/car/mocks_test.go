// Car mock repository helpers for handler and router tests.
package car_test

import (
	"context"

	"github.com/phides-code/go-multi-api/internal/car"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

type mockCarRepository struct {
	createFn func(ctx context.Context, c car.Car) (car.Car, error)
	getFn    func(ctx context.Context, id string) (car.Car, error)
	listFn   func(ctx context.Context) ([]car.Car, error)
	updateFn func(ctx context.Context, c car.Car) (car.Car, error)
	deleteFn func(ctx context.Context, id string) (car.Car, error)
}

func (m *mockCarRepository) Create(ctx context.Context, c car.Car) (car.Car, error) {
	return m.createFn(ctx, c)
}
func (m *mockCarRepository) GetByID(ctx context.Context, id string) (car.Car, error) {
	return m.getFn(ctx, id)
}
func (m *mockCarRepository) List(ctx context.Context) ([]car.Car, error) {
	return m.listFn(ctx)
}
func (m *mockCarRepository) Update(ctx context.Context, c car.Car) (car.Car, error) {
	return m.updateFn(ctx, c)
}
func (m *mockCarRepository) Delete(ctx context.Context, id string) (car.Car, error) {
	return m.deleteFn(ctx, id)
}

func emptyCarRepo() *mockCarRepository {
	return &mockCarRepository{
		createFn: func(_ context.Context, _ car.Car) (car.Car, error) { return car.Car{}, nil },
		getFn:    func(_ context.Context, _ string) (car.Car, error) { return car.Car{}, nil },
		listFn:   func(_ context.Context) ([]car.Car, error) { return nil, nil },
		updateFn: func(_ context.Context, _ car.Car) (car.Car, error) { return car.Car{}, nil },
		deleteFn: func(_ context.Context, _ string) (car.Car, error) { return car.Car{}, nil },
	}
}

func dispatchCarRepo() *mockCarRepository {
	return &mockCarRepository{
		getFn: func(_ context.Context, gotID string) (car.Car, error) {
			return car.Car{ID: gotID, Model: testutil.TestCarModel}, nil
		},
		listFn:   func(_ context.Context) ([]car.Car, error) { return nil, nil },
		createFn: func(_ context.Context, c car.Car) (car.Car, error) { return c, nil },
		updateFn: func(_ context.Context, c car.Car) (car.Car, error) { return c, nil },
		deleteFn: func(_ context.Context, _ string) (car.Car, error) { return car.Car{}, nil },
	}
}

func listCarRepo(items []car.Car) *mockCarRepository {
	return &mockCarRepository{
		listFn: func(_ context.Context) ([]car.Car, error) { return items, nil },
	}
}

func updateCarRepo(wantID string, updated car.Car) *mockCarRepository {
	return &mockCarRepository{
		updateFn: func(_ context.Context, c car.Car) (car.Car, error) {
			if c.ID != wantID {
				return car.Car{}, domain.ErrNotFound
			}
			return updated, nil
		},
	}
}

func panicCarRepo() *mockCarRepository {
	panicFn := func() { panic("repository must not be called") }
	return &mockCarRepository{
		createFn: func(context.Context, car.Car) (car.Car, error) { panicFn(); return car.Car{}, nil },
		getFn:    func(context.Context, string) (car.Car, error) { panicFn(); return car.Car{}, nil },
		listFn:   func(context.Context) ([]car.Car, error) { panicFn(); return nil, nil },
		updateFn: func(context.Context, car.Car) (car.Car, error) { panicFn(); return car.Car{}, nil },
		deleteFn: func(context.Context, string) (car.Car, error) { panicFn(); return car.Car{}, nil },
	}
}
