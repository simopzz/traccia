# Story 1.2: Activity & Food Event Creation

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a traveler,
I want to add Activity and Food events to specific days in my trip with smart time defaults,
so that I can build the shape of each day quickly without filling every field from scratch.

## Acceptance Criteria

1. **Given** a user is viewing a trip day, **When** they click "Add Event", **Then** a Sheet panel opens from the right (desktop) or bottom (mobile) with a TypeSelector showing 5 type icons and shared form fields.

2. **Given** the Sheet panel is open, **When** the user selects Activity or Food from the TypeSelector, **Then** the form displays shared fields: name (required), location/address, start time, end time, notes, and pinned toggle. No type-specific fields appear for Activity or Food.

3. **Given** the user is adding an event to a day with existing events, **When** the form loads, **Then** start time is pre-filled from the preceding event's end time, **And** end time is calculated from type-based duration (Activity ~2hr, Food ~1.5hr).

4. **Given** the user is adding the first event of a day, **When** the form loads, **Then** start time defaults to 9:00 AM, **And** end time defaults to start time + type-based duration.

5. **Given** the user submits a valid event form, **When** the server processes the request, **Then** the event appears in the timeline grouped by day in chronological order, **And** the event card shows type icon, name, time range, and location, **And** the day's HTML is replaced via HTMX day-level swap (`outerHTML` on `#day-{date}`).

6. **Given** the user toggles the pinned switch during creation, **When** the event is saved, **Then** the event displays a lock icon in the card header, **And** the drag handle is hidden.

7. **Given** the user submits a form with missing required fields, **When** validation fails, **Then** field-level error messages appear with rose borders on invalid fields, **And** all other entered values are preserved, **And** the Sheet panel stays open.

## Tasks / Subtasks

### Database & Schema (no changes needed)

- [x] Task 0: Verify schema readiness (AC: all)
  - [x] 0.1 Confirm `events` table has all required columns: `title`, `category`, `location`, `latitude`, `longitude`, `start_time`, `end_time`, `event_date`, `position`, `notes`, `pinned` — already done in Story 1.1
  - [x] 0.2 Confirm sqlc queries exist: `CreateEvent`, `ListEventsByTripAndDate`, `GetMaxPositionByTripAndDate`, `GetLastEventByTrip` — already done in Story 1.1
  - [x] 0.3 No new migrations needed. Activity and Food use the base events table only (no detail tables).

### Service Layer Enhancements

