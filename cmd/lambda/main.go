// Lambda entrypoint: builds the gateway and starts the Lambda handler.
package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/phides-code/go-multi-api/internal/app"
	"github.com/phides-code/go-multi-api/internal/platform"
)

func main() {
	logger := platform.NewLogger()

	g, err := app.Build(context.Background(), logger)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}

	lambda.Start(g.Handle)
}
