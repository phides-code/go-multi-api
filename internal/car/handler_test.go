// Unit tests for car HTTP handling using a mocked repository.
package car_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/car"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/platform"
	"github.com/phides-code/go-multi-api/internal/testutil"
)

func TestCarHandlerCreate(t *testing.T) {
	t.Parallel()

	validCreateBody := testutil.ValidCarBody().JSON(t)

	tests := []struct {
		name         string
		body         string
		setupRepo    func() *mockCarRepository
		wantStatus   int
		wantErrorMsg string
		wantModel    string
	}{
		{
			name: "success",
			body: validCreateBody,
			setupRepo: func() *mockCarRepository {
				return &mockCarRepository{
					createFn: func(_ context.Context, c car.Car) (car.Car, error) {
						return c, nil
					},
				}
			},
			wantStatus: http.StatusCreated,
			wantModel:  testutil.TestCarModel,
		},
		{
			name: "repo failure",
			body: validCreateBody,
			setupRepo: func() *mockCarRepository {
				return &mockCarRepository{
					createFn: func(_ context.Context, _ car.Car) (car.Car, error) {
						return car.Car{}, errors.New("db down")
					},
				}
			},
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: platform.InternalServerErrorMessage,
		},
		{
			name: "duplicate id",
			body: validCreateBody,
			setupRepo: func() *mockCarRepository {
				return &mockCarRepository{
					createFn: func(_ context.Context, _ car.Car) (car.Car, error) {
						return car.Car{}, domain.ErrAlreadyExists
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

			h := car.NewHandler(tt.setupRepo(), platform.NewLogger())

			resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
				Body:       tt.body,
			})
			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			envelope := testutil.RequireStatusAndEnvelope(t, resp, tt.wantStatus)

			if tt.wantErrorMsg != "" {
				testutil.AssertAPIError(t, envelope, tt.wantErrorMsg)
				return
			}

			c := decodeCarData(t, envelope)
			assertCarDataKeys(t, envelope)

			if c.Model != tt.wantModel {
				t.Fatalf("model = %q, want %q", c.Model, tt.wantModel)
			}

			if err := domain.ValidateID(c.ID); err != nil {
				t.Fatalf("expected generated uuid: %v", err)
			}
			if c.CreatedOn == 0 {
				t.Fatal("expected createdOn in response")
			}
			now := uint64(time.Now().UnixMilli())
			if c.CreatedOn > now || now-c.CreatedOn > 5000 {
				t.Fatalf("createdOn = %d, expected within 5s of %d", c.CreatedOn, now)
			}
		})
	}
}

