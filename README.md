# go-multi-api

A single AWS Lambda that serves a JSON HTTP API backed by DynamoDB. Each URL path maps to one DynamoDB table and one resource type. Today that is `/bananas`; more paths are added by registering new handlers on the same Lambda.

## How it works

```
API Gateway  →  Lambda (router)  →  resource handler  →  repository  →  DynamoDB
```

1. API Gateway forwards requests as `APIGatewayProxyRequest` events.
2. The router checks auth, matches the first path segment (e.g. `bananas`), and delegates to that resource's handler.
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

### Bananas (`/bananas`)

| Method | Path | Behavior |
|--------|------|----------|
| `GET` | `/bananas` | List bananas (paginated; see below) |
| `GET` | `/bananas/{id}` | Get one banana by UUID |
| `POST` | `/bananas` | Create a banana; server generates UUID v4 id |
| `PUT` | `/bananas/{id}` | Update `content` only; 404 if missing |
| `DELETE` | `/bananas/{id}` | Hard delete; returns the deleted item |

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

Validation: `content` is required, 1–1000 Unicode characters. Path `{id}` values must be valid UUIDs.

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

Work top-down: domain → handler → repository (only if updatable) → tests → API docs in this README.

### 1. Domain entity and validation

In `internal/domain/banana.go`:

- Add the field to `Banana` with matching `json` and `dynamodbav` tags (attribute names must match what is stored in DynamoDB).
- If clients may set the field on create or update, add it to `CreateBananaInput` / `UpdateBananaInput`.
- Add or extend validation functions; call them from `ValidateCreateInput` / `ValidateUpdateInput`.

Add table-driven cases in `internal/domain/banana_test.go` for each rule (empty, too long, invalid enum, etc.).

If the field is **server-owned** (e.g. `createdOn`), do not add it to create/update inputs — set it in the handler or repository instead.

### 2. HTTP handler

In `internal/handler/banana_handler.go`:

- Extend the anonymous JSON structs in `create` and `update` to parse the new field from the request body.
- Map parsed values into `CreateBananaInput` / `UpdateBananaInput` before calling domain validation.
- When building the `domain.Banana` passed to the repository, copy validated fields from the input.

Handlers parse JSON and delegate rules to domain — do not validate business rules inline in the handler.

### 3. DynamoDB adapter

In `internal/dynamodb/banana_repository.go`:

- **Create, GetByID, List, Delete** — usually no code change. `attributevalue.MarshalMap` / `UnmarshalMap` use struct tags; new attributes are read and written automatically.
- **Update** — if the field is updatable via PUT, extend `UpdateExpression`, `ExpressionAttributeNames`, and `ExpressionAttributeValues` (today only `content` is updated). Immutable fields (`id`, `createdOn`) stay out of the expression.

DynamoDB is schemaless: you do not alter `template.yml` for a new optional attribute on an existing table.

### 4. Tests

| Layer | File | What to add |
|-------|------|-------------|
| Domain | `internal/domain/banana_test.go` | Validation edge cases for the new field |
| Handler | `internal/handler/banana_handler_test.go` | Client-error row in `TestBananaHandlerClientErrors`; success assertions in Create/Update/GetByID/List if behavior changes |
| DynamoDB | `internal/dynamodb/banana_repository_test.go` | Update success path if the field is updatable; existing not-found / SDK-error patterns stay the same |

Run `make test` before opening a PR.

### 5. API contract

Update the **Banana shape**, **Create body**, and **Update body** sections above so HTTP docs match what the handler accepts and returns.

### Checklist

| Step | File(s) |
|------|---------|
| Struct + tags | `internal/domain/banana.go` |
| Validation | `internal/domain/banana.go`, `banana_test.go` |
| JSON parsing | `internal/handler/banana_handler.go` |
| Update expression (if updatable) | `internal/dynamodb/banana_repository.go`, `banana_repository_test.go` |
| HTTP tests | `internal/handler/banana_handler_test.go` |
| API docs | `README.md` (this file) |

Optional fields with no validation need only struct tags and handler JSON wiring — still add a handler test that proves the field round-trips on create/get.

