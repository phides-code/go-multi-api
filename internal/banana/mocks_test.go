// Banana mock repository helpers for handler and router tests.
package banana_test

import (
	"context"
	"errors"

	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/domain"
)

type mockBananaRepository struct {
	createFn func(ctx context.Context, b banana.Banana) (banana.Banana, error)
	getFn    func(ctx context.Context, id string) (banana.Banana, error)
	listFn   func(ctx context.Context, opts domain.ListOptions) (banana.Page, error)
	updateFn func(ctx context.Context, b banana.Banana) (banana.Banana, error)
	deleteFn func(ctx context.Context, id string) (banana.Banana, error)
}

func (m *mockBananaRepository) Create(ctx context.Context, b banana.Banana) (banana.Banana, error) {
	return m.createFn(ctx, b)
}

func (m *mockBananaRepository) GetByID(ctx context.Context, id string) (banana.Banana, error) {
	return m.getFn(ctx, id)
}

func (m *mockBananaRepository) List(ctx context.Context, opts domain.ListOptions) (banana.Page, error) {
	return m.listFn(ctx, opts)
}

func (m *mockBananaRepository) Update(ctx context.Context, b banana.Banana) (banana.Banana, error) {
	return m.updateFn(ctx, b)
}

func (m *mockBananaRepository) Delete(ctx context.Context, id string) (banana.Banana, error) {
	return m.deleteFn(ctx, id)
}

func emptyBananaRepo() *mockBananaRepository {
	return &mockBananaRepository{
		createFn: func(_ context.Context, _ banana.Banana) (banana.Banana, error) {
			return banana.Banana{}, nil
		},
		getFn: func(_ context.Context, _ string) (banana.Banana, error) {
			return banana.Banana{}, nil
		},
		listFn: func(_ context.Context, _ domain.ListOptions) (banana.Page, error) {
			return banana.Page{}, nil
		},
		updateFn: func(_ context.Context, _ banana.Banana) (banana.Banana, error) {
			return banana.Banana{}, nil
		},
		deleteFn: func(_ context.Context, _ string) (banana.Banana, error) {
			return banana.Banana{}, nil
		},
	}
}

// dispatchBananaRepo returns a permissive mock for router dispatch tests (GET by id succeeds).
func dispatchBananaRepo() *mockBananaRepository {
	return &mockBananaRepository{
		getFn: func(_ context.Context, gotID string) (banana.Banana, error) {
			return banana.Banana{ID: gotID, Content: "found"}, nil
		},
		listFn: func(_ context.Context, _ domain.ListOptions) (banana.Page, error) {
			return banana.Page{}, nil
		},
		createFn: func(_ context.Context, b banana.Banana) (banana.Banana, error) {
			return b, nil
		},
		updateFn: func(_ context.Context, b banana.Banana) (banana.Banana, error) {
			return b, nil
		},
		deleteFn: func(_ context.Context, _ string) (banana.Banana, error) {
			return banana.Banana{}, nil
		},
	}
}

func listBananaRepo(items []banana.Banana) *mockBananaRepository {
	return &mockBananaRepository{
		listFn: func(_ context.Context, opts domain.ListOptions) (banana.Page, error) {
			if opts.Limit != domain.DefaultListLimit {
				return banana.Page{}, errors.New("wrong limit")
			}
			return banana.Page{Items: items}, nil
		},
	}
}

func updateBananaRepo(wantID string, updated banana.Banana) *mockBananaRepository {
	return &mockBananaRepository{
		updateFn: func(_ context.Context, b banana.Banana) (banana.Banana, error) {
			if b.ID != wantID {
				return banana.Banana{}, domain.ErrNotFound
			}
			return updated, nil
		},
	}
}

func panicBananaRepo() *mockBananaRepository {
	panicFn := func() {
		panic("repository must not be called")
	}
	return &mockBananaRepository{
		createFn: func(context.Context, banana.Banana) (banana.Banana, error) {
			panicFn()
			return banana.Banana{}, nil
		},
		getFn: func(context.Context, string) (banana.Banana, error) {
			panicFn()
			return banana.Banana{}, nil
		},
		listFn: func(context.Context, domain.ListOptions) (banana.Page, error) {
			panicFn()
			return banana.Page{}, nil
		},
		updateFn: func(context.Context, banana.Banana) (banana.Banana, error) {
			panicFn()
			return banana.Banana{}, nil
		},
		deleteFn: func(context.Context, string) (banana.Banana, error) {
			panicFn()
			return banana.Banana{}, nil
		},
	}
}
