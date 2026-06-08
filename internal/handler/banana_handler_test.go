// Unit tests for banana HTTP handling using a mocked repository.
package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

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
			if banana.CreatedOn == 0 {
				t.Fatal("expected createdOn to be set on create")
			}
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
	if banana.CreatedOn == 0 {
		t.Fatal("expected createdOn in response")
	}
	now := uint64(time.Now().UnixMilli())
	if banana.CreatedOn > now || now-banana.CreatedOn > 5000 {
		t.Fatalf("createdOn = %d, expected within 5s of %d", banana.CreatedOn, now)
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

func TestBananaHandlerGetByID(t *testing.T) {
	t.Parallel()

	validUuid := uuid.NewString()
	validBanana := domain.Banana{
		ID:      validUuid,
		Content: "valid content",
	}

	tests := []struct {
		name         string
		pathID       string
		wantStatus   int
		wantBanana   *domain.Banana
		wantErrorMsg string
		setupRepo    func(pathID string) *mockBananaRepository
	}{
		{
			name:         "GET by ID success",
			pathID:       validUuid,
			wantStatus:   http.StatusOK,
			wantBanana:   &validBanana,
			wantErrorMsg: "",
			setupRepo: func(pathID string) *mockBananaRepository {
				return &mockBananaRepository{
					getFn: func(_ context.Context, id string) (domain.Banana, error) {
						if id != pathID {
							return domain.Banana{}, domain.ErrNotFound
						}
						return validBanana, nil
					},
				}
			},
		},
		{
			name:         "GET by ID invalid",
			pathID:       "bad id",
			wantStatus:   http.StatusBadRequest,
			wantBanana:   nil,
			wantErrorMsg: "invalid id",
			setupRepo:    func(pathID string) *mockBananaRepository { return stubRepo() },
		},
		{
			name:         "GET by ID not found",
			pathID:       validUuid,
			wantStatus:   http.StatusNotFound,
			wantBanana:   nil,
			wantErrorMsg: "not found",
			setupRepo: func(pathID string) *mockBananaRepository {
				return &mockBananaRepository{
					getFn: func(_ context.Context, id string) (domain.Banana, error) {
						if id == pathID {
							return domain.Banana{}, domain.ErrNotFound
						}
						return validBanana, nil
					},
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := handler.NewBananaHandler(tt.setupRepo(tt.pathID), platform.NewLogger())

			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
			}

			if tt.pathID != "" {
				req.PathParameters = map[string]string{"id": tt.pathID}
			}

			resp, err := h.Handle(context.Background(), req)
			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			var envelope platform.APIResponse
			if err := json.Unmarshal([]byte(resp.Body), &envelope); err != nil {
				t.Fatalf("unmarshal body: %v", err)
			}

			if tt.wantErrorMsg != "" {
				if envelope.Data != nil {
					t.Fatalf("expected nil data, got %v", envelope.Data)
				}

				if envelope.Error == nil || *envelope.Error != tt.wantErrorMsg {
					t.Fatalf("error = %v, want %q", envelope.Error, tt.wantErrorMsg)
				}
			} else {
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

				if banana != *tt.wantBanana {
					t.Fatalf("banana = %+v, want %+v", banana, tt.wantBanana)
				}
			}
		})
	}
}

func TestBananaHandlerClientErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		method       string
		body         string
		wantStatus   int
		wantErrorMsg string
	}{
		{
			name:         "POST invalid json",
			method:       "POST",
			body:         "{not json",
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "invalid json",
		},
		{
			name:         "POST empty content",
			method:       "POST",
			body:         "{\"content\":\"\"}",
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "invalid content",
		},
		{
			name:         "PATCH unsupported method",
			method:       "PATCH",
			body:         "",
			wantStatus:   http.StatusMethodNotAllowed,
			wantErrorMsg: "method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := handler.NewBananaHandler(stubRepo(), platform.NewLogger())

			req := events.APIGatewayProxyRequest{
				HTTPMethod: tt.method,
				Body:       tt.body,
			}

			resp, err := h.Handle(context.Background(), req)

			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			var envelope platform.APIResponse
			if err := json.Unmarshal([]byte(resp.Body), &envelope); err != nil {
				t.Fatalf("unmarshal body: %v", err)
			}

			if envelope.Data != nil {
				t.Fatalf("expected nil data, got %v", envelope.Data)
			}

			if envelope.Error == nil || *envelope.Error != tt.wantErrorMsg {
				t.Fatalf("error = %v, want %q", envelope.Error, tt.wantErrorMsg)
			}
		})
	}
}
