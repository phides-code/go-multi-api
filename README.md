# go-multi-api

A single AWS Lambda serving a JSON HTTP API backed by DynamoDB. Each URL path maps to one table and one resource type (`/bananas` today). Add resources by registering handlers on the same Lambda.

## How it works

```
API Gateway  →  Lambda (router)  →  resource handler  →  repository  →  DynamoDB
```

The router checks auth, routes by first path segment, and delegates. Handlers parse HTTP, run domain validation, call the repository. Shared HTTP concerns live in `internal/platform`.

## Project layout

```
cmd/lambda/main.go          entrypoint → app.NewRouter
internal/
  domain/                   entities, interfaces, id.go, validation.go, errors.go, pagination.go
  dynamodb/                 repository implementations (+ assert_test.go: assertUpdateSets; banana_assert_test.go)
  handler/                  handlers, router (+ assert_test.go shared envelope helpers; banana_assert_test.go, banana_mocks_test.go)
  platform/                 response envelope, errors, logging, auth
  app/wire.go               construct repos, Register handlers
  testutil/                 shared test constants (e.g. TestCFTToken for router/wire tests)
template.yml                SAM: API Gateway, Lambda, tables
Makefile                    build, test, local, deploy
```

Copy the banana layering for new resources: entity → interface → DynamoDB impl → handler → tests. Reuse shared domain helpers; per-resource files wire `ValidateCreateInput` / `ValidateUpdateInput`.

## API contract

### Authentication

Every request except `OPTIONS` requires `X-CF-Token: <token>` (deploy param `AwsCfToken` → env `AWS_CF_TOKEN`). `make local` sets `AWS_SAM_LOCAL=true` and skips the check.

### Response envelope

```json
{ "data": { ... } | [ ... ] | null, "error": "message" | null }
```

Success: `data` set, `error` null. Failure: opposite.

**Standard client errors** (`internal/platform/errors.go`):

| HTTP | `error` | Domain sentinel | Cause |
| ---- | ------- | --------------- | ----- |
| 400 | `invalid json` | `ErrInvalidJSON` | Bad body |
| 400 | `invalid id` | `ErrInvalidID` | Path `{id}` not UUID |
| 400 | `validation failed` | `ErrValidationFailed` | Domain rule failed |
| 400 | `invalid cursor` | `ErrInvalidCursor` | Bad `?cursor=` |
| 404 | `not found` | `ErrNotFound` | Missing item |
| 409 | `already exists` | `ErrAlreadyExists` | Duplicate create |
| 405 | `method not allowed` | `ErrMethodNotAllowed` | Unsupported method |
| 401 | `unauthorized` | — | Bad/missing token |
| 500 | `internal server error` | — | Unexpected failure |

Return `ErrValidationFailed` from validation; no per-field error strings unless you extend platform mapping and this table. Client-facing text comes from each sentinel's `Error()` in `domain/errors.go` via `platform.ClientErrorMessage`. New cross-cutting errors: add sentinel in `domain/errors.go`, add a row to `clientErrorMappings` in `platform/errors.go`, document here.

### Bananas (`/bananas`)

| Method | Path | Behavior |
| ------ | ---- | -------- |
| `GET` | `/bananas` | List (paginated) |
| `GET` | `/bananas/{id}` | Get by UUID |
| `POST` | `/bananas` | Create; server sets `id`, `createdOn` |
| `PUT` | `/bananas/{id}` | Update `content`; 404 if missing |
| `DELETE` | `/bananas/{id}` | Hard delete; returns deleted item |

**Item shape** (single banana in create/get/update/delete responses; list `items` use the same fields):

```json
{
  "id": "uuid",
  "content": "string",
  "createdOn": 1717516800000
}
```

**Create body** (POST): `{ "content": "string" }`

**Update body** (PUT): `{ "content": "string" }`

**List** (`GET /bananas`): `data.items` (array of item shape), optional `data.nextCursor`. Fixed page size of 50 (`domain.DefaultListLimit`) — not configurable via `?limit=`. Pagination uses `?cursor=<nextCursor>` only; omit `cursor` for the first page. Bad cursor → 400 `invalid cursor`.

**Validation:** `content` required on create/update, 1–1000 Unicode characters (`BananaMinContentLength`–`BananaMaxContentLength` in `banana.go`) → 400 `validation failed`. Path `{id}` must be UUID → 400 `invalid id`.

## Development

Go 1.23+, [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html).

