---
title: 'Database Seeder'
slug: 'database-seeder'
created: '2026-02-19'
status: 'completed'
stepsCompleted: [1, 2, 3, 4]
tech_stack: ['Go', 'PostgreSQL', 'pgx', 'log/slog']
files_to_modify: ['cmd/seed/main.go', 'go.mod']
code_patterns: ['Service Layer Pattern', 'Dependency Injection', 'Repository Pattern', 'CLI Tool']
test_patterns: ['Manual verification via CLI execution', 'UI verification']
---

# Tech-Spec: Database Seeder

**Created:** 2026-02-19

## Overview

### Problem Statement

Development requires realistic test data (Trips and Events) to verify UI flows and timelines without manual entry or fragile SQL scripts.

### Solution

Implement a dedicated `cmd/seed/main.go` CLI tool that initializes the application's Service Layer (`TripService`, `EventService`) to programmatically generate random trips and events, ensuring all business logic and validation rules are respected. The tool will use **structured logging** (`log/slog`) to provide clear feedback during the seeding process.

### Scope

**In Scope:**
- Create `cmd/seed/main.go` entry point.
- Implement logic to generate random trips (names, destinations, dates).
- Implement logic to generate random events within trip dates (titles, categories, times).
- Use `TripService` and `EventService` to persist data.
- **Implement a `--clean` flag to remove existing seed data before running.**
- **Implement explicit check for non-production environments.**

**Out of Scope:**
- SQL data loading.
- User interface for seeding (beyond CLI logs).
- User authentication for the seeder (it will run as a local administrative tool).

## Context for Development

### Codebase Patterns

- **Service Layer Pattern**: Business logic resides in `internal/service/`.
- **Repository Pattern**: Data access is abstracted in `internal/repository/`.
- **Configuration**: Uses `internal/infra/config` for database connection strings.
- **Dependency Injection**: Dependencies are manually wired in `main.go`.

### Files to Reference

| File | Purpose |
| ---- | ------- |
| `cmd/app/main.go` | Application entry point and DI reference. |
| `internal/service/trip.go` | Trip creation logic and validation. |
| `internal/service/event.go` | Event creation logic and validation. |
| `internal/infra/database/postgres.go` | Database connection setup. |
| `internal/infra/config/config.go` | Configuration loading. |

### Technical Decisions

- **Why Service Layer?**: The user requested avoiding direct DB writes to ensure data integrity and leverage existing validation logic (e.g., date ranges).
- **Why `cmd/seed/main.go`?**: Isolated executable that can be run on demand without polluting the main application binary.
- **Random Data**: Will use standard library `math/rand` or a helper to generate plausible-looking data.
- **Dependency Management**: Reuse `internal/infra/config` and `internal/infra/database` to ensure consistent connection settings.
- **Cleanup Strategy**: The `--clean` flag will delete trips created by this seeder (identified by a specific naming convention or just delete all trips in dev/test environment).
- **Why no TUI?**: Simplified for maintainability and robustness. Standard logging is sufficient for a developer tool.

## Implementation Plan

### Tasks

1.  **[x] Create Seeder Entry Point and Setup**
    *   **File:** `cmd/seed/main.go`
    *   **Action:** Create the `main` package and `main` function.
    *   **Details:**
        *   Load configuration using `config.Load()`.
        *   **CRITICAL:** Check `cfg.Environment`. If it is "production", LOG ERROR AND EXIT IMMEDIATELY.
        *   Establish database connection using `database.NewPool()`. **Defer `pool.Close()`**.
        *   Initialize `TripStore` and `EventStore`.
        *   Initialize `TripService` and `EventService`.
        *   Parse command line flags: `clean := flag.Bool("clean", false, "Clean up existing seed data")`.

2.  **[x] Implement Cleanup Logic**
    *   **File:** `cmd/seed/main.go`
    *   **Action:** Implement `cleanup` function.
    *   **Details:**
        *   If `--clean` is set, execute a cleanup routine.
        *   **Strategy:** Use `TripStore` (Repository) directly to execute a `DELETE FROM trips WHERE name LIKE '[SEED]%'` query (or similar pattern) to be efficient.
        *   Log the number of deleted records.

3.  **[x] Implement Data Generation Logic**
    *   **File:** `cmd/seed/main.go`
    *   **Action:** Add functions to generate and persist trips and events.
    *   **Details:**
        *   Create `seedTrips` function.
        *   **Loop:** Create N (e.g., 5) trips.
        *   **Constraints:**
            *   Name starts with "[SEED]" for easy identification/cleanup.
            *   Duration 3-14 days. Start date within next 365 days.
        *   **For each trip:**
            *   Call `tripService.Create`.
            *   Log success: `slog.Info("Created trip", "name", trip.Name, "id", trip.ID)`.
            *   Call `seedEvents(trip)`.
        *   **Events Logic:**
            *   Iterate through each day of the trip.
            *   Create M (e.g., 1-3) events per day.
            *   **Constraints:** Start time 08:00-20:00. Duration 1-3h.
            *   Call `eventService.Create`.
            *   Log success (debug level): `slog.Debug("Created event", "title", event.Title)`.

### Acceptance Criteria

*   **AC1: Prod Guard Works.**
    *   **Given** `ENVIRONMENT=production`,
    *   **When** I run `go run cmd/seed/main.go`,
    *   **Then** the command exits with an error and does NOT modify the database.
*   **AC2: Data is seeded correctly.**
    *   **Given** dev environment,
    *   **When** I run `go run cmd/seed/main.go`,
    *   **Then** valid trips and events exist in the DB.
*   **AC3: Cleanup Flag Works.**
    *   **Given** existing seed data,
    *   **When** I run `go run cmd/seed/main.go --clean`,
    *   **Then** old data matching the seed pattern is removed before new data is added.
*   **AC4: Realistic Constraints.**
    *   **Given** seeded events,
    *   **When** I inspect their times,
    *   **Then** they fall within reasonable hours (08:00-20:00) and durations (1-3h).

## Additional Context

### Dependencies

- `github.com/simopzz/traccia/internal/service`
- `github.com/simopzz/traccia/internal/repository`
- `github.com/simopzz/traccia/internal/infra/database`
- `github.com/simopzz/traccia/internal/infra/config`
- `math/rand`
- `log/slog`

### Testing Strategy

- **Manual Verification:** Run the tool with different flags and check logs/DB.

### Notes

- Use `[SEED]` prefix for trip names to make cleanup safe and easy.

## Review Notes
- Adversarial review completed
- Findings: 6 total, 4 fixed, 2 skipped
- Resolution approach: auto-fix
