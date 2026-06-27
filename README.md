# go-multi-api

A single AWS Lambda that serves a JSON HTTP API backed by DynamoDB. Each URL path maps to one DynamoDB table and one resource type. Today that is `/bananas`; more paths are added by registering new handlers on the same Lambda.

## How it works

```
API Gateway  →  Lambda (router)  →  resource handler  →  repository  →  DynamoDB
```

1. API Gateway forwards requests as `APIGatewayProxyRequest` events.
2. The router checks auth, takes the first path segment (e.g. `bananas`), looks it up in registered handlers, and delegates to that resource's handler.
3. The handler parses the HTTP method and body, runs domain validation, and calls the repository.
4. The repository reads and writes DynamoDB using AWS SDK v2.
5. All responses use the same JSON envelope (see below).

Shared concerns live in `internal/platform`: response formatting, error-to-status mapping, structured logging, and the `X-CF-Token` header check.

## Project layout

```
cmd/lambda/main.go          Lambda entrypoint: wires repos, router, starts handler
internal/
  domain/                   Entities, validation, repository interfaces
  dynamodb/                 DynamoDB repository implementations
  handler/                  HTTP handlers and router
  platform/                 Shared response, errors, logging, auth
template.yml                API Gateway, Lambda, DynamoDB tables (SAM)
Makefile                    build, test, local, deploy
```

Bananas are the reference implementation. When adding a resource, copy the same layering: entity → repository interface → DynamoDB impl → handler → tests.

## API contract

### Authentication

Every request except `OPTIONS` must include:

```
X-CF-Token: <token>
```

The token is set at deploy time (`AwsCfToken` parameter) and exposed to the Lambda as `AWS_CF_TOKEN`.

When you run via `make local` (`sam local start-api`), SAM sets `AWS_SAM_LOCAL=true` in the container and the API skips this check — you do not need `X-CF-Token` or `AWS_CF_TOKEN` in `env.json` for auth.

### Response envelope

All responses are JSON with `Content-Type: application/json`:

```json
{
  "data": { ... } | [ ... ] | null,
  "error": "message" | null
}
```

On success, `error` is null and `data` holds the result. On failure, `data` is null and `error` holds a short client-facing message.

**Standard client errors** (all resources; mapped in `internal/platform/errors.go`):

| HTTP status | `error` message | Domain sentinel | Typical cause |
| ----------- | --------------- | --------------- | ------------- |
| 400 | `invalid json` | `ErrInvalidJSON` | Malformed request body |
| 400 | `invalid id` | `ErrInvalidID` | Path `{id}` is not a UUID |
| 400 | `validation failed` | `ErrValidationFailed` | Field or business rule failed in domain validation |
| 400 | `invalid cursor` | `ErrInvalidCursor` | Bad `?cursor=` on list |
| 404 | `not found` | `ErrNotFound` | Item missing |
| 409 | `already exists` | `ErrAlreadyExists` | Duplicate create |
| 405 | `method not allowed` | `ErrMethodNotAllowed` | Unsupported HTTP method |
| 401 | `unauthorized` | — | Missing or invalid `X-CF-Token` |
| 500 | `internal server error` | — | Unexpected / infrastructure failure |

Return `domain.ErrValidationFailed` from any `Validate*` function when a payload field fails a rule. Do not add per-field sentinels unless you also extend platform mapping and accept a new public `error` string.

### Bananas (`/bananas`)

| Method   | Path            | Behavior                                     |
| -------- | --------------- | -------------------------------------------- |
| `GET`    | `/bananas`      | List bananas (paginated; see below)          |
| `GET`    | `/bananas/{id}` | Get one banana by UUID                       |
| `POST`   | `/bananas`      | Create a banana; server generates UUID v4 id |
| `PUT`    | `/bananas/{id}` | Update `content` only; 404 if missing        |
| `DELETE` | `/bananas/{id}` | Hard delete; returns the deleted item        |

**Banana shape:**

```json
{
    "id": "uuid",
    "content": "string",
    "createdOn": 1717516800000
}
```