```bash
make test      # unit tests + coverage gate
make build
make local     # API on :8000 (Docker); no auth header needed
```

```bash
curl http://localhost:8000/bananas
curl "http://localhost:8000/bananas?cursor=opaque-cursor"
```

**Deploy:** `make init` (first time), then `export AWS_CF_TOKEN=… && make deploy`. CI (`.github/workflows/go.yml`) tests, builds, deploys on push to `main`.

## Adding a field

Extend an existing resource (e.g. add `description` to `Banana`). **TDD:** failing test → minimum code → green. Domain first, HTTP second, persistence last.

| Step | What | Files |
| ---- | ---- | ----- |
| 1 | **Domain tests** — required string: one wiring row in create/update input tests (`validation_test.go` already covers generic string rules). Enum/format: resource-specific tests in `<resource>_test.go`. | `internal/domain/<resource>_test.go`, optionally `validation_test.go` |
| 2 | **Struct + validation** — field on `Banana` + `json`/`dynamodbav` tags; add to create/update inputs if client-set; wire `ValidateRequiredString` with resource-scoped bounds (e.g. `BananaMinContentLength`) or custom `Validate*`. Server-owned fields (e.g. `createdOn`): set in handler/repo, not inputs. | `internal/domain/<resource>.go` |
| 3 | **Handler tests** — client-error rows (400 `validation failed`; use `panic<Resource>Repo` so validation failures never call the repo); success + `assert<Resource>DataKeys` if wire shape changes. | `internal/handler/<resource>_handler_test.go`, `<resource>_mocks_test.go`, `<resource>_assert_test.go`, `assert_test.go` |
| 4 | **Handler** — parse JSON, build inputs, validate, copy fields to `domain.Banana`. No inline business rules. | `internal/handler/<resource>_handler.go` |
| 5 | **DynamoDB test** (if PUT-updatable) — table rows use `setupMock func(t *testing.T)`; success/error via `assertBananaRepoResult`; create success: `assertBananaPutItem`; update success: `assertUpdateSets(t, params, map[string]string{…})` for **all** updatable attrs. | `internal/dynamodb/<resource>_repository_test.go`, `<resource>_assert_test.go`, `assert_test.go` |
| 6 | **DynamoDB Update** (if PUT-updatable) — add field to SET expression, names, values (alphabetical order). Create/Get/List/Delete usually unchanged (struct tags). No `template.yml` change. | `internal/dynamodb/<resource>_repository.go` |
| 7 | **Docs** — update item/create/update sections above. | this file |

Skip 5–6 for read-only or create-only fields. Optional unvalidated fields: handler round-trip test on create/get.

`make test` before PR.

## Adding a new table

Each table gets its own entity, interface, DynamoDB repo, handler, and tests. Implement only the HTTP methods you need (handler **and** `template.yml`). File checklist and naming rules: **[docs/new-resource.md](docs/new-resource.md)**.

**TDD:** one vertical slice first (e.g. `GET /apples` → empty page), then expand method by method.

| Step | What | Files |
| ---- | ---- | ----- |
| 1 | Failing handler test + router dispatch test | `internal/handler/<resource>_handler_test.go`, `router_test.go` |
| 2 | Domain tests, entity, validation, repository interface | `internal/domain/<resource>_test.go`, `<resource>.go`, `<resource>_repository.go` |
| 3 | HTTP handler (+ tests per method, client errors, one 500 per op) | `internal/handler/<resource>_handler.go` |
| 4 | DynamoDB tests then impl — table-driven tests, `setupMock(t)`, `assertBananaRepoResult`; create success: `assertBananaPutItem`; update success: `assertUpdateSets` | `internal/dynamodb/<resource>_repository_test.go`, `<resource>_repository.go`, `<resource>_assert_test.go`, `assert_test.go` |
| 5 | Wire: `newRouter` constructs repo, `Register("<resources>", …)` | `internal/app/wire.go`, `wire_test.go` |
| 6 | SAM table, `DynamoDBCrudPolicy` per table, API events | `template.yml` |
| 7 | API docs | this file |

Reference: `banana_*` throughout. Errors: use `ErrValidationFailed` unless adding a new cross-cutting sentinel (see standard errors table).

**Second table in a copied project:** extend `wire.go` with another repo + `Register`, add table/policy/events in `template.yml`. Details: [docs/new-resource.md](docs/new-resource.md#second-table-in-a-copied-project).

`make test` before PR.