## Adding a new table

Each DynamoDB table gets its **own** entity, repository interface, DynamoDB implementation, HTTP handler, and tests. Reuse `internal/platform` and the router pattern; do not share handlers or repository interfaces across resources.

A resource only needs the HTTP methods it actually uses. Define those in the handler **and** in `template.yml` — do not implement unused CRUD operations just because bananas have them.

### 1. Domain entity and validation

Create `internal/domain/<resource>.go`:

- Struct with `json` and `dynamodbav` tags matching the DynamoDB item shape.
- Input types for create/update payloads.
- Validation functions (reuse `ValidateID` and `NewID` from `internal/domain/id.go` when using UUID keys).

Add `internal/domain/<resource>_test.go` for validation rules.

### 2. Repository interface

Create `internal/domain/<resource>_repository.go`:

- Define a `<Resource>Repository` interface with only the methods the handler needs (e.g. `List` and `Create`, or full CRUD like bananas).
- Define a `<Resource>Page` type if the resource supports listing (`Items`, `NextCursor`).
- Reuse `ListOptions` if pagination matches the banana pattern.

### 3. DynamoDB implementation

Create `internal/dynamodb/<resource>_repository.go`:

- Accept `*dynamodb.Client` in the constructor (same as `NewBananaRepository`).
- Hardcode the table name constant (must match `template.yml`).
- Implement each repository method using SDK v2 (`GetItem`, `PutItem`, `Scan`, etc.).
- Map `domain.ErrNotFound` when an item is missing.

Use `internal/dynamodb/banana_repository.go` as a reference for pagination cursor encoding and conditional writes.

### 4. HTTP handler

Create `internal/handler/<resource>_handler.go`:

- Struct holding the repository interface and `*platform.Logger`.
- `Handle(ctx, req)` switches on `req.HTTPMethod` and `req.PathParameters["id"]`.
- For list endpoints, read pagination from `req.QueryStringParameters` (e.g. `cursor`) and return a page object (`items`, `nextCursor`) in `data`.
- Parse JSON bodies, call domain validation, call the repository.
- Return `platform.SuccessResponse` / `platform.ErrorResponse` via a local `errorResponse` helper (see `banana_handler.go`).

Create `internal/handler/<resource>_handler_test.go`:

- Mock the repository interface in the test file.
- Test each supported method, status codes, and the response envelope.
- For router integration, add a dispatch test in `router_test.go`.

### 5. Wire it up

**Router** — add a case in `internal/handler/router.go` `matchResource`:

```go
case "apples":
    return "apples", true
```

**Lambda entrypoint** — in `cmd/lambda/main.go`:

```go
appleRepo := dynamodbrepo.NewAppleRepository(dynamodb.NewFromConfig(cfg))
router.Register("apples", handler.NewAppleHandler(appleRepo, logger))
```

### 6. Infrastructure (`template.yml`)

For each new table:

1. Add a `AWS::DynamoDB::Table` resource (partition key `id` unless the design differs).
2. Add a `DynamoDBCrudPolicy` (or a narrower policy) on `AppnameBackendFunction` for that table.
3. Add API Gateway `Events` only for the HTTP methods this resource exposes (include `OPTIONS` for CORS preflight on each path).

Redeploy after changes: `make deploy`.

### 7. Error mapping (if needed)

If the new resource introduces new client errors, add sentinel errors in `internal/domain/errors.go` and map them in `internal/platform/errors.go` (`HTTPStatusForError`, `ClientErrorMessage`).

### Checklist

| Step | File(s) |
|------|---------|
| Entity + validation | `internal/domain/<resource>.go`, `<resource>_test.go` |
| Repository interface | `internal/domain/<resource>_repository.go` |
| DynamoDB impl | `internal/dynamodb/<resource>_repository.go` |
| HTTP handler | `internal/handler/<resource>_handler.go`, `<resource>_handler_test.go` |
| Router | `internal/handler/router.go`, `router_test.go` |
| Wiring | `cmd/lambda/main.go` |
| AWS resources | `template.yml` |

Run `make test` before opening a PR.
