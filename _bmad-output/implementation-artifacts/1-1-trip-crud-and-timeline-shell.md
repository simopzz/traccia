# Story 1.1: Trip CRUD & Timeline Shell

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a traveler,
I want to create, view, edit, and delete trips and see them as a day-by-day timeline,
so that I can organize my travel plans with a clear structure from first day to last.

## Acceptance Criteria

1. **Given** a user is on the trip list page, **When** they click "Create Trip" and fill in name (required), date range (required), and destination (optional), **Then** a new trip is created and the user is redirected to the trip timeline page, which shows one section per day spanning the date range.

2. **Given** a user has created trips, **When** they visit the trip list page, **Then** all trips are displayed with name, destination, and dates.

3. **Given** a user is viewing a trip, **When** they edit the trip's name, destination, or date range, **Then** the changes are persisted and the timeline adjusts to the new date range.

4. **Given** a user is viewing a trip, **When** they delete the trip and confirm via the confirmation dialog which displays the number of associated events, **Then** the trip and all associated events are permanently removed, **And** the user is returned to the trip list.

5. **Given** a user is viewing a trip timeline, **When** no events exist for a day, **Then** an EmptyDayPrompt is displayed with an "Add Event" call-to-action.

6. **Given** a user edits a trip's date range to exclude days that contain events, **When** the edit is submitted, **Then** a validation error is shown listing the affected days and their event counts, explaining they must be cleared first.

7. **Given** a user has no trips, **When** they visit the trip list page, **Then** an empty state is displayed with a "Plan your first trip" call-to-action.

## Tasks / Subtasks

### Foundation (blocks all other tasks)

- [ ] Task 1: Rewrite database schema (AC: all)
  - [ ] 1.1 Replace `migrations/001_initial.up.sql` with target schema: `trips` table, `events` table with `event_date DATE`, `position INTEGER` (gap-based), `notes TEXT`, `pinned BOOLEAN`, `category` including `flight`; detail tables (`flight_details`, `lodging_details`, `transit_details`) with `ON DELETE CASCADE`; composite index `(trip_id, event_date, position)`
  - [ ] 1.2 Update `migrations/001_initial.down.sql` to drop all tables in reverse dependency order
  - [ ] 1.3 Update `sqlc.yaml` to output to `internal/repository/sqlcgen/` package
  - [ ] 1.4 Update SQL queries in `internal/repository/sql/trips.sql` for any schema column changes
  - [ ] 1.5 Update SQL queries in `internal/repository/sql/events.sql` — add `event_date` to inserts/updates, update position queries for gap-based algorithm
  - [ ] 1.6 Run `just generate` and verify sqlcgen output in new package location
  - [ ] 1.7 Run `just docker-down && just docker-up && just migrate-up` to verify clean schema

- [ ] Task 2: Update repository layer for sqlcgen isolation (AC: all)
  - [ ] 2.1 Update `internal/repository/trip_store.go` imports to use `sqlcgen` package
  - [ ] 2.2 Update `internal/repository/event_store.go` imports to use `sqlcgen` package
  - [ ] 2.3 Map sqlcgen-generated types to domain types in store adapters (no domain type leakage)
  - [ ] 2.4 Verify all repository methods compile and pass basic smoke test

- [ ] Task 3: Update domain layer (AC: all)
  - [ ] 3.1 Update `internal/domain/models.go` — ensure Trip has `StartDate`, `EndDate` (not optional); Event has `EventDate`, `Position`, `Notes`, `Pinned`; add `CategoryFlight` constant
  - [ ] 3.2 Update `internal/domain/ports.go` — ensure `TripRepository` and `EventRepository` interfaces align with updated store methods; add `CountEventsByTripAndDateRange` or equivalent for AC #6 validation
  - [ ] 3.3 Update `internal/domain/errors.go` if new error types needed (e.g., `ErrDateRangeConflict`)

- [ ] Task 4: Set up Tailwind CSS + templui design system (AC: all UI)
  - [ ] 4.1 Install Tailwind CSS CLI, create `static/css/input.css` with `@import` directives and theme overrides (teal brand `#008080`, Swiss signal colors, Inter font)
  - [ ] 4.2 Install templui components via CLI: Card, Dialog, Sheet, Toast, Tabs, Breadcrumb, Button, Input, Textarea, Form, Label, DatePicker, Skeleton, Separator, Badge, Icon
  - [ ] 4.3 Vendor JS dependencies to `static/js/`: `htmx.min.js`, `alpine.min.js` (with version + source URL comments)
  - [ ] 4.4 Update `Justfile` — add `just css` for Tailwind build, update `just dev` to run air + Tailwind watcher concurrently
  - [ ] 4.5 Update `internal/handler/layout.templ` — replace embedded CSS with Tailwind link, add vendored JS script tags, add base layout structure (nav, main, breadcrumb slots)
  - [ ] 4.6 Verify `just dev` runs both Go hot reload and Tailwind watcher

