# New resource checklist

Replace `<resource>` / `<Resource>` / `<resources>` (e.g. `apple`, `Apple`, `apples`).

Full walkthrough: [Adding a new table](../README.md#adding-a-new-table).

## Files to create

| File                                              | Reference                                     |
| ------------------------------------------------- | --------------------------------------------- |
| `internal/domain/<resource>.go`                   | `internal/domain/banana.go`                   |
| `internal/domain/<resource>_test.go`              | `internal/domain/banana_test.go`              |
| `internal/domain/<resource>_repository.go`        | `internal/domain/banana_repository.go`        |
| `internal/dynamodb/<resource>_repository.go`      | `internal/dynamodb/banana_repository.go`      |
| `internal/dynamodb/<resource>_repository_test.go` | `internal/dynamodb/banana_repository_test.go` |
| `internal/handler/<resource>_handler.go`          | `internal/handler/banana_handler.go`          |
| `internal/handler/<resource>_handler_test.go`     | `internal/handler/banana_handler_test.go`     |

## Files to edit

- [ ] `internal/app/wire.go` — construct repo, `router.Register("<resources>", ...)`
- [ ] `internal/handler/router_test.go` — dispatch test (or extend `TestRouterDispatchesRegisteredPrefix` pattern)
- [ ] `template.yml` — table, **one `DynamoDBCrudPolicy` per table**, API events

## Table naming (must match)

|                      | Value                     |
| -------------------- | ------------------------- |
| SAM logical ID       | `Appname<Resources>Table` |
| Physical `TableName` | `Appname<Resources>`      |
| Go constant          | `"Appname<Resources>"`    |

## Tests before PR

- [ ] `go test ./internal/domain/ ./internal/dynamodb/ ./internal/handler/ ./internal/app/`
- [ ] `make build`
