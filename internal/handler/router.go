// Request router: dispatches API Gateway paths to registered resource handlers.
package handler

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/platform"
)

type ResourceHandler interface {
	Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}

type Router struct {
	logger   *platform.Logger
	handlers map[string]ResourceHandler
}

func NewRouter(logger *platform.Logger) *Router {
	return &Router{
		logger:   logger,
		handlers: make(map[string]ResourceHandler),
	}
}

func (r *Router) Register(prefix string, handler ResourceHandler) {
	r.handlers[prefix] = handler
}

func (r *Router) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := r.logger.WithRequestID(req.RequestContext.RequestID)
	logger.InfoContext(ctx, "incoming request",
		"method", req.HTTPMethod,
		"path", req.Path,
	)

	resource, ok := matchResource(req.Path)
	if !ok {
		return platform.ErrorResponse(404, "not found")
	}

	handler, ok := r.handlers[resource]
	if !ok {
		return platform.ErrorResponse(404, "not found")
	}

	return handler.Handle(ctx, req)
}

func matchResource(path string) (string, bool) {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "", false
	}

	segment := strings.Split(trimmed, "/")[0]
	switch segment {
	case "bananas":
		return "bananas", true
	default:
		return "", false
	}
}
