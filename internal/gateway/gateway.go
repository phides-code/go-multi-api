// API gateway: auth gate and first-path-segment routing to registered resource handlers.
package gateway

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

type Gateway struct {
	logger   *platform.Logger
	cfToken  string
	handlers map[string]ResourceHandler
}

func NewGateway(logger *platform.Logger) *Gateway {
	return NewGatewayWithCFTToken(logger, os.Getenv("AWS_CF_TOKEN"))
}

func NewGatewayWithCFTToken(logger *platform.Logger, cfToken string) *Gateway {
	return &Gateway{
		logger:   logger,
		cfToken:  cfToken,
		handlers: make(map[string]ResourceHandler),
	}
}

func (g *Gateway) Register(prefix string, handler ResourceHandler) {
	g.handlers[prefix] = handler
}

func (g *Gateway) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod != http.MethodOptions &&
		!platform.LocalMode() &&
		!platform.ValidCFTToken(g.cfToken, req.Headers) {
		return platform.ErrorResponse(http.StatusUnauthorized, "unauthorized")
	}

	logger := g.logger.WithRequestID(req.RequestContext.RequestID)
	logger.InfoContext(ctx, "incoming request",
		"method", req.HTTPMethod,
		"path", req.Path,
	)

	segment, ok := firstPathSegment(req.Path)

	if !ok {
		return platform.ErrorResponse(404, "not found")
	}

	handler, ok := g.handlers[segment]
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
