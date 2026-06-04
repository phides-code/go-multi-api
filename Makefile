# Build and deploy the multi-resource API Lambda via AWS SAM.
.PHONY: build
build:
	sam build

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
