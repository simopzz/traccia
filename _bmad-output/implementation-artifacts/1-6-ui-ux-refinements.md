# Story 1.6: ui-ux-refinements
Status: ready-for-dev

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

## Tasks / Subtasks
- [ ] **Event Card Refactor**
    - [ ] Add time display to `EventCard` component.
    - [ ] Apply `min-h` and `max-h` classes (Tailwind).
    - [ ] Test with 15m and 6h events.
- [ ] **Smart Defaults (Alpine.js)**
    - [ ] Pass Trip Start Date to the View context.
    - [ ] Use Alpine `x-init` or simple JS to set the `value` of the date inputs on load/modal open.

## Dev Notes
- For min-height: Ensure the visual timeline still roughly correlates to time, but prioritize readability. The "64px/hr" rule can be "soft" for <30m events.
- For Smart Defaults: `input type="datetime-local"` requires `YYYY-MM-DDTHH:mm` format.
