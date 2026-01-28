# Story 1.2: trip-management-create-read-reset

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Planner (Sarah),
I want to create a new Trip with a name, destination, and dates,
so that I have a container to start organizing my itinerary.

## Acceptance Criteria

1. **Given** the user is on the home page
2. **When** they enter "Japan Trip" and dates and click "Start Planning"
3. **Then** a new Trip record is created in the database
4. **And** the user is redirected to the Trip Timeline view (e.g., `/trips/{uuid}`)
5. **And** the "Clear Trip" button deletes all associated events for that trip

## Tasks / Subtasks

- [x] **Database Schema Migration**
    - [x] Create `trips` table (id UUID PK, name TEXT, destination TEXT, start_date DATE, end_date DATE, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ)
    - [x] Create `events` table (id UUID PK, trip_id UUID FK, title TEXT, location TEXT, start_time TIMESTAMPTZ, end_time TIMESTAMPTZ, created_at TIMESTAMPTZ, updated_at TIMESTAMPTZ) - *Foundation for Reset functionality*
- [x] **Backend Implementation (Features/Timeline)**
    - [x] Define `Trip` and `Event` structs in `internal/features/timeline/models.go`
    - [x] Implement `CreateTrip(ctx, params)` service method in `internal/features/timeline/service.go`
    - [x] Implement `GetTrip(ctx, id)` service method
    - [x] Implement `ResetTrip(ctx, id)` service method (DELETE FROM events WHERE trip_id = ?)
    - [x] Implement `POST /trips` handler (Parse form, Call Service, Redirect)
    - [x] Implement `GET /trips/{id}` handler (Render Timeline View)
    - [x] Implement `POST /trips/{id}/reset` handler (Call Service, Refresh Page/Part)
- [x] **Frontend Implementation (Templ/HTMX)**
    - [x] Create `internal/features/timeline/home.templ`: Landing page with "Create Trip" form
    - [x] Create `internal/features/timeline/view.templ`: Timeline page showing Trip details + "Clear Trip" button
    - [x] Configure `GET /` to render `home.templ`
    - [x] Add `hx-post="/trips/{id}/reset"` and `hx-confirm="Are you sure?"` to Clear button

## Dev Notes

- **Context:** This is the first functional feature. Establish the pattern of "Feature Folder -> Service -> Handler -> Templ".
- **Phase 1 MVP:** Single-player mode. **No Authentication** required for this story.
- **Database:** Use `pgx/v5` or the driver provided by Blueprint. Ensure UUID generation is handled (e.g., in Go or DB default).
- **HTMX:** Use `hx-boost="true"` on the body or main container to enable SPA-like navigation between Home and Trip View.

### Technical Requirements

- **Language:** Go 1.25 (Chi Router)
- **Database:** Postgres
- **Frontend:** Templ + HTMX + Tailwind v4
- **Routing:**
  - `GET /` -> Home (Create Form)
  - `POST /trips` -> Create Action
  - `GET /trips/{id}` -> Timeline View
  - `POST /trips/{id}/reset` -> Reset Action

### Architecture Compliance

- **Feature Folders:** `internal/features/timeline/` is the ONLY place for this code.
- **Naming Conventions:**
  - DB Tables: `snake_case`, plural (`trips`, `events`).
  - Go Structs: `CamelCase` (`Trip`, `Event`).
  - JSON Tags: `camelCase` (`json:"tripId"`).
- **Styling:** Use Tailwind utility classes. No custom CSS files unless absolutely necessary (add to `input.css` if so).

### Library/Framework Requirements

- **Chi:** Use for routing.
- **Templ:** All HTML must be generated via Templ.
- **HTMX:** Use for form submission and partial updates.
- **Tailwind:** Ensure classes are picked up by the build process (re-run `make css` or ensure `air` is watching `.templ` files).

### File Structure Requirements

```bash
traccia/
├── internal/features/timeline/
│   ├── models.go        # Trip/Event structs
│   ├── service.go       # CreateTrip, GetTrip, ResetTrip logic
│   ├── handler.go       # HTTP handlers
│   ├── home.templ       # Landing page
│   └── view.templ       # Trip view
├── migrations/
│   └── YYYYMMDDHHMMSS_create_trips_events.sql
```

## Previous Story Intelligence

- **From Story 1.1:**
  - Tailwind v4 is active. Use `@source` in `input.css` if Templ files aren't being scanned, or rely on the configured content paths.
  - `internal/features/timeline` already exists. Use it.

## Git Intelligence Summary

- **Recent Activity:** Story 1.1 established the scaffolding. This is the first "real" code.

## Latest Tech Information

- **Tailwind v4:** Remember it scans files automatically if configured correctly. No `tailwind.config.js` might be present if using the CSS-first configuration.

## Project Context Reference

- [Epics: Story 1.2](_bmad-output/planning-artifacts/epics.md#story-12-trip-management-createreadreset)
- [Architecture: Feature Folders](_bmad-output/planning-artifacts/architecture.md#structure-patterns)

## Story Completion Status

- **Status:** ready-for-dev
- **Validation:** Ready for `dev-story` execution.

## Dev Agent Record

### Agent Model Used
opencode/gemini-3-pro

### Debug Log References
- Confirmed Story 1.1 completion.
- Confirmed "Single Player" mode (No Auth) for MVP.
- Identified target file structure.

### Completion Notes List
- Implemented `trips` and `events` tables via golang-migrate.
- Implemented `CreateTrip`, `GetTrip`, `ResetTrip` service methods with tests.
- Implemented Templ views for Home and Trip View using `web/layouts.Base()`.
- Implemented Chi handlers and registered them in `routes.go`.
- Added `hx-boost="true"` to base layout as per Dev Notes.
- Added unit and integration tests for service and handlers.
- Verified 100% test pass rate.
- Fixed 404 handling for missing trips.
- Improved date parsing robustness.
- Fixed UI rendering for missing End Date.

### File List
- internal/features/timeline/models.go
- internal/features/timeline/service.go
- internal/features/timeline/service_test.go
- internal/features/timeline/handler.go
- internal/features/timeline/handler_test.go
- internal/features/timeline/home.templ
- internal/features/timeline/home_templ.go
- internal/features/timeline/view.templ
- internal/features/timeline/view_templ.go
- migrations/000001_create_trips_events_tables.up.sql
- migrations/000001_create_trips_events_tables.down.sql
- internal/database/database.go
- internal/server/routes.go
- web/layouts/base.templ