**Create body** (POST only — `id` is not accepted from the client):

```json
{ "content": "string" }
```

**Update body** (PUT):

```json
{ "content": "string" }
```

**List response** (`GET /bananas`):

```json
{
    "data": {
        "items": [
            {
                "id": "uuid",
                "content": "string",
                "createdOn": 1717516800000
            }
        ],
        "nextCursor": "opaque-cursor"
    },
    "error": null
}
```

- Returns up to 50 items per page (`domain.DefaultListLimit`).
- `nextCursor` is omitted when there is no next page.
- Pass `?cursor=<nextCursor>` to fetch the next page.

Validation: `content` is required, 1–1000 Unicode characters. Violations return 400 with `"validation failed"`. Path `{id}` values must be valid UUIDs (400 `"invalid id"`).

## Development

**Requirements:** Go 1.23+, [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)

```bash
# Run unit tests
make test

# Build
make build

# Run API locally on port 8000 (requires Docker for sam local)
make local
```

Example local requests (no auth header — SAM local sets `AWS_SAM_LOCAL`):

```bash
curl http://localhost:8000/bananas

# Next page (use nextCursor value from the previous response)
curl "http://localhost:8000/bananas?cursor=opaque-cursor"
```

## Deploy

First-time deploy (interactive):

```bash
make init
```

Subsequent deploys:

```bash
export AWS_CF_TOKEN=your-production-token
make deploy
```

CI (`.github/workflows/go.yml`) runs tests, `sam build`, and deploy on pushes to `main`.

## Adding a field to an existing resource

Use this when extending bananas (or any resource that already follows the template). Example: add `ripeness` to `Banana`.

**TDD order:** write a failing test that specifies the behavior, then the minimum code to pass. Domain rules first; HTTP contract second; persistence last (only if updatable).

### 1. Domain tests

In `internal/domain/banana_test.go`, add table-driven cases for each new rule (empty, too long, invalid enum, etc.).

Run `go test ./internal/domain/ -run TestValidate…` — tests should **fail** until step 2.

### 2. Domain entity and validation

In `internal/domain/banana.go`:

- Add the field to `Banana` with matching `json` and `dynamodbav` tags (attribute names must match what is stored in DynamoDB).
- If clients may set the field on create or update, add it to `CreateBananaInput` / `UpdateBananaInput`.
- Add or extend validation functions; call them from `ValidateCreateInput` / `ValidateUpdateInput`. Return `domain.ErrValidationFailed` when a rule fails.

If the field is **server-owned** (e.g. `createdOn`), do not add it to create/update inputs — set it in the handler or repository instead.

Domain tests from step 1 should now pass.

### 3. Handler tests

In `internal/handler/banana_handler_test.go`:

- Client-error row in `TestBananaHandlerClientErrors` (POST) and/or `TestBananaHandlerUpdate` (PUT) for each rule clients can hit — expect 400 and `"validation failed"` unless the failure is JSON or ID shape (see standard client errors above).
- Success assertions in Create/Update/GetByID/List if the field changes responses; use `assertBananaDataKeys` when the wire shape changes.

Run the new handler tests — they should **fail** until step 4.

### 4. HTTP handler

In `internal/handler/banana_handler.go`:

- Extend the anonymous JSON structs in `create` and `update` to parse the new field from the request body.
- Map parsed values into `CreateBananaInput` / `UpdateBananaInput` before calling domain validation.
- When building the `domain.Banana` passed to the repository, copy validated fields from the input.

Handlers parse JSON and delegate rules to domain — do not validate business rules inline in the handler.

### 5. DynamoDB tests (if updatable)

If the field is updatable via PUT, add an update success case in `internal/dynamodb/banana_repository_test.go` (existing not-found / SDK-error patterns stay the same).

Test should **fail** until step 6.

### 6. DynamoDB adapter (if updatable)

In `internal/dynamodb/banana_repository.go`:

