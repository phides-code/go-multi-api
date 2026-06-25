// Lambda entrypoint: wires the HTTP router and starts the Lambda handler.
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

	router, err := app.NewRouter(context.Background(), logger)
	if err != nil {
		log.Fatalf("wire router: %v", err)
	}

	lambda.Start(router.Handle)
}
