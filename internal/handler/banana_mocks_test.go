// Banana mock repository helpers for handler and router tests.
package handler_test

import (
	"context"
	"errors"

	"github.com/phides-code/go-multi-api/internal/domain"
)

type mockBananaRepository struct {
	createFn func(ctx context.Context, banana domain.Banana) (domain.Banana, error)
	getFn    func(ctx context.Context, id string) (domain.Banana, error)
	listFn   func(ctx context.Context, opts domain.ListOptions) (domain.BananaPage, error)
	updateFn func(ctx context.Context, banana domain.Banana) (domain.Banana, error)
	deleteFn func(ctx context.Context, id string) (domain.Banana, error)
}

func (m *mockBananaRepository) Create(ctx context.Context, banana domain.Banana) (domain.Banana, error) {
	return m.createFn(ctx, banana)
}

func (m *mockBananaRepository) GetByID(ctx context.Context, id string) (domain.Banana, error) {
	return m.getFn(ctx, id)
}

func (m *mockBananaRepository) List(ctx context.Context, opts domain.ListOptions) (domain.BananaPage, error) {
	return m.listFn(ctx, opts)
}

func (m *mockBananaRepository) Update(ctx context.Context, banana domain.Banana) (domain.Banana, error) {
	return m.updateFn(ctx, banana)
}

func (m *mockBananaRepository) Delete(ctx context.Context, id string) (domain.Banana, error) {
	return m.deleteFn(ctx, id)
}

func emptyBananaRepo() *mockBananaRepository {
	return &mockBananaRepository{
		createFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
			return domain.Banana{}, nil
		},
		getFn: func(_ context.Context, id string) (domain.Banana, error) {
			return domain.Banana{}, nil
		},
		listFn: func(_ context.Context, opts domain.ListOptions) (domain.BananaPage, error) {
			return domain.BananaPage{}, nil
		},
		updateFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
			return domain.Banana{}, nil
		},
		deleteFn: func(_ context.Context, id string) (domain.Banana, error) {
			return domain.Banana{}, nil
		},
	}
}

// dispatchBananaRepo returns a permissive mock for router dispatch tests (GET by id succeeds).
func dispatchBananaRepo() *mockBananaRepository {
	return &mockBananaRepository{
		getFn: func(_ context.Context, gotID string) (domain.Banana, error) {
			return domain.Banana{ID: gotID, Content: "found"}, nil
		},
		listFn: func(_ context.Context, _ domain.ListOptions) (domain.BananaPage, error) {
			return domain.BananaPage{}, nil
		},
		createFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
			return banana, nil
		},
		updateFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
			return banana, nil
		},
		deleteFn: func(_ context.Context, _ string) (domain.Banana, error) {
			return domain.Banana{}, nil
		},
	}
}

func listBananaRepo(items []domain.Banana) *mockBananaRepository {
	return &mockBananaRepository{
		listFn: func(_ context.Context, opts domain.ListOptions) (domain.BananaPage, error) {
			if opts.Limit != domain.DefaultListLimit {
				return domain.BananaPage{}, errors.New("wrong limit")
			}
			return domain.BananaPage{Items: items}, nil
		},
	}
}

func updateBananaRepo(wantID string, updated domain.Banana) *mockBananaRepository {
	return &mockBananaRepository{
		updateFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
			if banana.ID != wantID {
				return domain.Banana{}, domain.ErrNotFound
			}
			return updated, nil
		}}
}

func panicBananaRepo() *mockBananaRepository {
	panicFn := func() {
		panic("repository must not be called")
	}
	return &mockBananaRepository{
		createFn: func(context.Context, domain.Banana) (domain.Banana, error) {
			panicFn()
			return domain.Banana{}, nil
		},
		getFn: func(context.Context, string) (domain.Banana, error) {
			panicFn()
			return domain.Banana{}, nil
		},
		listFn: func(context.Context, domain.ListOptions) (domain.BananaPage, error) {
			panicFn()
			return domain.BananaPage{}, nil
		},
		updateFn: func(context.Context, domain.Banana) (domain.Banana, error) {
			panicFn()
			return domain.Banana{}, nil
		},
		deleteFn: func(context.Context, string) (domain.Banana, error) {
			panicFn()
			return domain.Banana{}, nil
		},
	}
}
