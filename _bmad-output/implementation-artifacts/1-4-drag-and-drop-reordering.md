# Story 1.4: drag-and-drop-reordering

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a Planner,
I want to drag event cards to new timeslots,
so that I can reschedule my day intuitively without manual time entry.

## Acceptance Criteria

1. **Given** multiple events on the timeline
2. **When** I drag an event card to a new position using the drag handle
3. **Then** the new visual order is instantly reflected (Alpine.js)
4. **And** a request is sent to the backend with the new sequence of Event IDs (HTMX)
5. **And** the backend recalculates the Start/End times of *all* affected events based on the new sequence ("Ripple Update")
    - *Logic:* Event[n].Start = Event[n-1].End + (Transit or Default Buffer)
6. **And** the timeline is re-rendered with the new correct times
7. **And** the visual order persists after page refresh

## Tasks / Subtasks

- [x] **Frontend Implementation (Alpine + Sortable)**
    - [x] Import `Sortable.js` (via CDN or bundled asset)
    - [x] Update `internal/features/timeline/view.templ` to include `Sortable.js` script
    - [x] Create an Alpine.js wrapper for Sortable behavior on the Event List container
    - [x] Add a visible "Drag Handle" icon to `EventCard` component
    - [x] Configure `hx-post` on the container to trigger on `end` event (reorder complete)
- [x] **Backend Implementation (Features/Timeline)**
    - [x] Implement `ReorderEvents(ctx, tripID, orderedEventIDs)` in `service.go`
        - [x] Fetch all events for the trip
        - [x] Re-sequence them based on input IDs
        - [x] **Ripple Logic:**
            - Keep 1st event's Start Time fixed (unless it was moved, then use Trip Start or stay same?) -> *Decision: Keep the Start Time of the event that is NOW first, or anchor to 08:00? MVP: Keep the Start Time of the event that became first.*
            - Recalculate subsequent events: `Start = Prev.End + Buffer` (Buffer = 0 for now, or 15m default)
        - [x] Update all modified events in DB transaction
    - [x] Implement `POST /trips/{id}/events/reorder` handler
        - [x] Parse `[]string` of IDs from form values
        - [x] Call Service
        - [x] Return re-rendered Timeline List
- [x] **Testing**
    - [x] Unit Test `ReorderEvents`: Verify times ripple correctly
    - [x] Integration Test: Verify handler accepts IDs and returns HTML

## Dev Notes

- **Sortable.js + HTMX Pattern**:
    - Use the standard pattern where the container is a `<form>` (or uses `hx-include`) and the items have `<input type="hidden" name="event_id" value="...">`.
    - When Sortable moves the DOM elements, the hidden inputs move too.
    - `hx-trigger="end"` (from Sortable) submits the new order of inputs.

    ```html
    <form id="timeline-list" 
          hx-post="/trips/{id}/events/reorder" 
          hx-trigger="end" 
          hx-target="#timeline-list" 
          hx-swap="outerHTML"
          x-data="{ 
              init() { 
                  new Sortable(this.$el, { 
                      handle: '.drag-handle',
                      animation: 150,
                      onEnd: function (evt) { 
                          // triggers htmx
                      }
                  }) 
              } 
          }">
        <!-- Items -->
    </form>
    ```

- **Ripple Logic Detail**:
    - This is critical for the "Constraints-Based" promise.
    - If I swap Event A (1 hour) and Event B (2 hours), the timeline shifts.
    - *Edge Case:* If the first event is moved to the middle, what becomes the new first event's start time?
        - *Rule:* The new first event inherits the *original first event's start time*.
        - *Example:*
            - Original: A (10:00-11:00), B (11:00-12:00)
            - Swap B to top.
            - New: B (10:00-11:00 *Wait, B is 2 hours? No duration is fixed* -> 10:00-12:00), A (12:00-13:00).

### Technical Requirements

- **Library:** `SortableJS` (Client-side), `Go` (Server logic).
- **State:** Alpine.js handles the visual drag; HTMX syncs the truth.
- **Database:** Bulk update (or multiple tx updates) for re-calculated times.

### Architecture Compliance

- **Feature Folder:** `internal/features/timeline`
- **Isolation:** The `ReorderEvents` logic belongs in `service.go`, not the handler.

### Library/Framework Requirements

- **SortableJS:** Load via unpkg or vendored in `web/assets/js`. Use the lightweight version if possible.

### File Structure Requirements

```bash
traccia/
├── internal/features/timeline/
│   ├── service.go       # Add ReorderEvents
│   ├── service_test.go  # Add TestReorderEvents
│   ├── handler.go       # Add POST reorder handler
│   ├── view.templ       # Add x-data sortable wrapper
│   └── components.templ # Add Drag Handle to card
├── web/assets/js/
│   └── sortable.min.js  # Add library
```

## Previous Story Intelligence

- **From Story 1.3:**
    - Events are currently sorted by Start Time in the `GetEvents` query.
    - `EventCard` height is dynamic (`Duration * 64px`). Reordering must preserve this visual duration while changing vertical position.
    - The `Event` struct already has everything needed.

## Git Intelligence Summary

- **Recent Activity:** Story 1.3 added Event Creation.
- **Pattern:** `features/timeline` pattern is established.

## Latest Tech Information

- **Alpine + Sortable:** The `x-init` pattern is the cleanest way to bind Sortable to the HTMX element.
- **HTMX:** Ensure `hx-disabled-elt` is used if the request takes time, to prevent double-drags.

## Project Context Reference

- [Epics: Story 1.4](_bmad-output/planning-artifacts/epics.md#story-14-drag-and-drop-reordering)
- [PRD: FR3](_bmad-output/planning-artifacts/prd.md#timeline-orchestration)

## Story Completion Status

- **Status:** review
- **Validation:** Ready for `dev-story`.

## Dev Agent Record

### Agent Model Used

opencode/gemini-3-pro

### Debug Log References

### Completion Notes List

- Implemented full stack Drag and Drop reordering using Alpine.js + Sortable.js + HTMX + Go.
- Added `sortable.min.js` to assets.
- **Fixed:** Added `alpine.min.js` to `base.templ` (was missing from project scaffolding).
- **Fixed:** Added `select-none` to drag handle to prevent text selection.
- Created `ReorderEvents` service method with "Ripple Logic" to recalculate times based on new order.
- Added drag handles to UI.
- Verified with TDD: Unit tests for Service logic and Handler integration tests.

### Senior Developer Review (AI) - Fixes Applied
- **Critical Fix:** Wrapped `ReorderEvents` in a transaction with `SELECT ... FOR UPDATE` to prevent race conditions during reordering.
- **Refactor:** Extracted `DefaultEventDuration` constant to avoid magic numbers.
- **Testing:** Added `TestReorderEvents_EdgeCases` to cover duplicate IDs, invalid IDs, and nil start times.
- **Security:** Sanitized error messages in `handleReorderEvents` to prevent leaking DB details.
- **Architecture:** Moved `sortable.min.js` to `base.templ` head for consistent asset loading.

### File List

- internal/features/timeline/components.templ
- internal/features/timeline/view.templ
- internal/features/timeline/handler.go
- internal/features/timeline/service.go
- internal/features/timeline/service_test.go
- internal/features/timeline/handler_test.go
- internal/features/timeline/view_test.go
- web/assets/js/sortable.min.js
- web/assets/js/alpine.min.js
- web/layouts/base.templ
