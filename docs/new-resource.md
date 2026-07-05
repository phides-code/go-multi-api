# New resource checklist

Replace `<resource>` / `<Resource>` / `<resources>` (e.g. `apple`, `Apple`, `apples`).

Full walkthrough: [Adding a new table](../README.md#adding-a-new-table).

## TDD order

1. **Failing handler test** — one vertical slice (e.g. `GET /apples` → empty page) with a mock repo.
2. **Router dispatch test** — `Register("<resources>", …)` in `router_test.go`.
3. **Domain tests** → entity + validation + repository interface.
4. **Handler** — minimum code to pass step 1; expand tests per method.
5. **DynamoDB tests** → table-driven repository tests (`setupMock(t)`, `assertBananaRepoResult`, `assertBananaPutItem` on create success, `assertUpdateSets` on update success) → implementation.
6. **Wire** — `internal/app/wire.go` + wiring smoke test.
7. **Infrastructure** — `template.yml`.
8. **API docs** — `README.md` contract for the new resource.
9. **`make test`** — must pass before PR.

## Files to create

| File | Reference |
| ---- | --------- |
| `internal/handler/<resource>_handler_test.go` | `internal/handler/banana_handler_test.go` — **start here** |
| `internal/handler/<resource>_assert_test.go` | `internal/handler/banana_assert_test.go` — decode/assert wire shape |
| `internal/handler/<resource>_mocks_test.go` | `internal/handler/banana_mocks_test.go` — mock repo helpers |
| `internal/handler/<resource>_fixtures_test.go` | `internal/handler/banana_fixtures_test.go` — package-local fixture helpers (e.g. `existingBananaFixture`) |
| `internal/domain/<resource>_test.go` | `internal/domain/banana_test.go` |
| `internal/domain/<resource>.go` | `internal/domain/banana.go` |
| `internal/domain/<resource>_repository.go` | `internal/domain/banana_repository.go` |
| `internal/handler/<resource>_handler.go` | `internal/handler/banana_handler.go` |
| `internal/dynamodb/<resource>_repository_test.go` | `internal/dynamodb/banana_repository_test.go` |
| `internal/dynamodb/<resource>_fixtures_test.go` | `internal/dynamodb/banana_fixtures_test.go` — e.g. `storedBananaFixture` for Get/Delete mocks |
| `internal/dynamodb/<resource>_assert_test.go` | `internal/dynamodb/banana_assert_test.go` — `assert<Resource>RepoResult`, `assert<Resource>PutItem` |
| `internal/dynamodb/<resource>_repository.go` | `internal/dynamodb/banana_repository.go` |

## Shared helpers (reuse; do not duplicate per resource)

| File | Purpose |
| ---- | ------- |
| `internal/handler/assert_test.go` | Shared envelope helpers (`requireStatusAndEnvelope`, `assertAPIError`) — any resource |
| `internal/handler/banana_assert_test.go` | Banana wire helpers — **copy** to `<resource>_assert_test.go`: `decode<Resource>Data`, `decode<Resource>PageData`, `assert<Resource>DataKeys` |
| `internal/handler/banana_mocks_test.go` | Mock repo pattern — **copy** to `<resource>_mocks_test.go`: `mock<Resource>Repository`, `empty<Resource>Repo`, `list<Resource>Repo`, `update<Resource>Repo`, `dispatch<Resource>Repo` (router dispatch), `panic<Resource>Repo` (validation must not call repo) |
| `internal/dynamodb/assert_test.go` | Shared: `assertUpdateSets` — any updatable resource |
| `internal/dynamodb/banana_assert_test.go` | Banana reference — **copy** to `<resource>_assert_test.go`: `assert<Resource>RepoResult`, `assert<Resource>PutItem` |
| `internal/domain/validation.go` | `ValidateRequiredString` |
| `internal/domain/id.go` | `ValidateID`, `NewID` (UUID keys) |
| `internal/domain/pagination.go` | `ListOptions`, `DefaultListLimit` (cursor-based list) |
| `internal/testutil/consts.go` | Cross-package test constants (e.g. `TestCFTToken` for `router_test.go` and `wire_test.go`) |
| `internal/testutil/banana_fixtures.go` | Banana fixtures shared by handler and DynamoDB tests — **copy pattern** for new resources: `Test<Resource>Content`, `BananaWithID` → `<Resource>WithID`, `BananaCreateBody`, `ListBananaPage` → `List<Resource>Page` |

## Files to edit

- [ ] `internal/handler/router_test.go` — dispatch test (`TestRouterDispatchesRegisteredPrefix` or `TestRouterDispatches<Resources>`); use `dispatch<Resource>Repo()` for permissive mocks (see `dispatchBananaRepo`)
- [ ] `internal/app/wire.go` — construct repo, `router.Register("<resources>", …)`
- [ ] `internal/app/wire_test.go` — wiring smoke test (mirror `TestWiringSmokeGETBananas`; use `testutil.TestCFTToken`)
- [ ] `template.yml` — table, **one `DynamoDBCrudPolicy` per table**, API events
- [ ] `README.md` — API contract: endpoints, item shape, create/update bodies, validation

## Table naming (must match)

| | Value |
|---|--------|
| SAM logical ID | `Appname<Resources>Table` |
| Physical `TableName` | `Appname<Resources>` |
| Go constant | `"Appname<Resources>"` |

## SAM API event names

Match the logical ID to the HTTP method (see `template.yml` bananas): `PostBanana` + `Method: POST`, `UpdateBanana` + `Method: PUT`, `GetBanana` + `GET`, etc. Avoid names like `PutBanana` for a POST route.

## Second table in a copied project

This template ships one resource (bananas). When a **copied project** adds another table:

1. Create the new resource files (table above).
2. In `internal/app/wire.go` — construct the new repo and add `router.Register("<resources>", handler.New<Resource>Handler(...))` beside the existing registration.
3. In `template.yml` — add a table resource, append a `DynamoDBCrudPolicy` (do not replace existing policies), add API events for the methods you expose.
4. Add a wiring smoke test row in `wire_test.go` for the new path.

No shared-type refactor required — shared domain/platform code stays resource-neutral.

## DynamoDB test patterns (copy from banana)

- Table field: `setupMock func(t *testing.T) *mockDynamoClient` — pass subtest `t` at `tt.setupMock(t)`; use named `t` in mocks that call `t.Error` / assert helpers.
- Persisted row fixture: `storedBananaFixture(t)` in `<resource>_fixtures_test.go` — ID-linked entity + marshaled DynamoDB item for Get/Delete mocks.
- Single-entity methods (Get/Create/Update/Delete): `assertBananaRepoResult(t, "<Op>", got, err, want, wantErr)`.
- Create success mock: `assertBananaPutItem(t, params, want)`.
- Update success mock: `assertUpdateSets(t, params, map[string]string{…})`.
- List: assert returned items/cursors in the subtest loop (no `assertBananaRepoResult` — returns a page).
- Validation bounds: resource-scoped constants in `<resource>.go` (see `BananaMinContentLength` / `BananaMaxContentLength`).

## Handler test fixtures (copy from banana)

- Canonical valid content: `testutil.TestBananaContent` — use for create/get/update/delete success paths and matching JSON via `testutil.BananaCreateBody`.
- ID-linked entity: `existingBananaFixture()` in `<resource>_fixtures_test.go` for get/update/delete table setup.
- List pagination labels: `testutil.ListBananaPage(withTimestamps)` — handler tests use `false`; DynamoDB list tests use `true` when asserting `createdOn`.
- Domain validation tests may use any non-empty string (e.g. `"hello"`) — they test rules, not the HTTP narrative.

## Before PR

- [ ] `make test`
- [ ] `make build` (especially after `template.yml` changes)
