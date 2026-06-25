// Request router: dispatches API Gateway paths to registered resource handlers.
package handler

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/phides-code/go-multi-api/internal/platform"
)

type ResourceHandler interface {
	Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}

type Router struct {
	logger   *platform.Logger
	cfToken  string
	handlers map[string]ResourceHandler
}

func NewRouter(logger *platform.Logger) *Router {
	return NewRouterWithCFTToken(logger, os.Getenv("AWS_CF_TOKEN"))
}

func NewRouterWithCFTToken(logger *platform.Logger, cfToken string) *Router {
	return &Router{
		logger:   logger,
		cfToken:  cfToken,
		handlers: make(map[string]ResourceHandler),
	}
}

func (r *Router) Register(prefix string, handler ResourceHandler) {
	r.handlers[prefix] = handler
}

func (r *Router) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != http.MethodOptions &&
		!platform.LocalMode() &&
		!platform.ValidCFTToken(r.cfToken, req.Headers) {
		return platform.ErrorResponse(http.StatusUnauthorized, "unauthorized")
	}

	logger := r.logger.WithRequestID(req.RequestContext.RequestID)
	logger.InfoContext(ctx, "incoming request",
		"method", req.HTTPMethod,
		"path", req.Path,
	)

	segment, ok := firstPathSegment(req.Path)

	if !ok {
		return platform.ErrorResponse(404, "not found")
	}

	handler, ok := r.handlers[segment]
	if !ok {
		return platform.ErrorResponse(404, "not found")
	}

	return handler.Handle(ctx, req)
}

func firstPathSegment(path string) (string, bool) {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "", false
	}
	return strings.Split(trimmed, "/")[0], true
}