- [x] Task 1: Implement smart time defaults (AC: #3, #4)
  - [x] 1.1 Add `SuggestDefaults(ctx, tripID, eventDate, category)` method to `EventService` that returns suggested start time and end time based on:
    - Call `ListByTripAndDate(ctx, tripID, eventDate)` and take the last element's EndTime as the preceding event's end time, or 9:00 AM if the list is empty (no new SQL query needed)
    - Type-based duration: Activity = 2hr, Food = 1.5hr (applied to calculate end time from start time)
  - [x] 1.2 Deprecate or remove existing `SuggestStartTime()` — it queries trip-wide which is wrong for per-day defaults. Replace all usages with `SuggestDefaults()`
  - [x] 1.3 Add type-based duration constants: `DefaultActivityDuration = 2 * time.Hour`, `DefaultFoodDuration = 90 * time.Minute`

### Handler Layer — Sheet Panel & Event Creation

- [x] Task 2: Update event creation handler for HTMX Sheet delivery (AC: #1, #3, #4, #5, #7)
  - [x] 2.1 Update `NewPage` handler (GET `/trips/{tripID}/events/new?date={date}`) to:
    - Accept `date` query param to scope smart defaults to the specific day
    - Accept `category` query param to pre-select type (optional, defaults to `activity`)
    - Call `SuggestDefaults()` to populate start/end time
    - **Dual-path response:** If request has `HX-Request` header → return Sheet HTML fragment for HTMX. Otherwise (direct URL navigation) → return full page with layout as fallback. This keeps both paths working from one handler.
  - [x] 2.2 Update `Create` handler (POST `/trips/{tripID}/events`) to:
    - Parse form: `date` hidden field (format `2006-01-02`), `start_time` and `end_time` as `HH:MM` from time inputs, combine into full `time.Time` via `time.Parse("2006-01-02 15:04", date + " " + timeVal)`
    - Parse remaining fields: title, category, location, notes, pinned checkbox
    - On validation error (422): return Sheet form HTML with field-level errors, preserving all entered values. The form's default `hx-target` is the Sheet, so 422 responses naturally swap into the Sheet.
    - On success: return full day HTML for the created event's `event_date`. Use `HX-Retarget: #day-{eventDate}` and `HX-Reswap: outerHTML` response headers to redirect the swap from the Sheet target to the day container. Also send `HX-Trigger: {"closeSheet": true}` to close the Sheet.
  - [x] 2.3 Add unified `EventFormData` struct used for BOTH initial render and error re-render:
    ```go
    type EventFormData struct {
        TripID    int
        Date      string            // "2006-01-02"
        Category  string            // pre-selected type
        Title     string            // empty on first render, submitted value on error
        Location  string
        StartTime string            // "HH:MM" — smart default or submitted value
        EndTime   string            // "HH:MM" — smart default or submitted value
        Notes     string
        Pinned    bool
        Errors    map[string]string // nil on first render, populated on error
    }
    ```
    The template reads from this ONE struct regardless of render path. On initial render: fields have smart defaults, Errors is nil. On error re-render: fields have submitted values, Errors has field-level messages. This prevents the bug where validation errors clear form values.

### Template Layer — TypeSelector, Sheet Form, EventCard

- [x] Task 3: Create TypeSelector component (AC: #1, #2)
  - [x] 3.1 Create `TypeSelector` templ component in `internal/handler/event_form.templ`:
    - Horizontal row of 5 icon buttons (Activity, Food, Lodging, Transit, Flight)
    - `role="radiogroup"` with `role="radio"` on each option
    - Alpine.js `x-data` for selected state, `x-on:click` to set type
    - Selected state: teal ring/border. **Default: Activity pre-selected** (eliminates empty form dead state — shared fields are visible immediately on Sheet open)
    - Arrow key navigation for accessibility
    - On selection: updates hidden `category` input, triggers form field morphing
  - [x] 3.2 Only Activity and Food are selectable in this story. Lodging, Transit, and Flight icons are **visible but disabled** (grayed out, `opacity-40`, `cursor-not-allowed`, `aria-disabled="true"`). Clicking them does nothing. Tooltip on hover: "Coming in a future update". This prevents creating events without their required detail fields (Stories 1.4-1.6 will enable them).

- [x] Task 4: Create Sheet-based event creation form (AC: #1, #2, #3, #4, #7)
  - [x] 4.1 Create `EventCreateSheet` templ component in `internal/handler/event_form.templ`:
    - Uses templui Sheet component (slides from right on desktop, bottom on mobile)
    - Contains TypeSelector at top
    - Shared form fields: title (required), location, start_time, end_time, notes (textarea), pinned (templui Switch component — not checkbox. Label: "Pin this event". Check Switch's form submission behavior in `internal/components/` — it likely uses a hidden input)
    - Alpine.js `x-show` for future type-specific fields (empty for Activity/Food)
    - `hx-post="/trips/{tripID}/events"` with `hx-target="#sheet-form"` and `hx-swap="innerHTML"` (targets Sheet by default; success responses use `HX-Retarget`/`HX-Reswap` headers to redirect to day container)
    - Hidden field `date` carrying the day context (format `2006-01-02`)
    - Time inputs use `type="time"` HTML5 inputs (send `HH:MM` values; handler combines with `date` hidden field into full `time.Time`)
    - Pre-populated with smart defaults (start time, end time from handler)
  - [x] 4.2 Field-level validation display: rose border (`border-rose-500`) + inline error text below field
  - [x] 4.3 Form preserves all values on validation error (server re-renders form with submitted values + errors)
  - [x] 4.4 Add "Create Event" (primary teal) and "Cancel" (ghost) buttons. Cancel closes Sheet without submission via templui Sheet's built-in close mechanism (Alpine.js `@click` that sets Sheet's `open` state to false — check templui Sheet component API). No HTMX call needed for cancel.

- [x] Task 5: Update EventCard for timeline display (AC: #5, #6)
  - [x] 5.1 Update `EventTimelineItem` in `internal/handler/event.templ` to match UX spec:
    - Card with 1px solid border, 2px hard shadow (Swiss bordered style)
    - Type icon with colored background on left (Activity=teal, Food=amber)
    - Title, time range ("9:00 AM – 11:00 AM"), location
    - Lock icon if pinned (replaces text badge), drag handle hidden when pinned
    - Semantic HTML: `<li>` with `aria-label` combining event name and time
  - [x] 5.2 Wrap EventCard in a **Collapsible structure** (templui or custom) — collapsed view shows the summary (icon, title, time, location), expanded view is empty/read-only for now. Story 1.3 will fill the expanded section with inline edit fields. Building the Collapsible wrapper now prevents rework in 1.3.
  - [x] 5.3 Ensure the EventCard integrates properly within `TimelineDay` component from Story 1.1. **Critical:** Read the existing `TimelineDay` and `EventTimelineItem` templates before building. The timeline spine CSS wraps AROUND the card — the spine connector is in the outer wrapper, not inside the card. The EventCard redesign (Swiss borders, Collapsible) changes the card interior only. Preserve the spine wrapper structure: `TimelineDay > spine-wrapper > EventCard`.

- [x] Task 6: Wire "Add Event" button in EmptyDayPrompt and day headers (AC: #1)
  - [x] 6.1 Update `EmptyDayPrompt` from Story 1.1 — wire "Add Event" button to fetch Sheet via HTMX:
    `hx-get="/trips/{tripID}/events/new?date={date}"` `hx-target="#sheet-container"` `hx-swap="innerHTML"`
  - [x] 6.2 Add "Add Event" button to each day header (visible even when day has events) with same HTMX attributes
  - [x] 6.3 Add Sheet structure to `TripDetailPage` in `trip.templ` (NOT in `layout.templ`). **Critical:** The templui Sheet **wrapper** (with Alpine.js `x-data` for open/close state) must be rendered **statically** in the page HTML. Only the Sheet **content** (`#sheet-form` inside the wrapper) is loaded dynamically via HTMX. If the entire Sheet (wrapper + content) is loaded via HTMX, Alpine.js won't initialize the open/close state and HTMX won't process the inner form's `hx-post` attributes. Structure:
    ```
    <div id="sheet-container">              ← static in TripDetailPage
      <Sheet x-data="{ open: false }">     ← static templui wrapper, Alpine manages open/close
        <div id="sheet-form">              ← HTMX swap target, content loaded dynamically
          <!-- form loaded here via hx-get -->
        </div>
      </Sheet>
    </div>
    ```

### HTMX Integration

- [x] Task 7: Wire HTMX day-level swaps for event creation (AC: #5)
  - [x] 7.1 After successful event creation, handler returns the full day HTML with `id="day-{eventDate}"` containing all events for that date. Use the `event_date` from the **created event** (as computed by the service layer), not the form's original date param.
  - [x] 7.2 The form's `hx-target` points to `#sheet-form` (Sheet area) by default. On **success**, handler sets response headers `HX-Retarget: #day-{eventDate}` + `HX-Reswap: outerHTML` to redirect the swap to the correct day container. This eliminates the 422-vs-200 target confusion — errors naturally stay in Sheet, success redirects to day.
  - [x] 7.3 On success, also send `HX-Trigger: {"closeSheet": true}` response header. Add Alpine.js listener on `#sheet-container` that listens for `closeSheet` custom event and closes the templui Sheet.
  - [x] 7.4 On validation error (422), return the form HTML only (stays in Sheet due to default `hx-target`). No retarget headers needed.

### Testing

- [x] Task 8: Write tests (AC: all)
  - [x] 8.1 Service tests: `SuggestDefaults` — first event of day returns 9:00 AM + type duration, subsequent event returns prev end time + type duration
  - [x] 8.2 Service tests: `Create` with Activity category — valid input, missing title, missing times
  - [x] 8.3 Service tests: `Create` with Food category — valid input, verify EventDate derived from StartTime
  - [x] 8.4 Service tests: position assignment — first event gets 1000, second gets 2000
  - [x] 8.5 Run `just test` and `just lint` — all passing, zero violations
  - [x] 8.6 Run `just build` to verify compilation succeeds (catches DI wiring issues in `main.go` that mock-based tests miss)
  - [x] 8.7 Run `just generate` followed by `just dev` and verify the app starts without errors (smoke test: navigate to a trip, click "Add Event", create an Activity event, verify it appears in the timeline)

## Dev Notes

### Critical: Sheet Panel, Not Full Page

Story 1.1 used full-page navigation for event forms. Story 1.2 **MUST change this** to use a Sheet panel (templui Sheet component). The "Add Event" button triggers an HTMX GET that loads the form into a Sheet container. The form submits via HTMX POST and the response swaps the day's HTML. The Sheet closes on success. This is the pattern for ALL future event creation (Stories 1.3-1.6).

### Critical: Day-Level HTMX Swap Pattern

After creating an event, the server returns the **full day HTML** (all events for that date). The response replaces `#day-{date}` via `outerHTML`. This is the architecture-mandated swap strategy — no partial card inserts. The handler must:
1. Create the event via service
2. Re-fetch all events for that date via `ListByTripAndDate`
3. Render the full `TimelineDay` component
4. Return it as the HTMX response

### Critical: TypeSelector Is Alpine.js, Not Server Round-Trip

The TypeSelector uses Alpine.js `x-show` to toggle form sections. All 5 form variants exist in the DOM (even though Activity/Food share the same fields). Type selection is instant client-side — no HTMX fetch for form morphing. The hidden `category` input is updated by Alpine.js when a type is selected.

### Activity & Food Have No Detail Tables

Activity and Food events use the base `events` table only. There are NO `activity_details` or `food_details` tables. The form only shows shared fields (title, location, start/end time, notes, pinned). Type-specific detail tables are used by Flight (1.4), Lodging (1.5), and Transit (1.6).

### Smart Defaults Logic

```
suggestDefaults(tripID, eventDate, category):
  events = ListByTripAndDate(tripID, eventDate)  // ordered by position ASC
  if len(events) > 0:
    // Find the event with the LATEST EndTime, not last-by-position.
    // Position order ≠ chronological order (user could create events out of order).
    latestEnd = max(event.EndTime for event in events)
    startTime = latestEnd
  else:
    startTime = 9:00 AM on eventDate

  duration = durationForCategory(category)  // Activity=2h, Food=1.5h
  endTime = startTime + duration

  return { startTime, endTime }
```

**No new SQL query needed.** Use existing `ListByTripAndDate`, iterate to find max EndTime. Do NOT just take `events[last].EndTime` — position order doesn't guarantee chronological order. The existing `SuggestStartTime()` queries trip-wide (wrong for per-day defaults) — deprecate it and replace with `SuggestDefaults()`.

### Position Assignment

New events are appended at `max(position on that date) + 1000`. The `GetMaxPositionByTripAndDate` query already exists in `events.sql`. First event gets position 1000. This is already implemented in `event_store.go` Create method.

### Existing Event Handler Refactoring — Dual-Path

The current `event.go` handler uses full-page rendering (`EventNewPage`, `EventEditPage`). For Story 1.2:
- `NewPage` handler detects `HX-Request` header: if present → return Sheet HTML fragment; if absent → return full page with layout (fallback for direct URL access)
- `Create` handler changes to return day HTML with retarget headers (not a redirect)
- The existing full-page event templates (`EventNewPage`, `EventEditPage`) are **kept** as the fallback path. The Sheet templates are the primary path triggered by HTMX.
- Edit functionality stays as-is until Story 1.3

### End Time Default — No Client-Side Recalculation

End time smart default is calculated **server-side on Sheet open** based on the initial category (Activity=2h, Food=1.5h). If the user switches type in the TypeSelector (Activity → Food), the end time does NOT recalculate automatically. Users can manually adjust. This avoids duplicating duration constants in JavaScript and keeps the implementation simple. Can be enhanced later if user feedback demands it.

### EventCard Collapsible Structure — Future-Proofing for Story 1.3

Build EventCard with a Collapsible wrapper now. The collapsed view (default in 1.2) shows the summary: type icon, title, time range, location. The expanded section is empty or shows read-only details (notes, pinned status). Story 1.3 will fill the expanded section with inline edit fields. Building without the Collapsible structure would require refactoring in 1.3.

### Pinned Toggle — Use templui Switch, Not Checkbox

The pinned toggle uses templui Switch component (not a plain checkbox). Label: "Pin this event". Check the installed Switch component in `internal/components/` for its form submission behavior — it likely uses a hidden input with on/off or true/false values. The handler must parse accordingly (may differ from Story 1.1's checkbox parsing).

### Form Validation Pattern — Hybrid Approach

Field-level error display requires mapping errors to specific fields. The current service returns a single `domain.ErrInvalidInput` with a message string. Use a **hybrid approach**:

1. **Handler pre-validates** required fields (title non-empty, start_time present, end_time present) and builds `FormErrors map[string]string` with field-name → error-message mappings
2. If all fields present, call service for **business rule validation** (end >= start, valid category, etc.)
3. If service returns `ErrInvalidInput`, parse the message to map to the most likely field (e.g., message contains "end time" → `end_time` field)
4. Handler builds `FormData` struct with submitted values + `FormErrors` map
5. Handler re-renders Sheet form template with errors (HTTP 422)
6. Template iterates `FormErrors` — rose border on fields with errors, inline error text via `aria-describedby`
7. All other field values preserved from `FormData`

This gives field-level display without refactoring the service error pattern from Story 1.1.

### Critical: HTMX Retarget Pattern (422 vs 200)

The form's `hx-target` always points to `#sheet-form` (the Sheet content area). This solves the 422-vs-200 targeting problem:

**On success (200):**
- Response body: full day HTML for `#day-{eventDate}`
- Response headers: `HX-Retarget: #day-{eventDate}`, `HX-Reswap: outerHTML`, `HX-Trigger: {"closeSheet": true}`
- HTMX follows `HX-Retarget` and swaps day container instead of Sheet
- Alpine.js listener on Sheet detects `closeSheet` and closes panel

**On validation error (422):**
- Response body: Sheet form HTML with errors
- No retarget headers — HTMX swaps into `#sheet-form` naturally
- Sheet stays open, form shows field-level errors

**Do NOT use `hx-target-422`** — it's not a real HTMX attribute. The retarget-on-success pattern is the correct approach.

### Sheet Close Mechanism

**Cancel:** Closes the Sheet via templui's built-in Alpine.js close mechanism (e.g., `@click="open = false"` — check the installed templui Sheet component API in `internal/components/sheet/sheet.templ`). No HTMX call needed for cancel.

**Success close — order of operations matters:** The day HTML must swap BEFORE the Sheet closes. If the Sheet closes first (removing its DOM), the HTMX retarget swap may fail. Two safe approaches:

**Option A (preferred):** Use `hx-on::after-swap` attribute on the form element to close the Sheet AFTER the swap completes:
```html
hx-on::after-swap="if(event.detail.xhr.status === 200) { $dispatch('close-sheet') }"
```
This guarantees: retarget → swap day HTML → then close Sheet.

**Option B (fallback):** Use `HX-Trigger: {"closeSheet": true}` response header. HTMX processes triggers after swaps, so the order should be correct. But verify this in testing — if the Sheet's DOM removal interferes with HTMX lifecycle, switch to Option A.

Alpine.js listener on Sheet wrapper:
```
x-on:close-sheet.window="open = false"
```

### Critical: Form Date + Time Parsing

The Sheet form sends separate `date` (hidden, `2006-01-02`) and `start_time`/`end_time` (`HH:MM` from `type="time"` inputs). The handler must combine them:

```go
func parseDateAndTime(dateStr, timeStr string) (time.Time, error) {
    return time.Parse("2006-01-02 15:04", dateStr + " " + timeStr)
}
```

This replaces Story 1.1's `parseDateTime()` which parsed a single datetime string. The `date` hidden field carries the day context from when the Sheet was opened.

### Critical: Disabled Types in TypeSelector

Lodging, Transit, and Flight are **disabled** in the TypeSelector (grayed out, non-interactive). This prevents creating events without their required detail table rows. Stories 1.4-1.6 will enable each type when its detail form and transactional creation are implemented. Do NOT allow creating base-only events for types that require detail tables.

### Previous Story (1.1) Intelligence

Key patterns established in Story 1.1 that MUST be followed:
- **pgtype helpers**: Use `toPgDate()`, `toPgTimestamptz()`, `toPgText()`, `toPgFloat8()`, `toPgBool()` from `internal/repository/helpers.go` for all pgx type conversions
- **EventDate derivation**: Use `time.Date(y, m, d, 0, 0, 0, 0, startTime.Location())` to extract date from StartTime (NOT `Truncate(24*time.Hour)`)
- **Service validation**: All input validation in service layer. Handler only parses HTTP form → service input struct.
- **Error mapping**: `domain.ErrInvalidInput` → 422, `domain.ErrNotFound` → 404, else → 500
- **Method override**: Forms use `_method=PUT|DELETE` hidden field for PUT/DELETE via POST
- **Template style**: Tailwind CSS with Swiss bordered cards, teal brand color, Inter font
- **Delete returning rows**: `:execrows` in sqlc, check rows affected in store

### Accessibility Requirements

- TypeSelector: `role="radiogroup"`, `role="radio"` per option, arrow key navigation
- EventCard: `<li>` with `aria-label` = "{Type}: {Name}, {StartTime} to {EndTime}"
- Sheet: focus trapped when open, Escape closes, focus returns to trigger
- Form fields: labels explicitly associated with inputs via `for`/`id`
- Error messages: associated with fields via `aria-describedby`
- `aria-live="polite"` on day container for HTMX updates
- All interactive elements >= 44x44px touch targets

### Project Structure Notes

- All changes follow the established layered architecture: `handler/` → `service/` → `domain/` ← `repository/`
- New template file: `internal/handler/event_form.templ` for TypeSelector + EventCreateSheet (matches architecture.md)
- No new directories needed
- Run `just generate` after creating/modifying `.templ` files
- Run `just css` or let `just dev` handle Tailwind rebuild after template changes

### References

- [Source: _bmad-output/planning-artifacts/architecture.md#Data Architecture] — Activity and Food use base table only, no detail tables
- [Source: _bmad-output/planning-artifacts/architecture.md#Communication Patterns] — HTMX interaction contract: POST `/trips/{id}/events` → full day HTML swap
- [Source: _bmad-output/planning-artifacts/architecture.md#Structure Patterns] — Event type code organization, templ component hierarchy
- [Source: _bmad-output/planning-artifacts/architecture.md#Process Patterns] — Position management: NewEventPosition = max + 1000
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md#Event Creation] — Sheet panel, TypeSelector, smart defaults
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md#EventCard] — Card anatomy, type icons, pinned lock icon
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md#Form Patterns] — Field-level validation, rose borders, value preservation
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md#Accessibility] — radiogroup, aria-label, focus management
- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.2] — Acceptance criteria, BDD scenarios
- [Source: _bmad-output/implementation-artifacts/1-1-trip-crud-and-timeline-shell.md] — Previous story patterns, pgtype helpers, validation approach

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No blocking issues encountered.

### Completion Notes List

- Implemented `SuggestDefaults()` service method with per-day smart time defaults (9:00 AM first event, latest EndTime for subsequent). Finds max EndTime across all events on the day, not just last-by-position.
- Added `DefaultActivityDuration` (2h) and `DefaultFoodDuration` (90min) constants.
- Removed `SuggestStartTime()` (trip-wide, wrong granularity) and replaced with `SuggestDefaults()`.
- Updated `NewPage` handler with dual-path: HTMX requests get Sheet form fragment, direct URL gets full page fallback.
- Updated `Create` handler with: separate date+time parsing via `parseDateAndTime()`, handler-level field validation for field-level errors, HTMX retarget pattern (422→Sheet, 200→day container), `HX-Trigger: closeSheet`.
- Added `EventFormData` struct shared by initial render and error re-render paths.
- Created `event_form.templ` with `TypeSelector` (5 icons, Activity/Food enabled, Lodging/Transit/Flight disabled with tooltip) and `EventCreateForm` (Sheet-based form with all fields, CSS switch toggle for pinned, field-level error display with rose borders).
- Redesigned `EventTimelineItem` as `<li>` with Collapsible structure (Alpine.js x-collapse), type icon with colored background, lock icon for pinned, drag handle hidden when pinned, `aria-label`.
- Updated `TripDetailPage` with static Sheet container (`EventSheet` component using templui Sheet). Sheet opens via trigger click + HTMX content load.
- Updated `TimelineDay` with "Add Event" button in day header and `<ul>` wrapper for events.
- Updated `EmptyDayPrompt` from `<a>` to `<button>` with HTMX GET to load Sheet form.
- No Switch component installed; used custom CSS switch (sr-only checkbox + styled spans).
- **Course correction**: Replaced templui Sheet component with custom Alpine.js sheet panel. The templui Sheet/Dialog system relies on `tailwind-merge-go` which does not support Tailwind v4, causing class override failures (conflicting positioning classes pile up instead of being resolved). Custom implementation uses `x-data`, `x-show`, `x-transition` for open/close with backdrop, and `$dispatch('open-sheet')`/`$dispatch('close-sheet')` events for triggers. Removed `dialog.min.js` from layout (no longer needed).
- **Future action item**: Replace all remaining templui components with Tailwind v4-compatible alternatives or plain Alpine.js implementations. The `tailwind-merge-go` incompatibility affects any templui component that relies on class overrides via `TwMerge`.
- 26 tests passing, 0 lint violations, build succeeds.

### File List

- internal/service/event.go (modified — added SuggestDefaults, EventDefaults, durationForCategory, duration constants; removed SuggestStartTime)
- internal/service/event_test.go (modified — added TestEventService_SuggestDefaults, TestEventService_Create_ActivityCategory, TestEventService_Create_FoodCategory, TestEventService_Create_PositionAssignment)
- internal/handler/event.go (modified — added EventFormData, dual-path NewPage, HTMX Create with retarget/reswap/closeSheet, field-level validation)
- internal/handler/event.templ (modified — redesigned EventTimelineItem with Collapsible, type icons, lock icon, aria-label, li element)
- internal/handler/event_form.templ (new — TypeSelector, EventCreateForm components)
- internal/handler/trip.templ (modified — TripDetailPage with custom Alpine.js sheet panel, TimelineDay with day header Add Event button and ul wrapper, EmptyDayPrompt with HTMX, $dispatch triggers)
- internal/handler/layout.templ (modified — removed dialog.min.js script tag, no longer needed)
- internal/handler/helpers.go (modified — added parseDateAndTime helper)

### Change Log

- 2026-02-16: Story 1.2 implementation complete — Activity & Food event creation via Sheet panel with TypeSelector, smart time defaults, HTMX day-level swaps, redesigned EventCard with Collapsible structure
