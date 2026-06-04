# Build and deploy the multi-resource API Lambda via AWS SAM.
.PHONY: build
build:
	sam build

# env.json key must match template logical ID: AppnameBackendFunction (see env.json.example).
local: build
	sam local start-api --port 8000 --env-vars env.json

build-AppnameBackendFunction:
	GOOS=linux CGO_ENABLED=0 go build -tags lambda.norpc -o $(ARTIFACTS_DIR)/bootstrap ./cmd/lambda

.PHONY: test
test:
	go test ./...

.PHONY: init
init: build
	sam deploy --guided

.PHONY: deploy
deploy: build
	sam deploy --parameter-overrides AwsCfToken="$(AWS_CF_TOKEN)"

.PHONY: delete
delete:
	sam delete
