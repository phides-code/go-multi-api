# New resource checklist

Replace `<resource>` / `<Resource>` / `<resources>` (e.g. `apple`, `Apple`, `apples`).

Full walkthrough: [Adding a new table](../README.md#adding-a-new-table).

## TDD order

1. **Failing handler test** — one vertical slice (e.g. `GET /apples` → empty list) with a mock repo in `internal/<resource>/`.
2. **Router integration test** — `Register("<resources>", …)` in `internal/<resource>/router_test.go`.
3. **Entity + validation tests** — `internal/<resource>/<resource>_test.go`.
4. **Handler** — minimum code to pass step 1; expand tests per method.
5. **DynamoDB tests** → table-driven repository tests in `internal/<resource>/dynamodb_test.go` → `dynamodb.go` implementation.
6. **Compose** — `internal/app/app.go` + composition smoke test.
7. **Infrastructure** — `template.yml`.
8. **API docs** — `README.md` contract for the new resource.
9. **`make test`** — must pass before PR.

## Files to create (vertical slice)

Copy `internal/banana/` → `internal/<resource>/` and rename. One package per resource:

| File | Reference (banana) |
| ---- | ---------------- |
| `internal/<resource>/<resource>.go` | `banana.go` — entity, validation (default string bounds from `domain`) |
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
| `internal/platform/` | Response envelope, error mapping, logging, auth header |
| `internal/testutil/consts.go` | `TestCFTToken` for gateway and composition tests |
| `internal/testutil/handler_assert.go` | `RequireStatusAndEnvelope`, `AssertAPIError` |
| `internal/testutil/dynamodb_assert.go` | `AssertUpdateSets` for update success mocks |

## Files to edit

- [ ] `internal/app/app.go` — `<resource>.NewRepository(...)`, `d.Register("<resources>", <resource>.NewHandler(...))`
- [ ] `internal/app/app_test.go` — composition smoke test (mirror `TestWiringSmokeGETBananas`)
- [ ] `internal/gateway/gateway_test.go` — generic routing/auth only; resource integration lives in `internal/<resource>/router_test.go`
- [ ] `template.yml` — table, **one `DynamoDBCrudPolicy` per table**, API events
- [ ] `README.md` — API contract: endpoints, item shape, create/update bodies, validation

## Table naming (must match)

| | Value |
|---|--------|
| SAM logical ID | `Appname<Resources>Table` |
| Physical `TableName` | `Appname<Resources>` |
| Go constant | `"Appname<Resources>"` in `<resource>/dynamodb.go` |

## SAM API event names

Match the logical ID to the HTTP method (see `template.yml` bananas): `PostBanana` + `Method: POST`, `UpdateBanana` + `Method: PUT`, `GetBanana` + `GET`, etc. Avoid names like `PutBanana` for a POST route.

## Second table in the same project

1. Copy `internal/banana/` → `internal/<resource>/` and rename symbols.
2. In `internal/app/app.go` — construct the new repo and `d.Register("<resources>", <resource>.NewHandler(...))`.
3. In `template.yml` — add table, append `DynamoDBCrudPolicy`, add API events.
4. Add a composition smoke test in `app_test.go`.
5. Add `internal/testutil/<resource>_fixtures.go` if handler and DynamoDB tests share fixtures.

Shared `domain/` and `platform/` stay resource-neutral.

## Test patterns (copy from banana)

- Package: production code in `package <resource>`; tests in `package <resource>_test`.
- Handler tests: `testutil.RequireStatusAndEnvelope`, `testutil.AssertAPIError`; mock repo in `mocks_test.go`.
- DynamoDB tests: `setupMock func(t *testing.T) *mockDynamoClient`; `storedBananaFixture(t)` for Get/Delete; `assertBananaRepoResult`, `assertBananaPutItem` in `assert_test.go`; `testutil.AssertUpdateSets` on update success.
- Gateway integration: `router_test.go` in the resource package registers with `gateway.NewGatewayWithCFTToken`.
- Validation bounds: use `domain.DefaultMinStringLength` / `DefaultMaxStringLength` unless the field opts out.
- Avoid naming a function parameter `banana` when the package is `banana` — use `b` instead (shadowing breaks `banana.Banana{}` zero values).

## Before PR

- [ ] `make test`
- [ ] `make build` (especially after `template.yml` changes)
