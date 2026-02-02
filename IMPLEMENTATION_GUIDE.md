# traccia - Implementation Guide

This guide documents the implementation of the traccia trip planning app.

## Project Structure

```
traccia/
├── cmd/app/main.go           # Application entry point
├── internal/
│   ├── domain/               # Business entities and interfaces
│   │   ├── models.go         # Trip, Event entities
│   │   ├── ports.go          # Repository interfaces
│   │   └── errors.go         # Domain errors
│   ├── service/              # Business logic
│   │   ├── trip.go
│   │   └── event.go
│   ├── repository/           # Data access layer (single package)
│   │   ├── sql/
│   │   │   ├── trips.sql        # Trip queries
│   │   │   └── events.sql       # Event queries
│   │   ├── models.go            # Generated: Trip, Event structs
│   │   ├── trips.sql.go         # Generated: trip query functions
│   │   ├── events.sql.go        # Generated: event query functions
│   │   ├── db.go                # Generated: DBTX interface
│   │   ├── trip_store.go        # TripStore implementation
│   │   └── event_store.go       # EventStore implementation
│   ├── handler/              # HTTP handlers + templates
│   │   ├── routes.go
│   │   ├── trip.go
│   │   ├── trip.templ
│   │   ├── event.go
│   │   ├── event.templ
│   │   ├── layout.templ
│   │   └── helpers.go
│   └── infra/                # Infrastructure
│       ├── config/
│       ├── database/
│       └── server/
├── migrations/               # Database migrations
├── static/                   # Static assets
├── docker-compose.yml        # Local PostgreSQL
├── sqlc.yaml                 # sqlc configuration
├── Makefile                  # Development commands
└── .env.example              # Environment template
```

## Getting Started

### Prerequisites

Install required tools:

```bash
# Router
go get github.com/go-chi/chi/v5

# Templates
go install github.com/a-h/templ/cmd/templ@latest

# Database
go get github.com/jackc/pgx/v5
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Migrations
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Hot reload (optional)
go install github.com/air-verse/air@latest

# Linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Setup

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Start PostgreSQL:
   ```bash
   make docker-up
   ```

3. Run migrations:
   ```bash
   make migrate-up
   ```

4. Generate code:
   ```bash
   make generate
   ```

5. Start the server:
   ```bash
   make dev
   # or
   go run ./cmd/app
   ```

6. Open http://localhost:3000

## Key Design Decisions

### Domain Layer

- **Trip**: Core entity with name, destination, and date range
- **Event**: Belongs to a trip, has category, location, time range, and position for ordering
- **Repository interfaces**: Defined in `domain/ports.go` for dependency inversion

### Repository Pattern

- Uses sqlc for type-safe SQL queries
- pgx/v5 for PostgreSQL driver
- Stores implement domain interfaces with compile-time checks
- **Single source of truth**: sqlc reads schema from `migrations/` directory, queries from `internal/repository/sql/`
- **Single package**: All repository code lives in one package to avoid duplicate model generation

### Service Layer

- Contains business logic
- Services take repository interfaces (not concrete implementations)
- `EventService.SuggestStartTime()` - suggests start time based on last event

### Handler Layer

- Uses chi router
- templ for type-safe HTML templates
- HTMX 2.x for dynamic updates
- Method override middleware for PUT/DELETE from forms

## API Routes

### Trips

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | / | TripHandler.List | List all trips |
| GET | /trips/new | TripHandler.NewPage | New trip form |
| POST | /trips | TripHandler.Create | Create trip |
| GET | /trips/{id} | TripHandler.Detail | Trip detail + events |
| GET | /trips/{id}/edit | TripHandler.EditPage | Edit trip form |
| PUT | /trips/{id} | TripHandler.Update | Update trip |
| DELETE | /trips/{id} | TripHandler.Delete | Delete trip |

### Events

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | /trips/{tripID}/events/new | EventHandler.NewPage | New event form |
| POST | /trips/{tripID}/events | EventHandler.Create | Create event |
| GET | /trips/{tripID}/events/{id}/edit | EventHandler.EditPage | Edit event form |
| PUT | /trips/{tripID}/events/{id} | EventHandler.Update | Update event |
| DELETE | /trips/{tripID}/events/{id} | EventHandler.Delete | Delete event |

## Database Schema

### trips table

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL | Primary key |
| user_id | UUID | Future: Supabase auth user |
| name | TEXT | Trip name |
| destination | TEXT | Trip destination |
| start_date | TIMESTAMPTZ | Trip start date |
| end_date | TIMESTAMPTZ | Trip end date |
| created_at | TIMESTAMPTZ | Created timestamp |
| updated_at | TIMESTAMPTZ | Updated timestamp |

### events table

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL | Primary key |
| trip_id | INTEGER | Foreign key to trips |
| title | TEXT | Event title |
| category | TEXT | activity/food/lodging/transit |
| location | TEXT | Location name |
| latitude | DOUBLE PRECISION | Optional latitude |
| longitude | DOUBLE PRECISION | Optional longitude |
| start_time | TIMESTAMPTZ | Event start time |
| end_time | TIMESTAMPTZ | Event end time |
| pinned | BOOLEAN | Highlight in timeline |
| position | INTEGER | Order within trip |
| created_at | TIMESTAMPTZ | Created timestamp |
| updated_at | TIMESTAMPTZ | Updated timestamp |

## Future: Supabase Auth Integration

The codebase is prepared for Supabase auth:

1. `user_id` column exists in `trips` table (nullable)
2. Repository interfaces accept `userID *string` parameter
3. Routes are grouped for easy middleware addition
4. `getUserID()` helper stub in handlers

To enable auth:
1. Uncomment `r.Use(authMiddleware)` in routes.go
2. Implement `authMiddleware` to validate Supabase JWT
3. Implement `getUserID()` to extract user ID from JWT

## Development Commands

```bash
make dev          # Start with hot reload (requires air)
make build        # Build binary to bin/app
make test         # Run tests
make lint         # Run linters
make generate     # Generate sqlc + templ code
make docker-up    # Start PostgreSQL
make docker-down  # Stop PostgreSQL
make migrate-up   # Run migrations
make migrate-down # Rollback migrations
make clean        # Clean build artifacts
```
