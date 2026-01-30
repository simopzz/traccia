# Story 1.6: ui-ux-refinements
Status: done

## Story
As a Planner,
I want to see event details clearly and enter dates easily,
So that the planning experience is smooth and informative.

## Acceptance Criteria

### Event Card Visibility
1.  **Given** any event card
2.  **Then** it must strictly display the Start Time and End Time (e.g., "14:00 - 15:30").
3.  **Given** a very short event (e.g., 15 mins)
4.  **Then** the card must have a `min-height` sufficient to show the Title and Time (e.g., `min-h-[40px]`).
5.  **Given** a very long event (e.g., 5 hours)
6.  **Then** the card should not dominate the screen (e.g., `max-h-[300px]` with internal scroll or truncation).

### Smart Date Defaults
7.  **Given** I am creating a new Trip
8.  **When** I pick a Start Date
9.  **Then** the End Date picker should default to Start Date + 7 days (or at least Start Date).
10. **Given** I am adding an Event to a Trip
11. **Then** the Date Picker should default to the Trip's Start Date (not "Today").
12. **And** if previous events exist, it should ideally default to the day of the last event.
13. **Then** the End Time picker should default to Start Time + 1 hour.

## Tasks / Subtasks
- [x] **Event Card Refactor**
    - [x] Add time display to `EventCard` component.
    - [x] Apply `min-h` and `max-h` classes (Tailwind).
    - [x] Test with 15m and 6h events.
- [x] **Smart Defaults (Alpine.js)**
    - [x] Pass Trip Start Date to the View context.
    - [x] Use Alpine `x-init` or simple JS to set the `value` of the date inputs on load/modal open.
- [ ] **Review Follow-ups (AI)**
    - [x] **Smart Default End Time**: Set default End Time to Start Time + 1 hour.
    - [x] **UX Fix**: Remove scrollbar from short events (conditional overflow).
    - [x] **Fix Short Event Visibility**: Increase `min-height` to ensure Time is visible without scrolling.
    - [x] [AI-Review][High] Remove ignored `_templ.go` files from Story File List [1-6-ui-ux-refinements.md]
    - [x] [AI-Review][Medium] Fix Visual Timeline Distortion: Reduce `min-h` to 40px [internal/features/timeline/components.templ]
    - [x] [AI-Review][Medium] Fix Default Start Time: Use previous event EndTime or StartTime+1h [internal/features/timeline/view.templ]
    - [x] [AI-Review][Low] Investigate Timezone Fragility in Alpine.js defaults [internal/features/timeline/view.templ]

## Dev Notes
- For min-height: Ensure the visual timeline still roughly correlates to time, but prioritize readability. The "64px/hr" rule can be "soft" for <30m events.
- For Smart Defaults: `input type="datetime-local"` requires `YYYY-MM-DDTHH:mm` format.

## Dev Agent Record
### Implementation Notes
- **Event Card Refactor**:
    - Updated `EventCard` in `components.templ` to display "Start - End" time range.
    - Added CSS classes `min-h-[40px]`, `max-h-[300px]`, and `overflow-y-auto` to handle short and long events.
    - Verified with unit tests in `components_test.go` covering 15m and 6h scenarios.
- **Smart Defaults**:
    - **Trip Creation**: Added Alpine.js logic in `home.templ` to automatically set End Date to Start Date + 7 days when Start Date changes.
    - **Event Creation**: Updated `view.templ` to pre-fill the "Add Event" form's Start Time.
        - Defaults to Trip Start Date (9 AM) if no events exist.
        - Defaults to Last Event's Start Time if events exist.
    - Added tests in `home_test.go` and `view_test.go` to verify this logic.

## File List
- internal/features/timeline/components.templ
- internal/features/timeline/components_test.go
- internal/features/timeline/home.templ
- internal/features/timeline/home_test.go
- internal/features/timeline/view.templ
- internal/features/timeline/view_test.go

## Change Log
- 2026-01-30: Completed Event Card Refactor. Added time range display and height constraints.
- 2026-01-30: Implemented Smart Defaults for date pickers using Alpine.js and Templ logic.


