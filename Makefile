# Build and deploy the multi-resource API Lambda via AWS SAM.

COVERAGE_MIN_TOTAL ?= 75
COVERAGE_MIN_HANDLER ?= 85
COVERAGE_MIN_DYNAMODB ?= 85

.PHONY: test
test:
	@go test ./internal/... -coverprofile=coverage.out
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub(/%/,"",$$3); print $$3}'); \
	awk -v total="$$total" -v min=$(COVERAGE_MIN_TOTAL) 'BEGIN { if (total+0 < min+0) exit 1 }' || \
		(echo "total coverage $$total% < $(COVERAGE_MIN_TOTAL)%"; exit 1); \
	handler=$$(go test ./internal/handler/ -cover 2>&1 | awk '/coverage:/ {gsub(/%/,""); print $$5}'); \
	awk -v cov="$$handler" -v min=$(COVERAGE_MIN_HANDLER) 'BEGIN { if (cov+0 < min+0) exit 1 }' || \
		(echo "handler coverage $$handler% < $(COVERAGE_MIN_HANDLER)%"; exit 1); \
	dynamo=$$(go test ./internal/dynamodb/ -cover 2>&1 | awk '/coverage:/ {gsub(/%/,""); print $$5}'); \
	awk -v cov="$$dynamo" -v min=$(COVERAGE_MIN_DYNAMODB) 'BEGIN { if (cov+0 < min+0) exit 1 }' || \
		(echo "dynamodb coverage $$dynamo% < $(COVERAGE_MIN_DYNAMODB)%"; exit 1); \
	echo "coverage OK (total $$total%, handler $$handler%, dynamodb $$dynamo%)"

.PHONY: build
build:
	sam build

local: build
	sam local start-api --port 8000

build-AppnameBackendFunction:
	GOOS=linux CGO_ENABLED=0 go build -tags lambda.norpc -o $(ARTIFACTS_DIR)/bootstrap ./cmd/lambda

.PHONY: init
init: build
	sam deploy --guided

.PHONY: deploy
deploy: build
	sam deploy --parameter-overrides AwsCfToken="$(AWS_CF_TOKEN)"

.PHONY: delete
delete:
	sam delete
