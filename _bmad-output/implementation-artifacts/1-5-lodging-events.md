# Story 1.5: Lodging Events

Status: review

## Story

As a traveler,
I want to add Lodging events with check-in/check-out times and booking reference,
so that I can track accommodation details alongside my daily activities.

## Acceptance Criteria

1. **Given** the user selects Lodging from the TypeSelector, **When** the form morphs, **Then** additional fields appear: check-in time, check-out time, booking reference.

2. **Given** the user submits a valid Lodging event, **When** the event is saved, **Then** the `lodging_details` are persisted in the `lodging_details` table within the same transaction as the base event, **And** the `LodgingCardContent` displays lodging-specific metadata (check-in/out, booking ref).

3. **Given** a user edits a Lodging event, **When** they modify lodging-specific fields, **Then** the `lodging_details` are updated and the card reflects the changes.

4. **Given** a user deletes a Lodging event, **When** the event is deleted, **Then** both the base event and `lodging_details` are removed (CASCADE — no explicit code needed, DB handles it).

## Tasks / Subtasks

### SQL Queries for lodging_details

- [x] Task 1: Add sqlc queries for lodging_details write and read paths (AC: #1, #2, #3)
  - [ ] 1.1 Create `internal/repository/sql/lodging_details.sql` with:
    ```sql
    -- name: CreateLodgingDetails :one
    INSERT INTO lodging_details (event_id, check_in_time, check_out_time, booking_reference)
    VALUES ($1, $2, $3, $4)
    RETURNING *;

    -- name: GetLodgingDetailsByEventID :one
    SELECT * FROM lodging_details WHERE event_id = $1;

    -- name: GetLodgingDetailsByEventIDs :many
    SELECT * FROM lodging_details WHERE event_id = ANY(@event_ids::int[]);

    -- name: UpdateLodgingDetails :one
    UPDATE lodging_details
    SET check_in_time = $2, check_out_time = $3, booking_reference = $4
    WHERE event_id = $1
    RETURNING *;
    ```
  - [ ] 1.2 Run `just generate` — generates `internal/repository/sqlcgen/lodging_details.sql.go`
  - [ ] 1.3 Verify generated file: `sqlcgen.CreateLodgingDetailsParams`, `sqlcgen.UpdateLodgingDetailsParams`, `sqlcgen.LodgingDetail` structs exist. Note the exact field types for `CheckInTime`, `CheckOutTime` (will be `pgtype.Timestamptz` — nullable), and `BookingReference` (will be `pgtype.Text` — nullable with default '').

### Repository Helpers

- [ ] Task 2: Add pgtype helpers for optional timestamps (AC: #2, #3)
  - [ ] 2.1 Add to `internal/repository/helpers.go`:
    ```go
    // toOptionalPgTimestamptz converts a nullable *time.Time to pgtype.Timestamptz.
    func toOptionalPgTimestamptz(t *time.Time) pgtype.Timestamptz {
        if t == nil {
            return pgtype.Timestamptz{Valid: false}
        }
        return pgtype.Timestamptz{Time: *t, Valid: true}
    }

    // fromPgTimestamptz converts a pgtype.Timestamptz to a nullable *time.Time.
    func fromPgTimestamptz(t pgtype.Timestamptz) *time.Time {
        if !t.Valid {
            return nil
        }
        return &t.Time
    }
    ```
    These are needed by `LodgingDetailsStore` to handle nullable timestamp columns.

### Domain Model

- [ ] Task 3: Add LodgingDetails to domain (AC: #2)
  - [ ] 3.1 Add to `internal/domain/models.go` (after `FlightDetails`):
    ```go
    type LodgingDetails struct {
        CheckInTime      *time.Time
        CheckOutTime     *time.Time
        BookingReference string
        ID               int
        EventID          int
    }
    ```
    `CheckInTime` and `CheckOutTime` are optional — the lodging_details schema has nullable TIMESTAMPTZ columns. `BookingReference` is a plain string (DB default is '' not NULL, but sqlc will generate pgtype.Text; extract `.String` in the row mapper).
  - [ ] 3.2 Add `Lodging *LodgingDetails` field to the existing `Event` struct (alongside `Flight *FlightDetails`):
    ```go
    Lodging *LodgingDetails // nil for all non-lodging events
    ```
    Field ordering will be checked by the `fieldalignment` linter — run `fieldalignment -fix ./internal/domain/...` if lint fails.

### Repository — LodgingDetailsStore

- [ ] Task 4: Create LodgingDetailsStore adapter (AC: #2, #3)
  - [ ] 4.1 Create `internal/repository/lodging_details_store.go`:
    ```go
    package repository

    import (
        "context"
        "errors"
        "fmt"

        "github.com/jackc/pgx/v5"

        "github.com/simopzz/traccia/internal/domain"
        "github.com/simopzz/traccia/internal/repository/sqlcgen"
    )

    type LodgingDetailsStore struct{}

    func NewLodgingDetailsStore() *LodgingDetailsStore {
        return &LodgingDetailsStore{}
    }

    // Create inserts lodging_details within the caller's transaction (q is tx-scoped).
    func (s *LodgingDetailsStore) Create(ctx context.Context, q *sqlcgen.Queries, eventID int, ld *domain.LodgingDetails) (*domain.LodgingDetails, error) {
        row, err := q.CreateLodgingDetails(ctx, sqlcgen.CreateLodgingDetailsParams{
            EventID:          int32(eventID),
            CheckInTime:      toOptionalPgTimestamptz(ld.CheckInTime),
            CheckOutTime:     toOptionalPgTimestamptz(ld.CheckOutTime),
            BookingReference: toPgText(ld.BookingReference),
        })
        if err != nil {
            return nil, fmt.Errorf("inserting lodging_details for event %d: %w", eventID, err)
        }
        result := lodgingRowToDomain(&row)
        return &result, nil
    }

    // GetByEventID loads lodging_details. Returns domain.ErrNotFound if the row doesn't exist.
    func (s *LodgingDetailsStore) GetByEventID(ctx context.Context, q *sqlcgen.Queries, eventID int) (*domain.LodgingDetails, error) {
        row, err := q.GetLodgingDetailsByEventID(ctx, int32(eventID))
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                return nil, domain.ErrNotFound
            }
            return nil, fmt.Errorf("fetching lodging_details for event %d: %w", eventID, err)
        }
        result := lodgingRowToDomain(&row)
        return &result, nil
    }

    // GetByEventIDs fetches lodging_details for multiple events in a single query.
    func (s *LodgingDetailsStore) GetByEventIDs(ctx context.Context, q *sqlcgen.Queries, eventIDs []int) (map[int]*domain.LodgingDetails, error) {
        if len(eventIDs) == 0 {
            return nil, nil
        }
        ids := make([]int32, len(eventIDs))
        for i, id := range eventIDs {
            ids[i] = int32(id)
        }
        rows, err := q.GetLodgingDetailsByEventIDs(ctx, ids)
        if err != nil {
            return nil, fmt.Errorf("fetching lodging_details by ids: %w", err)
        }
        results := make(map[int]*domain.LodgingDetails)
        for i := range rows {
            ld := lodgingRowToDomain(&rows[i])
            results[int(rows[i].EventID)] = &ld
        }
        return results, nil
    }

    // Update updates existing lodging_details. Uses the caller-provided queries (can be tx-scoped).
    func (s *LodgingDetailsStore) Update(ctx context.Context, q *sqlcgen.Queries, eventID int, ld *domain.LodgingDetails) (*domain.LodgingDetails, error) {
        row, err := q.UpdateLodgingDetails(ctx, sqlcgen.UpdateLodgingDetailsParams{
            EventID:          int32(eventID),
            CheckInTime:      toOptionalPgTimestamptz(ld.CheckInTime),
            CheckOutTime:     toOptionalPgTimestamptz(ld.CheckOutTime),
            BookingReference: toPgText(ld.BookingReference),
        })
        if err != nil {
            return nil, fmt.Errorf("updating lodging_details for event %d: %w", eventID, err)
        }
        result := lodgingRowToDomain(&row)
        return &result, nil
    }

    func lodgingRowToDomain(row *sqlcgen.LodgingDetail) domain.LodgingDetails {
        return domain.LodgingDetails{
            ID:               int(row.ID),
            EventID:          int(row.EventID),
            CheckInTime:      fromPgTimestamptz(row.CheckInTime),
            CheckOutTime:     fromPgTimestamptz(row.CheckOutTime),
            BookingReference: row.BookingReference.String,
        }
    }
    ```
    **Note:** Verify the exact field names in the generated `sqlcgen.LodgingDetail` struct after running `just generate` in Task 1. Adjust if sqlc uses different names.

### Repository — EventStore Updates

- [ ] Task 5: Update EventStore to add `lodging` field and wire lodging into Create/Update/load paths (AC: #2, #3)
  - [ ] 5.1 Update `EventStore` struct and constructor in `internal/repository/event_store.go`:
    ```go
    type EventStore struct {
        db      *pgxpool.Pool
        queries *sqlcgen.Queries
        flight  *FlightDetailsStore
        lodging *LodgingDetailsStore
    }

    func NewEventStore(db *pgxpool.Pool, flightStore *FlightDetailsStore, lodgingStore *LodgingDetailsStore) *EventStore {
        return &EventStore{
            db:      db,
            queries: sqlcgen.New(db),
            flight:  flightStore,
            lodging: lodgingStore,
        }
    }
    ```
  - [ ] 5.2 Add `loadLodgingDetails` helper in `event_store.go` (mirrors `loadFlightDetails`):
    ```go
    func (s *EventStore) loadLodgingDetails(ctx context.Context, events []domain.Event) []domain.Event {
        var lodgingIDs []int
        for i := range events {
            if events[i].Category == domain.CategoryLodging {
                lodgingIDs = append(lodgingIDs, events[i].ID)
            }
        }
        if len(lodgingIDs) == 0 {
            return events
        }
        details, err := s.lodging.GetByEventIDs(ctx, s.queries, lodgingIDs)
        if err != nil {
            slog.WarnContext(ctx, "failed to load lodging_details", "error", err)
            return events
        }
        for i := range events {
            if events[i].Category == domain.CategoryLodging {
                events[i].Lodging = details[events[i].ID]
            }
        }
        return events
    }
    ```
  - [ ] 5.3 In `Create` method, add the lodging transactional path. Insert it BEFORE the non-transactional fallback block (after the existing flight block):
    ```go
    if event.Category == domain.CategoryLodging && event.Lodging != nil {
        tx, txErr := s.db.Begin(ctx)
        if txErr != nil {
            return fmt.Errorf("beginning transaction: %w", txErr)
        }
        defer func() { _ = tx.Rollback(ctx) }()

        txq := sqlcgen.New(tx)
        row, txErr := txq.CreateEvent(ctx, sqlcgen.CreateEventParams{
            TripID:    int32(event.TripID),
            EventDate: toPgDate(event.EventDate),
            Title:     event.Title,
            Category:  string(event.Category),
            Location:  toPgText(event.Location),
            Latitude:  toPgFloat8(event.Latitude),
            Longitude: toPgFloat8(event.Longitude),
            StartTime: toPgTimestamptz(event.StartTime),
            EndTime:   toPgTimestamptz(event.EndTime),
            Pinned:    toPgBool(event.Pinned),
            Position:  position,
            Notes:     toPgText(event.Notes),
        })
        if txErr != nil {
            return fmt.Errorf("inserting event: %w", txErr)
        }

        lodgingDetails := event.Lodging
        *event = eventRowToDomain(&row)

        ld, txErr := s.lodging.Create(ctx, txq, event.ID, lodgingDetails)
        if txErr != nil {
            return txErr
        }
        event.Lodging = ld

        return tx.Commit(ctx)
    }
    ```
  - [ ] 5.4 In `Update` method, add the lodging transactional path (after the flight block, before the non-transactional fallback):
    ```go
    if updated.Category == domain.CategoryLodging && updated.Lodging != nil {
        tx, txErr := s.db.Begin(ctx)
        if txErr != nil {
            return nil, fmt.Errorf("beginning transaction: %w", txErr)
        }
        defer func() { _ = tx.Rollback(ctx) }()

        txq := sqlcgen.New(tx)
        row, txErr := txq.UpdateEvent(ctx, sqlcgen.UpdateEventParams{
            ID:        int32(id),
            Title:     updated.Title,
            Category:  string(updated.Category),
            Location:  toPgText(updated.Location),
            Latitude:  toPgFloat8(updated.Latitude),
            Longitude: toPgFloat8(updated.Longitude),
            StartTime: toPgTimestamptz(updated.StartTime),
            EndTime:   toPgTimestamptz(updated.EndTime),
            Pinned:    toPgBool(updated.Pinned),
            Position:  int32(updated.Position),
            EventDate: toPgDate(updated.EventDate),
            Notes:     toPgText(updated.Notes),
        })
        if txErr != nil {
            return nil, fmt.Errorf("updating event: %w", txErr)
        }
        result := eventRowToDomain(&row)

        ld, txErr := s.lodging.Update(ctx, txq, id, updated.Lodging)
        if txErr != nil {
            return nil, txErr
        }
        result.Lodging = ld

        if txErr = tx.Commit(ctx); txErr != nil {
            return nil, fmt.Errorf("committing transaction: %w", txErr)
        }
        return &result, nil
    }
    ```
  - [ ] 5.5 Update `GetByID` to also load lodging details:
    ```go
    // After loading event and flight details, add:
    if event.Category == domain.CategoryLodging {
        events := s.loadLodgingDetails(ctx, []domain.Event{event})
        event = events[0]
    }
    ```
    Insert this after the existing flight loading block.
  - [ ] 5.6 Update `ListByTrip` and `ListByTripAndDate` to call `loadLodgingDetails` after `loadFlightDetails`:
    ```go
    events = s.loadFlightDetails(ctx, events)
    return s.loadLodgingDetails(ctx, events), nil
    ```

### Service Updates

- [ ] Task 6: Update service layer for lodging events (AC: #2, #3)
  - [ ] 6.1 Add `LodgingDetails *domain.LodgingDetails` to `CreateEventInput` in `internal/service/event.go`:
    ```go
    type CreateEventInput struct {
        // ... existing fields ...
        LodgingDetails *domain.LodgingDetails // nil for non-lodging events
    }
    ```
    Run `fieldalignment -fix ./internal/service/...` if the linter complains about struct field order.
  - [ ] 6.2 In `EventService.Create`, after the existing flight details assignment block, add:
    ```go
    if input.Category == domain.CategoryLodging {
        event.Lodging = input.LodgingDetails
        if event.Lodging == nil {
            event.Lodging = &domain.LodgingDetails{} // empty details are valid
        }
    }
    ```
    Place this just before `s.repo.Create(ctx, event)`.
  - [ ] 6.3 Add `LodgingDetails *domain.LodgingDetails` to `UpdateEventInput`:
    ```go
    type UpdateEventInput struct {
        // ... existing fields ...
        LodgingDetails *domain.LodgingDetails // nil means "don't change lodging details"
    }
    ```
  - [ ] 6.4 In the `Update` updater closure, pass through lodging details (alongside the existing flight block):
    ```go
    if input.LodgingDetails != nil {
        event.Lodging = input.LodgingDetails
    }
    ```

### Handler Updates

- [ ] Task 7: Update handler to parse and route lodging form data (AC: #1, #2, #3)
  - [ ] 7.1 Add lodging fields to `EventFormData` in `internal/handler/event.go`:
    ```go
    type EventFormData struct {
        // ... existing fields ...
        // Lodging-specific
        CheckInTime  string // "2006-01-02T15:04" format, empty string = not provided
        CheckOutTime string
        // BookingReference already exists — shared between flight and lodging
    }
    ```
    `BookingReference` is already in `EventFormData` from Story 1.4. Lodging reuses the same form field name (`booking_reference`) since both categories use it for the same concept.
    Run `fieldalignment -fix ./internal/handler/...` if the linter complains.
  - [ ] 7.2 Add `parseLodgingDetails` helper in `internal/handler/event.go`:
    ```go
    func parseLodgingDetails(formData *EventFormData) *domain.LodgingDetails {
        ld := &domain.LodgingDetails{
            BookingReference: formData.BookingReference,
        }
        if formData.CheckInTime != "" {
            t, err := time.ParseInLocation("2006-01-02T15:04", formData.CheckInTime, time.UTC)
            if err == nil {
                ld.CheckInTime = &t
            }
        }
        if formData.CheckOutTime != "" {
            t, err := time.ParseInLocation("2006-01-02T15:04", formData.CheckOutTime, time.UTC)
            if err == nil {
                ld.CheckOutTime = &t
            }
        }
        return ld
    }
    ```
  - [ ] 7.3 In `Create` handler, parse lodging-specific form values (alongside the existing flight parsing block):
    ```go
    formData.CheckInTime  = r.FormValue("check_in_time")
    formData.CheckOutTime = r.FormValue("check_out_time")
    // BookingReference is already parsed for flight — it covers lodging too
    ```
  - [ ] 7.4 In `Create` handler, build `LodgingDetails` and pass to service (alongside the flight block):
    ```go
    var lodgingDetails *domain.LodgingDetails
    if category == string(domain.CategoryLodging) {
        lodgingDetails = parseLodgingDetails(&formData)
    }

    input := &service.CreateEventInput{
        // ... existing fields ...
        LodgingDetails: lodgingDetails,
    }
    ```
  - [ ] 7.5 In `Update` handler, parse lodging fields and build `LodgingDetails`:
    ```go
    formData.CheckInTime  = r.FormValue("check_in_time")
    formData.CheckOutTime = r.FormValue("check_out_time")

    var lodgingDetails *domain.LodgingDetails
    if event.Category == domain.CategoryLodging {
        lodgingDetails = parseLodgingDetails(&formData)
    }
    input := &service.UpdateEventInput{
        // ... existing fields ...
        LodgingDetails: lodgingDetails,
    }
    ```

### Template — Enable Lodging in TypeSelector

- [ ] Task 8: Enable Lodging in the creation form (AC: #1)
  - [ ] 8.1 In `internal/handler/event_form.templ`, update `enabledCategories`:
    ```go
    var enabledCategories = map[string]bool{
        "activity": true,
        "food":     true,
        "flight":   true,
        "lodging":  true, // ADD THIS
    }
    ```
  - [ ] 8.2 Create `internal/handler/lodging_form.templ` for the lodging-specific form fields. Follow the exact same pattern as `internal/handler/flight_form.templ`:
    ```templ
    package handler

    import "github.com/simopzz/traccia/internal/domain"

    func lodgingDataFromDomain(event domain.Event) EventFormData {
        if event.Lodging == nil {
            return EventFormData{}
        }
        data := EventFormData{
            BookingReference: event.Lodging.BookingReference,
        }
        if event.Lodging.CheckInTime != nil {
            data.CheckInTime = event.Lodging.CheckInTime.Format("2006-01-02T15:04")
        }
        if event.Lodging.CheckOutTime != nil {
            data.CheckOutTime = event.Lodging.CheckOutTime.Format("2006-01-02T15:04")
        }
        return data
    }

    // LodgingFormFields renders lodging-specific form inputs.
    // Used in both the create sheet and the full-page fallback.
    templ LodgingFormFields(data *EventFormData) {
        <div x-show="selected === 'lodging'" class="mb-4 border-t-2 border-slate-100 pt-4">
            <p class="text-xs font-bold uppercase tracking-wide text-amber-700 mb-3">Lodging Details</p>
            <div class="grid grid-cols-2 gap-3 mb-3">
                <div>
                    <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Check-in</label>
                    <input
                        type="datetime-local"
                        name="check_in_time"
                        value={ data.CheckInTime }
                        class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand focus:shadow-[2px_2px_0px_0px_#008080]"
                    />
                </div>
                <div>
                    <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Check-out</label>
                    <input
                        type="datetime-local"
                        name="check_out_time"
                        value={ data.CheckOutTime }
                        class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand focus:shadow-[2px_2px_0px_0px_#008080]"
                    />
                </div>
            </div>
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Booking Reference</label>
                <input
                    type="text"
                    name="booking_reference"
                    value={ data.BookingReference }
                    class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand focus:shadow-[2px_2px_0px_0px_#008080]"
                />
            </div>
        </div>
    }
    ```
  - [ ] 8.3 Add `@LodgingFormFields(data)` call in `EventCreateForm` in `event_form.templ`. Place it AFTER the `@FlightFormFields(data)` call (or in the same logical group of type-specific fields).
  - [ ] 8.4 Add `@LodgingFormFields(data)` call in `EventNewPage` (the full-page fallback form), also after `@FlightFormFields(data)`.
  - [ ] 8.5 Run `just generate` after creating `lodging_form.templ`.

### Template — LodgingCardContent

- [ ] Task 9: Create lodging card display component (AC: #2)
  - [ ] 9.1 Create `internal/handler/lodging_card.templ`:
    ```templ
    package handler

    import (
        "fmt"
        "github.com/simopzz/traccia/internal/domain"
    )

    // LodgingCardContent renders lodging-specific detail in an expanded event card.
    // Called from EventTimelineItem when event.Category == domain.CategoryLodging.
    templ LodgingCardContent(ld *domain.LodgingDetails) {
        if ld == nil {
            return
        }
        <div class="mt-3 pt-3 border-t border-slate-100 space-y-2">
            if ld.CheckInTime != nil || ld.CheckOutTime != nil {
                <div class="flex items-center gap-3 text-sm">
                    if ld.CheckInTime != nil {
                        <div class="text-xs text-slate-500">
                            <span class="font-medium text-slate-600">Check-in:</span>
                            <span class="font-mono ml-1">{ ld.CheckInTime.Format("Mon 2 Jan, 15:04") }</span>
                        </div>
                    }
                    if ld.CheckOutTime != nil {
                        <div class="text-xs text-slate-500">
                            <span class="font-medium text-slate-600">Check-out:</span>
                            <span class="font-mono ml-1">{ ld.CheckOutTime.Format("Mon 2 Jan, 15:04") }</span>
                        </div>
                    }
                </div>
            }
            if ld.BookingReference != "" {
                <div class="text-xs text-slate-500">
                    <span class="font-medium text-slate-600">Ref:</span>
                    <span class="font-mono ml-1 tracking-wider">{ ld.BookingReference }</span>
                </div>
            }
        </div>
    }
    ```
    **Note:** The `fmt` import is included for consistency with the pattern; remove it if unused.
  - [ ] 9.2 Run `just generate` after creating this file.

### Template — EventTimelineItem Updates for Lodging

- [ ] Task 10: Update EventTimelineItem to render lodging details in view and edit modes (AC: #2, #3)
  - [ ] 10.1 In `internal/handler/event.templ`, in the **view mode** section (`x-show="!editing"`), add after the flight block:
    ```templ
    if event.Category == domain.CategoryLodging {
        @LodgingCardContent(event.Lodging)
    }
    ```
  - [ ] 10.2 In the **edit mode** section (`x-show="editing"`), add lodging inline edit fields after the flight edit block:
    ```templ
    if event.Category == domain.CategoryLodging {
        <div class="mb-3 pt-3 border-t-2 border-slate-100">
            <p class="text-xs font-bold uppercase tracking-wide text-amber-700 mb-3">Lodging Details</p>
            <div class="grid grid-cols-2 gap-3 mb-3">
                <div>
                    <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1">Check-in</label>
                    if props != nil {
                        <input type="datetime-local" name="check_in_time" value={ props.FormValues.CheckInTime }
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
                    } else {
                        <input type="datetime-local" name="check_in_time" value={ lodgingDataFromDomain(event).CheckInTime }
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
                    }
                </div>
                <div>
                    <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1">Check-out</label>
                    if props != nil {
                        <input type="datetime-local" name="check_out_time" value={ props.FormValues.CheckOutTime }
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
                    } else {
                        <input type="datetime-local" name="check_out_time" value={ lodgingDataFromDomain(event).CheckOutTime }
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
                    }
                </div>
            </div>
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1">Booking Reference</label>
                if props != nil {
                    <input type="text" name="booking_reference" value={ props.FormValues.BookingReference }
                        class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
                } else {
                    <input type="text" name="booking_reference" value={ lodgingDataFromDomain(event).BookingReference }
                        class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
                }
            </div>
        </div>
    }
    ```
    Lodging edit fields are shown via server-side Go condition (same pattern as flight — category is fixed at creation).

### Seed Command

- [ ] Task 11: Update `cmd/seed/main.go` to support lodging events (AC: #2)
  - [ ] 11.1 The seed has its own `NewEventStore` call on line 60 that must be updated alongside `cmd/app/main.go`:
    ```go
    flightDetailsStore := repository.NewFlightDetailsStore()
    lodgingDetailsStore := repository.NewLodgingDetailsStore()
    eventStore := repository.NewEventStore(pool, flightDetailsStore, lodgingDetailsStore)
    ```
  - [ ] 11.2 Add a `createLodgingEvent` function (mirrors `createFlightEvent`):
    ```go
    func createLodgingEvent(ctx context.Context, eventService *service.EventService, tripID int, checkIn time.Time, checkOut time.Time) error {
        hotelNames := []string{"Grand Hotel", "City Inn", "Palace Suites", "Central Hotel", "Boutique Stay"}
        refLetters := []string{"HTL", "BKG", "RSV", "CNF"}
        ref := fmt.Sprintf("%s%d", refLetters[rand.Intn(len(refLetters))], rand.Intn(99999))

        details := &domain.LodgingDetails{
            CheckInTime:      &checkIn,
            CheckOutTime:     &checkOut,
            BookingReference: ref,
        }

        input := &service.CreateEventInput{
            TripID:         tripID,
            Title:          hotelNames[rand.Intn(len(hotelNames))],
            Category:       domain.CategoryLodging,
            StartTime:      checkIn,
            EndTime:        checkOut,
            Location:       "Hotel Address",
            Notes:          "Generated by seeder",
            LodgingDetails: details,
        }

        _, err := eventService.Create(ctx, input)
        return err
    }
    ```
  - [ ] 11.3 In `seedEvents`, remove `domain.CategoryLodging` from the random categories slice (line ~189) — lodging events should be created with proper details via `createLodgingEvent`, not as bare base events:
    ```go
    categories := []domain.EventCategory{
        domain.CategoryActivity,
        domain.CategoryFood,
        domain.CategoryTransit,
        // CategoryLodging removed — created explicitly below
    }
    ```
  - [ ] 11.4 In `seedEvents`, add one lodging event per trip spanning the full trip duration (call after the day loop, not inside it):
    ```go
    checkIn := time.Date(
        trip.StartDate.Year(), trip.StartDate.Month(), trip.StartDate.Day(),
        15, 0, 0, 0, trip.StartDate.Location(), // standard 3pm check-in
    )
    checkOut := time.Date(
        trip.EndDate.Year(), trip.EndDate.Month(), trip.EndDate.Day(),
        11, 0, 0, 0, trip.EndDate.Location(), // standard 11am check-out
    )
    if err := createLodgingEvent(ctx, eventService, trip.ID, checkIn, checkOut); err != nil {
        slog.Warn("Failed to create lodging event", "error", err, "trip_id", trip.ID)
        failureCount++
    } else {
        slog.Debug("Created lodging event", "trip_id", trip.ID)
        successCount++
    }
    ```

### DI Wiring

- [ ] Task 12: Update `cmd/app/main.go` to wire LodgingDetailsStore (AC: #2)
  - [ ] 11.1 Add `lodgingDetailsStore` and update `NewEventStore` call:
    ```go
    // Repositories
    tripStore := repository.NewTripStore(pool)
    flightDetailsStore := repository.NewFlightDetailsStore()
    lodgingDetailsStore := repository.NewLodgingDetailsStore()
    eventStore := repository.NewEventStore(pool, flightDetailsStore, lodgingDetailsStore)
    ```

### Testing

- [ ] Task 13: Write service tests and verify build (AC: #1, #2, #3, #4)
  - [ ] 13.1 Service test: `Create` lodging event → `event.Lodging` populated with correct field values
    - Mock `EventRepository.Create` to capture the `*domain.Event` passed; verify `event.Lodging != nil` and `BookingReference` matches input.
  - [ ] 13.2 Service test: `Create` lodging event with nil `LodgingDetails` → defaults to empty `LodgingDetails{}` (no nil panic).
  - [ ] 13.3 Service test: `Update` lodging event → `event.Lodging` updated to new values.
    - Input: existing event with `Lodging.BookingReference = "ABC123"`, update with `LodgingDetails{BookingReference: "XYZ789"}`.
    - Verify updater sets `event.Lodging.BookingReference = "XYZ789"`.
  - [ ] 13.4 Service test: `Update` non-lodging event with nil `LodgingDetails` → no change to `event.Lodging`.
  - [ ] 13.5 Run `just test` — all passing, no races.
  - [ ] 13.6 Run `just lint` — zero violations. Apply `fieldalignment -fix` if struct field ordering violations appear.
  - [ ] 13.7 Run `just build` — binary compiles.
  - [ ] 13.8 Manual smoke tests:
    - Click "Add Event" → select Lodging in TypeSelector → lodging fields appear (check-in, check-out, booking ref); flight fields hidden.
    - Fill in check-in/check-out datetimes and booking ref → Save → event appears in timeline with amber/bed icon.
    - Expand card → `LodgingCardContent` shows formatted check-in/out times and booking reference in monospace.
    - Click Edit → lodging fields appear pre-filled with existing values → modify → Save → card updates.
    - Click Delete → event removed → Undo → event restored with lodging details intact.
    - Create Activity → no lodging fields appear in TypeSelector or card.
    - Create Flight → lodging fields hidden, flight fields visible.

## Dev Notes

### Architecture: No Migration Needed

The `lodging_details` table already exists in `migrations/001_initial.up.sql`:
```sql
CREATE TABLE lodging_details (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL UNIQUE REFERENCES events(id) ON DELETE CASCADE,
    check_in_time TIMESTAMPTZ,
    check_out_time TIMESTAMPTZ,
    booking_reference TEXT DEFAULT ''
);
```
**No new migration is needed for Story 1.5.** The schema is already correct.

### Critical: Nullable TIMESTAMPTZ vs non-nullable TEXT

`check_in_time` and `check_out_time` are nullable TIMESTAMPTZ → sqlc generates `pgtype.Timestamptz`.
`booking_reference TEXT DEFAULT ''` has no NOT NULL → sqlc generates `pgtype.Text` (`.String` accessor).

The `helpers.go` already has `toPgTimestamptz(time.Time)`. For **optional** timestamps, we add two new helpers (Task 2):
- `toOptionalPgTimestamptz(t *time.Time) pgtype.Timestamptz`
- `fromPgTimestamptz(t pgtype.Timestamptz) *time.Time`

### Critical: EventStore Signature Change

`NewEventStore` gains a third parameter: `lodgingStore *LodgingDetailsStore`. This breaks the existing calls in both `cmd/app/main.go` (Task 12) and `cmd/seed/main.go` (Task 11). If there are any other callers (e.g., test files using `NewEventStore`), update those too. Check `event_test.go` files.

### Critical: loadLodgingDetails After loadFlightDetails

`ListByTrip` and `ListByTripAndDate` must call both loaders:
```go
events = s.loadFlightDetails(ctx, events)
return s.loadLodgingDetails(ctx, events), nil
```
Order doesn't matter, but both must be called to avoid nil `Lodging` fields on lodging events.

### Critical: Soft Delete + CASCADE

Same pattern as Flight (from Story 1.4):
- Soft delete (`set deleted_at = NOW()`) only updates the base event — `lodging_details` row stays in DB.
- Restore (`set deleted_at = NULL`) — `lodging_details` row still present. Undo flow works correctly.
- Hard delete (trip deletion via CASCADE) → `lodging_details` deleted by DB CASCADE.

No code change needed for delete/restore — the existing handlers work for all categories.

### Pattern: BookingReference Reuse

`EventFormData.BookingReference` is already present from Story 1.4 (used for flights). Lodging reuses the same form field (`name="booking_reference"`) since the concept is identical. This means `EventFormData` needs no new field for booking_reference — just `CheckInTime` and `CheckOutTime`.

### Pattern: Monospace for Structured Data

Per UX spec: monospace for booking references. Use `font-mono` Tailwind class for:
- Booking reference values
- Check-in/check-out formatted times

Lodging category color convention: use amber (`text-amber-700`, `bg-amber-50`) — consistent with the `Bed` icon already in `categoryIcons` map in `event_form.templ`.

### Pattern: datetime-local Input Format

`<input type="datetime-local">` produces values in `"2006-01-02T15:04"` format (no seconds). Parse with:
```go
time.ParseInLocation("2006-01-02T15:04", formData.CheckInTime, time.UTC)
```
If the field is empty, `parseLodgingDetails` returns `CheckInTime: nil` — this is valid; lodging events don't require check-in/out times.

### Pattern: All Existing Patterns from Stories 1.1–1.4 Apply

- **pgtype helpers**: `toPgText()`, `toPgTimestamptz()`, `toPgFloat8()`, etc. in `helpers.go`
- **Error mapping**: `domain.ErrNotFound` → 404, `domain.ErrInvalidInput` → 422, else → 500
- **HTMX success**: Day-level swap via `HX-Retarget: #day-{date}` + `HX-Reswap: outerHTML` — unchanged
- **Sheet close**: `HX-Trigger: {"close-sheet": true}` on successful create — unchanged
- **`just generate`**: Run after every `.templ` or `.sql` file change
- **Never edit generated files**: `sqlcgen/*.go`, `*_templ.go`
- **fieldalignment linter**: Run `fieldalignment -fix` on changed packages if struct ordering violations appear

### Project Structure Notes

New files:
- `internal/repository/sql/lodging_details.sql`
- `internal/repository/sqlcgen/lodging_details.sql.go` (generated)
- `internal/repository/lodging_details_store.go`
- `internal/handler/lodging_card.templ`
- `internal/handler/lodging_card_templ.go` (generated)
- `internal/handler/lodging_form.templ`
- `internal/handler/lodging_form_templ.go` (generated)

Modified files:
- `internal/repository/helpers.go` — add `toOptionalPgTimestamptz`, `fromPgTimestamptz`
- `internal/domain/models.go` — add `LodgingDetails` struct, `Lodging *LodgingDetails` on `Event`
- `internal/repository/event_store.go` — add `lodging` field, update constructor, Create/Update transactional paths, GetByID/ListByTrip/ListByTripAndDate loading
- `internal/service/event.go` — add `LodgingDetails` to inputs, populate in Create/Update
- `internal/handler/event.go` — add `CheckInTime`/`CheckOutTime` to `EventFormData`, add `parseLodgingDetails`, update Create/Update handlers
- `internal/handler/event_form.templ` — enable lodging in `enabledCategories`, add `@LodgingFormFields` call
- `internal/handler/event.templ` — add `LodgingCardContent` call in view mode, lodging edit fields in edit mode
- `cmd/app/main.go` — wire `LodgingDetailsStore`, update `NewEventStore` call
- `internal/service/event_test.go` — add lodging service tests

Do NOT modify:
- `migrations/` — no new migration needed
- `internal/repository/sqlcgen/*` — always regenerated
- `internal/handler/*_templ.go` — always regenerated

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.5] — Acceptance criteria: check-in/out, booking reference, transactional persist, LodgingCardContent, CASCADE delete
- [Source: _bmad-output/planning-artifacts/architecture.md#Data Architecture] — Base + Detail Tables: `lodging_details` with 1:1 event_id FK; CASCADE DELETE; detail stores handle write-path via DBTX
- [Source: _bmad-output/planning-artifacts/architecture.md#Structure Patterns] — Event type code organization: 8-step layer ordering; `LodgingDetailsStore` adapter; `lodging_card.templ` per architecture
- [Source: _bmad-output/planning-artifacts/architecture.md#Enforcement Guidelines] — Rule 7: add new event types as separate files, never by modifying existing type components
- [Source: migrations/001_initial.up.sql] — lodging_details: TIMESTAMPTZ nullable (check_in/out), TEXT DEFAULT '' (booking_reference)
- [Source: _bmad-output/implementation-artifacts/1-4-flight-events.md] — Full pattern reference: TransactionStore, loadDetails batch loader, FlightDetailsStore, form.templ separation, fieldalignment linter behavior, soft-delete + CASCADE interaction

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- Implemented full lodging event support following the established flight event pattern
- Created `lodging_details.sql` with 4 queries; `just generate` produced `sqlcgen/lodging_details.sql.go`
- Added `toOptionalPgTimestamptz`/`fromPgTimestamptz` helpers to `repository/helpers.go` for nullable TIMESTAMPTZ
- Added `LodgingDetails` domain struct and `Lodging *LodgingDetails` field on `Event`
- Created `LodgingDetailsStore` adapter with Create/GetByEventID/GetByEventIDs/Update methods
- Updated `EventStore`: added `lodging` field, updated `NewEventStore`, added lodging transactional paths, `loadLodgingDetails` batch loader, and refactored base event creation/update into helpers to eliminate duplication.
- Updated service layer: added validation for CheckOut > CheckIn, updated inputs and closures.
- Updated handler: improved form parsing (conditional per category), added error handling for lodging time parsing, and fixed UI color inconsistency (amber used consistently).
- Added `internal/repository/lodging_details_store_test.go` and updated `internal/service/event_test.go` with new validation tests.
- All tests pass; lint clean; build successful.

### File List

**New files:**
- `internal/repository/sql/lodging_details.sql`
- `internal/repository/sqlcgen/lodging_details.sql.go` (generated)
- `internal/repository/lodging_details_store.go`
- `internal/repository/lodging_details_store_test.go`
- `internal/handler/lodging_card.templ`
- `internal/handler/lodging_card_templ.go` (generated)
- `internal/handler/lodging_form.templ`
- `internal/handler/lodging_form_templ.go` (generated)

**Modified files:**
- `internal/repository/helpers.go`
- `internal/domain/models.go`
- `internal/repository/event_store.go`
- `internal/service/event.go`
- `internal/handler/event.go`
- `internal/handler/event_form.templ`
- `internal/handler/event.templ`
- `cmd/app/main.go`
- `cmd/seed/main.go`
- `internal/service/event_test.go`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