- **Create, GetByID, List, Delete** — usually no code change. `attributevalue.MarshalMap` / `UnmarshalMap` use struct tags; new attributes are read and written automatically.
- **Update** — extend `UpdateExpression`, `ExpressionAttributeNames`, and `ExpressionAttributeValues` (today only `content` is updated). Immutable fields (`id`, `createdOn`) stay out of the expression.

DynamoDB is schemaless: you do not alter `template.yml` for a new optional attribute on an existing table.

Skip steps 5–6 when the field is read-only or set only on create.

### 7. API contract

Update the **Banana shape**, **Create body**, and **Update body** sections above so HTTP docs match what the handler accepts and returns.

Run `make test` before opening a PR.

### Checklist (TDD order)

| Step | File(s) |
| ---- | ------- |
| Domain tests | `internal/domain/banana_test.go` |
| Struct + validation | `internal/domain/banana.go` |
| Handler tests | `internal/handler/banana_handler_test.go` |
| JSON parsing | `internal/handler/banana_handler.go` |
| DynamoDB tests (if updatable) | `internal/dynamodb/banana_repository_test.go` |
| Update expression (if updatable) | `internal/dynamodb/banana_repository.go` |
| API docs | `README.md` (this file) |

Optional fields with no validation: still add a handler test that proves the field round-trips on create/get.

## Adding a new table

Copy the file list from [docs/new-resource.md](docs/new-resource.md) when starting a new resource.

Each DynamoDB table gets its **own** entity, repository interface, DynamoDB implementation, HTTP handler, and tests. Reuse `internal/platform` and the router pattern; do not share handlers or repository interfaces across resources.

A resource only needs the HTTP methods it actually uses. Define those in the handler **and** in `template.yml` — do not implement unused CRUD operations just because bananas have them.

**TDD order:** pick one vertical slice (e.g. `GET /apples` returns an empty page). Write a failing test for the public contract, then the minimum code to pass. Expand method by method.

### 1. First failing test (HTTP contract)

In `internal/handler/<resource>_handler_test.go`:

- Mock the repository interface in the test file.
- Add one test for the thinnest slice you are shipping first (often `List` or `Create`).
- Assert status, envelope shape, and response body.

In `internal/handler/router_test.go`:

- Register the new handler and add a dispatch test (see `TestRouterDispatchesRegisteredPrefix`).

Tests should **fail** — handler, interface, and wiring do not exist yet.

### 2. Domain tests and entity

Create `internal/domain/<resource>_test.go` with validation table cases.

Create `internal/domain/<resource>.go`:

- Struct with `json` and `dynamodbav` tags matching the DynamoDB item shape.
- Input types for create/update payloads.
- Validation functions (reuse `ValidateID` and `NewID` from `internal/domain/id.go` when using UUID keys).

Create `internal/domain/<resource>_repository.go`:

- Define a `<Resource>Repository` interface with only the methods the handler needs.
- Define a `<Resource>Page` type if the resource supports listing (`Items`, `NextCursor`).
- Reuse `ListOptions` if pagination matches the banana pattern.

Domain tests should pass after validation is implemented.

### 3. HTTP handler

Create `internal/handler/<resource>_handler.go`:

- Struct holding the repository interface and `*platform.Logger`.
- `Handle(ctx, req)` switches on `req.HTTPMethod` and `req.PathParameters["id"]`.
- For list endpoints, read pagination from `req.QueryStringParameters` (e.g. `cursor`) and return a page object (`items`, `nextCursor`) in `data`.
- Parse JSON bodies, call domain validation, call the repository.
- Return `platform.SuccessResponse` / `platform.ErrorResponse` via a local `errorResponse` helper (see `banana_handler.go`).

Handler tests from step 1 should pass with the mock repo. Add tests for each supported method, client errors, and one repo-failure → 500 per operation.

### 4. DynamoDB tests and implementation

Create `internal/dynamodb/<resource>_repository_test.go` first:

- CRUD success paths you implement, plus not-found and one SDK error per method (copy the banana patterns).

Create `internal/dynamodb/<resource>_repository.go`:

