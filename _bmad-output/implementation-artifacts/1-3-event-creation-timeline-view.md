# Story 1.3: event-creation-timeline-view

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Planner,
I want to add events with times and locations to my trip and see them in a linear vertical list,
so that I can visualize the flow of my day.

## Acceptance Criteria

1. **Given** a Trip exists
2. **When** I fill out the "Add Event" form (Title, Address, Category, Start Time, End Time)
3. **Then** the event is saved to the database with UTC timestamps
4. **And** the timeline updates via HTMX to show the new event card
5. **And** the vertical height of the card is proportional to its duration (approx 64px per hour)
6. **And** the Start Time must be before End Time (Validation)

## Tasks / Subtasks

- [x] **Database Schema Migration**
    - [x] Create/Update `events` table to support new requirements:
        - [x] Ensure `category` (TEXT) exists (or add it)
        - [x] Add `geo_lat` (DOUBLE PRECISION, NULLABLE) - *Foundation for Story 2.1*
        - [x] Add `geo_lng` (DOUBLE PRECISION, NULLABLE) - *Foundation for Story 2.1*
        - [x] *Note: `location` (TEXT) already exists from Story 1.2 foundation*
- [x] **Backend Implementation (Features/Timeline)**
    - [x] Update `Event` struct in `internal/features/timeline/models.go` (add Category, GeoLat, GeoLng)
    - [x] Implement `CreateEvent(ctx, params)` in `service.go`
        - [x] Validate EndTime > StartTime
        - [x] Store times in UTC
    - [x] Implement `POST /trips/{id}/events` handler
        - [x] Parse form data
        - [x] Call Service
        - [x] Return HTML fragment (Event Card) or OOB Swap
- [x] **Frontend Implementation (Templ/HTMX)**
    - [x] Create `internal/features/timeline/components.templ`:
        - [x] `EventCard(event Event)` component
        - [x] Height calculation logic: `style="height: { DurationHours * 64 }px"`
    - [x] Update `internal/features/timeline/view.templ`:
        - [x] Add "Add Event" Form (Title, Address, Category Select, Start, End)
        - [x] Add Event List container
    - [x] **UX Detail**: Ensure date inputs handle timezone offsets correctly (send ISO/UTC to backend or handle conversion)

## Dev Notes

- **Timezones**: This is the trickiest part. The browser sends local time in `<input type="datetime-local">`.
    - **Strategy**: 
        1. Receive local time string from input.
        2. Parse it as "Trip Local Time" (for now, assume User Local Time = Trip Time, or just store as UTC).
        3. *Better*: Store as TIMESTAMPTZ.
- **Visuals**:
    - 64px/hour means a 15 min event is 16px high. Ensure text fits or truncates.
    - Category can be a simple `<select>`: "Activity", "Lodging", "Food", "Transit", "Other".
- **HTMX**:
    - On form submit, return the *newly created event card* and append it to the list `hx-swap="beforeend" hx-target="#event-list"`.
    - OR re-render the whole list `hx-swap="innerHTML" hx-target="#timeline-container"`. Re-rendering the list is safer for sorting (Story 1.4 requires order, but for now just time-based sort).
    - **Decision**: Re-render the list sorted by Start Time.

### Technical Requirements

- **Language:** Go 1.25
- **Database:** Postgres
- **Frontend:** Templ + HTMX + Tailwind v4
- **Validation:** Server-side validation is mandatory. Return 422 with HTML error snippet if validation fails.

### Architecture Compliance

- **Feature Folders:** Keep everything in `internal/features/timeline`.
- **Naming:** `category` in DB, `Category` in Struct.
- **Future Proofing:** Add `geo_lat`/`geo_lng` now to avoid schema changes in Story 2.1, even if unused.

### Library/Framework Requirements

- **Tailwind v4:** Use arbitrary values if needed for height `h-[64px]` or dynamic inline styles for variable height. Inline style is better for calculated height: `style={ fmt.Sprintf("height: %dpx", height) }`.

### File Structure Requirements

```bash
traccia/
├── internal/features/timeline/
│   ├── models.go        # Update Event struct
│   ├── service.go       # Add CreateEvent
│   ├── handler.go       # Add POST event handler
│   ├── view.templ       # Update with Form + List
│   └── components.templ # NEW: EventCard component
├── migrations/
│   └── YYYYMMDDHHMMSS_add_event_details.sql # NEW
```

## Previous Story Intelligence

- **From Story 1.2:**
    - The `events` table was created but minimal. You likely need to `ALTER TABLE` to add `category`, `geo_lat`, `geo_lng`.
    - `ResetTrip` implementation in Story 1.2 deletes events. Ensure this still works (Cascading delete or manual).

## Git Intelligence Summary

- **Recent Activity:** Story 1.2 completed basic Trip CRUD.
- **Configuration:** Tailwind v4 is active.

## Latest Tech Information

- **Templ**: Ensure `internal/features/timeline/components.templ` is compiled. Run `templ generate`.
- **Tailwind**: Dynamic height via inline styles is standard for data-driven visualization.

## Project Context Reference

- [Epics: Story 1.3](_bmad-output/planning-artifacts/epics.md#story-13-event-creation--timeline-view)
- [Architecture: Feature Folders](_bmad-output/planning-artifacts/architecture.md#structure-patterns)

## Story Completion Status

- **Status:** ready-for-dev
- **Validation:** Ready for `dev-story`.

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

- Completed Database Schema Migration (added category, geo_lat, geo_lng to events table)
- Validated migration with schema_test.go using Testcontainers
- Updated Event struct with new fields
- Implemented CreateEvent service method with validation and UTC enforcement
- Added unit tests for CreateEvent and validation
- Implemented POST /trips/{id}/events handler
- Created basic EventCard component for handler integration
- Updated view.templ to include Add Event form and Event List
- Implemented GetEvents to populate the view with existing events
- Verified EventCard height calculation with unit tests

### File List

- migrations/000002_add_event_details.up.sql
- migrations/000002_add_event_details.down.sql
- internal/features/timeline/schema_test.go
- internal/features/timeline/models.go
- internal/features/timeline/service.go
- internal/features/timeline/service_test.go
- internal/features/timeline/models_test.go
- internal/features/timeline/handler.go
- internal/features/timeline/handler_test.go
- internal/features/timeline/components.templ
- internal/features/timeline/components_templ.go
- internal/features/timeline/components_test.go
- internal/features/timeline/view.templ
- internal/features/timeline/view_templ.go
