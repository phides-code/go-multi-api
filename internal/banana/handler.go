// HTTP handler for /bananas: maps API Gateway requests to repository operations.
package banana

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/platform"
)

type Handler struct {
	repo   Repository
	logger *platform.Logger
}

func NewHandler(repo Repository, logger *platform.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	const op = "banana request"

	id := strings.TrimSpace(req.PathParameters[AttrID])

	switch req.HTTPMethod {
	case http.MethodGet:
		if id == "" {
			return h.list(ctx, req)
		}
		return h.getByID(ctx, id)
	case http.MethodPost:
		return h.create(ctx, req.Body)
	case http.MethodPut:
		return h.update(ctx, id, req.Body)
	case http.MethodDelete:
		return h.delete(ctx, id)
	default:
		return h.errorResponse(ctx, domain.ErrMethodNotAllowed, op)
	}
}

func (h *Handler) list(ctx context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	const op = "list bananas"

	items, err := h.repo.List(ctx)
	if err != nil {
		return h.errorResponse(ctx, err, op)
	}

	return platform.SuccessResponse(http.StatusOK, items)
}

func (h *Handler) getByID(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	const op = "get banana"

	if err := domain.ValidateID(id); err != nil {
		return h.errorResponse(ctx, err, op)
	}

	banana, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return h.errorResponse(ctx, err, op)
	}

	return platform.SuccessResponse(http.StatusOK, banana)
}

type writePayload struct {
	Descriptor string `json:"descriptor"`
	Rating     int    `json:"rating"`
}

func (h *Handler) create(ctx context.Context, body string) (events.APIGatewayProxyResponse, error) {
	const op = "create banana"

	var payload writePayload
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return h.errorResponse(ctx, domain.ErrInvalidJSON, op)
	}

	input := CreateInput{Descriptor: payload.Descriptor, Rating: payload.Rating}
	if err := ValidateCreateInput(input); err != nil {
		return h.errorResponse(ctx, err, op)
	}

	banana := Banana{
		ID:         domain.NewID(),
		Descriptor: payload.Descriptor,
		Rating:     payload.Rating,
		CreatedOn:  uint64(time.Now().UnixMilli()),
	}

	created, err := h.repo.Create(ctx, banana)
	if err != nil {
		return h.errorResponse(ctx, err, op)
	}

	return platform.SuccessResponse(http.StatusCreated, created)
}

func (h *Handler) update(ctx context.Context, id, body string) (events.APIGatewayProxyResponse, error) {
	const op = "update banana"

	if err := domain.ValidateID(id); err != nil {
		return h.errorResponse(ctx, err, op)
	}

	var payload writePayload
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return h.errorResponse(ctx, domain.ErrInvalidJSON, op)
	}

	input := UpdateInput{ID: id, Descriptor: payload.Descriptor, Rating: payload.Rating}
	if err := ValidateUpdateInput(input); err != nil {
		return h.errorResponse(ctx, err, op)
	}

	updated, err := h.repo.Update(ctx, Banana{
		ID:         id,
		Descriptor: payload.Descriptor,
		Rating:     payload.Rating,
	})
	if err != nil {
		return h.errorResponse(ctx, err, op)
	}

	return platform.SuccessResponse(http.StatusOK, updated)
}

func (h *Handler) delete(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	const op = "delete banana"

	if err := domain.ValidateID(id); err != nil {
		return h.errorResponse(ctx, err, op)
	}

	deleted, err := h.repo.Delete(ctx, id)
	if err != nil {
		return h.errorResponse(ctx, err, op)
	}

	return platform.SuccessResponse(http.StatusOK, deleted)
}

func (h *Handler) errorResponse(ctx context.Context, err error, operation string) (events.APIGatewayProxyResponse, error) {
	if platform.IsClientError(err) {
		h.logger.InfoContext(ctx, operation+" client error", "error", err.Error())
	} else {
		h.logger.LogError(ctx, operation+" failed", err)
	}

	return platform.ClientErrorResponse(err)
}
