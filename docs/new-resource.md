# New resource checklist

Add a second (or third) table and URL prefix to this Lambda by copying the banana vertical slice.

Replace placeholders consistently:

| Placeholder | Example |
| --- | --- |
| `<resource>` | `apple` (package / directory) |
| `<Resource>` | `Apple` (exported types) |
| `<resources>` | `apples` (`PathPrefix`, plural path) |

Summary table: [README — Adding a new resource](../README.md#adding-a-new-resource).

## Steps (file by file)

Work one HTTP method at a time when possible (e.g. `GET /apples` → empty list), then expand.

### 1. Copy and rename

| File(s) | Do this |
| --- | --- |
| `internal/banana/` → `internal/<resource>/` | `cp -R internal/banana internal/<resource>` |
| Every file under `internal/<resource>/` | Rename files and symbols: `banana`→`<resource>`, `Banana`→`<Resource>`, `bananas`→`<resources>` |
| `internal/<resource>/<resource>.go` | Set `PathPrefix` (plural, no slash) and `TableName` (must match SAM physical name) |

In production code, prefer the local/parameter name `<resource>` so find-replace stays mechanical. In `package <resource>_test`, use a short local (e.g. `a`) so you do not shadow the imported package.

### 2. First failing tests

| File(s) | Do this |
| --- | --- |
| `internal/<resource>/handler_test.go` | One failing test for the first method (mock repo) |
| `internal/<resource>/mocks_test.go` | Mock `Repository` helpers for that method |
| `internal/<resource>/router_test.go` | `Register(<resource>.PathPrefix, …)` and assert gateway dispatch |

### 3. Entity and validation

| File(s) | Do this |
| --- | --- |
| `internal/<resource>/<resource>_test.go` | Validation cases (`validCreateInput` / `validUpdateInput` locals) |
| `internal/<resource>/<resource>.go` | Entity, `CreateInput` / `UpdateInput`, validation funcs |
| `internal/<resource>/repository.go` | Keep / trim `Repository` methods to what you implement |

### 4. Handler

| File(s) | Do this |
| --- | --- |
| `internal/<resource>/handler.go` | Implement only the HTTP methods you need |
| `internal/<resource>/handler_test.go` | Success, client errors, one 500 per op as you add methods |
| `internal/<resource>/fixtures_test.go` | `existing<Resource>Fixture`, `new<Resource>ValidationBodies` |
| `internal/<resource>/assert_test.go` | Wire decode helpers + `assert<Resource>DataKeys` |

### 5. DynamoDB

| File(s) | Do this |
| --- | --- |
| `internal/<resource>/dynamodb_test.go` | Table-driven repo tests with mocked client |
| `internal/<resource>/dynamodb_fixtures_test.go` | `stored<Resource>Fixture(t)` (entity + marshaled item) |
| `internal/<resource>/assert_test.go` | `assert<Resource>PutItem`, `assert<Resource>RepoResult` |
| `internal/<resource>/dynamodb.go` | `Attr*`, conditions, CRUD impl (`NewRepository`) |
| `internal/testutil/<resource>_fixtures.go` | Shared body/entity fixtures (mirror `banana_fixtures.go`) when handler and DynamoDB tests both need them |

### 6. Compose and deploy config

| File(s) | Do this |
| --- | --- |
| `internal/app/app.go` | Reuse shared `dynamodb.NewFromConfig(cfg)`; `<resource>.NewRepository(client)`; `g.Register(PathPrefix, NewHandler(...))` |
| `internal/app/<resource>_stub_test.go` | No-op `Repository` (mirror `banana_stub_test.go`; must live under `app`, not the resource package) |
| `internal/app/app_test.go` | Extend `testGateway` with the stub; add `assertWiringSmokeGET(..., "/"+PathPrefix)` |
| `template.yml` | Table, **one `DynamoDBCrudPolicy` per table**, API events only for methods you implemented |
| `README.md` | Endpoints, item shape, create/update bodies, validation |
| `Makefile` | Optional per-package coverage gate (same pattern as banana) |

Do **not** add resource-specific cases to `internal/gateway/gateway_test.go`. Gateway tests stay generic; resource routing belongs in `router_test.go`.

### 7. Before PR

- [ ] `make test`
- [ ] `make build` (required after `template.yml` changes)
- [ ] README documents the new resource
- [ ] Only the HTTP methods you implemented have matching SAM events

## Naming (must match across Go and SAM)

| Piece | Convention | Example |
| --- | --- | --- |
| SAM logical ID | `Appname<Resources>Table` | `AppnameApplesTable` |
| Physical table name | `Appname<Resources>` | `AppnameApples` |
| Go `TableName` in `<resource>.go` | same physical name | `const TableName = "AppnameApples"` |
| Go `PathPrefix` in `<resource>.go` | plural, no leading slash | `const PathPrefix = "apples"` |

SAM API event **logical IDs** should match the HTTP verb (see banana events in `template.yml`): e.g. `PostApple` + `POST`, `UpdateApple` + `PUT`. Avoid `PutApple` for a POST route.

## Vertical slice file map

| File | Role |
| --- | --- |
| `<resource>.go` | `PathPrefix`, `TableName`, entity, inputs, validation |
| `repository.go` | `Repository` interface |
| `handler.go` | HTTP → repo; `NewHandler(repo, logger)` |
| `dynamodb.go` | DynamoDB `Repository` |
| `<resource>_test.go` | Validation unit tests |
| `handler_test.go` | HTTP tests (`package <resource>_test`) |
| `dynamodb_test.go` | Repository tests |
| `assert_test.go` | Wire decode + put/repo asserts |
| `fixtures_test.go` | Handler fixtures |
| `dynamodb_fixtures_test.go` | Stored item + marshaled map |
| `mocks_test.go` | Mock `Repository` helpers |
| `router_test.go` | Gateway + resource integration |
| `internal/testutil/<resource>_fixtures.go` | Cross-package fixtures |

## Reuse (do not copy per resource)

| Package | Use for |
| --- | --- |
| `internal/domain` | Sentinels, `ValidateID`, `ValidateRequiredString` / `ValidateRequiredInt` |
| `internal/gateway` | `Register(prefix, handler)`, auth gate |
| `internal/platform` | Envelope, `ClientErrorResponse`, logger, CF token helpers |
| `internal/testutil` | `RequireHandle`, `AssertAPIError`, `AssertWantErr`, `AssertUpdateSets`, `CFTokenHeaders` |

## Test patterns (copy from banana)

**Packages:** production code in `package <resource>`; tests in `package <resource>_test`.

**`handler_test.go`**

- Prefer `testutil.RequireHandle` and `testutil.AssertAPIError`.
- Request JSON from `testutil.<Resource>Body` / `Valid<Resource>Body().JSON(t)` (separate from the entity so tag drift fails).
- Invalid bodies from `fixtures_test.go` (`new<Resource>ValidationBodies`) with shape names (`EmptyValue`, `Whitespace`, …), not field names — reuse for POST and PUT.
- Entities via `testutil.<Resource>WithID(...)`.

**`<resource>_test.go`**

- Local `validCreateInput` / `validUpdateInput`; clone and break one field per case.
- Prefer `testutil.AssertWantErr` and shared canonical values.
- Default bounds from `domain` unless the field opts out.

**`dynamodb_test.go`**

- `setupMock func(t *testing.T) *mockDynamoClient`
- `stored<Resource>Fixture` from `dynamodb_fixtures_test.go` for Get/Delete/Update success
- Create: shared fixture + `assert<Resource>PutItem` in `assert_test.go`
- Update success: `testutil.AssertUpdateSets` (expected `SET` attrs sorted alphabetically)

**`app_test.go` / `router_test.go`**

- Smoke: `testGateway` + `assertWiringSmokeGET`
- Integration: `gateway.NewGatewayWithCFTToken` + `testutil.CFTokenHeaders`
