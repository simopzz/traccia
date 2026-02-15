# Story 1.1: Trip CRUD & Timeline Shell

Status: review

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

- [x] Task 1: Rewrite database schema (AC: all)
  - [x] 1.1 Replace `migrations/001_initial.up.sql` with target schema: `trips` table, `events` table with `event_date DATE`, `position INTEGER` (gap-based), `notes TEXT`, `pinned BOOLEAN`, `category` including `flight`; detail tables (`flight_details`, `lodging_details`, `transit_details`) with `ON DELETE CASCADE`; composite index `(trip_id, event_date, position)`
  - [x] 1.2 Update `migrations/001_initial.down.sql` to drop all tables in reverse dependency order
  - [x] 1.3 Update `sqlc.yaml` to output to `internal/repository/sqlcgen/` package
  - [x] 1.4 Update SQL queries in `internal/repository/sql/trips.sql` for any schema column changes
  - [x] 1.5 Update SQL queries in `internal/repository/sql/events.sql` — add `event_date` to inserts/updates, update position queries for gap-based algorithm
  - [x] 1.6 Run `just generate` and verify sqlcgen output in new package location
  - [x] 1.7 Run `just docker-down && just docker-up && just migrate-up` to verify clean schema

- [x] Task 2: Update repository layer for sqlcgen isolation (AC: all)
  - [x] 2.1 Update `internal/repository/trip_store.go` imports to use `sqlcgen` package
  - [x] 2.2 Update `internal/repository/event_store.go` imports to use `sqlcgen` package
  - [x] 2.3 Map sqlcgen-generated types to domain types in store adapters (no domain type leakage)
  - [x] 2.4 Verify all repository methods compile and pass basic smoke test

- [x] Task 3: Update domain layer (AC: all)
  - [x] 3.1 Update `internal/domain/models.go` — ensure Trip has `StartDate`, `EndDate` (not optional); Event has `EventDate`, `Position`, `Notes`, `Pinned`; add `CategoryFlight` constant
  - [x] 3.2 Update `internal/domain/ports.go` — ensure `TripRepository` and `EventRepository` interfaces align with updated store methods; add `CountEventsByTripAndDateRange` or equivalent for AC #6 validation
  - [x] 3.3 Update `internal/domain/errors.go` if new error types needed (e.g., `ErrDateRangeConflict`)

- [x] Task 4: Set up Tailwind CSS + templui design system (AC: all UI)
  - [x] 4.1 Install Tailwind CSS CLI, create `static/css/input.css` with `@import` directives and theme overrides (teal brand `#008080`, Swiss signal colors, Inter font)
  - [x] 4.2 Install templui components via CLI: Card, Dialog, Sheet, Toast, Tabs, Breadcrumb, Button, Input, Textarea, Form, Label, DatePicker, Skeleton, Separator, Badge, Icon
  - [x] 4.3 Vendor JS dependencies to `static/js/`: `htmx.min.js`, `alpine.min.js` (with version + source URL comments)
  - [x] 4.4 Update `Justfile` — add `just css` for Tailwind build, update `just dev` to run air + Tailwind watcher concurrently
  - [x] 4.5 Update `internal/handler/layout.templ` — replace embedded CSS with Tailwind link, add vendored JS script tags, add base layout structure (nav, main, breadcrumb slots)
  - [x] 4.6 Verify `just dev` runs both Go hot reload and Tailwind watcher

### Trip CRUD

