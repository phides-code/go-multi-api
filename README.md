# go-multi-api

An AWS Lambda serving a JSON HTTP API backed by DynamoDB. Each URL path maps to one table and one resource type (`/bananas` is the example implementation). Add resources by registering handlers on the same Lambda.

```
API Gateway → Lambda (gateway) → resource handler → repository → DynamoDB
```

The gateway authenticates, routes on the first path segment, and delegates. Each resource owns its entity, validation, HTTP handler, and DynamoDB code under `internal/<resource>/`. Cross-cutting rules live in `domain` and `platform`.

## Quick links

| I want to…                          | Go here                                      |
| ----------------------------------- | -------------------------------------------- |
| Run tests / local API               | [Development](#development)                  |
| Understand `/bananas`               | [Bananas](#bananas-bananas)                  |
| Add a field to an existing resource | [Adding a field](#adding-a-field)            |
| Add a new table / URL prefix        | [docs/new-resource.md](docs/new-resource.md) |

## Project layout

```
cmd/lambda/main.go       Lambda entry → app.Build
internal/
  app/                   Composition root (wire repos + Register)
  banana/                Reference vertical slice (copy this)
  domain/                Shared errors, UUID rules, string/int validation
  gateway/               Auth gate + first-segment routing
  platform/              Response envelope, error mapping, logging, CF token
  testutil/              Shared test helpers and banana fixtures
template.yml             SAM: API, Lambda, tables
Makefile                 test, build, local, deploy
```

**Dependency direction:** `app.Build` → `gateway` → handler → repository → DynamoDB. Handlers never call DynamoDB directly. Resources do not import each other.

Copy `internal/banana/` for a new resource. Keep composition stubs in `internal/app/*_stub_test.go` (Go cannot import another package’s tests).

## API contract

### Authentication

Every request except `OPTIONS` needs header `X-CF-Token` (`platform.CFTTokenHeader`). SAM parameter `AwsCfToken` maps to env `AWS_CF_TOKEN` (`platform.CFTTokenEnvVar`).

Under `sam local` (`AWS_SAM_LOCAL` = `true` or `1`), the token check is skipped.

### Response envelope

```json
{ "data": { ... } | [ ... ] | null, "error": "message" | null }
```

Success sets `data` and leaves `error` null. Failure does the opposite.

| HTTP | `error`                 | Domain sentinel       | When                         |
| ---- | ----------------------- | --------------------- | ---------------------------- |
| 400  | `invalid json`          | `ErrInvalidJSON`      | Body is not JSON             |
| 400  | `invalid id`            | `ErrInvalidID`        | Path `{id}` is not a UUID    |
| 400  | `validation failed`     | `ErrValidationFailed` | Domain rule failed           |
| 401  | `unauthorized`          | `ErrUnauthorized`     | Missing/wrong token          |
| 404  | `not found`             | `ErrNotFound`         | Missing item / unknown route |
| 405  | `method not allowed`    | `ErrMethodNotAllowed` | Unsupported method           |
| 409  | `already exists`        | `ErrAlreadyExists`    | Duplicate create             |
| 500  | `internal server error` | —                     | Unexpected failure           |

Client-facing text is the sentinel’s `Error()` string (`domain/errors.go`), mapped in `platform/errors.go`. Prefer `ErrValidationFailed` for field rules; add a new sentinel only when you need a new HTTP status or message for every resource.

### Bananas (`/bananas`)

| Method   | Path            | Behavior                                    |
| -------- | --------------- | ------------------------------------------- |
| `GET`    | `/bananas`      | List all                                    |
| `GET`    | `/bananas/{id}` | Get by UUID                                 |
| `POST`   | `/bananas`      | Create (`id` and `createdOn` set by server) |
| `PUT`    | `/bananas/{id}` | Update `descriptor` and `rating`            |
| `DELETE` | `/bananas/{id}` | Hard delete; returns the deleted item       |

**Item** (list returns an array of the same shape):

```json
{
    "id": "uuid",
    "descriptor": "string",
    "rating": 50,
    "createdOn": 1717516800000
}
```

**Create / update body:** `{ "descriptor": "string", "rating": 0 }`

**Validation**

- `descriptor`: required, 1–100 Unicode characters (`domain.DefaultMinStringLength`–`DefaultMaxStringLength`)
- `rating`: required integer 0–100 (`domain.DefaultMinInt`–`DefaultMaxInt`)
- Path `{id}`: UUID, or 400 `invalid id`

List scans the full table. DynamoDB pagination stays inside the repository; it is not exposed over HTTP.

## Development

Requires Go 1.23+ and [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html).

```bash
make test      # unit tests + coverage gates (see Makefile)
make build
make local     # API on :8000; no auth header required
```

```bash
curl http://localhost:8000/bananas
```

**Deploy:** `make init` once, then `export AWS_CF_TOKEN=… && make deploy`. CI (`.github/workflows/go.yml`) tests, builds, and deploys on push to `main`.

## Adding a field

Example: add `origin` to banana. Stay inside `internal/<resource>/` (plus shared fixtures / README). Prefer TDD: failing test → smallest fix → green.

1. **Validation test** — In `<resource>_test.go`, extend the local `validCreateInput` / `validUpdateInput` helpers and add a case that blanks (or otherwise breaks) the new field.
2. **Entity + validation** — Add the field to the entity (`json` + `dynamodbav` tags). If clients set it, add it to create/update inputs and validate with `domain.ValidateRequiredString` / `ValidateRequiredInt` (or a custom rule). Server-owned fields are set in the handler, not in inputs.
3. **Fixtures** — Extend `testutil.<Resource>Body` / `Valid<Resource>Body`, package validation bodies, and wire-key asserts (`assert<Resource>DataKeys` uses `Attr*` constants).
4. **Handler tests** — Client-error rows for bad values; success paths that assert the new field when it appears in the response.
5. **Handler** — Parse, validate, pass through to the repository.
6. **DynamoDB** — If the field is updatable, extend the Update `SET` expression (keep attribute names alphabetical) and the update test’s `AssertUpdateSets` map. Create usually needs no expression change (full `PutItem`).
7. **Docs** — Update the resource section in this README.

Skip steps 6 for read-only or create-only fields. Run `make test` before opening a PR.

## Adding a new resource

Copy `internal/banana/` and follow the checklist: **[docs/new-resource.md](docs/new-resource.md)**.

Short version:

1. Copy the package; rename symbols (`PathPrefix`, `TableName`, types, tests).
2. Implement only the HTTP methods you need (handler **and** matching `template.yml` events).
3. Wire `NewRepository` + `Register` in `app.Build` (reuse the shared DynamoDB client).
4. Add `internal/app/<resource>_stub_test.go` and a smoke path in `app_test.go`.
5. Add the SAM table, one `DynamoDBCrudPolicy` per table, and API events.
6. Document the resource in this README.

Use `domain.ErrValidationFailed` unless you are introducing a new cross-cutting sentinel (see the errors table above).
