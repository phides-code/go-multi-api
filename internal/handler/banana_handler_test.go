// Unit tests for banana HTTP handling using a mocked repository.
package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/handler"
	"github.com/phides-code/go-multi-api/internal/platform"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func TestBananaHandlerCreate(t *testing.T) {
	t.Parallel()

	validCreateBody := testutil.BananaCreateBody(testutil.TestBananaContent)

	tests := []struct {
		name         string
		body         string
		setupRepo    func() *mockBananaRepository
		wantStatus   int
		wantErrorMsg string
		wantContent  string
	}{
		{
			name: "success",
			body: validCreateBody,
			setupRepo: func() *mockBananaRepository {
				return &mockBananaRepository{
					createFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
						return banana, nil
					},
				}
			},
			wantStatus:  http.StatusCreated,
			wantContent: testutil.TestBananaContent,
		},
		{
			name: "repo failure",
			body: validCreateBody,
			setupRepo: func() *mockBananaRepository {
				return &mockBananaRepository{
					createFn: func(_ context.Context, _ domain.Banana) (domain.Banana, error) {
						return domain.Banana{}, errors.New("db down")
					},
				}
			},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: platform.InternalServerErrorMessage,
		},
		{
			name: "duplicate id",
			body: validCreateBody,
			setupRepo: func() *mockBananaRepository {
				return &mockBananaRepository{
					createFn: func(_ context.Context, _ domain.Banana) (domain.Banana, error) {
						return domain.Banana{}, domain.ErrAlreadyExists
					},
				}
			},
			wantStatus:   http.StatusConflict,
			wantErrorMsg: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := handler.NewBananaHandler(tt.setupRepo(), platform.NewLogger())

			resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
				Body:       tt.body,
			})
			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			envelope := requireStatusAndEnvelope(t, resp, tt.wantStatus)

			if tt.wantErrorMsg != "" {
				assertAPIError(t, envelope, tt.wantErrorMsg)
				return
			}

			banana := decodeBananaData(t, envelope)
			assertBananaDataKeys(t, envelope)

			if banana.Content != tt.wantContent {
				t.Fatalf("content = %q, want %q", banana.Content, tt.wantContent)
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
		})
	}
}

