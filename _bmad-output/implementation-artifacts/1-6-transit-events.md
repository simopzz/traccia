# Story 1.6: Transit Events

Status: done

## Story

As a traveler,
I want to add Transit events with origin, destination, and transport mode,
so that I can plan how I get between places and see transit legs in my timeline.

## Acceptance Criteria

1. **Given** the user selects Transit from the TypeSelector, **When** the form morphs, **Then** additional fields appear: origin, destination, transport mode.

2. **Given** the user submits a valid Transit event, **When** the event is saved, **Then** the `transit_details` are persisted in the `transit_details` table within the same transaction as the base event, **And** the `TransitCardContent` displays transit-specific metadata (origin → destination, mode).

3. **Given** a user edits a Transit event, **When** they modify transit-specific fields, **Then** the `transit_details` are updated and the card reflects the changes.

4. **Given** a user deletes a Transit event, **When** the event is deleted, **Then** both the base event and `transit_details` are removed (CASCADE — no explicit code needed, DB handles it).

## Tasks / Subtasks

### SQL Queries for transit_details

- [x] Task 1: Add sqlc queries for transit_details write and read paths (AC: #1, #2, #3)
  - [x] 1.1 Create `internal/repository/sql/transit_details.sql`:
    ```sql
    -- name: CreateTransitDetails :one
    INSERT INTO transit_details (event_id, origin, destination, transport_mode)
    VALUES ($1, $2, $3, $4)
    RETURNING *;

    -- name: GetTransitDetailsByEventID :one
    SELECT * FROM transit_details WHERE event_id = $1;

    -- name: GetTransitDetailsByEventIDs :many
    SELECT * FROM transit_details WHERE event_id = ANY(@event_ids::int[]);

    -- name: UpdateTransitDetails :one
    UPDATE transit_details
    SET origin = $2, destination = $3, transport_mode = $4
    WHERE event_id = $1
    RETURNING *;
    ```
  - [x] 1.2 Run `just generate` — generates `internal/repository/sqlcgen/transit_details.sql.go`
  - [x] 1.3 Verify generated file: `sqlcgen.CreateTransitDetailsParams`, `sqlcgen.UpdateTransitDetailsParams`, `sqlcgen.TransitDetail` structs exist. All three TEXT fields (`Origin`, `Destination`, `TransportMode`) will be `pgtype.Text` — use `.String` accessor when mapping back to domain. Existing `toPgText()` helper covers the write path.

### Domain Model

- [x] Task 2: Add TransitDetails to domain (AC: #2)
  - [x] 2.1 Add to `internal/domain/models.go` (after `LodgingDetails`):
    ```go
    type TransitDetails struct {
        Origin        string
        Destination   string
        TransportMode string
        ID            int
        EventID       int
    }
    ```
    All fields are plain strings — no nullable types. Run `fieldalignment -fix ./internal/domain/...` if the linter complains about struct field order.
  - [x] 2.2 Add `Transit *TransitDetails` field to the existing `Event` struct (alongside `Flight *FlightDetails` and `Lodging *LodgingDetails`):
    ```go
    Transit *TransitDetails // nil for all non-transit events
    ```
    Field ordering will be checked by the `fieldalignment` linter — run `fieldalignment -fix ./internal/domain/...` if lint fails.

### Repository — TransitDetailsStore

- [x] Task 3: Create TransitDetailsStore adapter (AC: #2, #3)
  - [x] 3.1 Create `internal/repository/transit_details_store.go`:
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

    type TransitDetailsStore struct{}

    func NewTransitDetailsStore() *TransitDetailsStore {
        return &TransitDetailsStore{}
    }

    // Create inserts transit_details within the caller's transaction (q is tx-scoped).
    func (s *TransitDetailsStore) Create(ctx context.Context, q *sqlcgen.Queries, eventID int, td *domain.TransitDetails) (*domain.TransitDetails, error) {
        row, err := q.CreateTransitDetails(ctx, sqlcgen.CreateTransitDetailsParams{
            EventID:       int32(eventID),
            Origin:        toPgText(td.Origin),
            Destination:   toPgText(td.Destination),
            TransportMode: toPgText(td.TransportMode),
        })
        if err != nil {
            return nil, fmt.Errorf("inserting transit_details for event %d: %w", eventID, err)
        }
        result := transitRowToDomain(&row)
        return &result, nil
    }

    // GetByEventID loads transit_details. Returns domain.ErrNotFound if the row doesn't exist.
    func (s *TransitDetailsStore) GetByEventID(ctx context.Context, q *sqlcgen.Queries, eventID int) (*domain.TransitDetails, error) {
        row, err := q.GetTransitDetailsByEventID(ctx, int32(eventID))
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                return nil, domain.ErrNotFound
            }
            return nil, fmt.Errorf("fetching transit_details for event %d: %w", eventID, err)
        }
        result := transitRowToDomain(&row)
        return &result, nil
    }

    // GetByEventIDs fetches transit_details for multiple events in a single query.
    func (s *TransitDetailsStore) GetByEventIDs(ctx context.Context, q *sqlcgen.Queries, eventIDs []int) (map[int]*domain.TransitDetails, error) {
        if len(eventIDs) == 0 {
            return nil, nil
        }
        ids := make([]int32, len(eventIDs))
        for i, id := range eventIDs {
            ids[i] = int32(id)
        }
        rows, err := q.GetTransitDetailsByEventIDs(ctx, ids)
        if err != nil {
            return nil, fmt.Errorf("fetching transit_details by ids: %w", err)
        }
        results := make(map[int]*domain.TransitDetails)
        for i := range rows {
            td := transitRowToDomain(&rows[i])
            results[int(rows[i].EventID)] = &td
        }
        return results, nil
    }

    // Update updates existing transit_details. Uses the caller-provided queries (can be tx-scoped).
    func (s *TransitDetailsStore) Update(ctx context.Context, q *sqlcgen.Queries, eventID int, td *domain.TransitDetails) (*domain.TransitDetails, error) {
        row, err := q.UpdateTransitDetails(ctx, sqlcgen.UpdateTransitDetailsParams{
            EventID:       int32(eventID),
            Origin:        toPgText(td.Origin),
            Destination:   toPgText(td.Destination),
            TransportMode: toPgText(td.TransportMode),
        })
        if err != nil {
            return nil, fmt.Errorf("updating transit_details for event %d: %w", eventID, err)
        }
        result := transitRowToDomain(&row)
        return &result, nil
    }

    func transitRowToDomain(row *sqlcgen.TransitDetail) domain.TransitDetails {
        return domain.TransitDetails{
            ID:            int(row.ID),
            EventID:       int(row.EventID),
            Origin:        row.Origin.String,
            Destination:   row.Destination.String,
            TransportMode: row.TransportMode.String,
        }
    }
    ```
    **Note:** Verify exact field names in generated `sqlcgen.TransitDetail` struct after running `just generate` in Task 1. The generated struct may use `TransportMode pgtype.Text` — adjust `.String` accessor accordingly.

### Repository — EventStore Updates

- [x] Task 4: Update EventStore to add `transit` field and wire transit into Create/Update/load paths (AC: #2, #3)
  - [x] 4.1 Update `EventStore` struct and constructor in `internal/repository/event_store.go`:
    ```go
    type EventStore struct {
        db      *pgxpool.Pool
        queries *sqlcgen.Queries
        flight  *FlightDetailsStore
        lodging *LodgingDetailsStore
        transit *TransitDetailsStore
    }

    func NewEventStore(db *pgxpool.Pool, flightStore *FlightDetailsStore, lodgingStore *LodgingDetailsStore, transitStore *TransitDetailsStore) *EventStore {
        return &EventStore{
            db:      db,
            queries: sqlcgen.New(db),
            flight:  flightStore,
            lodging: lodgingStore,
            transit: transitStore,
        }
    }
    ```
  - [x] 4.2 Add `loadTransitDetails` helper in `event_store.go` (mirrors `loadFlightDetails` and `loadLodgingDetails`):
    ```go
    func (s *EventStore) loadTransitDetails(ctx context.Context, events []domain.Event) []domain.Event {
        var transitIDs []int
        for i := range events {
            if events[i].Category == domain.CategoryTransit {
                transitIDs = append(transitIDs, events[i].ID)
            }
        }
        if len(transitIDs) == 0 {
            return events
        }
        details, err := s.transit.GetByEventIDs(ctx, s.queries, transitIDs)
        if err != nil {
            slog.WarnContext(ctx, "failed to load transit_details", "error", err)
            return events
        }
        for i := range events {
            if events[i].Category == domain.CategoryTransit {
                events[i].Transit = details[events[i].ID]
            }
        }
        return events
    }
    ```
  - [x] 4.3 In `Create` method, add the transit transactional path (after the lodging block, before the non-transactional fallback):
    ```go
    if event.Category == domain.CategoryTransit && event.Transit != nil {
        tx, txErr := s.db.Begin(ctx)
        if txErr != nil {
            return fmt.Errorf("beginning transaction: %w", txErr)
        }
        defer func() { _ = tx.Rollback(ctx) }()

        txq := sqlcgen.New(tx)
        row, txErr := txq.CreateEvent(ctx, params)
        if txErr != nil {
            return fmt.Errorf("inserting event: %w", txErr)
        }

        transitDetails := event.Transit
        *event = eventRowToDomain(&row)

        td, txErr := s.transit.Create(ctx, txq, event.ID, transitDetails)
        if txErr != nil {
            return txErr
        }
        event.Transit = td

        return tx.Commit(ctx)
    }
    ```
  - [x] 4.4 In `Update` method, add the transit transactional path (after the lodging block, before the non-transactional fallback):
    ```go
    if updated.Category == domain.CategoryTransit && updated.Transit != nil {
        tx, txErr := s.db.Begin(ctx)
        if txErr != nil {
            return nil, fmt.Errorf("beginning transaction: %w", txErr)
        }
        defer func() { _ = tx.Rollback(ctx) }()

        txq := sqlcgen.New(tx)
        row, txErr := txq.UpdateEvent(ctx, toUpdateEventParams(id, updated))
        if txErr != nil {
            return nil, fmt.Errorf("updating event: %w", txErr)
        }
        result := eventRowToDomain(&row)

        td, txErr := s.transit.Update(ctx, txq, id, updated.Transit)
        if txErr != nil {
            return nil, txErr
        }
        result.Transit = td

        if txErr = tx.Commit(ctx); txErr != nil {
            return nil, fmt.Errorf("committing transaction: %w", txErr)
        }
        return &result, nil
    }
    ```
    **Note:** Check whether `event_store.go` uses a `toUpdateEventParams` helper (added in Story 1.5 refactor) or inline `sqlcgen.UpdateEventParams{}` struct. Match the existing pattern exactly.
  - [x] 4.5 Update `GetByID` to also load transit details (after the existing flight and lodging loading blocks):
    ```go
    if event.Category == domain.CategoryTransit {
        events := s.loadTransitDetails(ctx, []domain.Event{event})
        event = events[0]
    }
    ```
  - [x] 4.6 Update `ListByTrip` and `ListByTripAndDate` to call `loadTransitDetails` after `loadLodgingDetails`:
    ```go
    events = s.loadFlightDetails(ctx, events)
    events = s.loadLodgingDetails(ctx, events)
    return s.loadTransitDetails(ctx, events), nil
    ```

### Service Updates

- [x] Task 5: Update service layer for transit events (AC: #2, #3)
  - [x] 5.1 Add `TransitDetails *domain.TransitDetails` to `CreateEventInput` in `internal/service/event.go`:
    ```go
    type CreateEventInput struct {
        // ... existing fields ...
        TransitDetails *domain.TransitDetails // nil for non-transit events
    }
    ```
    Run `fieldalignment -fix ./internal/service/...` if the linter complains about struct field order.
  - [x] 5.2 In `EventService.Create`, after the existing lodging details assignment block, add:
    ```go
    if input.Category == domain.CategoryTransit {
        event.Transit = input.TransitDetails
        if event.Transit == nil {
            event.Transit = &domain.TransitDetails{} // empty details are valid
        }
    }
    ```
    Place this just before `s.repo.Create(ctx, event)`.
  - [x] 5.3 Add `TransitDetails *domain.TransitDetails` to `UpdateEventInput`:
    ```go
    type UpdateEventInput struct {
        // ... existing fields ...
        TransitDetails *domain.TransitDetails // nil means "don't change transit details"
    }
    ```
  - [x] 5.4 In the `Update` updater closure, pass through transit details (alongside the existing flight and lodging blocks):
    ```go
    if input.TransitDetails != nil {
        event.Transit = input.TransitDetails
    }
    ```
    No business-rule validation needed for transit fields (no time relationships to enforce unlike lodging check-in/check-out).

### Handler Updates

- [x] Task 6: Update handler to parse and route transit form data (AC: #1, #2, #3)
  - [x] 6.1 Add transit fields to `EventFormData` in `internal/handler/event.go`:
    ```go
    type EventFormData struct {
        // ... existing fields ...
        // Transit-specific
        Origin        string
        Destination   string
        TransportMode string
    }
    ```
    Run `fieldalignment -fix ./internal/handler/...` if the linter complains about struct field order.
  - [x] 6.2 Add `parseTransitDetails` helper in `internal/handler/event.go`:
    ```go
    func parseTransitDetails(formData *EventFormData) *domain.TransitDetails {
        return &domain.TransitDetails{
            Origin:        formData.Origin,
            Destination:   formData.Destination,
            TransportMode: formData.TransportMode,
        }
    }
    ```
  - [x] 6.3 In `Create` handler, parse transit-specific form values (alongside the existing flight and lodging parsing blocks):
    ```go
    formData.Origin        = r.FormValue("origin")
    formData.Destination   = r.FormValue("destination")
    formData.TransportMode = r.FormValue("transport_mode")
    ```
  - [x] 6.4 In `Create` handler, build `TransitDetails` and pass to service (alongside the flight and lodging blocks):
    ```go
    var transitDetails *domain.TransitDetails
    if category == string(domain.CategoryTransit) {
        transitDetails = parseTransitDetails(&formData)
    }

    input := &service.CreateEventInput{
        // ... existing fields ...
        TransitDetails: transitDetails,
    }
    ```
  - [x] 6.5 In `Update` handler, parse transit fields and build `TransitDetails`:
    ```go
    formData.Origin        = r.FormValue("origin")
    formData.Destination   = r.FormValue("destination")
    formData.TransportMode = r.FormValue("transport_mode")

    var transitDetails *domain.TransitDetails
    if event.Category == domain.CategoryTransit {
        transitDetails = parseTransitDetails(&formData)
    }
    input := &service.UpdateEventInput{
        // ... existing fields ...
        TransitDetails: transitDetails,
    }
    ```

### Template — Enable Transit in TypeSelector

- [x] Task 7: Enable Transit in the creation form (AC: #1)
  - [x] 7.1 In `internal/handler/event_form.templ`, update `enabledCategories`:
    ```go
    var enabledCategories = map[string]bool{
        "activity": true,
        "food":     true,
        "flight":   true,
        "lodging":  true,
        "transit":  true, // ADD THIS
    }
    ```
  - [x] 7.2 Create `internal/handler/transit_form.templ` for transit-specific form fields. Follow the exact same pattern as `internal/handler/lodging_form.templ`:
    ```templ
    package handler

    import "github.com/simopzz/traccia/internal/domain"

    func transitDataFromDomain(event domain.Event) EventFormData {
        if event.Transit == nil {
            return EventFormData{}
        }
        return EventFormData{
            Origin:        event.Transit.Origin,
            Destination:   event.Transit.Destination,
            TransportMode: event.Transit.TransportMode,
        }
    }

    // TransitFormFields renders transit-specific form inputs.
    // Used in both the create sheet and the full-page fallback.
    templ TransitFormFields(data EventFormData) {
        <div x-show="selected === 'transit'" class="mb-4 border-t-2 border-slate-100 pt-4">
            <p class="text-xs font-bold uppercase tracking-wide text-purple-700 mb-3">Transit Details</p>
            <div class="grid grid-cols-2 gap-3 mb-3">
                <div>
                    <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">From</label>
                    <input
                        type="text"
                        name="origin"
                        value={ data.Origin }
                        placeholder="e.g. Shibuya Station"
                        class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm focus:outline-none focus:border-brand focus:shadow-[2px_2px_0px_0px_#008080]"
                    />
                </div>
                <div>
                    <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">To</label>
                    <input
                        type="text"
                        name="destination"
                        value={ data.Destination }
                        placeholder="e.g. Asakusa Station"
                        class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm focus:outline-none focus:border-brand focus:shadow-[2px_2px_0px_0px_#008080]"
                    />
                </div>
            </div>
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Mode</label>
                <input
                    type="text"
                    name="transport_mode"
                    value={ data.TransportMode }
                    placeholder="e.g. Metro, Bus, Walk, Taxi"
                    list="transport-mode-options"
                    class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm focus:outline-none focus:border-brand focus:shadow-[2px_2px_0px_0px_#008080]"
                />
                <datalist id="transport-mode-options">
                    <option value="Metro"/>
                    <option value="Bus"/>
                    <option value="Train"/>
                    <option value="Taxi"/>
                    <option value="Walk"/>
                    <option value="Ferry"/>
                    <option value="Car"/>
                    <option value="Tram"/>
                </datalist>
            </div>
        </div>
    }
    ```
    **Note on TransportMode UX:** The architecture specifies `transport_mode TEXT` as a free-form field. A `<datalist>` provides common suggestions while still allowing any value — this is the right approach for a free-text field with likely values. Duration default for transit is ~30min (per UX spec).
  - [x] 7.3 Add `@TransitFormFields(data)` call in `EventCreateForm` in `event_form.templ`. Place it in the same group as `@FlightFormFields(data)` and `@LodgingFormFields(data)`.
  - [x] 7.4 Add `@TransitFormFields(data)` call in `EventNewPage` (the full-page fallback form), also in the same group.
  - [x] 7.5 Run `just generate` after creating `transit_form.templ`.

### Template — TransitCardContent

- [x] Task 8: Create transit card display component (AC: #2)
  - [x] 8.1 Create `internal/handler/transit_card.templ`:
    ```templ
    package handler

    import "github.com/simopzz/traccia/internal/domain"

    // TransitCardContent renders transit-specific detail in an expanded event card.
    // Called from EventTimelineItem when event.Category == domain.CategoryTransit.
    templ TransitCardContent(td *domain.TransitDetails) {
        if td == nil {
            return
        }
        <div class="mt-3 pt-3 border-t border-slate-100 space-y-2">
            if td.Origin != "" || td.Destination != "" {
                <div class="flex items-center gap-2 text-sm">
                    if td.Origin != "" {
                        <span class="text-xs text-slate-500 font-mono">{ td.Origin }</span>
                    }
                    if td.Origin != "" && td.Destination != "" {
                        <span class="text-xs text-slate-400">→</span>
                    }
                    if td.Destination != "" {
                        <span class="text-xs text-slate-500 font-mono">{ td.Destination }</span>
                    }
                </div>
            }
            if td.TransportMode != "" {
                <div class="text-xs text-slate-500">
                    <span class="font-medium text-slate-600">Mode:</span>
                    <span class="ml-1">{ td.TransportMode }</span>
                </div>
            }
        </div>
    }
    ```
  - [x] 8.2 Run `just generate` after creating this file.

### Template — EventTimelineItem Updates for Transit

- [x] Task 9: Update EventTimelineItem to render transit details in view and edit modes (AC: #2, #3)
  - [x] 9.1 In `internal/handler/event.templ`, in the **view mode** section (`x-show="!editing"`), add after the lodging block:
    ```templ
    if event.Category == domain.CategoryTransit {
        @TransitCardContent(event.Transit)
    }
    ```
  - [x] 9.2 In the **edit mode** section (`x-show="editing"`), add transit inline edit fields after the lodging edit block:
    ```templ
    if event.Category == domain.CategoryTransit {
        <div class="mb-3 pt-3 border-t-2 border-slate-100">
            <p class="text-xs font-bold uppercase tracking-wide text-purple-700 mb-3">Transit Details</p>
            <div class="grid grid-cols-2 gap-3 mb-3">
                <div>
                    <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1">From</label>
                    if props != nil {
                        <input type="text" name="origin" value={ props.FormValues.Origin }
                            placeholder="e.g. Shibuya Station"
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm focus:outline-none focus:border-brand"/>
                    } else {
                        <input type="text" name="origin" value={ transitDataFromDomain(event).Origin }
                            placeholder="e.g. Shibuya Station"
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm focus:outline-none focus:border-brand"/>
                    }
                </div>
                <div>
                    <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1">To</label>
                    if props != nil {
                        <input type="text" name="destination" value={ props.FormValues.Destination }
                            placeholder="e.g. Asakusa Station"
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm focus:outline-none focus:border-brand"/>
                    } else {
                        <input type="text" name="destination" value={ transitDataFromDomain(event).Destination }
                            placeholder="e.g. Asakusa Station"
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm focus:outline-none focus:border-brand"/>
                    }
                </div>
            </div>
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1">Mode</label>
                if props != nil {
                    <input type="text" name="transport_mode" value={ props.FormValues.TransportMode }
                        placeholder="e.g. Metro, Bus, Walk"
                        class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm focus:outline-none focus:border-brand"/>
                } else {
                    <input type="text" name="transport_mode" value={ transitDataFromDomain(event).TransportMode }
                        placeholder="e.g. Metro, Bus, Walk"
                        class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm focus:outline-none focus:border-brand"/>
                }
            </div>
        </div>
    }
    ```
    Transit edit fields are shown via server-side Go condition (category is fixed at creation, same pattern as flight/lodging).
  - [x] 9.3 In `event.templ`, find the full-page fallback category `<select>` and remove the `disabled` attribute from the transit option:
    ```templ
    // Change from:
    <option value="transit" disabled>Transit (coming soon)</option>
    // Change to:
    <option value="transit">Transit</option>
    ```

### Seed Command

- [x] Task 10: Update `cmd/seed/main.go` to support transit events (AC: #2)
  - [x] 10.1 Update `NewEventStore` call in `cmd/seed/main.go` (alongside existing `cmd/app/main.go`):
    ```go
    flightDetailsStore := repository.NewFlightDetailsStore()
    lodgingDetailsStore := repository.NewLodgingDetailsStore()
    transitDetailsStore := repository.NewTransitDetailsStore()
    eventStore := repository.NewEventStore(pool, flightDetailsStore, lodgingDetailsStore, transitDetailsStore)
    ```
  - [x] 10.2 Remove `domain.CategoryTransit` from the random categories slice (it produces bare base events without details):
    ```go
    categories := []domain.EventCategory{
        domain.CategoryActivity,
        domain.CategoryFood,
        // CategoryTransit removed — created explicitly via createTransitEvent
        // CategoryLodging removed — created explicitly via createLodgingEvent
    }
    ```
  - [x] 10.3 Add a `createTransitEvent` function:
    ```go
    func createTransitEvent(ctx context.Context, eventService *service.EventService, tripID int, startTime time.Time) error {
        origins := []string{"Shibuya Station", "Shinjuku Station", "Tokyo Station", "Kyoto Station", "Osaka Station"}
        destinations := []string{"Asakusa Station", "Akihabara", "Ginza", "Harajuku", "Ueno"}
        modes := []string{"Metro", "Train", "Bus", "Walk", "Taxi"}

        details := &domain.TransitDetails{
            Origin:        origins[rand.Intn(len(origins))],
            Destination:   destinations[rand.Intn(len(destinations))],
            TransportMode: modes[rand.Intn(len(modes))],
        }

        input := &service.CreateEventInput{
            TripID:         tripID,
            Title:          fmt.Sprintf("%s to %s", details.Origin, details.Destination),
            Category:       domain.CategoryTransit,
            StartTime:      startTime,
            EndTime:        startTime.Add(30 * time.Minute),
            Notes:          "Generated by seeder",
            TransitDetails: details,
        }

        _, err := eventService.Create(ctx, input)
        return err
    }
    ```
  - [x] 10.4 In `seedEvents`, call `createTransitEvent` once per day (or per trip) to seed realistic transit data. Mirror the lodging event seeding pattern.

### DI Wiring

- [x] Task 11: Update `cmd/app/main.go` to wire TransitDetailsStore (AC: #2)
  - [x] 11.1 Add `transitDetailsStore` and update `NewEventStore` call:
    ```go
    // Repositories
    tripStore := repository.NewTripStore(pool)
    flightDetailsStore := repository.NewFlightDetailsStore()
    lodgingDetailsStore := repository.NewLodgingDetailsStore()
    transitDetailsStore := repository.NewTransitDetailsStore()
    eventStore := repository.NewEventStore(pool, flightDetailsStore, lodgingDetailsStore, transitDetailsStore)
    ```

### Testing

- [x] Task 12: Write service tests and verify build (AC: #1, #2, #3, #4)
  - [x] 12.1 Service test: `Create` transit event → `event.Transit` populated with correct field values
    - Mock `EventRepository.Create` to capture the `*domain.Event` passed; verify `event.Transit != nil`, `Origin`, `Destination`, `TransportMode` match input.
  - [x] 12.2 Service test: `Create` transit event with nil `TransitDetails` → defaults to empty `TransitDetails{}` (no nil panic).
  - [x] 12.3 Service test: `Update` transit event → `event.Transit` updated to new values.
    - Input: existing event with `Transit.Origin = "A"`, update with `TransitDetails{Origin: "B"}`.
    - Verify updater sets `event.Transit.Origin = "B"`.
  - [x] 12.4 Service test: `Update` non-transit event with nil `TransitDetails` → no change to `event.Transit`.
  - [x] 12.5 Run `just test` — all passing, no races.
  - [x] 12.6 Run `just lint` — zero violations. Apply `fieldalignment -fix` if struct field ordering violations appear.
  - [x] 12.7 Run `just build` — binary compiles.
  - [x] 12.8 Manual smoke tests:
    - Click "Add Event" → select Transit in TypeSelector → transit fields appear (From, To, Mode with datalist); flight/lodging fields hidden.
    - Fill in origin, destination, transport mode → Save → event appears in timeline with purple Bus icon.
    - Expand card → `TransitCardContent` shows origin → destination and mode.
    - Click Edit → transit fields appear pre-filled → modify → Save → card updates.
    - Click Delete → event removed → Undo → event restored with transit details intact.
    - Create Activity/Food → no transit fields appear.
    - Verify seed command: `just seed` creates transit events alongside lodging events.

## Dev Notes

### Architecture: No Migration Needed

The `transit_details` table already exists in `migrations/001_initial.up.sql`:
```sql
CREATE TABLE transit_details (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL UNIQUE REFERENCES events(id) ON DELETE CASCADE,
    origin TEXT DEFAULT '',
    destination TEXT DEFAULT '',
    transport_mode TEXT DEFAULT ''
);
```
**No new migration is needed for Story 1.6.** The schema is already correct.

### Critical: All TEXT Fields — Simpler than Lodging

All three transit fields (`origin`, `destination`, `transport_mode`) are `TEXT DEFAULT ''` without `NOT NULL`. sqlc will generate `pgtype.Text` for these fields. Use:
- Write path: `toPgText(td.Origin)` — helper already exists in `internal/repository/helpers.go`
- Read path: `row.Origin.String` — the `.String` accessor on `pgtype.Text`

No new pgtype helpers are needed. This is strictly simpler than the lodging story (no optional timestamps).

### Critical: EventStore Signature Change — Three Callsites

`NewEventStore` gains a fourth parameter: `transitStore *TransitDetailsStore`. This breaks existing calls in:
1. `cmd/app/main.go` (Task 11)
2. `cmd/seed/main.go` (Task 10)
3. **Check for test files** that construct `EventStore` directly — search for `NewEventStore` across the codebase and update all callsites before running `just build`.

### Critical: loadTransitDetails Chained After loadLodgingDetails

`ListByTrip` and `ListByTripAndDate` must chain all three loaders:
```go
events = s.loadFlightDetails(ctx, events)
events = s.loadLodgingDetails(ctx, events)
return s.loadTransitDetails(ctx, events), nil
```
All three must be called to avoid nil detail fields on respective category events.

### Critical: Soft Delete + CASCADE — Same as Flight/Lodging

- Soft delete (`set deleted_at = NOW()`) only updates the base event — `transit_details` row stays in DB.
- Restore (`set deleted_at = NULL`) — `transit_details` row still present. Undo flow works correctly.
- Hard delete (trip deletion via CASCADE) → `transit_details` deleted by DB CASCADE.

No code change needed for delete/restore handlers.

### Pattern: EventStore Create/Update Helper Functions

Story 1.5 refactored `event_store.go` to extract `toCreateEventParams` and `toUpdateEventParams` helpers to eliminate duplication across the flight/lodging transactional blocks. Verify these helpers exist and use them in the transit block — do NOT inline the full `sqlcgen.CreateEventParams{}` struct if the helpers already exist.

```go
// Use existing helpers, not inline structs:
row, txErr := txq.CreateEvent(ctx, params)  // params = toCreateEventParams(event, position)
row, txErr := txq.UpdateEvent(ctx, toUpdateEventParams(id, updated))
```

### Pattern: Transit Color Convention

Transit color is `bg-purple-100 text-purple-700` (already set in `event.templ` `categoryBgColors` map). Use `text-purple-700` for the "Transit Details" section header in `transit_form.templ` and the edit mode header in `event.templ`.

### Pattern: `transitDataFromDomain` for Edit Mode

`transit_form.templ` defines `transitDataFromDomain(event domain.Event) EventFormData` — this is called in the edit mode section of `event.templ` (same as `lodgingDataFromDomain` and `flightDataFromDomain`). It projects `event.Transit` fields into `EventFormData` for the inline edit inputs when `props == nil`.

### Pattern: All Established Patterns from Stories 1.1–1.5 Apply

- **pgtype helpers**: `toPgText()` in `helpers.go` — no new helpers needed for transit
- **Error mapping**: `domain.ErrNotFound` → 404, `domain.ErrInvalidInput` → 422, else → 500
- **HTMX success**: Day-level swap via `HX-Retarget: #day-{date}` + `HX-Reswap: outerHTML` — unchanged
- **Sheet close**: `HX-Trigger: {"close-sheet": true}` on successful create — unchanged
- **`just generate`**: Run after every `.templ` or `.sql` file change
- **Never edit generated files**: `sqlcgen/*.go`, `*_templ.go`
- **fieldalignment linter**: Run `fieldalignment -fix` on changed packages if struct ordering violations appear

### Project Structure Notes

New files:
- `internal/repository/sql/transit_details.sql`
- `internal/repository/sqlcgen/transit_details.sql.go` (generated)
- `internal/repository/transit_details_store.go`
- `internal/handler/transit_card.templ`
- `internal/handler/transit_card_templ.go` (generated)
- `internal/handler/transit_form.templ`
- `internal/handler/transit_form_templ.go` (generated)

Modified files:
- `internal/domain/models.go` — add `TransitDetails` struct, `Transit *TransitDetails` on `Event`
- `internal/repository/event_store.go` — add `transit` field, update constructor, Create/Update transactional paths, GetByID/ListByTrip/ListByTripAndDate loading
- `internal/service/event.go` — add `TransitDetails` to inputs, populate in Create/Update
- `internal/handler/event.go` — add `Origin`/`Destination`/`TransportMode` to `EventFormData`, add `parseTransitDetails`, update Create/Update handlers
- `internal/handler/event_form.templ` — enable transit in `enabledCategories`, add `@TransitFormFields` call
- `internal/handler/event.templ` — add `TransitCardContent` call in view mode, transit edit fields in edit mode, remove `disabled` from transit select option
- `cmd/app/main.go` — wire `TransitDetailsStore`, update `NewEventStore` call
- `cmd/seed/main.go` — wire `TransitDetailsStore`, add `createTransitEvent`, remove transit from random categories
- `internal/service/event_test.go` — add transit service tests

Do NOT modify:
- `migrations/` — no new migration needed
- `internal/repository/sqlcgen/*` — always regenerated
- `internal/handler/*_templ.go` — always regenerated

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.6] — Acceptance criteria: origin, destination, transport mode, transactional persist, TransitCardContent, CASCADE delete
- [Source: _bmad-output/planning-artifacts/architecture.md#Data Architecture] — Base + Detail Tables: `transit_details` with 1:1 event_id FK; CASCADE DELETE; detail stores handle write-path via DBTX
- [Source: _bmad-output/planning-artifacts/architecture.md#Structure Patterns] — Event type code organization: 8-step layer ordering (migration → sql → domain → repository → service → handler → templates → generate)
- [Source: _bmad-output/planning-artifacts/architecture.md#Enforcement Guidelines] — Rule 7: add new event types as separate files, never by modifying existing type components
- [Source: migrations/001_initial.up.sql] — transit_details: all TEXT DEFAULT '', no NOT NULL constraints
- [Source: _bmad-output/implementation-artifacts/1-5-lodging-events.md] — Full pattern reference: EventStore helpers (toCreateEventParams/toUpdateEventParams), loadDetails batch loader, store adapter, form.templ separation, fieldalignment linter, soft-delete + CASCADE interaction
- [Source: internal/handler/event.templ] — Transit color: `bg-purple-100 text-purple-700`; icon: `icon.Bus`; "coming soon" option to be enabled

## Change Log

- 2026-02-23: Story 1.6 implemented — transit events with origin, destination, transport mode; full create/edit/view/seed support

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6[1m] (2026-02-23)

### Debug Log References

### Completion Notes List

- Implemented transit events end-to-end following Story 1.6 spec
- SQL queries generated via sqlc; `TransitDetail` struct uses `pgtype.Text` for all three fields
- `TransitDetailsStore` created following `LodgingDetailsStore` pattern
- `EventStore` wired with `TransitDetailsStore` (4th arg); Create/Update transactional paths added
- `GetByID`, `ListByTrip`, `ListByTripAndDate` all chain `loadTransitDetails` after existing loaders
- Service `CreateEventInput` and `UpdateEventInput` gain `TransitDetails` field; nil defaults to `&TransitDetails{}`
- Handler parses `origin`, `destination`, `transport_mode` form fields in Create and Update
- `transit_form.templ` provides `TransitFormFields` with datalist suggestions and `transitDataFromDomain` helper
- `transit_card.templ` provides `TransitCardContent` for view mode
- `event.templ` inline edit mode gains transit edit fields; full-page fallback select option enabled
- `event_form.templ` enables transit in TypeSelector
- `cmd/app/main.go` and `cmd/seed/main.go` wired with `NewTransitDetailsStore`
- Seed command adds explicit `createTransitEvent`; `CategoryTransit` removed from random pool
- 5 service tests pass (including Transit duration); all existing tests pass; zero lint violations; build clean
- **AI-Review Fixes (Round 1)**:
  - Added 30-minute default duration for transit category in `EventService`.
  - Refactored `EventStore.Update` to use shared `params` variable in transit path.
  - Added `SuggestDefaults` test case for Transit events.
- **AI-Review Fixes (Round 2)**:
  - Added `pgx.ErrNoRows → domain.ErrNotFound` mapping in `TransitDetailsStore.Update` (matches `GetByEventID` pattern).
  - Created `transit_details_store_test.go` with `Test_transitRowToDomain` (matches `lodging_details_store_test.go` parity).
  - Updated File List to document `lodging_details_store_test.go` and `static/css/app.css`.

### File List

**New files:**
- `internal/repository/sql/transit_details.sql`
- `internal/repository/sqlcgen/transit_details.sql.go` (generated)
- `internal/repository/transit_details_store.go`
- `internal/repository/transit_details_store_test.go`
- `internal/handler/transit_card.templ`
- `internal/handler/transit_card_templ.go` (generated)
- `internal/handler/transit_form.templ`
- `internal/handler/transit_form_templ.go` (generated)

**Modified files:**
- `internal/domain/models.go`
- `internal/repository/event_store.go`
- `internal/service/event.go`
- `internal/handler/event.go`
- `internal/handler/event_form.templ`
- `internal/handler/event_form_templ.go` (generated)
- `internal/handler/event.templ`
- `internal/handler/event_templ.go` (generated)
- `cmd/app/main.go`
- `cmd/seed/main.go`
- `internal/service/event_test.go`
- `internal/repository/lodging_details_store_test.go` (lodgingRowToDomain test added for parity)
- `static/css/app.css` (Tailwind rebuild — new transit/purple classes)
