# New resource checklist

Add a second (or third) table and URL prefix to this Lambda by copying the banana vertical slice.

Replace placeholders consistently:

| Placeholder | Example |
| --- | --- |
| `<resource>` | `apple` (package / directory) |
| `<Resource>` | `Apple` (exported types) |
| `<resources>` | `apples` (`PathPrefix`, plural path) |

Walkthrough summary: [README — Adding a new resource](../README.md#adding-a-new-resource).

## 1. Copy the slice

```bash
cp -R internal/banana internal/<resource>
```

Rename files and symbols (`banana` → `<resource>`, `Banana` → `<Resource>`, `bananas` → `<resources>`). Prefer the local/parameter name `<resource>` in production code so find-replace stays mechanical. In `package <resource>_test`, use a short local (e.g. `a`) so you do not shadow the imported package.

## 2. TDD order

Work one vertical slice at a time (e.g. `GET /apples` → empty list), then expand method by method.

1. Failing handler test with a mock repo
2. Router integration test (`router_test.go` registers `PathPrefix`)
3. Entity + validation tests
4. Handler implementation for that method
5. DynamoDB tests, then `dynamodb.go`
6. Composition + SAM + README
7. `make test` (and `make build` after template changes)

## 3. Files in the vertical slice

| File | Role (see banana equivalent) |
| --- | --- |
| `<resource>.go` | `PathPrefix`, `TableName`, entity, create/update inputs, validation |
| `repository.go` | `Repository` interface |
| `handler.go` | HTTP → repo; `NewHandler(repo, logger)` |
| `dynamodb.go` | DynamoDB `Repository` (`Attr*`, conditions, CRUD) |
| `<resource>_test.go` | Validation unit tests |
| `handler_test.go` | HTTP tests (`package <resource>_test`) |
| `dynamodb_test.go` | Repository tests with mocked DynamoDB client |
| `assert_test.go` | Wire decode, put/update/repo result asserts |
| `fixtures_test.go` | Handler fixtures (existing item, validation bodies) |
| `dynamodb_fixtures_test.go` | Stored item + marshaled DynamoDB map |
| `mocks_test.go` | Mock `Repository` helpers |
| `router_test.go` | Gateway + resource integration |
| `internal/testutil/<resource>_fixtures.go` | Shared fixtures if handler and DynamoDB tests both need them |

## 4. Wire into the app

- [ ] `internal/app/app.go` — reuse shared `dynamodb.NewFromConfig(cfg)`; `<resource>.NewRepository(client)`; `g.Register(<resource>.PathPrefix, <resource>.NewHandler(...))`
- [ ] `internal/app/<resource>_stub_test.go` — no-op `Repository` (mirror `banana_stub_test.go`; must live under `app`, not the resource package)
- [ ] `internal/app/app_test.go` — extend `testGateway` with the stub; add `assertWiringSmokeGET(..., "/"+<resource>.PathPrefix)`
- [ ] `template.yml` — table, **one `DynamoDBCrudPolicy` per table**, API events for each method you support
- [ ] `README.md` — endpoints, item shape, create/update bodies, validation
- [ ] `Makefile` — optional per-package coverage gate if you want the same bar as banana

Do **not** add resource-specific cases to `gateway_test.go`. Gateway tests stay generic; resource routing belongs in `router_test.go`.

## 5. Naming (must match across Go and SAM)

| Piece | Convention | Example |
| --- | --- | --- |
| SAM logical ID | `Appname<Resources>Table` | `AppnameApplesTable` |
| Physical table name | `Appname<Resources>` | `AppnameApples` |
| Go `TableName` | same physical name | `const TableName = "AppnameApples"` |
| Go `PathPrefix` | plural, no leading slash | `const PathPrefix = "apples"` |

SAM API event **logical IDs** should match the HTTP verb (see banana events in `template.yml`): e.g. `PostApple` + `POST`, `UpdateApple` + `PUT`. Avoid `PutApple` for a POST route.

## 6. Reuse these packages

Do not copy domain/platform/gateway per resource.

| Package | Use for |
| --- | --- |
| `internal/domain` | Sentinels, `ValidateID`, `ValidateRequiredString` / `ValidateRequiredInt` |
| `internal/gateway` | `Register(prefix, handler)`, auth gate |
| `internal/platform` | Envelope, `ClientErrorResponse`, logger, CF token helpers |
| `internal/testutil` | `RequireHandle`, `AssertAPIError`, `AssertWantErr`, `AssertUpdateSets`, `CFTokenHeaders` |

## 7. Test patterns (copy from banana)

**Packages:** production code in `package <resource>`; tests in `package <resource>_test`.

**Handler tests**

- Prefer `testutil.RequireHandle` (err + status + envelope) and `testutil.AssertAPIError` for client errors.
- Request JSON: `testutil.<Resource>Body` / `Valid<Resource>Body().JSON(t)` — declared separately from the entity so tag drift fails tests.
- Invalid bodies: package-local `new<Resource>ValidationBodies(t)` with shape names (`EmptyValue`, `Whitespace`, `ValueTooLong`, `ValueBelowMin`, `ValueAboveMax`), not field names. Reuse the same fixtures for POST and PUT.
- Entity for get/update/delete: `testutil.<Resource>WithID(Valid<Resource>Body(), createdOn)`.

**Validation tests**

- Local `validCreateInput` / `validUpdateInput` helpers; clone and break one field per case.
- Prefer `testutil.AssertWantErr` and shared canonical values over one-off literals.
- Default string/int bounds come from `domain` unless the field opts out.

**DynamoDB tests**

- `setupMock func(t *testing.T) *mockDynamoClient`
- `stored<Resource>Fixture(t)` for Get/Delete/Update success (entity + marshaled item)
- Create: `BananaWithID`-style fixture + `assert<Resource>PutItem`
- Update success: `testutil.AssertUpdateSets` (keys are sorted alphabetically in the expected `SET`)

**Composition / gateway**

- Smoke: `testGateway` + `assertWiringSmokeGET` in `app_test.go`
- Integration: `router_test.go` uses `gateway.NewGatewayWithCFTToken` and `testutil.CFTokenHeaders`

## 8. Before PR

- [ ] `make test`
- [ ] `make build` (required after `template.yml` changes)
- [ ] README documents the new resource
- [ ] Only the HTTP methods you implemented have matching SAM events
