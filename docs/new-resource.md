# New resource checklist

Replace `<resource>` / `<Resource>` / `<resources>` (e.g. `apple`, `Apple`, `apples`).

Full walkthrough: [Adding a new table](../README.md#adding-a-new-table).

## TDD order

1. **Failing handler test** — one vertical slice (e.g. `GET /apples` → empty list) with a mock repo in `internal/<resource>/`.
2. **Router integration test** — `Register(<resource>.PathPrefix, …)` in `internal/<resource>/router_test.go`.
3. **Entity + validation tests** — `internal/<resource>/<resource>_test.go`.
4. **Handler** — minimum code to pass step 1; expand tests per method.
5. **DynamoDB tests** → table-driven repository tests in `internal/<resource>/dynamodb_test.go` → `dynamodb.go` implementation.
6. **Compose** — `internal/app/app.go` (shared client + `Register`), `<resource>_stub_test.go`, smoke via `testGateway` / `assertWiringSmokeGET`.
7. **Infrastructure** — `template.yml`.
8. **API docs** — `README.md` contract for the new resource.
9. **`make test`** — must pass before PR.

## Files to create (vertical slice)

Copy `internal/banana/` → `internal/<resource>/` and rename. One package per resource:

| File | Reference (banana) |
| ---- | ---------------- |
| `internal/<resource>/<resource>.go` | `banana.go` — `PathPrefix`, `TableName`, entity, validation (default string bounds from `domain`) |
| `internal/<resource>/repository.go` | `repository.go` — `Repository` interface |
| `internal/<resource>/handler.go` | `handler.go` — HTTP handler; `NewHandler(repo, logger)` |
| `internal/<resource>/dynamodb.go` | `dynamodb.go` — `NewRepository(client)` DynamoDB impl |
| `internal/<resource>/<resource>_test.go` | `banana_test.go` — validation tests |
| `internal/<resource>/handler_test.go` | `handler_test.go` — HTTP tests (`package <resource>_test`) |
| `internal/<resource>/dynamodb_test.go` | `dynamodb_test.go` — repository tests |
| `internal/<resource>/assert_test.go` | `assert_test.go` — wire decode + repo result/put asserts |
| `internal/<resource>/fixtures_test.go` | `fixtures_test.go` — e.g. `existingAppleFixture()` |
| `internal/<resource>/dynamodb_fixtures_test.go` | `dynamodb_fixtures_test.go` — e.g. `storedAppleFixture(t)` |
| `internal/<resource>/mocks_test.go` | `mocks_test.go` — mock repo helpers |
| `internal/<resource>/router_test.go` | `router_test.go` — router + resource integration |
| `internal/testutil/<resource>_fixtures.go` | `banana_fixtures.go` — optional shared fixtures if needed cross-package |

## Shared packages (reuse; do not duplicate per resource)

| Package / file | Purpose |
| ---- | ------- |
| `internal/domain/` | Cross-cutting only: `errors.go`, `id.go`, `validation.go` |
| `internal/gateway/gateway.go` | Auth gate + path routing; `Register(prefix, ResourceHandler)` |
| `internal/platform/` | Response envelope, error mapping, logging, auth (`CFTTokenHeader`, `CFTTokenEnvVar`, `ExpectedCFTToken`) |
| `internal/testutil/consts.go` | `TestCFTToken`, `CFTokenHeaders` (pair token with `platform.CFTTokenEnvVar` in `t.Setenv`) |
| `internal/app/banana_stub_test.go` | Pattern for composition no-op repos (stay in `app`, not the resource package — see below) |
| `internal/testutil/handler_assert.go` | `RequireHandle`, `RequireStatusAndEnvelope`, `AssertAPIError` |
| `internal/testutil/error_assert.go` | `AssertWantErr` for table-driven validation tests |
| `internal/testutil/dynamodb_assert.go` | `AssertUpdateSets` for update success mocks |

## Files to edit

- [ ] `internal/app/app.go` — reuse shared `client := dynamodb.NewFromConfig(cfg)`; `<resource>.NewRepository(client)`; `g.Register(<resource>.PathPrefix, <resource>.NewHandler(...))`
- [ ] `internal/app/<resource>_stub_test.go` — no-op `Repository` for composition smoke tests (mirror `banana_stub_test.go`; keep under `app`, not `internal/<resource>/`)
- [ ] `internal/app/app_test.go` — extend `testGateway` with the new stub; add `assertWiringSmokeGET(t, testGateway(t), "/"+<resource>.PathPrefix)`
- [ ] `internal/gateway/gateway_test.go` — generic routing/auth only; resource integration lives in `internal/<resource>/router_test.go`
- [ ] `template.yml` — table, **one `DynamoDBCrudPolicy` per table**, API events
- [ ] `README.md` — API contract: endpoints, item shape, create/update bodies, validation

## Table naming (must match)

| | Value |
|---|--------|
| SAM logical ID | `Appname<Resources>Table` |
| Physical `TableName` | `Appname<Resources>` |
| Go constant | `<resource>.TableName` in `<resource>.go` (used by `dynamodb.go`) |
| Path prefix | `<resource>.PathPrefix` (used by `app.Build` / `Register`) |

## SAM API event names

Match the logical ID to the HTTP method (see `template.yml` bananas): `PostBanana` + `Method: POST`, `UpdateBanana` + `Method: PUT`, `GetBanana` + `GET`, etc. Avoid names like `PutBanana` for a POST route.

## Second table in the same project

1. Copy `internal/banana/` → `internal/<resource>/` and rename symbols.
2. In `internal/app/app.go` — `<resource>.NewRepository(client)` on the existing shared client; `g.Register(<resource>.PathPrefix, <resource>.NewHandler(...))`.
3. In `template.yml` — add table, append `DynamoDBCrudPolicy`, add API events.
4. Add `internal/app/<resource>_stub_test.go`; extend `testGateway` and add a smoke path in `app_test.go`.
5. Add `internal/testutil/<resource>_fixtures.go` if handler and DynamoDB tests share fixtures.

Composition stubs stay in `internal/app/` (`*_stub_test.go`). They cannot live in the resource’s `_test.go` files and still be used by `app` tests (Go does not allow importing another package’s tests).

Shared `domain/` and `platform/` stay resource-neutral.

## Test patterns (copy from banana)

- Package: production code in `package <resource>`; tests in `package <resource>_test`.
- Handler tests: prefer `testutil.RequireHandle` (err + status + envelope); `testutil.AssertAPIError` for client errors; mock repo in `mocks_test.go`. Shared request payloads: `testutil.<Resource>Body` / `Valid<Resource>Body()` / `.JSON(t)` (independent of the entity so tag regressions fail). Reuse package-local `new<Resource>ValidationBodies(t)` for generic invalid shapes (`…EmptyValue`, `…Whitespace`, `…ValueTooLong`, `…ValueBelowMin`, `…ValueAboveMax`) across POST and PUT — names describe the invalidation, not which field was mutated.
- Entity fixtures: `testutil.<Resource>WithID(Valid<Resource>Body(), createdOn)` — client fields via named body struct, not positional args.
- DynamoDB tests: `setupMock func(t *testing.T) *mockDynamoClient`; `storedBananaFixture(t)` for Get/Delete; `assertBananaRepoResult`, `assertBananaPutItem` in `assert_test.go`; `testutil.AssertUpdateSets` on update success.
- Composition smoke: `testGateway` + `assertWiringSmokeGET` in `app_test.go`; one `*_stub_test.go` per resource under `internal/app/`.
- Gateway integration: `router_test.go` in the resource package registers with `gateway.NewGatewayWithCFTToken`; use `testutil.CFTokenHeaders` and `platform.SAMLocalEnvVar` when needed.
- Validation tests: define `validCreateInput` / `validUpdateInput` as local funcs inside each test; clone and tweak one field per case. Prefer `testutil.AssertWantErr` and `testutil` canonical values over package-local literals.
- Validation bounds: use `domain.DefaultMinStringLength` / `DefaultMaxStringLength` unless the field opts out.
- In production package `banana`, prefer the parameter/local name `banana` so find-replace of the template resource name rewrites it. In `package banana_test`, use a short local (e.g. `b`) so you do not shadow the imported `banana` package.

## Before PR

- [ ] `make test`
- [ ] `make build` (especially after `template.yml` changes)