func TestBananaHandlerDelete(t *testing.T) {
	t.Parallel()

	validUuid, deletedBanana, _ := existingBananaFixture()

	tests := []struct {
		name         string
		pathID       string
		wantStatus   int
		wantBanana   *domain.Banana
		wantErrorMsg string
		setupRepo    func(pathID string) *mockBananaRepository
	}{
		{
			name:         "DELETE success",
			pathID:       validUuid,
			wantStatus:   http.StatusOK,
			wantBanana:   &deletedBanana,
			wantErrorMsg: "",
			setupRepo: func(pathID string) *mockBananaRepository {
				return &mockBananaRepository{
					deleteFn: func(_ context.Context, id string) (domain.Banana, error) {
						if id != pathID {
							return domain.Banana{}, domain.ErrNotFound
						}
						return deletedBanana, nil
					},
				}
			},
		},
		{
			name:         "DELETE invalid ID",
			pathID:       "bad id",
			wantStatus:   http.StatusBadRequest,
			wantBanana:   nil,
			wantErrorMsg: "invalid id",
			setupRepo:    func(pathID string) *mockBananaRepository { return emptyBananaRepo() },
		},
		{
			name:         "DELETE ID not found",
			pathID:       validUuid,
			wantStatus:   http.StatusNotFound,
			wantBanana:   nil,
			wantErrorMsg: "not found",
			setupRepo: func(pathID string) *mockBananaRepository {
				return &mockBananaRepository{
					deleteFn: func(_ context.Context, id string) (domain.Banana, error) {
						if id == pathID {
							return domain.Banana{}, domain.ErrNotFound
						}
						return deletedBanana, nil
					},
				}
			},
		},
		{
			name:         "DELETE repo failure",
			pathID:       validUuid,
			wantStatus:   http.StatusInternalServerError,
			wantBanana:   nil,
			wantErrorMsg: platform.InternalServerErrorMessage,
			setupRepo: func(pathID string) *mockBananaRepository {
				return &mockBananaRepository{
					deleteFn: func(_ context.Context, _ string) (domain.Banana, error) {
						return domain.Banana{}, errors.New("db down")
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
				HTTPMethod: http.MethodDelete,
			}

			if tt.pathID != "" {
				req.PathParameters = map[string]string{"id": tt.pathID}
			}

			resp, err := h.Handle(context.Background(), req)
			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			envelope := requireStatusAndEnvelope(t, resp, tt.wantStatus)

			if tt.wantErrorMsg != "" {
				assertAPIError(t, envelope, tt.wantErrorMsg)
				return
			}

			banana := decodeBananaData(t, envelope)

			if banana != *tt.wantBanana {
				t.Fatalf("banana = %+v, want %+v", banana, tt.wantBanana)
			}
		})
	}
}

func TestBananaHandlerGetByID(t *testing.T) {
	t.Parallel()

	validUuid, validBanana, _ := existingBananaFixture()

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
			setupRepo:    func(pathID string) *mockBananaRepository { return emptyBananaRepo() },
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
		{
			name:         "GET by ID repo failure",
			pathID:       validUuid,
			wantStatus:   http.StatusInternalServerError,
			wantBanana:   nil,
			wantErrorMsg: platform.InternalServerErrorMessage,
			setupRepo: func(pathID string) *mockBananaRepository {
				return &mockBananaRepository{
					getFn: func(_ context.Context, _ string) (domain.Banana, error) {
						return domain.Banana{}, errors.New("db down")
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
			envelope := requireStatusAndEnvelope(t, resp, tt.wantStatus)

			if tt.wantErrorMsg != "" {
				assertAPIError(t, envelope, tt.wantErrorMsg)
				return
			}

			banana := decodeBananaData(t, envelope)
			assertBananaDataKeys(t, envelope)

			if banana != *tt.wantBanana {
				t.Fatalf("banana = %+v, want %+v", banana, tt.wantBanana)
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
		setupRepo    func() *mockBananaRepository
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
			wantErrorMsg: "validation failed",
			setupRepo:    panicBananaRepo,
		},
		{
			name:         "PATCH unsupported method",
			method:       "PATCH",
			body:         "",
			wantStatus:   http.StatusMethodNotAllowed,
			wantErrorMsg: "method not allowed",
		},
		{
			name:         "POST whitespace content",
			method:       "POST",
			body:         `{"content":"   "}`,
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "validation failed",
			setupRepo:    panicBananaRepo,
		},
		{
			name:         "POST content too long",
			method:       "POST",
			body:         fmt.Sprintf(`{"content":%q}`, strings.Repeat("a", domain.BananaMaxContentLength+1)),
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "validation failed",
			setupRepo:    panicBananaRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := emptyBananaRepo()
			if tt.setupRepo != nil {
				repo = tt.setupRepo()
			}

			h := handler.NewBananaHandler(repo, platform.NewLogger())

			req := events.APIGatewayProxyRequest{
				HTTPMethod: tt.method,
				Body:       tt.body,
			}

			resp, err := h.Handle(context.Background(), req)
			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			assertAPIError(t, requireStatusAndEnvelope(t, resp, tt.wantStatus), tt.wantErrorMsg)
		})
	}
}

func TestBananaHandlerList(t *testing.T) {
	t.Parallel()

	bananaOne, bananaTwo, page2Item := testutil.ListBananaPage(false)
	wantItems := []domain.Banana{bananaOne, bananaTwo}
	nextCursor := "abc123"

	tests := []struct {
		name           string
		wantStatus     int
		wantItems      []domain.Banana
		wantErrorMsg   string
		wantNextCursor string
		setupRepo      func() *mockBananaRepository
		queryCursor    string
	}{
		{
			name:           "GET list returns items",
			wantStatus:     http.StatusOK,
			wantItems:      wantItems,
			wantErrorMsg:   "",
			wantNextCursor: "",
			setupRepo:      func() *mockBananaRepository { return listBananaRepo(wantItems) },
		},
		{
			name:           "GET list empty",
			wantStatus:     http.StatusOK,
			wantItems:      []domain.Banana{},
			wantErrorMsg:   "",
			wantNextCursor: "",
			setupRepo:      func() *mockBananaRepository { return listBananaRepo([]domain.Banana{}) },
		},
		{
			name:           "GET list returns next cursor",
			wantStatus:     http.StatusOK,
			wantItems:      wantItems,
			wantErrorMsg:   "",
			wantNextCursor: nextCursor,
			setupRepo: func() *mockBananaRepository {
				return &mockBananaRepository{
					listFn: func(_ context.Context, opts domain.ListOptions) (domain.BananaPage, error) {
						return domain.BananaPage{
							Items:      wantItems,
							NextCursor: nextCursor,
						}, nil
					},
				}
			},
		},
		{
			name:        "GET list passes cursor query param",
			wantStatus:  http.StatusOK,
			wantItems:   []domain.Banana{page2Item},
			queryCursor: nextCursor,
			setupRepo: func() *mockBananaRepository {
				return &mockBananaRepository{
					listFn: func(_ context.Context, opts domain.ListOptions) (domain.BananaPage, error) {
						if opts.Cursor != nextCursor {
							return domain.BananaPage{}, errors.New("wrong cursor")
						}
						return domain.BananaPage{Items: []domain.Banana{page2Item}}, nil
					},
				}
			},
		},
		{
			name:         "GET list repo failure",
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: platform.InternalServerErrorMessage,
			setupRepo: func() *mockBananaRepository {
				return &mockBananaRepository{
					listFn: func(_ context.Context, _ domain.ListOptions) (domain.BananaPage, error) {
						return domain.BananaPage{}, errors.New("db down")
					},
				}
			},
		},
		{
			name:         "GET list invalid cursor",
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "invalid cursor",
			queryCursor:  "!!!not-base64!!!",
			setupRepo: func() *mockBananaRepository {
				return &mockBananaRepository{
					listFn: func(_ context.Context, _ domain.ListOptions) (domain.BananaPage, error) {
						return domain.BananaPage{}, domain.ErrInvalidCursor
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := handler.NewBananaHandler(tt.setupRepo(), platform.NewLogger())

			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
			}

			if tt.queryCursor != "" {
				req.QueryStringParameters = map[string]string{
					"cursor": tt.queryCursor,
				}
			}

			resp, err := h.Handle(context.Background(), req)
			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			envelope := requireStatusAndEnvelope(t, resp, tt.wantStatus)

			if tt.wantErrorMsg != "" {
				assertAPIError(t, envelope, tt.wantErrorMsg)
				return
			}

			page := decodeBananaPageData(t, envelope)

			if len(page.Items) != len(tt.wantItems) {
				t.Fatalf("len(page.Items) = %d, want %d", len(page.Items), len(tt.wantItems))
			}

			for i := range tt.wantItems {
				if page.Items[i] != tt.wantItems[i] {
					t.Fatalf("items[%d] = %+v, want %+v", i, page.Items[i], tt.wantItems[i])
				}
			}

			if page.NextCursor != tt.wantNextCursor {
				t.Fatalf("nextCursor = %q, want %q", page.NextCursor, tt.wantNextCursor)
			}
		})
	}
}

func TestBananaHandlerUpdate(t *testing.T) {
	t.Parallel()

	validUuid, updatedBanana, validUpdateBody := existingBananaFixture()

	tests := []struct {
		name         string
		pathID       string
		body         string
		wantStatus   int
		wantBanana   *domain.Banana
		wantErrorMsg string
		setupRepo    func(pathID string) *mockBananaRepository
	}{
		{
			name:         "PUT success",
			pathID:       validUuid,
			body:         validUpdateBody,
			wantStatus:   http.StatusOK,
			wantBanana:   &updatedBanana,
			wantErrorMsg: "",
			setupRepo: func(pathID string) *mockBananaRepository {
				return updateBananaRepo(pathID, updatedBanana)
			},
		},
		{
			name:         "PUT invalid ID",
			pathID:       "bad id",
			body:         validUpdateBody,
			wantStatus:   http.StatusBadRequest,
			wantBanana:   nil,
			wantErrorMsg: "invalid id",
			setupRepo: func(pathID string) *mockBananaRepository {
				return emptyBananaRepo()
			},
		},
		{
			name:         "PUT invalid JSON",
			pathID:       validUuid,
			body:         "not json",
			wantStatus:   http.StatusBadRequest,
			wantBanana:   nil,
			wantErrorMsg: "invalid json",
			setupRepo: func(pathID string) *mockBananaRepository {
				return emptyBananaRepo()
			},
		},
		{
			name:         "PUT empty content",
			pathID:       validUuid,
			body:         `{"content":""}`,
			wantStatus:   http.StatusBadRequest,
			wantBanana:   nil,
			wantErrorMsg: "validation failed",
			setupRepo:    func(pathID string) *mockBananaRepository { return emptyBananaRepo() },
		},
		{
			name:         "PUT banana not found",
			pathID:       validUuid,
			wantStatus:   http.StatusNotFound,
			body:         validUpdateBody,
			wantBanana:   nil,
			wantErrorMsg: "not found",
			setupRepo: func(pathID string) *mockBananaRepository {
				return &mockBananaRepository{
					updateFn: func(_ context.Context, banana domain.Banana) (domain.Banana, error) {
						if banana.ID == pathID {
							return domain.Banana{}, domain.ErrNotFound
						}
						return updatedBanana, nil
					},
				}
			},
		},
		{
			name:         "PUT repo failure",
			pathID:       validUuid,
			body:         validUpdateBody,
			wantStatus:   http.StatusInternalServerError,
			wantBanana:   nil,
			wantErrorMsg: platform.InternalServerErrorMessage,
			setupRepo: func(pathID string) *mockBananaRepository {
				return &mockBananaRepository{
					updateFn: func(_ context.Context, _ domain.Banana) (domain.Banana, error) {
						return domain.Banana{}, errors.New("db down")
					},
				}
			},
		},
		{
			name:         "PUT whitespace content",
			pathID:       validUuid,
			body:         `{"content":"   "}`,
			wantStatus:   http.StatusBadRequest,
			wantBanana:   nil,
			wantErrorMsg: "validation failed",
			setupRepo:    func(pathID string) *mockBananaRepository { return emptyBananaRepo() },
		},
		{
			name:         "PUT content too long",
			pathID:       validUuid,
			body:         fmt.Sprintf(`{"content":%q}`, strings.Repeat("a", domain.BananaMaxContentLength+1)),
			wantStatus:   http.StatusBadRequest,
			wantBanana:   nil,
			wantErrorMsg: "validation failed",
			setupRepo:    func(pathID string) *mockBananaRepository { return emptyBananaRepo() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := handler.NewBananaHandler(tt.setupRepo(tt.pathID), platform.NewLogger())

			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPut,
				Body:       tt.body,
			}

			if tt.pathID != "" {
				req.PathParameters = map[string]string{"id": tt.pathID}
			}

			resp, err := h.Handle(context.Background(), req)
			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			envelope := requireStatusAndEnvelope(t, resp, tt.wantStatus)

			if tt.wantErrorMsg != "" {
				assertAPIError(t, envelope, tt.wantErrorMsg)
				return
			}

			banana := decodeBananaData(t, envelope)
			assertBananaDataKeys(t, envelope)

			if banana != *tt.wantBanana {
				t.Fatalf("banana = %+v, want %+v", banana, tt.wantBanana)
			}
		})
	}
}