### Trip CRUD

- [ ] Task 5: Update service layer (AC: #1, #3, #6)
  - [ ] 5.1 Update `internal/service/trip.go` — validate name (required) and date range (required), destination optional
  - [ ] 5.2 Add date range shrink validation: query events with `event_date` outside new range, return `ErrValidation` with affected day details and event counts (AC #6)
  - [ ] 5.3 Ensure `Update` uses updater pattern with date range validation before persistence
  - [ ] 5.4 Add `GetEventCountByTrip` or similar for delete confirmation dialog (AC #4)

- [ ] Task 6: Update handler layer — Trip CRUD (AC: #1, #2, #3, #4, #7)
  - [ ] 6.1 Update `internal/handler/trip.go` — `Create` handler redirects to trip timeline page (not list)
  - [ ] 6.2 Update `List` handler to pass event counts per trip (for delete dialog) and handle empty state
  - [ ] 6.3 Update `Detail` handler to compute date range slice (`StartDate` to `EndDate` inclusive) and pass to template
  - [ ] 6.4 Update `Update` handler to handle validation errors from date range shrink (return form with inline errors)
  - [ ] 6.5 Update `Delete` handler to use templui Dialog component with event count

### Timeline Shell & Templates

- [ ] Task 7: Redesign trip list templates (AC: #2, #7)
  - [ ] 7.1 Rewrite `internal/handler/trip.templ` — `TripListPage` using templui Card components with Swiss bordered style (1px border, 2px hard shadow), teal accent
  - [ ] 7.2 `TripCard` shows name, destination (if present), date range
  - [ ] 7.3 Empty state component: "Plan your first trip" with primary teal button (AC #7)

- [ ] Task 8: Build timeline shell templates (AC: #1, #5)
  - [ ] 8.1 Create `TimelineDay` component — day heading format "Day N — Weekday, Month Day" (e.g., "Day 3 — Wednesday, May 14"), vertical spine connector, event slot area
  - [ ] 8.2 Create `EmptyDayPrompt` component — centered message "Add your first event to Day N" + "Add Event" secondary button wired to `GET /trips/{id}/events/new?date={date}` (placeholder until Story 1.2)
  - [ ] 8.3 `TripDetailPage` renders day tabs (templui Tabs) + breadcrumb (Trip List → Trip Name) + full date range as `TimelineDay` sections
  - [ ] 8.4 Ensure responsive layout: single column, max-width ~800px centered, mobile full-width

- [ ] Task 9: Trip create/edit form templates (AC: #1, #3, #6)
  - [ ] 9.1 Rewrite `TripNewPage` / `TripEditPage` using templui Form, Input, DatePicker components
  - [ ] 9.2 Name field required, destination optional, date range required (start + end date pickers)
  - [ ] 9.3 Field-level validation display: rose border + inline error message for invalid fields
  - [ ] 9.4 Delete confirmation using templui Dialog: "Delete {trip name}? This will remove all {N} events." with Cancel (ghost) + Delete (destructive rose) buttons

### Testing

- [ ] Task 10: Write tests (AC: all)
  - [ ] 10.1 Service tests: `internal/service/trip_test.go` — table-driven tests for Create (valid, missing name, missing dates), Update (valid, date range shrink with events, date range expand), Delete
  - [ ] 10.2 Service tests: date range shrink validation — verify returns error with affected day details when events exist outside new range
  - [ ] 10.3 Run `just test` and `just lint` — all passing, zero violations

## Dev Notes

### Critical: Schema Is the First Task

The existing `001_initial.up.sql` has a flat `events` table without detail tables, `event_date`, gap-based positions, `notes`, or Flight category. **Replace it entirely** with the target schema from architecture.md. No data migration — no production data exists. This unblocks all downstream stories (1.2-1.6). Do NOT try to preserve the old schema and layer migrations on top.

### Critical: Never Edit Generated Files

Files in `internal/repository/sqlcgen/` are generated by sqlc. NEVER edit `db.go`, `models.go`, `*_sql.go` in that directory. All changes flow: `migrations/*.sql` (schema) → `internal/repository/sql/*.sql` (queries) → `just generate` → `sqlcgen/` output. Store adapters in `internal/repository/*_store.go` map generated types to domain types.

### Timeline Day Generation

The handler MUST generate a slice of all dates from `trip.StartDate` to `trip.EndDate` inclusive. This is computed in the handler, not queried from the database. Events are fetched separately via `ListByTripAndDate` and distributed into their respective day slots. Days with zero events render `EmptyDayPrompt`. The timeline shell is the trip's date range skeleton — it exists even with no events.

### Day Heading Format

Day section headings use format: **"Day N — Weekday, Month Day"** (e.g., "Day 3 — Wednesday, May 14"). Day number is 1-indexed from trip start date. Include both the relative day number and the absolute date with weekday for context.

### Destination Is Optional

Trip creation requires name + date range. Destination is optional — users may not know their destination when starting to plan. The trip list and detail page handle missing destination gracefully (omit or show placeholder).

### Date Range Shrink Validation

When updating a trip's date range, the service layer must check for events with `event_date` outside the new range before allowing the update. If events exist on excluded days, return `domain.ErrValidation` with a message listing affected days and their event counts. Example: "Cannot shorten trip: Day 5 (May 16) has 3 events. Remove or move them first."

### Delete Confirmation Dialog

Use templui Dialog component, not browser `confirm()`. The dialog displays: "Delete {trip name}? This will remove all {N} events." Buttons: Cancel (ghost style) + Delete (destructive rose style). Event count requires a query — pass it to the template.

### EmptyDayPrompt "Add Event" Button

Wire the "Add Event" button to `GET /trips/{id}/events/new?date={date}`. This route will be implemented in Story 1.2. For Story 1.1, the button links to the route — it may 404 or show a placeholder until 1.2 lands. Do NOT omit the button or disable it.

### Tailwind + templui Setup

- Install Tailwind CSS CLI build pipeline: `static/css/input.css` → `static/css/app.css`
- Theme overrides in `input.css`: brand teal `#008080`, signal colors (emerald/amber/rose), Inter font, tabular numerals on time displays
- templui components installed via CLI (`templui add ...`) — they become owned source in the project
- `just dev` must run air (Go + templ hot reload) AND Tailwind watcher concurrently
- Vendor `htmx.min.js` and `alpine.min.js` to `static/js/` with version + source URL comments in the files
- Replace embedded CSS in `layout.templ` with Tailwind stylesheet link

### HTMX Patterns for Trip CRUD

- Trip create: standard form POST → redirect (HTTP 303) to `/trips/{id}`
- Trip edit: standard form POST with `_method=PUT` → redirect to `/trips/{id}`
- Trip delete: Dialog confirm → form POST with `_method=DELETE` → redirect to `/trips` (or HTMX `hx-delete` with `HX-Redirect` header)
- No HTMX partial swaps needed for Trip CRUD in Story 1.1 — full page navigations are appropriate here

### Existing Code Reuse

The existing codebase has working Trip CRUD logic in handlers, services, and repositories. Adapt and refactor — do not rewrite from scratch. Key changes:
- Repository: update imports to `sqlcgen` package, adjust column mappings for schema changes
- Service: add date range shrink validation, make destination optional
- Handler: update redirect on create, add date range computation for timeline, update templates
- Templates: full rewrite to Tailwind + templui — the Go template logic can inform the new implementation but the HTML/CSS is all new

### Project Structure Notes

- All files follow the established layered architecture: `handler/` → `service/` → `domain/` ← `repository/`
- sqlc output isolated in `internal/repository/sqlcgen/` (separate Go package) — this is a CHANGE from the current flat `repository/` layout
- templ output (`*_templ.go`) co-located with `.templ` source (tool constraint)
- Static assets: `static/css/` for Tailwind, `static/js/` for vendored JS
- No new directories beyond what architecture.md specifies

### References

- [Source: _bmad-output/planning-artifacts/architecture.md#Data Architecture] — schema design, detail tables, gap-based positions
- [Source: _bmad-output/planning-artifacts/architecture.md#Project Structure & Boundaries] — complete directory tree, sqlcgen isolation
- [Source: _bmad-output/planning-artifacts/architecture.md#Implementation Patterns & Consistency Rules] — naming, HTMX contract, error handling
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md#Design Direction] — Swiss Bordered Cards, timeline spine, event type icons
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md#Component Strategy] — TimelineDay, EmptyDayPrompt, DayOverview specs
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md#UX Consistency Patterns] — button hierarchy, feedback patterns, form patterns, navigation
- [Source: _bmad-output/planning-artifacts/prd.md#Functional Requirements] — FR1-FR5 (Trip CRUD), FR13 (timeline display)
- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.1] — acceptance criteria, BDD scenarios

## Dev Agent Record

### Agent Model Used

<!-- To be filled by dev agent -->

### Debug Log References

### Completion Notes List

### File List
