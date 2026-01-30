# Story 1.5: pinned-events-logic
Status: review

## Story
As a Planner,
I want to mark specific events as "Pinned" (Fixed Time),
So that other events flow around them during reordering instead of pushing them.

## Acceptance Criteria
1.  **Given** an event on the timeline
2.  **When** I click a "Pin" toggle/icon
3.  **Then** the event visual style changes (e.g., lock icon, border change) to indicate it is pinned
4.  **And** the `is_pinned` status is saved to the database
5.  **Given** a Pinned event exists at 14:00
6.  **When** I drag an unpinned event (duration 1h) to a slot before it
7.  **Then** the dragged event's end time must NOT push the Pinned event
    *   *Constraint:* If the new position would overlap or push the Pinned event, the reorder should either:
        *   (MVP) Allow the move but creating a "Conflict" (visual overlap) - *simpler*.
        *   (Ideal) Stop the ripple effect at the Pinned event.
    *   *Decision for MVP:* The "Ripple" calculation (Service Layer) must STOP updating subsequent events when it hits a Pinned event.
8.  **And** if I try to drag the Pinned event itself, it should behave normally (I can move it explicitly), but its new position is now the new "Anchor".

## Tasks / Subtasks
- [x] **Database Migration**
    - [x] Add `is_pinned` (BOOLEAN DEFAULT FALSE) to `events` table.
- [x] **Backend Implementation**
    - [x] Update `Event` struct.
    - [x] Update `ReorderEvents` service logic:
        - [x] Iterate through events.
        - [x] If `is_pinned` is true, use its existing/requested Start Time as a hard anchor.
        - [x] Ensure `Start Time` of subsequent events = `Max(Previous.End, Pinned.Start)`.
    - [x] Add `TogglePin` endpoint (or include in Update Event).
- [x] **Frontend Implementation**
    - [x] Add Pin/Unpin button to Event Card.
    - [x] Visual indicator for Pinned state.

## Dev Notes
- This is the foundation for Epic 2.
- The Reorder logic needs to be smarter than "Previous.End". It needs to look ahead.

## File List
- migrations/000003_add_is_pinned_to_events.up.sql
- migrations/000003_add_is_pinned_to_events.down.sql
- internal/features/timeline/schema_test.go
- internal/features/timeline/models.go
- internal/features/timeline/models_test.go
- internal/features/timeline/service.go
- internal/features/timeline/service_test.go
- internal/features/timeline/handler.go
- internal/features/timeline/components.templ
- _bmad-output/implementation-artifacts/sprint-status.yaml

## Dev Agent Record
### Implementation Notes
- **[Database Migration]** Added `is_pinned` column to `events` table via migration `000003`.
- **[Tests]** Updated `schema_test.go` to verify the existence of the new column.
- **[Backend]** Updated `Event` struct with `IsPinned`.
- **[Backend]** Updated `ReorderEvents` to respect pinned events (fixed time).
- **[Backend]** Implemented `TogglePin` service and handler.
- **[Frontend]** Updated `EventCard` to show Pin button and visual state.
- **[Tests]** Added tests for pinned reordering and pin toggling.
### Debug Log
- **[Issue]** User reported 500 error on Trip Creation (GetTrip). Cause: Missing `is_pinned` column in running DB.
- **[Fix]** Applied pending migrations (000003) to local database using `migrate` tool.
