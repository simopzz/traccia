# CLAUDE.md

Keep your replies extremely concise and focus on conveying the key information. No unnecessary fluff, no long code snippets.

Whenever working with any third-party library or something similar, you MUST look up the official documentation to ensure that you're working with up-to-date information.
Use the DocsExplorer subagent for efficient documentation lookup.

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Traccia

Traccia is a trip planning web app built with Go. Users create trips and populate them with a timeline of events (activities, food, lodging, transit). The project follows a layered architecture with dependency inversion.

## Commands

```bash
make dev              # Hot reload via air
make build            # Build binary to bin/app
make test             # go test -v -race ./...
make lint             # golangci-lint run
make generate         # sqlc generate && templ generate (run after changing .sql queries or .templ files)
make docker-up        # Start PostgreSQL container
make docker-down      # Stop PostgreSQL container
make migrate-up       # Apply database migrations
make migrate-down     # Rollback database migrations
```

Run a single test: `go test -v -race -run TestName ./internal/service/...`

## Tech Stack

- **Go 1.25** with chi router, templ templates, HTMX 2.0 for frontend interactivity
- **PostgreSQL 16** via Docker, pgx/v5 driver, sqlc for type-safe query generation
- **golang-migrate** for schema migrations
- **golangci-lint** for linting (generated `*_templ.go` and `*.gen.go` files are excluded)

## Architecture

```
cmd/app/main.go          → Entry point, wires dependencies manually (no DI framework)
internal/domain/          → Entities (Trip, Event), repository interfaces (ports.go), domain errors
internal/service/         → Business logic, input validation, takes repository interfaces
internal/repository/      → Implements domain interfaces; sqlc-generated code + store adapters
internal/repository/sql/  → Raw SQL queries (sqlc source of truth)
internal/handler/         → HTTP handlers + .templ templates, chi routing (routes.go)
internal/infra/           → Config (env vars), database pool, HTTP server with graceful shutdown
migrations/               → PostgreSQL migration files (also used as sqlc schema source)
```

**Request flow**: HTTP → chi router (handler/routes.go) → Handler → Service → Repository → PostgreSQL

**Dependency direction**: handler → service → domain ← repository (domain has zero deps on other layers)

## Code Generation

Two tools generate code — always run `make generate` after changes:

1. **sqlc**: Reads schema from `migrations/`, queries from `internal/repository/sql/*.sql`. Generates `internal/repository/{models,db,*_sql}.go`. Query format: `-- name: FuncName :one|:many|:exec`
2. **templ**: Compiles `.templ` files into `*_templ.go` alongside them. Templates are type-safe Go components.

Never edit generated files (`*_templ.go`, `*_sql.go`, `models.go`, `db.go` in repository/).

## Key Patterns

- **Method override middleware**: HTML forms use `_method=PUT|DELETE` hidden field since browsers only support GET/POST
- **Updater pattern for updates**: Services pass `func(*Entity) *Entity` closures to repository Update methods for partial updates
- **Event positioning**: New events auto-increment position; `SuggestStartTime()` returns end time of last event in trip
- **Auth stubs**: `user_id` column exists, repository interfaces accept `userID *string`, `getUserID()` is a stub — prepared for future Supabase auth
- **goimports local prefix**: `github.com/simopzz/traccia` (configured in `.golangci.yml`)

## Go Conventions

**Imports**: standard library first, then third-party, then local (`github.com/simopzz/traccia`). Enforced by goimports.

**Naming**: `PascalCase` for exported types/methods, `camelCase` for local variables. Group related logic into descriptively-named files.

**Error handling**:
- Always wrap errors with context: `fmt.Errorf("fetching trip %s: %w", id, err)`
- Use `errors.Is()` / `errors.As()` for comparisons, never `==`
- Domain errors live in `internal/domain/errors.go`

**Testing**:
- Table-driven tests with explicit input/output expectations
- Use external test packages (`package foo_test`) for black-box testing
- Test all exported functions and error paths
- Run with `-race` flag (already in `make test`)

**Modern Go**:
- Use `any` instead of `interface{}`
- Pass `context.Context` as first parameter where applicable
- Use `log/slog` for structured logging

## Environment

Copy `.env.example` to `.env`. Required vars: `SERVER_ADDRESS`, `DATABASE_URL`, `ENVIRONMENT`.

Default DB connection: `postgres://traccia:traccia@localhost:5432/traccia?sslmode=disable`
