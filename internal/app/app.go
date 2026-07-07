// Composition root: loads AWS config, constructs repositories, and registers resource handlers on the gateway.
package app

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/phides-code/go-multi-api/internal/banana"
	"github.com/phides-code/go-multi-api/internal/gateway"
	"github.com/phides-code/go-multi-api/internal/platform"
)

func Build(ctx context.Context, logger *platform.Logger) (*gateway.Gateway, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	bananaRepo := banana.NewRepository(dynamodb.NewFromConfig(cfg))
	return buildGateway(logger, bananaRepo), nil
}

func buildGateway(logger *platform.Logger, bananaRepo banana.Repository) *gateway.Gateway {
	g := gateway.NewGateway(logger)
	g.Register("bananas", banana.NewHandler(bananaRepo, logger))
	return g
}
