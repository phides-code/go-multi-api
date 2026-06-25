package handler_test

import (
	"context"
	"errors"

	"github.com/phides-code/go-multi-api/internal/domain"
)

type mockBananaRepository struct {
	createFn func(ctx context.Context, banana domain.Banana) (domain.Banana, error)
	getFn    func(ctx context.Context, id string) (domain.Banana, error)
	listFn   func(ctx context.Context, opts domain.ListOptions) (domain.Page, error)
	updateFn func(ctx context.Context, banana domain.Banana) (domain.Banana, error)
	deleteFn func(ctx context.Context, id string) (domain.Banana, error)
}

func (m *mockBananaRepository) Create(ctx context.Context, banana domain.Banana) (domain.Banana, error) {
	return m.createFn(ctx, banana)
}

func (m *mockBananaRepository) GetByID(ctx context.Context, id string) (domain.Banana, error) {
	return m.getFn(ctx, id)
}

func (m *mockBananaRepository) List(ctx context.Context, opts domain.ListOptions) (domain.Page, error) {
	return m.listFn(ctx, opts)
}

func (m *mockBananaRepository) Update(ctx context.Context, banana domain.Banana) (domain.Banana, error) {
	return m.updateFn(ctx, banana)
}

func (m *mockBananaRepository) Delete(ctx context.Context, id string) (domain.Banana, error) {
	return m.deleteFn(ctx, id)
}

func stubRepo() *mockBananaRepository {
	return &mockBananaRepository{
		createFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
			return domain.Banana{}, nil
		},
		getFn: func(_ context.Context, id string) (domain.Banana, error) {
			return domain.Banana{}, nil
		},
		listFn: func(_ context.Context, opts domain.ListOptions) (domain.Page, error) {
			return domain.Page{}, nil
		},
		updateFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
			return domain.Banana{}, nil
		},
		deleteFn: func(_ context.Context, id string) (domain.Banana, error) {
			return domain.Banana{}, nil
		},
	}
}

func listRepo(items []domain.Banana) *mockBananaRepository {
	return &mockBananaRepository{
		listFn: func(_ context.Context, opts domain.ListOptions) (domain.Page, error) {
			if opts.Limit != domain.DefaultListLimit {
				return domain.Page{}, errors.New("wrong limit")
			}
			return domain.Page{Items: items}, nil
		},
	}
}

func updateRepo(wantID string, updated domain.Banana) *mockBananaRepository {
	return &mockBananaRepository{
		updateFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
			if banana.ID != wantID {
				return domain.Banana{}, domain.ErrNotFound
			}
			return updated, nil
		}}
}