func TestCarHandlerDelete(t *testing.T) {
	t.Parallel()

	validUuid, deletedCar, _ := existingCarFixture(t)

	tests := []struct {
		name         string
		pathID       string
		wantStatus   int
		wantCar      *car.Car
		wantErrorMsg string
		setupRepo    func(pathID string) *mockCarRepository
	}{
		{
			name:         "DELETE success",
			pathID:       validUuid,
			wantStatus:   http.StatusOK,
			wantCar:      &deletedCar,
			wantErrorMsg: "",
			setupRepo: func(pathID string) *mockCarRepository {
				return &mockCarRepository{
					deleteFn: func(_ context.Context, id string) (car.Car, error) {
						if id != pathID {
							return car.Car{}, domain.ErrNotFound
						}
						return deletedCar, nil
					},
				}
			},
		},
		{
			name:         "DELETE invalid ID",
			pathID:       "bad id",
			wantStatus:   http.StatusBadRequest,
			wantCar:      nil,
			wantErrorMsg: "invalid id",
			setupRepo:    func(pathID string) *mockCarRepository { return emptyCarRepo() },
		},
		{
			name:         "DELETE ID not found",
			pathID:       validUuid,
			wantStatus:   http.StatusNotFound,
			wantCar:      nil,
			wantErrorMsg: "not found",
			setupRepo: func(pathID string) *mockCarRepository {
				return &mockCarRepository{
					deleteFn: func(_ context.Context, id string) (car.Car, error) {
						if id == pathID {
							return car.Car{}, domain.ErrNotFound
						}
						return deletedCar, nil
					},
				}
			},
		},
		{
			name:         "DELETE repo failure",
			pathID:       validUuid,
			wantStatus:   http.StatusInternalServerError,
			wantCar:      nil,
			wantErrorMsg: platform.InternalServerErrorMessage,
			setupRepo: func(pathID string) *mockCarRepository {
				return &mockCarRepository{
					deleteFn: func(_ context.Context, _ string) (car.Car, error) {
						return car.Car{}, errors.New("db down")
					},
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := car.NewHandler(tt.setupRepo(tt.pathID), platform.NewLogger())

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

			envelope := testutil.RequireStatusAndEnvelope(t, resp, tt.wantStatus)

			if tt.wantErrorMsg != "" {
				testutil.AssertAPIError(t, envelope, tt.wantErrorMsg)
				return
			}

			c := decodeCarData(t, envelope)

			if c != *tt.wantCar {
				t.Fatalf("car = %+v, want %+v", c, tt.wantCar)
			}
		})
	}
}

func TestCarHandlerGetByID(t *testing.T) {
	t.Parallel()

	validUuid, validCar, _ := existingCarFixture(t)

	tests := []struct {
		name         string
		pathID       string
		wantStatus   int
		wantCar      *car.Car
		wantErrorMsg string
		setupRepo    func(pathID string) *mockCarRepository
	}{
		{
			name:         "GET by ID success",
			pathID:       validUuid,
			wantStatus:   http.StatusOK,
			wantCar:      &validCar,
			wantErrorMsg: "",
			setupRepo: func(pathID string) *mockCarRepository {
				return &mockCarRepository{
					getFn: func(_ context.Context, id string) (car.Car, error) {
						if id != pathID {
							return car.Car{}, domain.ErrNotFound
						}
						return validCar, nil
					},
				}
			},
		},
		{
			name:         "GET by ID invalid",
			pathID:       "bad id",
			wantStatus:   http.StatusBadRequest,
			wantCar:      nil,
			wantErrorMsg: "invalid id",
			setupRepo:    func(pathID string) *mockCarRepository { return emptyCarRepo() },
		},
		{
			name:         "GET by ID not found",
			pathID:       validUuid,
			wantStatus:   http.StatusNotFound,
			wantCar:      nil,
			wantErrorMsg: "not found",
			setupRepo: func(pathID string) *mockCarRepository {
				return &mockCarRepository{
					getFn: func(_ context.Context, id string) (car.Car, error) {
						if id == pathID {
							return car.Car{}, domain.ErrNotFound
						}
						return validCar, nil
					},
				}
			},
		},
		{
			name:         "GET by ID repo failure",
			pathID:       validUuid,
			wantStatus:   http.StatusInternalServerError,
			wantCar:      nil,
			wantErrorMsg: platform.InternalServerErrorMessage,
			setupRepo: func(pathID string) *mockCarRepository {
				return &mockCarRepository{
					getFn: func(_ context.Context, _ string) (car.Car, error) {
						return car.Car{}, errors.New("db down")
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := car.NewHandler(tt.setupRepo(tt.pathID), platform.NewLogger())

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
			envelope := testutil.RequireStatusAndEnvelope(t, resp, tt.wantStatus)

			if tt.wantErrorMsg != "" {
				testutil.AssertAPIError(t, envelope, tt.wantErrorMsg)
				return
			}

			c := decodeCarData(t, envelope)
			assertCarDataKeys(t, envelope)

			if c != *tt.wantCar {
				t.Fatalf("car = %+v, want %+v", c, tt.wantCar)
			}
		})
	}
}

func TestCarHandlerClientErrors(t *testing.T) {
	t.Parallel()

	validationBodies := newCarValidationBodies(t)

	tests := []struct {
		name         string
		method       string
		body         string
		wantStatus   int
		wantErrorMsg string
		setupRepo    func() *mockCarRepository
	}{
		{
			name:         "POST invalid json",
			method:       "POST",
			body:         "{not json",
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "invalid json",
		},
		{
			name:         "POST empty model",
			method:       "POST",
			body:         validationBodies.carWithEmptyValue,
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "validation failed",
			setupRepo:    panicCarRepo,
		},
		{
			name:         "PATCH unsupported method",
			method:       "PATCH",
			body:         "",
			wantStatus:   http.StatusMethodNotAllowed,
			wantErrorMsg: "method not allowed",
		},
		{
			name:         "POST whitespace model",
			method:       "POST",
			body:         validationBodies.carWithWhitespace,
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "validation failed",
			setupRepo:    panicCarRepo,
		},
		{
			name:         "POST model too long",
			method:       "POST",
			body:         validationBodies.carWithValueTooLong,
			wantStatus:   http.StatusBadRequest,
			wantErrorMsg: "validation failed",
			setupRepo:    panicCarRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := emptyCarRepo()
			if tt.setupRepo != nil {
				repo = tt.setupRepo()
			}

			h := car.NewHandler(repo, platform.NewLogger())

			req := events.APIGatewayProxyRequest{
				HTTPMethod: tt.method,
				Body:       tt.body,
			}

			resp, err := h.Handle(context.Background(), req)
			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			testutil.AssertAPIError(t, testutil.RequireStatusAndEnvelope(t, resp, tt.wantStatus), tt.wantErrorMsg)
		})
	}
}

func TestCarHandlerList(t *testing.T) {
	t.Parallel()

	carOne, carTwo, _ := testutil.ListCars(false)
	wantItems := []car.Car{carOne, carTwo}

	tests := []struct {
		name         string
		wantStatus   int
		wantItems    []car.Car
		wantErrorMsg string
		setupRepo    func() *mockCarRepository
	}{
		{
			name:       "GET list returns items",
			wantStatus: http.StatusOK,
			wantItems:  wantItems,
			setupRepo:  func() *mockCarRepository { return listCarRepo(wantItems) },
		},
		{
			name:       "GET list empty",
			wantStatus: http.StatusOK,
			wantItems:  []car.Car{},
			setupRepo:  func() *mockCarRepository { return listCarRepo([]car.Car{}) },
		},
		{
			name:         "GET list repo failure",
			wantStatus:   http.StatusInternalServerError,
			wantErrorMsg: platform.InternalServerErrorMessage,
			setupRepo: func() *mockCarRepository {
				return &mockCarRepository{
					listFn: func(_ context.Context) ([]car.Car, error) {
						return nil, errors.New("db down")
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := car.NewHandler(tt.setupRepo(), platform.NewLogger())

			resp, err := h.Handle(context.Background(), events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
			})
			if err != nil {
				t.Fatalf("handle: %v", err)
			}

			envelope := testutil.RequireStatusAndEnvelope(t, resp, tt.wantStatus)

			if tt.wantErrorMsg != "" {
				testutil.AssertAPIError(t, envelope, tt.wantErrorMsg)
				return
			}

			items := decodeCarListData(t, envelope)

			if len(items) != len(tt.wantItems) {
				t.Fatalf("len(items) = %d, want %d", len(items), len(tt.wantItems))
			}

			for i := range tt.wantItems {
				if items[i] != tt.wantItems[i] {
					t.Fatalf("items[%d] = %+v, want %+v", i, items[i], tt.wantItems[i])
				}
			}
		})
	}
}

func TestCarHandlerUpdate(t *testing.T) {
	t.Parallel()

	validUuid, updatedCar, validUpdateBody := existingCarFixture(t)
	validationBodies := newCarValidationBodies(t)

	tests := []struct {
		name         string
		pathID       string
		body         string
		wantStatus   int
		wantCar      *car.Car
		wantErrorMsg string
		setupRepo    func(pathID string) *mockCarRepository
	}{
		{
			name:         "PUT success",
			pathID:       validUuid,
			body:         validUpdateBody,
			wantStatus:   http.StatusOK,
			wantCar:      &updatedCar,
			wantErrorMsg: "",
			setupRepo: func(pathID string) *mockCarRepository {
				return updateCarRepo(pathID, updatedCar)
			},
		},
		{
			name:         "PUT invalid ID",
			pathID:       "bad id",
			body:         validUpdateBody,
			wantStatus:   http.StatusBadRequest,
			wantCar:      nil,
			wantErrorMsg: "invalid id",
			setupRepo: func(pathID string) *mockCarRepository {
				return emptyCarRepo()
			},
		},
		{
			name:         "PUT invalid JSON",
			pathID:       validUuid,
			body:         "not json",
			wantStatus:   http.StatusBadRequest,
			wantCar:      nil,
			wantErrorMsg: "invalid json",
			setupRepo: func(pathID string) *mockCarRepository {
				return emptyCarRepo()
			},
		},
		{
			name:         "PUT empty model",
			pathID:       validUuid,
			body:         validationBodies.carWithEmptyValue,
			wantStatus:   http.StatusBadRequest,
			wantCar:      nil,
			wantErrorMsg: "validation failed",
			setupRepo:    func(pathID string) *mockCarRepository { return emptyCarRepo() },
		},
		{
			name:         "PUT car not found",
			pathID:       validUuid,
			wantStatus:   http.StatusNotFound,
			body:         validUpdateBody,
			wantCar:      nil,
			wantErrorMsg: "not found",
			setupRepo: func(pathID string) *mockCarRepository {
				return &mockCarRepository{
					updateFn: func(_ context.Context, c car.Car) (car.Car, error) {
						if c.ID == pathID {
							return car.Car{}, domain.ErrNotFound
						}
						return updatedCar, nil
					},
				}
			},
		},
		{
			name:         "PUT repo failure",
			pathID:       validUuid,
			body:         validUpdateBody,
			wantStatus:   http.StatusInternalServerError,
			wantCar:      nil,
			wantErrorMsg: platform.InternalServerErrorMessage,
			setupRepo: func(pathID string) *mockCarRepository {
				return &mockCarRepository{
					updateFn: func(_ context.Context, _ car.Car) (car.Car, error) {
						return car.Car{}, errors.New("db down")
					},
				}
			},
		},
		{
			name:         "PUT whitespace model",
			pathID:       validUuid,
			body:         validationBodies.carWithWhitespace,
			wantStatus:   http.StatusBadRequest,
			wantCar:      nil,
			wantErrorMsg: "validation failed",
			setupRepo:    func(pathID string) *mockCarRepository { return emptyCarRepo() },
		},
		{
			name:         "PUT model too long",
			pathID:       validUuid,
			body:         validationBodies.carWithValueTooLong,
			wantStatus:   http.StatusBadRequest,
			wantCar:      nil,
			wantErrorMsg: "validation failed",
			setupRepo:    func(pathID string) *mockCarRepository { return emptyCarRepo() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := car.NewHandler(tt.setupRepo(tt.pathID), platform.NewLogger())

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

			envelope := testutil.RequireStatusAndEnvelope(t, resp, tt.wantStatus)

			if tt.wantErrorMsg != "" {
				testutil.AssertAPIError(t, envelope, tt.wantErrorMsg)
				return
			}

			c := decodeCarData(t, envelope)
			assertCarDataKeys(t, envelope)

			if c != *tt.wantCar {
				t.Fatalf("car = %+v, want %+v", c, tt.wantCar)
			}
		})
	}
}
