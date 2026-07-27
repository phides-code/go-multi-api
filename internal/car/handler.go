// HTTP handler for /cars: maps API Gateway requests to repository operations.
package car

import (
	"context"
	"encoding/json"
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
	id := strings.TrimSpace(req.PathParameters["id"])

	switch req.HTTPMethod {
	case "GET":
		if id == "" {
			return h.list(ctx, req)
		}
		return h.getByID(ctx, id)
	case "POST":
		return h.create(ctx, req.Body)
	case "PUT":
		return h.update(ctx, id, req.Body)
	case "DELETE":
		return h.delete(ctx, id)
	default:
		return h.errorResponse(ctx, domain.ErrMethodNotAllowed, "car request")
	}
}

func (h *Handler) list(ctx context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	items, err := h.repo.List(ctx)
	if err != nil {
		return h.errorResponse(ctx, err, "list cars")
	}

	return platform.SuccessResponse(200, items)
}

func (h *Handler) getByID(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	if err := domain.ValidateID(id); err != nil {
		return h.errorResponse(ctx, err, "get car")
	}

	c, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return h.errorResponse(ctx, err, "get car")
	}

	return platform.SuccessResponse(200, c)
}

func (h *Handler) create(ctx context.Context, body string) (events.APIGatewayProxyResponse, error) {
	var payload struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return h.errorResponse(ctx, domain.ErrInvalidJSON, "create car")
	}

	input := CreateInput{Model: payload.Model}
	if err := ValidateCreateInput(input); err != nil {
		return h.errorResponse(ctx, err, "create car")
	}

	c := Car{
		ID:        domain.NewID(),
		Model:     payload.Model,
		CreatedOn: uint64(time.Now().UnixMilli()),
	}

	created, err := h.repo.Create(ctx, c)
	if err != nil {
		return h.errorResponse(ctx, err, "create car")
	}

	return platform.SuccessResponse(201, created)
}

func (h *Handler) update(ctx context.Context, id, body string) (events.APIGatewayProxyResponse, error) {
	if err := domain.ValidateID(id); err != nil {
		return h.errorResponse(ctx, err, "update car")
	}

	var payload struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return h.errorResponse(ctx, domain.ErrInvalidJSON, "update car")
	}

	input := UpdateInput{ID: id, Model: payload.Model}
	if err := ValidateUpdateInput(input); err != nil {
		return h.errorResponse(ctx, err, "update car")
	}

	updated, err := h.repo.Update(ctx, Car{
		ID:    id,
		Model: payload.Model,
	})
	if err != nil {
		return h.errorResponse(ctx, err, "update car")
	}

	return platform.SuccessResponse(200, updated)
}

func (h *Handler) delete(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	if err := domain.ValidateID(id); err != nil {
		return h.errorResponse(ctx, err, "delete car")
	}

	deleted, err := h.repo.Delete(ctx, id)
	if err != nil {
		return h.errorResponse(ctx, err, "delete car")
	}

	return platform.SuccessResponse(200, deleted)
}

func (h *Handler) errorResponse(ctx context.Context, err error, operation string) (events.APIGatewayProxyResponse, error) {
	if platform.IsClientError(err) {
		h.logger.InfoContext(ctx, operation+" client error", "error", err.Error())
	} else {
		h.logger.LogError(ctx, operation+" failed", err)
	}

	return platform.ErrorResponse(platform.HTTPStatusForError(err), platform.ClientErrorMessage(err))
}
