# go-multi-api

A single AWS Lambda serving a JSON HTTP API backed by DynamoDB. Each URL path maps to one table and one resource type (`/bananas` today). Add resources by registering handlers on the same Lambda.

## How it works

```
API Gateway  →  Lambda (router)  →  resource handler  →  repository  →  DynamoDB
```

The router checks auth, routes by first path segment, and delegates. Each resource is a vertical slice under `internal/<resource>/`. Shared cross-cutting rules live in `internal/domain`; HTTP envelope and auth in `internal/platform`.

## Project layout

```
cmd/lambda/main.go          entrypoint → app.Build
internal/
  domain/                   cross-cutting: errors, id, validation
  gateway/                  auth gate + path routing; Register(prefix, ResourceHandler)
  banana/                   vertical slice: entity, repository, handler, dynamodb
  platform/                 response envelope, errors, logging, auth
  app/app.go                composition root: construct repos, register handlers
  testutil/                 shared test helpers (TestCFTToken, envelope/dynamodb asserts, fixtures)
template.yml                SAM: API Gateway, Lambda, tables
Makefile                    build, test, local, deploy
```

Copy `internal/banana/` for new resources. Reuse `domain.ValidateRequiredString` and `domain.ValidateID`; wire resource-specific validation in `<resource>.go`.

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
| 404 | `not found` | `ErrNotFound` | Missing item |
| 409 | `already exists` | `ErrAlreadyExists` | Duplicate create |
| 405 | `method not allowed` | `ErrMethodNotAllowed` | Unsupported method |
| 401 | `unauthorized` | — | Bad/missing token |
| 500 | `internal server error` | — | Unexpected failure |

Return `ErrValidationFailed` from validation; no per-field error strings unless you extend platform mapping and this table. Client-facing text comes from each sentinel's `Error()` in `domain/errors.go` via `platform.ClientErrorMessage`. New cross-cutting errors: add sentinel in `domain/errors.go`, add a row to `clientErrorMappings` in `platform/errors.go`, document here.

### Bananas (`/bananas`)

| Method | Path | Behavior |
| ------ | ---- | -------- |
| `GET` | `/bananas` | List all |
| `GET` | `/bananas/{id}` | Get by UUID |
| `POST` | `/bananas` | Create; server sets `id`, `createdOn` |
| `PUT` | `/bananas/{id}` | Update `content`; 404 if missing |
| `DELETE` | `/bananas/{id}` | Hard delete; returns deleted item |

**Item shape** (single banana in create/get/update/delete responses; list returns an array of the same shape):

```json
{
  "id": "uuid",
  "content": "string",
  "createdOn": 1717516800000
}
```

**Create body** (POST): `{ "content": "string" }`

**Update body** (PUT): `{ "content": "string" }`

**List** (`GET /bananas`): `data` is an array of item shape. The repository scans the full table (DynamoDB pagination is handled internally, not exposed over HTTP).

**Validation:** `content` required on create/update, 1–1000 Unicode characters (`MinContentLength`–`MaxContentLength` in `internal/banana/banana.go`) → 400 `validation failed`. Path `{id}` must be UUID → 400 `invalid id`.

## Development

Go 1.23+, [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html).

```bash
make test      # unit tests + coverage gate (69% total, 85% gateway, 85% banana)
make build
make local     # API on :8000 (Docker); no auth header needed
```

```bash
curl http://localhost:8000/bananas
```

**Deploy:** `make init` (first time), then `export AWS_CF_TOKEN=… && make deploy`. CI (`.github/workflows/go.yml`) tests, builds, deploys on push to `main`.

## Adding a field

Extend an existing resource (e.g. add `description` to `Banana`). **TDD:** failing test → minimum code → green. Validation first, HTTP second, persistence last — all within `internal/<resource>/`.

| Step | What | Files |
| ---- | ---- | ----- |
| 1 | **Validation tests** — wiring row in create/update input tests (`domain/validation_test.go` already covers generic string rules). | `internal/<resource>/<resource>_test.go` |
| 2 | **Struct + validation** — field on entity + `json`/`dynamodbav` tags; add to create/update inputs if client-set; wire `domain.ValidateRequiredString` or custom rules. Server-owned fields: set in handler/dynamodb, not inputs. | `internal/<resource>/<resource>.go` |
| 3 | **Handler tests** — client-error rows (400 `validation failed`; use `panic<Resource>Repo`); success + `assert<Resource>DataKeys` if wire shape changes. | `internal/<resource>/handler_test.go`, `mocks_test.go`, `assert_test.go` |
| 4 | **Handler** — parse JSON, validate, call repo. | `internal/<resource>/handler.go` |
| 5 | **DynamoDB test** (if PUT-updatable) — `setupMock(t)`; `assert<Resource>RepoResult`; create: `assert<Resource>PutItem`; update: `testutil.AssertUpdateSets` (copy from `internal/banana/assert_test.go`). | `internal/<resource>/dynamodb_test.go`, `assert_test.go` |
| 6 | **DynamoDB impl** — add field to SET expression (alphabetical). Usually no `template.yml` change. | `internal/<resource>/dynamodb.go` |
| 7 | **Docs** — update item/create/update sections above. | this file |

Skip 5–6 for read-only or create-only fields. Optional unvalidated fields: handler round-trip test on create/get.

`make test` before PR.

## Adding a new table

Each table gets its own package under `internal/<resource>/`. Implement only the HTTP methods you need (handler **and** `template.yml`). Checklist: **[docs/new-resource.md](docs/new-resource.md)**.

**TDD:** one vertical slice first (e.g. `GET /apples` → empty page), then expand method by method.

| Step | What | Files |
| ---- | ---- | ----- |
| 1 | Copy `internal/banana/` → `internal/<resource>/`; failing handler + router integration tests | `internal/<resource>/handler_test.go`, `router_test.go` |
| 2 | Entity, validation, repository interface | `internal/<resource>/<resource>.go`, `repository.go` |
| 3 | HTTP handler (+ tests per method, client errors, one 500 per op) | `internal/<resource>/handler.go` |
| 4 | DynamoDB tests then impl | `internal/<resource>/dynamodb_test.go`, `dynamodb.go` |
| 5 | Compose: construct repo, `Register("<resources>", …)` on gateway | `internal/app/app.go`, `app_test.go` |
| 6 | SAM table, `DynamoDBCrudPolicy` per table, API events | `template.yml` |
| 7 | API docs | this file |

Reference: `internal/banana/`. Errors: use `domain.ErrValidationFailed` unless adding a new cross-cutting sentinel (see standard errors table).

**Second table:** copy the banana package, register in `app/app.go`, extend `template.yml`. Details: [docs/new-resource.md](docs/new-resource.md#second-table-in-the-same-project).

`make test` before PR.
