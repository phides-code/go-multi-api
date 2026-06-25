package app

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/phides-code/go-multi-api/internal/domain"
	dynamodbrepo "github.com/phides-code/go-multi-api/internal/dynamodb"
	"github.com/phides-code/go-multi-api/internal/handler"
	"github.com/phides-code/go-multi-api/internal/platform"
)

func NewRouter(ctx context.Context, logger *platform.Logger) (*handler.Router, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	bananaRepo := dynamodbrepo.NewBananaRepository(dynamodb.NewFromConfig(cfg))
	return newRouter(logger, bananaRepo), nil
}

func newRouter(logger *platform.Logger, bananaRepo domain.BananaRepository) *handler.Router {
	router := handler.NewRouter(logger)
	router.Register("bananas", handler.NewBananaHandler(bananaRepo, logger))
	return router
}