- [x] Task 5: Update service layer (AC: #1, #3, #6)
  - [x] 5.1 Update `internal/service/trip.go` — validate name (required) and date range (required), destination optional
  - [x] 5.2 Add date range shrink validation: query events with `event_date` outside new range, return `ErrValidation` with affected day details and event counts (AC #6)
  - [x] 5.3 Ensure `Update` uses updater pattern with date range validation before persistence
  - [x] 5.4 Add `GetEventCountByTrip` or similar for delete confirmation dialog (AC #4)

- [x] Task 6: Update handler layer — Trip CRUD (AC: #1, #2, #3, #4, #7)
  - [x] 6.1 Update `internal/handler/trip.go` — `Create` handler redirects to trip timeline page (not list)
  - [x] 6.2 Update `List` handler to pass event counts per trip (for delete dialog) and handle empty state
  - [x] 6.3 Update `Detail` handler to compute date range slice (`StartDate` to `EndDate` inclusive) and pass to template
  - [x] 6.4 Update `Update` handler to handle validation errors from date range shrink (return form with inline errors)
  - [x] 6.5 Update `Delete` handler to use templui Dialog component with event count

### Timeline Shell & Templates

- [x] Task 7: Redesign trip list templates (AC: #2, #7)
  - [x] 7.1 Rewrite `internal/handler/trip.templ` — `TripListPage` using templui Card components with Swiss bordered style (1px border, 2px hard shadow), teal accent
  - [x] 7.2 `TripCard` shows name, destination (if present), date range
  - [x] 7.3 Empty state component: "Plan your first trip" with primary teal button (AC #7)

- [x] Task 8: Build timeline shell templates (AC: #1, #5)
  - [x] 8.1 Create `TimelineDay` component — day heading format "Day N — Weekday, Month Day" (e.g., "Day 3 — Wednesday, May 14"), vertical spine connector, event slot area
  - [x] 8.2 Create `EmptyDayPrompt` component — centered message "Add your first event to Day N" + "Add Event" secondary button wired to `GET /trips/{id}/events/new?date={date}` (placeholder until Story 1.2)
  - [x] 8.3 `TripDetailPage` renders day tabs (templui Tabs) + breadcrumb (Trip List → Trip Name) + full date range as `TimelineDay` sections
  - [x] 8.4 Ensure responsive layout: single column, max-width ~800px centered, mobile full-width

- [x] Task 9: Trip create/edit form templates (AC: #1, #3, #6)
  - [x] 9.1 Rewrite `TripNewPage` / `TripEditPage` using templui Form, Input, DatePicker components
  - [x] 9.2 Name field required, destination optional, date range required (start + end date pickers)
  - [x] 9.3 Field-level validation display: rose border + inline error message for invalid fields
  - [x] 9.4 Delete confirmation using templui Dialog: "Delete {trip name}? This will remove all {N} events." with Cancel (ghost) + Delete (destructive rose) buttons

### Testing

- [x] Task 10: Write tests (AC: all)
  - [x] 10.1 Service tests: `internal/service/trip_test.go` — table-driven tests for Create (valid, missing name, missing dates), Update (valid, date range shrink with events, date range expand), Delete
  - [x] 10.2 Service tests: date range shrink validation — verify returns error with affected day details when events exist outside new range
  - [x] 10.3 Run `just test` and `just lint` — all passing, zero violations

### Review Follow-ups (AI)

- [x] [AI-Review][HIGH] AC #4: Delete dialog must display actual event count — wire `CountByTrip` in `EditPage` handler, pass count to template, update dialog text to "This will remove all {N} events" [`internal/handler/trip.go:106`, `internal/handler/trip.templ:237`]
- [x] [AI-Review][HIGH] AC #6: Date range shrink must list affected days with event counts — wire `CountEventsByTripGroupedByDate` into repository interface and service, return per-day details in error message [`internal/service/trip.go:88`, `internal/repository/sql/events.sql:44`]
- [x] [AI-Review][HIGH] Trip Update service must validate input — add name non-empty, dates non-zero, end >= start validation before applying updater [`internal/service/trip.go:69`]
- [x] [AI-Review][MEDIUM] TripNewPage must preserve form values on validation error — pass submitted input to template, add `value` attributes to form fields [`internal/handler/trip.go:59`, `internal/handler/trip.templ:62`]
- [x] [AI-Review][MEDIUM] Event delete handler should redirect to trip timeline instead of returning 200 with empty body [`internal/handler/event.go:154`]
- [x] [AI-Review][MEDIUM] Add Notes textarea to event new/edit forms [`internal/handler/event.templ:59`, `internal/handler/event.templ:155`]
- [x] [AI-Review][MEDIUM] ValidateDateRangeShrink should only run when range actually shrinks — compare old vs new dates first [`internal/handler/trip.go:146`]
- [x] [AI-Review][LOW] Move shared pgtype helper functions to `internal/repository/helpers.go`
- [x] [AI-Review][LOW] EventDate derivation via `Truncate(24*time.Hour)` is fragile for future timezone support [`internal/service/event.go:51`]

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

Claude Opus 4.6

### Debug Log References

No debug issues encountered.

### Completion Notes List

- Rewrote database schema with target architecture: trips (DATE columns), events (event_date, notes, gap-based positions), flight_details, lodging_details, transit_details
- Migrated sqlc output to isolated `internal/repository/sqlcgen/` package
- Updated domain models with EventDate, Notes, CategoryFlight
- Updated repository interfaces with CountEventsByTripAndDateRange, ListByTripAndDate, CountByTrip
- Added ErrDateRangeConflict domain error
- Set up Tailwind CSS v4 standalone CLI with brand teal theme
- Installed templui v1.5.0 components (card, dialog, button, input, textarea, form, label, badge, separator, icon, tabs, breadcrumb, toast, skeleton, datepicker, sheet)
- Vendored HTMX 2.0.8 and Alpine.js 3.15.8
- Updated Justfile with css/css-watch recipes and concurrent dev workflow
- Updated service layer: destination optional, date range required, date range shrink validation
- Updated handler layer: timeline day computation from date range, form error handling, delete confirmation via Alpine.js dialog
- Rewrote all templates with Tailwind CSS: Swiss bordered cards, timeline spine, empty day prompts, breadcrumb navigation
- Wrote 11 table-driven service tests covering Create, Update, Delete, and ValidateDateRangeShrink
- All tests pass, all lint checks pass
- Resolved review finding [HIGH]: Delete dialog now displays actual event count from CountByTrip
- Resolved review finding [HIGH]: Date range shrink now returns per-day details with event counts via CountEventsByTripGroupedByDate
- Resolved review finding [HIGH]: Trip Update service now validates name non-empty, dates non-zero, end >= start before applying updater
- Resolved review finding [MEDIUM]: TripNewPage preserves form values (name, destination, dates) on validation error
- Resolved review finding [MEDIUM]: Event delete handler now redirects to trip timeline (303) instead of returning 200
- Resolved review finding [MEDIUM]: Added Notes textarea to both event new and edit forms
- Resolved review finding [MEDIUM]: ValidateDateRangeShrink now compares old vs new dates first, only queries DB when range actually shrinks
- Resolved review finding [LOW]: Moved shared pgtype helpers (toPgDate, toPgTimestamptz, toPgText, toPgFloat8, toPgBool) to internal/repository/helpers.go
- Resolved review finding [LOW]: Replaced Truncate(24*time.Hour) with time.Date() for safer EventDate derivation

### File List

- `migrations/001_initial.up.sql` — rewritten with target schema
- `migrations/001_initial.down.sql` — updated drop order
- `sqlc.yaml` — output to sqlcgen package
- `internal/repository/sql/trips.sql` — updated queries, added CountEventsByTripAndDateRange
- `internal/repository/sql/events.sql` — added event_date, notes, gap-based position queries, CountEventsByTrip, CountEventsByTripGroupedByDate, ListEventsByTripAndDate
- `internal/repository/sqlcgen/` — new generated package (db.go, models.go, events.sql.go, trips.sql.go)
- `internal/repository/trip_store.go` — updated imports, DATE types, CountEventsByTripAndDateRange method
- `internal/repository/event_store.go` — updated imports, event_date/notes mapping, gap-based positions, CountByTrip/ListByTripAndDate methods
- `internal/repository/db.go` — deleted (moved to sqlcgen)
- `internal/repository/models.go` — deleted (moved to sqlcgen)
- `internal/repository/events.sql.go` — deleted (moved to sqlcgen)
- `internal/repository/trips.sql.go` — deleted (moved to sqlcgen)
- `internal/domain/models.go` — added EventDate, Notes, CategoryFlight
- `internal/domain/ports.go` — added CountEventsByTripAndDateRange, ListByTripAndDate, CountByTrip
- `internal/domain/errors.go` — added ErrDateRangeConflict
- `internal/service/trip.go` — destination optional, date range validation, ValidateDateRangeShrink
- `internal/service/event.go` — added EventDate derivation, Notes, CountByTrip, ListByTripAndDate, pointer receiver for Update
- `internal/service/trip_test.go` — new: 11 table-driven tests
- `internal/handler/trip.go` — TimelineDayData, buildTimelineDays, FormErrors, date range shrink validation
- `internal/handler/event.go` — pointer UpdateEventInput
- `internal/handler/layout.templ` — Tailwind CSS, vendored JS
- `internal/handler/trip.templ` — full rewrite with Tailwind: TripListPage, TripCard, TripNewPage, TripEditPage, TripDetailPage, TimelineDay, EmptyDayPrompt
- `internal/handler/event.templ` — full rewrite with Tailwind: EventTimelineItem, EventNewPage, EventEditPage
- `internal/repository/helpers.go` — new: shared pgtype helper functions (toPgDate, toPgTimestamptz, toPgText, toPgFloat8, toPgBool)
- `internal/handler/helpers.go` — unchanged
- `internal/handler/routes.go` — unchanged
- `Justfile` — added css, css-watch recipes, updated dev/build
- `.golangci.yml` — excluded internal/components and test files from some linters
- `.templui.json` — new: templui configuration
- `static/css/input.css` — new: Tailwind config with brand theme
- `static/css/app.css` — new: generated Tailwind output
- `static/js/htmx.min.js` — new: vendored HTMX 2.0.8
- `static/js/alpine.min.js` — new: vendored Alpine.js 3.15.8
- `static/js/*.min.js` — new: templui component JS files
- `internal/components/` — new: templui component source files

## Change Log

- 2026-02-15: Story implementation complete — all 10 tasks done, 11 tests passing, lint clean
- 2026-02-15: Code review — 3 HIGH, 4 MEDIUM, 2 LOW issues found. 9 action items created. Status → in-progress
- 2026-02-15: Addressed all 9 code review findings — 13 tests passing, lint clean. Status → review
