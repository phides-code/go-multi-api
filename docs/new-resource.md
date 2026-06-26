# New resource checklist

Replace `<resource>` / `<Resource>` / `<resources>` (e.g. `apple`, `Apple`, `apples`).

Full walkthrough: [Adding a new table](../README.md#adding-a-new-table).

## TDD order

1. **Failing handler test** — one vertical slice (e.g. `GET /apples` → empty page) with a mock repo.
2. **Router dispatch test** — `Register("<resources>", …)` in `router_test.go`.
3. **Domain tests** → entity + validation + repository interface.
4. **Handler** — minimum code to pass step 1; expand tests per method.
5. **DynamoDB tests** → repository implementation.
6. **Wire** — `internal/app/wire.go` + wiring smoke test.
7. **Infrastructure** — `template.yml`.
8. **`make test`** — must pass before PR.

## Files to create

| File | Reference |
| ---- | --------- |
| `internal/handler/<resource>_handler_test.go` | `internal/handler/banana_handler_test.go` — **start here** |
| `internal/domain/<resource>_test.go` | `internal/domain/banana_test.go` |
| `internal/domain/<resource>.go` | `internal/domain/banana.go` |
| `internal/domain/<resource>_repository.go` | `internal/domain/banana_repository.go` |
| `internal/handler/<resource>_handler.go` | `internal/handler/banana_handler.go` |
| `internal/dynamodb/<resource>_repository_test.go` | `internal/dynamodb/banana_repository_test.go` |
| `internal/dynamodb/<resource>_repository.go` | `internal/dynamodb/banana_repository.go` |

## Files to edit

- [ ] `internal/handler/router_test.go` — dispatch test
- [ ] `internal/app/wire.go` — construct repo, `router.Register("<resources>", ...)`
- [ ] `template.yml` — table, **one `DynamoDBCrudPolicy` per table**, API events

## Table naming (must match)

| | Value |
|---|--------|
| SAM logical ID | `Appname<Resources>Table` |
| Physical `TableName` | `Appname<Resources>` |
| Go constant | `"Appname<Resources>"` |

## Before PR

- [ ] `make test`
- [ ] `make build`
