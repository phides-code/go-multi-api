# Build and deploy the multi-resource API Lambda via AWS SAM.

COVERAGE_MIN_TOTAL ?= 69
COVERAGE_MIN_GATEWAY ?= 85
COVERAGE_MIN_BANANA ?= 85

.PHONY: test
test:
	@go test ./internal/... -coverprofile=coverage.out
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub(/%/,"",$$3); print $$3}'); \
	awk -v total="$$total" -v min=$(COVERAGE_MIN_TOTAL) 'BEGIN { if (total+0 < min+0) exit 1 }' || \
		(echo "total coverage $$total% < $(COVERAGE_MIN_TOTAL)%"; exit 1); \
	gateway=$$(go test ./internal/gateway/ -cover 2>&1 | awk '/coverage:/ {gsub(/%/,""); print $$5}'); \
	awk -v cov="$$gateway" -v min=$(COVERAGE_MIN_GATEWAY) 'BEGIN { if (cov+0 < min+0) exit 1 }' || \
		(echo "gateway coverage $$gateway% < $(COVERAGE_MIN_GATEWAY)%"; exit 1); \
	banana=$$(go test ./internal/banana/ -cover 2>&1 | awk '/coverage:/ {gsub(/%/,""); print $$5}'); \
	awk -v cov="$$banana" -v min=$(COVERAGE_MIN_BANANA) 'BEGIN { if (cov+0 < min+0) exit 1 }' || \
		(echo "banana coverage $$banana% < $(COVERAGE_MIN_BANANA)%"; exit 1); \
	echo "coverage OK (total $$total%, gateway $$gateway%, banana $$banana%)"

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