- Accept `*dynamodb.Client` in the constructor (same as `NewBananaRepository`).
- Hardcode the table name constant (must match `template.yml`).
- Implement each repository method using SDK v2 (`GetItem`, `PutItem`, `Scan`, etc.).
- Map `domain.ErrNotFound` when an item is missing.

Use `internal/dynamodb/banana_repository.go` as a reference for pagination cursor encoding and conditional writes.

### 5. Wire it up

The router dispatches by the first URL path segment. A resource is reachable only after `Register("<prefix>", ...)` — you do not edit `router.go` when adding a table.

**Composition** — in `internal/app/wire.go`, inside `newRouter`:

```go
appleRepo := dynamodbrepo.NewAppleRepository(dynamodb.NewFromConfig(cfg))
router.Register("apples", handler.NewAppleHandler(appleRepo, logger))
```

`cmd/lambda/main.go` calls `app.NewRouter` and starts the Lambda handler.

Wiring is smoke-tested in `internal/app/wire_test.go` — extend or mirror `TestWiringSmokeGETBananas` when appropriate.

### 6. Infrastructure (`template.yml`)

For each new table:

1. Add a `AWS::DynamoDB::Table` resource (partition key `id` unless the design differs).
2. Add a `DynamoDBCrudPolicy` (or a narrower policy) on `AppnameBackendFunction` for that table.
3. Add API Gateway `Events` only for the HTTP methods this resource exposes (include `OPTIONS` for CORS preflight on each path).

**Table naming** — keep these three in sync for each resource:

| What                                               | Pattern                   | Banana example                        |
| -------------------------------------------------- | ------------------------- | ------------------------------------- |
| SAM logical ID                                     | `Appname<Resources>Table` | `AppnameBananasTable`                 |
| Physical table name (`TableName`)                  | `Appname<Resources>`      | `AppnameBananas`                      |
| Go constant in `dynamodb/<resource>_repository.go` | same as physical name     | `bananasTableName = "AppnameBananas"` |

`DynamoDBCrudPolicy` references the SAM logical ID: `TableName: !Ref AppnameBananasTable`.

**IAM** — add one `DynamoDBCrudPolicy` per table on `AppnameBackendFunction`. Reference the table resource (`!Ref AppnameBananasTable`), not a wildcard. When you add apples, append a second policy entry; do not replace or broaden the bananas policy.

```yaml
Policies:
  - AWSLambdaExecute
  - DynamoDBCrudPolicy:
      TableName: !Ref AppnameBananasTable
  - DynamoDBCrudPolicy:
      TableName: !Ref AppnameApplesTable
```

Redeploy after changes: `make deploy`.

### 7. Error mapping

Most errors are already shared — see **Standard client errors** under [API contract](#api-contract).

**When adding a resource or field:**

- **Field / business rules** — return `domain.ErrValidationFailed` from validation; no platform change.
- **New cross-cutting concern** (rare) — add a sentinel in `internal/domain/errors.go` and map status + client message in `internal/platform/errors.go` (`HTTPStatusForError`, `ClientErrorMessage`). Add a row to the standard errors table in this README.

Per-field client messages (e.g. `"invalid ripeness"`) are not supported today; all validation failures surface as `"validation failed"`.

### Checklist (TDD order)

| Step | File(s) |
| ---- | ------- |
| Handler + router tests (first slice) | `internal/handler/<resource>_handler_test.go`, `internal/handler/router_test.go` |
| Domain tests + entity + interface | `internal/domain/<resource>_test.go`, `<resource>.go`, `<resource>_repository.go` |
| HTTP handler | `internal/handler/<resource>_handler.go` |
| DynamoDB tests + impl | `internal/dynamodb/<resource>_repository_test.go`, `<resource>_repository.go` |
| Wiring | `internal/app/wire.go`; smoke test in `internal/app/wire_test.go` |
| AWS resources | `template.yml` |
| Error mapping (if needed) | Usually none — use `ErrValidationFailed`; extend `domain/errors.go` + `platform/errors.go` only for new cross-cutting errors |
| API docs | `README.md` (this file) |

Run `make test` before opening a PR.
