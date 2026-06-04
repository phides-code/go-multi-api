// Lambda entrypoint: loads AWS config, wires DynamoDB repositories, and starts the HTTP router.
package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbrepo "github.com/phides-code/go-multi-api/internal/dynamodb"
	"github.com/phides-code/go-multi-api/internal/handler"
	"github.com/phides-code/go-multi-api/internal/platform"
)

func main() {
	logger := platform.NewLogger()

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}

	bananaRepo := dynamodbrepo.NewBananaRepository(dynamodb.NewFromConfig(cfg))

	router := handler.NewRouter(logger)
	router.Register("bananas", handler.NewBananaHandler(bananaRepo, logger))

	lambda.Start(router.Handle)
}
