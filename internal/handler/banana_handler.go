// HTTP handler for /bananas: maps API Gateway requests to domain operations.
package handler

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/domain"
	"github.com/phides-code/go-multi-api/internal/platform"
)

type BananaHandler struct {
	repo   domain.BananaRepository
	logger *platform.Logger
}

func NewBananaHandler(repo domain.BananaRepository, logger *platform.Logger) *BananaHandler {
	return &BananaHandler{repo: repo, logger: logger}
}

func (h *BananaHandler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
		return h.errorResponse(ctx, domain.ErrMethodNotAllowed, "banana request")
	}
}

func (h *BananaHandler) list(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cursor := strings.TrimSpace(req.QueryStringParameters["cursor"])

	page, err := h.repo.List(ctx, domain.ListOptions{
		Limit:  domain.DefaultListLimit,
		Cursor: cursor,
	})
	if err != nil {
		return h.errorResponse(ctx, err, "list bananas")
	}

	return platform.SuccessResponse(200, page)
}

func (h *BananaHandler) getByID(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	if err := domain.ValidateID(id); err != nil {
		return h.errorResponse(ctx, err, "get banana")
	}

	banana, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return h.errorResponse(ctx, err, "get banana")
	}

	return platform.SuccessResponse(200, banana)
}

func (h *BananaHandler) create(ctx context.Context, body string) (events.APIGatewayProxyResponse, error) {
	var payload struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return h.errorResponse(ctx, domain.ErrInvalidJSON, "create banana")
	}

	input := domain.CreateBananaInput{Content: payload.Content}
	if err := domain.ValidateCreateInput(input); err != nil {
		return h.errorResponse(ctx, err, "create banana")
	}

	banana := domain.Banana{
		ID:        domain.NewID(),
		Content:   payload.Content,
		CreatedOn: uint64(time.Now().UnixMilli()),
	}

	created, err := h.repo.Create(ctx, banana)
	if err != nil {
		return h.errorResponse(ctx, err, "create banana")
	}

	return platform.SuccessResponse(201, created)
}

func (h *BananaHandler) update(ctx context.Context, id, body string) (events.APIGatewayProxyResponse, error) {
	if err := domain.ValidateID(id); err != nil {
		return h.errorResponse(ctx, err, "update banana")
	}

	var payload struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return h.errorResponse(ctx, domain.ErrInvalidJSON, "update banana")
	}

	input := domain.UpdateBananaInput{ID: id, Content: payload.Content}
	if err := domain.ValidateUpdateInput(input); err != nil {
		return h.errorResponse(ctx, err, "update banana")
	}

	updated, err := h.repo.Update(ctx, domain.Banana{
		ID:      id,
		Content: payload.Content,
	})
	if err != nil {
		return h.errorResponse(ctx, err, "update banana")
	}

	return platform.SuccessResponse(200, updated)
}

func (h *BananaHandler) delete(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	if err := domain.ValidateID(id); err != nil {
		return h.errorResponse(ctx, err, "delete banana")
	}

	deleted, err := h.repo.Delete(ctx, id)
	if err != nil {
		return h.errorResponse(ctx, err, "delete banana")
	}

	return platform.SuccessResponse(200, deleted)
}

func (h *BananaHandler) errorResponse(ctx context.Context, err error, operation string) (events.APIGatewayProxyResponse, error) {
	if platform.IsClientError(err) {
		h.logger.InfoContext(ctx, operation+" client error", "error", err.Error())
	} else {
		h.logger.LogError(ctx, operation+" failed", err)
	}

	return platform.ErrorResponse(platform.HTTPStatusForError(err), platform.ClientErrorMessage(err))
}
