# Story 1.3: Event Edit, Delete & Detail View

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a traveler,
I want to edit event details, delete events, and expand cards to see full information,
so that I can refine my plan and access all details without leaving the timeline.

## Acceptance Criteria

1. **Given** a user clicks/taps on an event card in the timeline, **When** the card expands, **Then** full event details are shown (all shared attributes including notes), **And** the expansion uses progressive disclosure (Collapsible).

2. **Given** a user is viewing an expanded event card, **When** they edit a field (name, time, location, notes, pinned status), **Then** the change is saved and the day's timeline updates via HTMX swap.

3. **Given** a user deletes an event, **When** the deletion is processed, **Then** the event is removed from the timeline, **And** a toast appears with "Event removed. [Undo]" for 8 seconds, **And** the day's HTML updates via HTMX swap.

4. **Given** the undo toast is visible, **When** the user clicks "Undo" within 8 seconds, **Then** the event is restored to its original position in the timeline.

5. **Given** a user edits an event's start time, **When** the change is saved, **Then** the `event_date` is recalculated from the new start time by the service layer.

## Tasks / Subtasks

### Schema — Soft Delete

- [x] Task 1: Add soft delete column to events table (AC: #3, #4)
  - [x] 1.1 Create `migrations/002_soft_delete_events.up.sql`:
    ```sql
    ALTER TABLE events ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;
    ```
  - [x] 1.2 Create `migrations/002_soft_delete_events.down.sql`:
    ```sql
    ALTER TABLE events DROP COLUMN deleted_at;
    ```
  - [x] 1.3 Update `internal/repository/sql/events.sql` — add `AND deleted_at IS NULL` to every SELECT query that lists or fetches events (`ListEventsByTripAndDate`, `GetEventsByTrip`, `GetEventByID`, `GetMaxPositionByTripAndDate`, `GetLastEventByTrip`, `CountEventsByTrip` — all of them)
  - [x] 1.4 Add two new sqlc queries in `events.sql`:
    ```sql
    -- name: SoftDeleteEvent :exec
    UPDATE events SET deleted_at = NOW() WHERE id = $1;

    -- name: RestoreEvent :one
    UPDATE events SET deleted_at = NULL WHERE id = $1
    RETURNING *;
    ```
  - [x] 1.5 Run `just generate` to regenerate `internal/repository/sqlcgen/`
  - [x] 1.6 Run `just migrate-up` to apply the migration

### Repository & Service — Soft Delete + Restore

- [x] Task 2: Wire soft delete and restore through the stack (AC: #3, #4, #5)
  - [x] 2.1 Update `internal/domain/ports.go` — extend `EventRepository` interface:
    ```go
    Delete(ctx context.Context, id int) error         // now soft-deletes
    Restore(ctx context.Context, id int) (*Event, error)
    ```
  - [x] 2.2 Update `internal/repository/event_store.go`:
    - `Delete()` → calls `sqlcgen.SoftDeleteEvent(ctx, id)` instead of hard delete. Add a doc comment:
      ```go
      // Delete soft-deletes the event (sets deleted_at). Events are permanently removed
      // when their parent trip is deleted via ON DELETE CASCADE.
      func (s *EventStore) Delete(ctx context.Context, id int) error {
      ```
    - Add `Restore()` → calls `sqlcgen.RestoreEvent(ctx, id)`, maps result to `*domain.Event`
  - [x] 2.3 Update `internal/service/event.go`:
    - `Delete()` delegates to `repo.Delete()` as before — no service-layer changes needed (repo handles soft vs hard transparently)
    - Add `Restore(ctx context.Context, id int) (*domain.Event, error)` method that calls `repo.Restore(ctx, id)` — wrap with context: `fmt.Errorf("restoring event %d: %w", id, err)`
  - [x] 2.4 Verify `event_date` recalculation in the `Update` service method updater closure:
    ```go
    // When input.StartTime != nil, ALSO update EventDate:
    if input.StartTime != nil {
        e.StartTime = *input.StartTime
        st := *input.StartTime
        e.EventDate = time.Date(st.Year(), st.Month(), st.Day(), 0, 0, 0, 0, st.Location())
    }
    ```
    Use `time.Date(y, m, d, 0, 0, 0, 0, loc)` — NOT `Truncate(24*time.Hour)`. If this recalculation is missing from the existing service code, add it now. This is AC #5.

### Expanded Card — Detail View + Inline Edit

- [x] Task 3: Redesign EventTimelineItem for detail view and inline editing (AC: #1, #2)
  - [x] 3.1 Update Alpine.js state on the `<li>` element:
    ```
    x-data="{ expanded: false, editing: false }"
    ```
    Collapsing the card (`expanded = false`) also resets editing: bind `x-on:click` on the header button to `expanded = !expanded; if (!expanded) editing = false`
  - [x] 3.2 Expanded section — **View mode** (`x-show="expanded && !editing"`):
    - Full location (even if long — not truncated like collapsed header)
    - Notes: full text, or muted italic placeholder "No notes" if empty
    - Pinned status: small badge or text ("Pinned" / "Flexible")
    - **Edit button**: `type="button"` `x-on:click="editing = true"` — styled as teal secondary/outline
    - **Delete button**: HTMX delete (see Task 4 for handler-side); styled as rose text button
      ```
      hx-delete="/trips/{tripID}/events/{id}"
      hx-target="#day-{eventDate}"
      hx-swap="outerHTML"
      ```
  - [x] 3.3 Expanded section — **Edit mode** (`x-show="expanded && editing"`):
    - Inline form with shared fields: title (required), location, notes (textarea), start time (`type="time"`), end time (`type="time"`), pinned toggle (custom CSS switch, same as EventCreateForm)
    - Hidden date field — value depends on render path:
      - **Initial render** (`props == nil`): `value="{event.StartTime.Format("2006-01-02")}"` (pre-fill from event)
      - **422 re-render** (`props != nil`): `value="{props.FormValues.Date}"` (preserve submitted value — user may have changed the date by editing start_time)
      - Template conditional: `if props != nil { use props.FormValues.Date } else { use event.StartTime.Format("2006-01-02") }`
    - Note: the edit form fields are intentionally duplicated from `EventCreateForm` (not extracted to a shared component). The contexts differ enough (sheet vs. inline, different HTMX targets, no TypeSelector) that a shared `EventFormFields` component would require too many parameters. Extract as a shared component after Story 1.6 when the full field set is stable.
    - Form submission:
      ```
      hx-post="/trips/{tripID}/events/{id}"
      hx-target="#event-{id}"
      hx-swap="outerHTML"
      ```
      Hidden `_method=PUT` field for method override
    - **Save button**: `type="submit"` — teal primary
    - **Cancel button**: `type="button"` `x-on:click="editing = false"` — ghost style; does NOT send any request
    - Field-level error display: rose border + inline error text via `aria-describedby` (same pattern as EventCreateForm)
  - [x] 3.4 Use a **single parametric** `EventTimelineItem` — do NOT create a second component. Add an `EventCardProps` type and an optional `props` parameter:
    ```go
    // EventCardProps carries edit-mode state for 422 re-renders.
    // Nil means normal view-mode rendering (the common path from TimelineDay).
    type EventCardProps struct {
        Editing    bool
        FormValues EventFormData // submitted field values + Errors map
    }

    templ EventTimelineItem(event domain.Event, props *EventCardProps) { ... }
    ```
    - **Normal path** (called from `TimelineDay`): `@EventTimelineItem(event, nil)` — `props` is nil, component renders in view mode, `x-data="{ expanded: false, editing: false }"`
    - **422 error path** (called from `Update` handler on validation failure): `@EventTimelineItem(event, &EventCardProps{Editing: true, FormValues: formData})` — component renders with `x-data="{ expanded: true, editing: true }"` and form fields populated from `props.FormValues` instead of `event`
    - Template logic: `if props != nil && props.Editing` selects form field values from `props.FormValues`; otherwise reads directly from `event`
    - This eliminates duplicate markup — single component handles both states, maintained in one place
  - [x] 3.5 Verify `x-collapse` (Alpine.js Collapse plugin) is loaded in `layout.templ`. It was referenced in Story 1.2's `EventTimelineItem`. If missing, add the vendored or CDN version alongside `alpine.min.js`.

### Handler Updates — Update, Delete, Restore

- [x] Task 4: Update event handlers for HTMX (AC: #2, #3, #4, #5)
  - [x] 4.1 Update `Update` handler (PUT `/trips/{tripID}/events/{id}`):
    - **First**: fetch the event via `service.GetByID(ctx, id)` — needed for (a) capturing `oldEventDate` and (b) rendering the card shell on 422 error. Return 404 if not found.
    - Capture `oldEventDate := event.EventDate` immediately after fetch, before any mutation
    - Parse form: `date` (hidden, "2006-01-02") + `start_time`/`end_time` (HH:MM) combined via `parseDateAndTime()`. **Do NOT use the legacy `datetime-local` parsing from Story 1.1's `EditPage`.**
    - Apply handler-level required-field validation first; call service for business rules
    - **On validation error (422)**: render `EventTimelineItem(event, &EventCardProps{Editing: true, FormValues: formData})` — set `HX-Retarget: #event-{id}`, `HX-Reswap: outerHTML`, return HTTP 422
    - **On success, same day** (`newEventDate == oldEventDate`): fetch day events, render `TimelineDay` — set `HX-Retarget: #day-{newEventDate}`, `HX-Reswap: outerHTML`
    - **On success, cross-day** (`newEventDate != oldEventDate`): set `HX-Redirect: /trips/{tripIDStr}` header (HTMX full-page redirect). No OOB swap — full reload is the correct MVP response. OOB cross-day handling belongs to Story 2.2 (drag-and-drop cross-day moves).
    - Non-HTMX fallback: `http.Redirect` to `/trips/{tripID}` (keep for backward compatibility)
  - [x] 4.2 Update `Delete` handler (DELETE `/trips/{tripID}/events/{id}`):
    - Fetch event before deleting (need `EventDate`, `TripID` for response)
    - Call `service.Delete(ctx, id)` — now a soft delete
    - Fetch remaining events: `service.ListByTripAndDate(ctx, tripID, eventDate)`
    - Render `TimelineDay` (without the soft-deleted event, which is filtered by `deleted_at IS NULL`)
    - Set `HX-Trigger` header with structured undo data:
      ```
      HX-Trigger: {"showUndoToast": {"eventId": 42, "tripId": 1, "eventDate": "2026-03-10"}}
      ```
    - Response body is the day HTML (no HX-Retarget needed — the delete button's `hx-target` already points to `#day-{date}`)
    - Non-HTMX fallback: `http.Redirect` to `/trips/{tripID}`
  - [x] 4.3 Add `Restore` handler (POST `/trips/{tripID}/events/{id}/restore`):
    - Call `service.Restore(ctx, id)` — returns `*domain.Event` with tripID and eventDate
    - Fetch all events for the restored event's day (including the just-restored one — `deleted_at IS NULL`)
    - Render `TimelineDay` for the event's date
    - Set `HX-Retarget: #day-{eventDate}`, `HX-Reswap: outerHTML`
    - Set `HX-Trigger: {"hideUndoToast": true}` to dismiss the toast
    - Non-HTMX fallback: `http.Redirect` to `/trips/{tripID}`
  - [x] 4.4 Add restore route in `internal/handler/routes.go`:
    ```go
    r.Post("/trips/{tripID}/events/{id}/restore", eventHandler.Restore)
    ```
  - [x] 4.5 Keep the existing `EditPage` handler unchanged — it remains the non-HTMX fallback for direct URL access. No HTMX path needed since inline edit is always in the DOM.

### Undo Toast Component

- [x] Task 5: Add undo toast to TripDetailPage (AC: #3, #4)
  - [x] 5.1 Add the undo toast component inside `TripDetailPage` in `internal/handler/trip.templ`:
    ```html
    <div x-data="{
             showToast: false,
             eventId: null,
             tripId: null,
             eventDate: null,
             toastTimer: null
         }"
         x-on:showundotoast.window="
             eventId = $event.detail.eventId;
             tripId = $event.detail.tripId;
             eventDate = $event.detail.eventDate;
             showToast = true;
             clearTimeout(toastTimer);
             toastTimer = setTimeout(() => { showToast = false; }, 8000);
         "
         x-on:hideundotoast.window="showToast = false; clearTimeout(toastTimer);"
         x-show="showToast"
         x-transition:enter="transition ease-out duration-200"
         x-transition:enter-start="opacity-0 translate-y-2"
         x-transition:enter-end="opacity-100 translate-y-0"
         x-transition:leave="transition ease-in duration-150"
         x-transition:leave-start="opacity-100 translate-y-0"
         x-transition:leave-end="opacity-0 translate-y-2"
         class="fixed bottom-4 left-1/2 -translate-x-1/2 z-50
                flex items-center gap-3 px-4 py-3
                bg-slate-800 text-white rounded-lg shadow-lg
                text-sm font-medium whitespace-nowrap">
        <span>Event removed.</span>
        <button type="button"
                class="text-teal-400 hover:text-teal-300 font-semibold underline"
                x-on:click="
                    htmx.ajax('POST',
                        '/trips/' + tripId + '/events/' + eventId + '/restore',
                        { target: '#day-' + eventDate, swap: 'outerHTML' }
                    );
                    showToast = false;
                    clearTimeout(toastTimer);
                ">
            Undo
        </button>
    </div>
    ```
  - [x] 5.2 `HX-Trigger` header format: HTMX converts `{"showUndoToast": {...}}` to a custom DOM event `showUndoToast` that bubbles to `window`. Alpine's `x-on:showundotoast.window` (lowercase, HTMX lowercases event names) catches it. **Verify**: HTMX 2.0 fires triggers as `CustomEvent` with `detail` set to the value object. Test that `$event.detail.eventId` is accessible — if not, the value may be at `$event.detail.value.eventId` depending on HTMX version.
  - [x] 5.3 The `hideUndoToast` HX-Trigger from the Restore handler also lowercases to `hideundotoast`. The `x-on:hideundotoast.window` listener dismisses the toast cleanly after a successful undo.

### Testing

- [x] Task 6: Write/update tests (AC: all)
  - [x] 6.1 Service test: `Update` — `EventDate` recalculates when `StartTime` changes date:
    - Input: event with `EventDate = 2026-03-10T00:00`, `StartTime = 2026-03-10T09:00`
    - Update `StartTime` to `2026-03-11T10:00`
    - Expect: updated event `EventDate = 2026-03-11T00:00`
  - [x] 6.2 Service test: `Update` — `EventDate` unchanged when only `Title` updated
  - [x] 6.3 Service test: `Update` — returns `domain.ErrInvalidInput` when end time ≤ start time
  - [x] 6.4 Service test: `Update` — returns `domain.ErrNotFound` for non-existent ID
  - [x] 6.5 Service test: `Delete` + `Restore` round-trip — delete then restore, verify event returns from `ListByTripAndDate`
  - [x] 6.6 Service test: `Restore` — returns `domain.ErrNotFound` for a non-existent or already-hard-deleted ID
  - [x] 6.7 Service test: `ListByTripAndDate` — does NOT return soft-deleted events
  - [x] 6.8 Run `just test` — all tests passing, no races
  - [x] 6.9 Run `just lint` — zero violations
  - [x] 6.10 Run `just build` — binary compiles
  - [ ] 6.11 Manual smoke tests:
    - Expand event card → verify all details shown (location, notes, pinned, Edit + Delete buttons)
    - Click Edit → form appears in expanded section; change title → Save → day updates via HTMX, card collapses in new day HTML
    - Click Edit → submit invalid data → card stays open in edit mode with rose borders + error messages
    - Click Edit → change start time to different date → Save → full-page redirect to trip (both days shown correctly)
    - Click Cancel → editing mode closes, no request sent
    - Click Delete → day HTML swaps (event gone), undo toast appears
    - Wait 8 seconds → toast auto-hides (event remains soft-deleted)
    - Click Delete → toast appears → click Undo within 8s → event restored to original position, toast dismisses

### Bug Fixes

- [x] Task 7: Fix undo toast visibility bug (AC: #3)
  - [x] 7.1 Modify `Delete` handler to dispatch `showUndoToast` event to `window` via script injection, ensuring the event fires even if the triggering element is removed from the DOM during the HTMX swap.

## Dev Notes

### Bug Fix: Undo Toast Visibility

The undo toast was failing to appear because the `HX-Trigger` header dispatches events on the triggering element (the delete button). Since the delete button is inside the `TimelineDay` which is swapped out (replaced) by the response, the element is detached from the DOM before the event can bubble up to the `window` listener.

**Fix:** Instead of using the `HX-Trigger` header, the `Delete` handler now appends a `<script>` tag to the response body:
```html
<script>window.dispatchEvent(new CustomEvent('showundotoast', {detail: {...}}));</script>
```
HTMX executes this script after the swap, ensuring the event is dispatched directly to `window` regardless of the button's existence.

### Critical: Inline Edit, Not Sheet

The UX specification states: "Sheet panel for event creation. **Inline editing for modifications.**" Story 1.3 must honour this distinction. The expanded card section IS the edit form, toggled via Alpine.js — no HTMX GET request is needed to load an edit form. The edit form is always in the DOM, hidden by `x-show="editing"`.

**Consequence**: The `EditPage` handler (GET `/trips/{tripID}/events/{id}/edit`) is NOT used in the HTMX path for Story 1.3. It remains only as the fallback for direct URL access. Do NOT wire Edit button to HTMX GET.

### Critical: Soft Delete — Why, Not Deferred Timer

The undo requirement (AC #4: "event is restored to its **original position**") mandates server-side state. A client-side deferred timer approach has two correctness bugs:

1. If any HTMX day swap fires during the 8s window (e.g., user creates another event), the deleted card's `id` in the DOM is replaced. When the timer fires `htmx.ajax('DELETE', ...)`, the target `#day-{date}` may contain events including a new one — resulting in a confusing re-swap.
2. If the user navigates away, the timer is cancelled and the event is never deleted. The event silently survives.

Soft delete (set `deleted_at = NOW()`, filter in all queries) is correct: position is preserved in DB, undo is a single SQL `UPDATE events SET deleted_at = NULL`, and there is no client-side timer managing state.

### Critical: Cross-Day Edit Uses HX-Redirect, Not OOB Swap

If a user edits an event's start time such that `event_date` changes (e.g., from March 10 to March 11), the `Update` handler must update two day containers. For Story 1.3, use `HX-Redirect: /trips/{tripID}` — HTMX performs a full-page GET, reloading the entire trip timeline which naturally shows both days correctly.

OOB swap (returning `hx-swap-oob` for two days simultaneously) is reserved for Story 2.2 (drag-and-drop cross-day moves), where `HX-Redirect` is not viable because the drag-drop interaction must not lose the user's current scroll position and visual context.

**Handler logic:**
```go
oldDate := event.EventDate  // captured before Update call
updatedEvent, err := h.service.Update(ctx, id, input)
// ...
if !updatedEvent.EventDate.Equal(oldDate) {
    w.Header().Set("HX-Redirect", "/trips/"+tripIDStr)
    w.WriteHeader(http.StatusOK)
    return
}
// same-day path: render TimelineDay + HX-Retarget
```

### Critical: event_date Recalculation in Service Update

The existing service `Update` updater closure MUST set `event.EventDate` when `StartTime` changes:

```go
updater := func(e *domain.Event) *domain.Event {
    if input.StartTime != nil {
        e.StartTime = *input.StartTime
        st := *input.StartTime
        // Derive date from new start time — NOT Truncate(24h)
        e.EventDate = time.Date(st.Year(), st.Month(), st.Day(), 0, 0, 0, 0, st.Location())
    }
    if input.EndTime != nil { e.EndTime = *input.EndTime }
    if input.Title != nil   { e.Title = *input.Title }
    if input.Location != nil { e.Location = *input.Location }
    if input.Notes != nil   { e.Notes = *input.Notes }
    if input.Pinned != nil  { e.Pinned = *input.Pinned }
    return e
}
```

If this is not present in the existing code, add it. This is AC #5.

### Inline Edit Form — HTMX Target Strategy

The inline edit form uses a two-target strategy (same as Story 1.2's create form):

| Outcome | Handler action | HTMX swap target |
|---|---|---|
| Validation error (422) | Render `EventTimelineItem(event, &EventCardProps{Editing: true, FormValues: formData})` | `#event-{id}` (`hx-target` on form) `outerHTML` |
| Success, same day | Render `TimelineDay` | `HX-Retarget: #day-{date}`, `HX-Reswap: outerHTML` |
| Success, cross-day | `HX-Redirect: /trips/{id}` | Full-page reload |

The form's `hx-target="#event-{id}"` is its DEFAULT target. On error, the handler returns the card HTML and HTMX swaps only that card. On success, `HX-Retarget` overrides the form's default target.

The card re-rendered on 422 calls `EventTimelineItem(event, &EventCardProps{Editing: true, FormValues: formData})`:
- `x-data="{ expanded: true, editing: true }"` — card opens directly in edit mode
- Form fields populated from `props.FormValues` (submitted values, not original event values)
- `props.FormValues.Errors` drives rose borders and `aria-describedby` error messages
- Normal `TimelineDay` path calls `EventTimelineItem(event, nil)` — no change to existing `TimelineDay` template logic beyond updating the function signature

### Inline Edit Form — Time Input Alignment with Create

Use the **same date+time input pattern** as `EventCreateForm`:
- `<input type="hidden" name="date">` — carries the current date
- `<input type="time" name="start_time">` — HH:MM
- `<input type="time" name="end_time">` — HH:MM

**Field value source depends on render path** — the template must distinguish:

| Field | Initial render (`props == nil`) | 422 re-render (`props != nil`) |
|---|---|---|
| `date` | `event.StartTime.Format("2006-01-02")` | `props.FormValues.Date` |
| `start_time` | `event.StartTime.Format("15:04")` | `props.FormValues.StartTime` |
| `end_time` | `event.EndTime.Format("15:04")` | `props.FormValues.EndTime` |
| `title` | `event.Title` | `props.FormValues.Title` |
| `location` | `event.Location` | `props.FormValues.Location` |
| `notes` | `event.Notes` | `props.FormValues.Notes` |
| `pinned` | `event.Pinned` | `props.FormValues.Pinned` |

Using `event` fields on a 422 re-render silently discards the user's edits. Using `props.FormValues` on an initial render causes a nil-pointer panic. The `if props != nil` guard in the template controls which source is used for all form fields.

The `Update` handler uses `parseDateAndTime(r.FormValue("date"), r.FormValue("start_time"))` from `handler/helpers.go`. **Do NOT use `datetime-local` input type** (legacy from Story 1.1's `EditPage`).

### HX-Trigger Event Name Casing

HTMX lowercases event names when dispatching as DOM events. The handler sets:
```
HX-Trigger: {"showUndoToast": {"eventId": 42, "tripId": 1, "eventDate": "2026-03-10"}}
```

HTMX dispatches a DOM `CustomEvent` named `showundotoast` (all lowercase). Alpine.js listens with `x-on:showundotoast.window`. Verify this in smoke testing — if the toast doesn't appear, check browser devtools for the actual dispatched event name.

The `$event.detail` object from HTMX 2.0 `HX-Trigger` contains the value directly: `$event.detail.eventId`, `$event.detail.tripId`, `$event.detail.eventDate`. If the event detail is nested (e.g., `$event.detail.value`), adjust accordingly.

### Delete Button — Real HTTP DELETE via hx-delete

The delete button uses HTMX `hx-delete` directly (not method override):
```templ
<button type="button"
        hx-delete={ fmt.Sprintf("/trips/%d/events/%d", event.TripID, event.ID) }
        hx-target={ fmt.Sprintf("#day-%s", event.EventDate.Format("2006-01-02")) }
        hx-swap="outerHTML">
    Delete
</button>
```

`hx-delete` sends a real HTTP DELETE. The chi route `r.Delete("/trips/{tripID}/events/{id}", ...)` handles it directly without needing the method override middleware. No `_method` hidden field needed here.

### Soft Delete — Query Update Checklist

Every SELECT query in `internal/repository/sql/events.sql` that reads events must add `AND deleted_at IS NULL`. Go through each query name and verify:
- `ListEventsByTripAndDate` ✓
- `GetEventsByTrip` (if it exists) ✓
- `GetEventByID` ✓
- `GetMaxPositionByTripAndDate` ✓
- `GetLastEventByTrip` (if it exists) ✓
- `CountEventsByTrip` ✓

Missing even one will cause soft-deleted events to ghost-appear in certain contexts. After running `just generate`, verify the generated `events.sql.go` methods all include the `deleted_at IS NULL` condition.

### Previous Story (1.1 + 1.2) Key Patterns

- **pgtype helpers**: `toPgDate()`, `toPgTimestamptz()`, `toPgText()`, `toPgFloat8()`, `toPgBool()` in `internal/repository/helpers.go`
- **EventDate derivation**: `time.Date(y, m, d, 0, 0, 0, 0, loc)` — NOT `Truncate(24*time.Hour)`
- **Error mapping**: `domain.ErrInvalidInput` → 422, `domain.ErrNotFound` → 404, else → 500
- **Date+time parsing**: `parseDateAndTime(dateStr, timeStr)` in `handler/helpers.go`
- **Method override**: HTML forms use `_method=PUT` hidden field. `hx-delete` does NOT need this (sends real DELETE).
- **Day HTML swap**: `HX-Retarget: #day-{date}` + `HX-Reswap: outerHTML` on success
- **Form error pattern**: 422 + form HTML with rose borders + `aria-describedby` for inline messages
- **No templui components**: Custom Alpine.js implementations (incompatible with Tailwind v4 `tailwind-merge-go`)
- **Custom CSS switch**: For pinned toggle — `sr-only` checkbox + styled spans (from Story 1.2)

### Project Structure Notes

Files to modify:
- `migrations/` — add `002_soft_delete_events.up.sql` + `.down.sql` (NEW)
- `internal/repository/sql/events.sql` — add `deleted_at IS NULL` to all SELECTs; add `SoftDeleteEvent`, `RestoreEvent` queries
- `internal/repository/sqlcgen/` — regenerated by `just generate` (do NOT edit manually)
- `internal/domain/ports.go` — add `Restore(ctx, id) (*Event, error)` to `EventRepository`
- `internal/repository/event_store.go` — update `Delete()` to soft-delete, add `Restore()`
- `internal/service/event.go` — add `Restore()` method; verify `EventDate` recalculation in `Update`
- `internal/handler/event.go` — update `Update` handler (HTMX), `Delete` handler (HTMX + HX-Trigger); add `Restore` handler
- `internal/handler/event.templ` — redesign `EventTimelineItem(event, props *EventCardProps)` (single component, handles view + edit modes); add `EventCardProps` type in `event.go`
- `internal/handler/routes.go` — add restore route
- `internal/handler/trip.templ` — add undo toast component inside `TripDetailPage`

Files NOT to modify (generated):
- `internal/repository/sqlcgen/*` — always regenerated by `just generate`
- `internal/handler/*_templ.go` — always regenerated by `templ generate`

Run `just generate` after any `.templ` or `.sql` file change. Run `just migrate-up` after adding the migration.

### References

- [Source: _bmad-output/planning-artifacts/architecture.md#Communication Patterns] — HTMX interaction contract: PUT `/events/{id}` → full day HTML swap; DELETE → full day HTML swap
- [Source: _bmad-output/planning-artifacts/architecture.md#Frontend Architecture] — HTMX day-level swaps; OOB swap reserved for cross-day drag-and-drop (Story 2.2)
- [Source: _bmad-output/planning-artifacts/architecture.md#Process Patterns] — Error handling layers; service returns domain errors, handler translates to HTTP
- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.3] — Acceptance criteria and BDD scenarios
- [Source: _bmad-output/planning-artifacts/epics.md#Additional Requirements from UX] — "Inline editing for modifications" (not Sheet); undo toast 8s; no confirmation dialog for single events
- [Source: _bmad-output/implementation-artifacts/1-2-activity-and-food-event-creation.md] — parseDateAndTime, HTMX retarget pattern, field-level validation, custom Alpine sheet, TypeSelector, custom CSS switch
- [Source: _bmad-output/implementation-artifacts/1-2-activity-and-food-event-creation.md#Completion Notes] — templui incompatible with Tailwind v4; HX-Trigger closeSheet pattern; no `hx-target-422`

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- Soft delete implemented via `deleted_at TIMESTAMPTZ` column (migration 002). All SELECT queries filter `AND deleted_at IS NULL`.
- `EventRepository` interface extended with `Restore(ctx, id) (*Event, error)`. `EventStore.Delete` now calls `SoftDeleteEvent` (no more hard delete). `EventStore.Restore` calls `RestoreEvent`.
- `EventService.Update` validates end time > start time when both are provided. `EventDate` recalculates from new `StartTime` using `time.Date(y,m,d,0,0,0,0,loc)` (existing code confirmed correct).
- `EventTimelineItem` signature updated to `(event domain.Event, props *EventCardProps)`. Single component handles view mode (props=nil) and 422 edit mode (props!=nil with Editing:true).
- Update handler: fetches event first, captures oldEventDate, uses `parseDateAndTime()` (not `datetime-local`), sends `HX-Redirect` on cross-day, `HX-Retarget: #day-*` on same-day success.
- Delete handler: soft deletes, re-fetches day events, sends `HX-Trigger: {"showUndoToast": {...}}` with JSON event data.
- Restore handler: restores event, sends `HX-Retarget: #day-*` + `HX-Trigger: {"hideUndoToast": true}`.
- Alpine Collapse plugin vendored to `static/js/alpine-collapse.min.js` and loaded before `alpine.min.js` in layout.
- `parseDateTime` (unused after Update handler rewrite) removed from `helpers.go`.
- 7 new service tests added covering EventDate recalculation, Update validation, Delete+Restore round-trip, ErrNotFound paths, and soft-delete filtering.

### File List

- `migrations/002_soft_delete_events.up.sql` (new)
- `migrations/002_soft_delete_events.down.sql` (new)
- `internal/repository/sql/events.sql` (modified)
- `internal/repository/sqlcgen/events.sql.go` (regenerated)
- `internal/repository/sqlcgen/models.go` (regenerated)
- `internal/domain/ports.go` (modified)
- `internal/repository/event_store.go` (modified)
- `internal/service/event.go` (modified)
- `internal/service/event_test.go` (modified)
- `internal/handler/event.go` (modified)
- `internal/handler/event.templ` (modified)
- `internal/handler/event_templ.go` (regenerated)
- `internal/handler/trip.templ` (modified)
- `internal/handler/trip_templ.go` (regenerated)
- `internal/handler/layout.templ` (modified)
- `internal/handler/layout_templ.go` (regenerated)
- `internal/handler/routes.go` (modified)
- `internal/handler/helpers.go` (modified)
- `static/js/alpine-collapse.min.js` (new)
- `static/css/app.css` (modified)
- `internal/handler/event_test.go` (new)

## Change Log

- 2026-02-19: Implemented story 1.3 — event edit, delete, detail view, and undo toast. Added soft delete via migration 002, extended domain/repository/service with Restore, redesigned EventTimelineItem with inline edit mode, updated Update/Delete handlers with HTMX support, added Restore handler and route, added undo toast component, vendored Alpine Collapse plugin. 7 new service tests. (claude-sonnet-4-6)
- 2026-02-19: Code review fixes — added `hx-disabled-elt` and disabled styling to inline edit/delete buttons for loading states; removed redundant POST route handlers for Update. (gemini-cli)
- 2026-02-19: Code review fixes (adversarial) — added non-HTMX fallback redirect to Restore handler (H1); added HTMX check to Update handler 422 paths via renderCardError helper, non-HTMX redirects to edit page (M2); fixed service.Update() to validate end>start when only one time is provided by pre-fetching event (M3); documented static/css/app.css in File List (M1); added TestEventService_Update_OnlyStartTimeMovedPastEndTime. (claude-sonnet-4-6)
- 2026-02-19: Bug fix — undo toast visibility. Modified `Delete` handler to use script injection for `showUndoToast` event dispatching to ensure it fires after the triggering element is removed. Added `internal/handler/event_test.go`. (dev-agent)
