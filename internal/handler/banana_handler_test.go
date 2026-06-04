// Unit tests for banana HTTP handling and router dispatch using a mocked repository.
package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/handler"
	"github.com/phides-code/go-multi-api/internal/platform"
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

func TestBananaHandlerCreate(t *testing.T) {
	t.Parallel()

	repo := &mockBananaRepository{
		createFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
			return banana, nil
		},
	}
	h := handler.NewBananaHandler(repo, platform.NewLogger())

	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Body:       `{"content":"ripe"}`,
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	var envelope platform.APIResponse
	if err := json.Unmarshal([]byte(resp.Body), &envelope); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if envelope.Error != nil {
		t.Fatalf("unexpected error: %s", *envelope.Error)
	}

	data, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}

	var banana domain.Banana
	if err := json.Unmarshal(data, &banana); err != nil {
		t.Fatalf("unmarshal banana: %v", err)
	}
	if banana.Content != "ripe" {
		t.Fatalf("content = %q, want %q", banana.Content, "ripe")
	}
	if err := domain.ValidateID(banana.ID); err != nil {
		t.Fatalf("expected generated uuid: %v", err)
	}
}

func TestBananaHandlerGetByIDNotFound(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	repo := &mockBananaRepository{
		getFn: func(_ context.Context, gotID string) (domain.Banana, error) {
			if gotID != id {
				t.Fatalf("id = %q, want %q", gotID, id)
			}
			return domain.Banana{}, domain.ErrNotFound
		},
	}
	h := handler.NewBananaHandler(repo, platform.NewLogger())

	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		PathParameters: map[string]string{"id": id},
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestBananaHandlerInvalidID(t *testing.T) {
	t.Parallel()

	repo := &mockBananaRepository{
		getFn: func(_ context.Context, _ string) (domain.Banana, error) {
			t.Fatal("repository should not be called for invalid id")
			return domain.Banana{}, nil
		},
	}
	h := handler.NewBananaHandler(repo, platform.NewLogger())

	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		PathParameters: map[string]string{"id": "bad-id"},
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestBananaHandlerDeleteReturnsDeletedObject(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	deleted := domain.Banana{ID: id, Content: "gone"}
	repo := &mockBananaRepository{
		deleteFn: func(_ context.Context, gotID string) (domain.Banana, error) {
			if gotID != id {
				t.Fatalf("id = %q, want %q", gotID, id)
			}
			return deleted, nil
		},
	}
	h := handler.NewBananaHandler(repo, platform.NewLogger())

	resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "DELETE",
		PathParameters: map[string]string{"id": id},
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var envelope platform.APIResponse
	if err := json.Unmarshal([]byte(resp.Body), &envelope); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	data, err := json.Marshal(envelope.Data)
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}

	var banana domain.Banana
	if err := json.Unmarshal(data, &banana); err != nil {
		t.Fatalf("unmarshal banana: %v", err)
	}
	if banana != deleted {
		t.Fatalf("deleted = %+v, want %+v", banana, deleted)
	}
}

func TestRouterDispatchesBananas(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	repo := &mockBananaRepository{
		getFn: func(_ context.Context, gotID string) (domain.Banana, error) {
			return domain.Banana{ID: gotID, Content: "found"}, nil
		},
		listFn:   func(_ context.Context, _ domain.ListOptions) (domain.Page, error) { return domain.Page{}, nil },
		createFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) { return banana, nil },
		updateFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) { return banana, nil },
		deleteFn: func(_ context.Context, _ string) (domain.Banana, error) { return domain.Banana{}, nil },
	}

	router := handler.NewRouter(platform.NewLogger())
	router.Register("bananas", handler.NewBananaHandler(repo, platform.NewLogger()))

	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:     "GET",
		Path:           "/bananas/" + id,
		PathParameters: map[string]string{"id": id},
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRouterUnknownResource(t *testing.T) {
	t.Parallel()

	router := handler.NewRouter(platform.NewLogger())
	resp, err := router.Handle(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/apples",
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}
