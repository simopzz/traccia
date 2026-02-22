# Story 1.4: Flight Events

Status: review

## Story

As a traveler,
I want to add Flight events with airline, flight number, airports, terminals, and gates,
so that I can capture all flight details in my trip timeline with the right level of specificity.

## Acceptance Criteria

1. **Given** the user selects Flight from the TypeSelector, **When** the form morphs, **Then** additional fields appear: airline, flight number, departure airport, arrival airport, departure terminal, arrival terminal, departure gate, arrival gate, booking reference.

2. **Given** the user submits a valid Flight event, **When** the event is saved, **Then** the `flight_details` are persisted in the `flight_details` table within the same transaction as the base event, **And** the `FlightCardContent` displays flight-specific metadata (airline, flight number, airports).

3. **Given** a user edits a Flight event, **When** they modify flight-specific fields, **Then** the `flight_details` are updated and the card reflects the changes.

4. **Given** a user deletes a Flight event, **When** the event is deleted, **Then** both the base event and `flight_details` are removed (CASCADE — no explicit code needed, DB handles it).

## Tasks / Subtasks

### SQL Queries for flight_details

- [x] Task 1: Add sqlc queries for flight_details write and read paths (AC: #1, #2, #3)
  - [x] 1.1 Create `internal/repository/sql/flight_details.sql` with the following queries:
    ```sql
    -- name: CreateFlightDetails :one
    INSERT INTO flight_details (event_id, airline, flight_number, departure_airport, arrival_airport, departure_terminal, arrival_terminal, departure_gate, arrival_gate, booking_reference)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    RETURNING *;

    -- name: GetFlightDetailsByEventID :one
    SELECT * FROM flight_details WHERE event_id = $1;

    -- name: UpdateFlightDetails :one
    UPDATE flight_details
    SET airline = $2, flight_number = $3, departure_airport = $4, arrival_airport = $5,
        departure_terminal = $6, arrival_terminal = $7, departure_gate = $8, arrival_gate = $9,
        booking_reference = $10
    WHERE event_id = $1
    RETURNING *;
    ```
  - [x] 1.2 Run `just generate` — generates `internal/repository/sqlcgen/flight_details.sql.go`
  - [x] 1.3 Verify generated file: `sqlcgen.CreateFlightDetailsParams`, `sqlcgen.UpdateFlightDetailsParams`, `sqlcgen.FlightDetail` structs exist

### Domain Model

- [x] Task 2: Add FlightDetails to domain (AC: #2)
  - [x] 2.1 Add to `internal/domain/models.go`:
    ```go
    type FlightDetails struct {
        Airline           string
        FlightNumber      string
        DepartureAirport  string
        ArrivalAirport    string
        DepartureTerminal string
        ArrivalTerminal   string
        DepartureGate     string
        ArrivalGate       string
        BookingReference  string
        EventID           int
        ID                int
    }
    ```
  - [x] 2.2 Add `Flight *FlightDetails` field to the existing `Event` struct:
    ```go
    type Event struct {
        // ... all existing fields stay unchanged ...
        Flight *FlightDetails // nil for all non-flight events
    }
    ```
    This field is populated by the repository layer for flight events. Templates check `event.Flight != nil` before accessing.

### Repository — FlightDetailsStore

- [x] Task 3: Create FlightDetailsStore adapter (AC: #2, #3)
  - [x] 3.1 Create `internal/repository/flight_details_store.go`:
    ```go
    package repository

    import (
        "context"
        "errors"

        "github.com/jackc/pgx/v5"

        "github.com/simopzz/traccia/internal/domain"
        "github.com/simopzz/traccia/internal/repository/sqlcgen"
    )

    type FlightDetailsStore struct{}

    func NewFlightDetailsStore() *FlightDetailsStore {
        return &FlightDetailsStore{}
    }

    // Create inserts flight_details within the caller's transaction (q is tx-scoped).
    func (s *FlightDetailsStore) Create(ctx context.Context, q *sqlcgen.Queries, eventID int, fd *domain.FlightDetails) (*domain.FlightDetails, error) {
        row, err := q.CreateFlightDetails(ctx, sqlcgen.CreateFlightDetailsParams{
            EventID:           int32(eventID),
            Airline:           toPgText(fd.Airline),
            FlightNumber:      toPgText(fd.FlightNumber),
            DepartureAirport:  toPgText(fd.DepartureAirport),
            ArrivalAirport:    toPgText(fd.ArrivalAirport),
            DepartureTerminal: toPgText(fd.DepartureTerminal),
            ArrivalTerminal:   toPgText(fd.ArrivalTerminal),
            DepartureGate:     toPgText(fd.DepartureGate),
            ArrivalGate:       toPgText(fd.ArrivalGate),
            BookingReference:  toPgText(fd.BookingReference),
        })
        if err != nil {
            return nil, fmt.Errorf("inserting flight_details for event %d: %w", eventID, err)
        }
        result := flightRowToDomain(&row)
        return &result, nil
    }

    // GetByEventID loads flight_details. Returns domain.ErrNotFound if the row doesn't exist.
    func (s *FlightDetailsStore) GetByEventID(ctx context.Context, q *sqlcgen.Queries, eventID int) (*domain.FlightDetails, error) {
        row, err := q.GetFlightDetailsByEventID(ctx, int32(eventID))
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                return nil, domain.ErrNotFound
            }
            return nil, fmt.Errorf("fetching flight_details for event %d: %w", eventID, err)
        }
        result := flightRowToDomain(&row)
        return &result, nil
    }

    // Update updates existing flight_details. Uses the caller-provided queries (can be tx-scoped).
    func (s *FlightDetailsStore) Update(ctx context.Context, q *sqlcgen.Queries, eventID int, fd *domain.FlightDetails) (*domain.FlightDetails, error) {
        row, err := q.UpdateFlightDetails(ctx, sqlcgen.UpdateFlightDetailsParams{
            EventID:           int32(eventID),
            Airline:           toPgText(fd.Airline),
            FlightNumber:      toPgText(fd.FlightNumber),
            DepartureAirport:  toPgText(fd.DepartureAirport),
            ArrivalAirport:    toPgText(fd.ArrivalAirport),
            DepartureTerminal: toPgText(fd.DepartureTerminal),
            ArrivalTerminal:   toPgText(fd.ArrivalTerminal),
            DepartureGate:     toPgText(fd.DepartureGate),
            ArrivalGate:       toPgText(fd.ArrivalGate),
            BookingReference:  toPgText(fd.BookingReference),
        })
        if err != nil {
            return nil, fmt.Errorf("updating flight_details for event %d: %w", eventID, err)
        }
        result := flightRowToDomain(&row)
        return &result, nil
    }

    func flightRowToDomain(row *sqlcgen.FlightDetail) domain.FlightDetails {
        return domain.FlightDetails{
            ID:                int(row.ID),
            EventID:           int(row.EventID),
            Airline:           row.Airline.String,
            FlightNumber:      row.FlightNumber.String,
            DepartureAirport:  row.DepartureAirport.String,
            ArrivalAirport:    row.ArrivalAirport.String,
            DepartureTerminal: row.DepartureTerminal.String,
            ArrivalTerminal:   row.ArrivalTerminal.String,
            DepartureGate:     row.DepartureGate.String,
            ArrivalGate:       row.ArrivalGate.String,
            BookingReference:  row.BookingReference.String,
        }
    }
    ```
  - [x] 3.2 Add `"fmt"` import to the file (alongside `"context"`, `"errors"`, pgx, domain, sqlcgen imports).

### Repository — EventStore Updates

- [x] Task 4: Update EventStore to store `db *pgxpool.Pool` and `flight *FlightDetailsStore`, enable transactional creates and detail loading (AC: #2, #3)
  - [x] 4.1 Update `EventStore` struct and constructor in `internal/repository/event_store.go`:
    ```go
    type EventStore struct {
        db      *pgxpool.Pool
        queries *sqlcgen.Queries
        flight  *FlightDetailsStore
    }

    func NewEventStore(db *pgxpool.Pool, flightStore *FlightDetailsStore) *EventStore {
        return &EventStore{
            db:      db,
            queries: sqlcgen.New(db),
            flight:  flightStore,
        }
    }
    ```
    Add `"github.com/jackc/pgx/v5/pgxpool"` import (it was passed in before but not stored).
  - [x] 4.2 Update `Create` method to use a transaction for flight events:
    ```go
    func (s *EventStore) Create(ctx context.Context, event *domain.Event) error {
        maxPos, err := s.queries.GetMaxPositionByTripAndDate(ctx, sqlcgen.GetMaxPositionByTripAndDateParams{
            TripID:    int32(event.TripID),
            EventDate: toPgDate(event.EventDate),
        })
        if err != nil {
            return err
        }
        position := maxPos + 1000
        if event.Position > 0 {
            position = int32(event.Position)
        }

        if event.Category == domain.CategoryFlight && event.Flight != nil {
            // Transactional: insert base event + flight_details atomically
            tx, err := s.db.Begin(ctx)
            if err != nil {
                return fmt.Errorf("beginning transaction: %w", err)
            }
            defer tx.Rollback(ctx)

            txq := sqlcgen.New(tx)
            row, err := txq.CreateEvent(ctx, sqlcgen.CreateEventParams{
                TripID: int32(event.TripID), EventDate: toPgDate(event.EventDate),
                Title: event.Title, Category: string(event.Category),
                Location: toPgText(event.Location), Latitude: toPgFloat8(event.Latitude),
                Longitude: toPgFloat8(event.Longitude), StartTime: toPgTimestamptz(event.StartTime),
                EndTime: toPgTimestamptz(event.EndTime), Pinned: toPgBool(event.Pinned),
                Position: position, Notes: toPgText(event.Notes),
            })
            if err != nil {
                return fmt.Errorf("inserting event: %w", err)
            }
            *event = eventRowToDomain(&row)

            fd, err := s.flight.Create(ctx, txq, event.ID, event.Flight)
            if err != nil {
                return err
            }
            event.Flight = fd

            return tx.Commit(ctx)
        }

        // Non-transactional path for Activity, Food (no detail table)
        row, err := s.queries.CreateEvent(ctx, sqlcgen.CreateEventParams{
            TripID: int32(event.TripID), EventDate: toPgDate(event.EventDate),
            Title: event.Title, Category: string(event.Category),
            Location: toPgText(event.Location), Latitude: toPgFloat8(event.Latitude),
            Longitude: toPgFloat8(event.Longitude), StartTime: toPgTimestamptz(event.StartTime),
            EndTime: toPgTimestamptz(event.EndTime), Pinned: toPgBool(event.Pinned),
            Position: position, Notes: toPgText(event.Notes),
        })
        if err != nil {
            return err
        }
        *event = eventRowToDomain(&row)
        return nil
    }
    ```
  - [x] 4.3 Add `loadFlightDetails` helper and update `GetByID` and `ListByTripAndDate`:
    ```go
    // loadFlightDetails enriches a flight event with its detail row.
    // No-op for non-flight events. Errors are logged but not fatal (event still displayed).
    func (s *EventStore) loadFlightDetails(ctx context.Context, events []domain.Event) []domain.Event {
        for i := range events {
            if events[i].Category != domain.CategoryFlight {
                continue
            }
            fd, err := s.flight.GetByEventID(ctx, s.queries, events[i].ID)
            if err != nil {
                slog.WarnContext(ctx, "failed to load flight_details", "event_id", events[i].ID, "error", err)
                continue
            }
            events[i].Flight = fd
        }
        return events
    }
    ```
    - In `GetByID`: after `event := eventRowToDomain(&row)`, add:
      ```go
      if event.Category == domain.CategoryFlight {
          events := s.loadFlightDetails(ctx, []domain.Event{event})
          event = events[0]
      }
      return &event, nil
      ```
    - In `ListByTripAndDate`: after building `events` slice, call:
      ```go
      return s.loadFlightDetails(ctx, events), nil
      ```
    - In `ListByTrip`: same — call `s.loadFlightDetails(ctx, events)` before returning.
  - [x] 4.4 Update `Update` to also update flight_details when category is flight:
    ```go
    func (s *EventStore) Update(ctx context.Context, id int, updater func(*domain.Event) *domain.Event) (*domain.Event, error) {
        event, err := s.GetByID(ctx, id) // now loads Flight details for flight events
        if err != nil {
            return nil, err
        }

        updated := updater(event)

        if updated.Category == domain.CategoryFlight && updated.Flight != nil {
            tx, err := s.db.Begin(ctx)
            if err != nil {
                return nil, fmt.Errorf("beginning transaction: %w", err)
            }
            defer tx.Rollback(ctx)

            txq := sqlcgen.New(tx)
            row, err := txq.UpdateEvent(ctx, sqlcgen.UpdateEventParams{
                ID: int32(id), Title: updated.Title, Category: string(updated.Category),
                Location: toPgText(updated.Location), Latitude: toPgFloat8(updated.Latitude),
                Longitude: toPgFloat8(updated.Longitude), StartTime: toPgTimestamptz(updated.StartTime),
                EndTime: toPgTimestamptz(updated.EndTime), Pinned: toPgBool(updated.Pinned),
                Position: int32(updated.Position), EventDate: toPgDate(updated.EventDate),
                Notes: toPgText(updated.Notes),
            })
            if err != nil {
                return nil, fmt.Errorf("updating event: %w", err)
            }
            result := eventRowToDomain(&row)

            fd, err := s.flight.Update(ctx, txq, id, updated.Flight)
            if err != nil {
                return nil, err
            }
            result.Flight = fd

            if err := tx.Commit(ctx); err != nil {
                return nil, fmt.Errorf("committing transaction: %w", err)
            }
            return &result, nil
        }

        // Non-transactional for Activity, Food
        row, err := s.queries.UpdateEvent(ctx, sqlcgen.UpdateEventParams{
            ID: int32(id), Title: updated.Title, Category: string(updated.Category),
            Location: toPgText(updated.Location), Latitude: toPgFloat8(updated.Latitude),
            Longitude: toPgFloat8(updated.Longitude), StartTime: toPgTimestamptz(updated.StartTime),
            EndTime: toPgTimestamptz(updated.EndTime), Pinned: toPgBool(updated.Pinned),
            Position: int32(updated.Position), EventDate: toPgDate(updated.EventDate),
            Notes: toPgText(updated.Notes),
        })
        if err != nil {
            return nil, err
        }
        result := eventRowToDomain(&row)
        return &result, nil
    }
    ```
  - [x] 4.5 Add `"fmt"` and `"log/slog"` imports to `event_store.go` (needed for new error wrapping and loadFlightDetails warning).

### Service Updates

- [x] Task 5: Update service layer for flight events (AC: #2, #3)
  - [x] 5.1 Add flight fields to `CreateEventInput` in `internal/service/event.go`:
    ```go
    type CreateEventInput struct {
        // ... existing fields ...
        FlightDetails *domain.FlightDetails // nil for non-flight events
    }
    ```
  - [x] 5.2 In `EventService.Create`, after setting `event.Category`, populate details:
    ```go
    if input.Category == domain.CategoryFlight {
        event.Flight = input.FlightDetails
        if event.Flight == nil {
            event.Flight = &domain.FlightDetails{} // empty details are valid
        }
    }
    ```
    Place this just before `s.repo.Create(ctx, event)`.
  - [x] 5.3 Add flight fields to `UpdateEventInput`:
    ```go
    type UpdateEventInput struct {
        // ... existing fields ...
        FlightDetails *domain.FlightDetails // nil means "don't change flight details"
    }
    ```
  - [x] 5.4 In `EventService.Update` updater closure, pass through flight details:
    ```go
    return s.repo.Update(ctx, id, func(event *domain.Event) *domain.Event {
        // ... existing field updates ...
        if input.FlightDetails != nil {
            event.Flight = input.FlightDetails
        }
        return event
    })
    ```
  - [x] 5.5 Add flight duration default in `durationForCategory`:
    ```go
    case domain.CategoryFlight:
        return 3 * time.Hour
    ```

### Handler Updates

- [x] Task 6: Update handler to parse and route flight form data (AC: #1, #2, #3)
  - [x] 6.1 Add flight fields to `EventFormData` in `internal/handler/event.go`:
    ```go
    type EventFormData struct {
        // ... existing fields ...
        // Flight-specific
        Airline           string
        FlightNumber      string
        DepartureAirport  string
        ArrivalAirport    string
        DepartureTerminal string
        ArrivalTerminal   string
        DepartureGate     string
        ArrivalGate       string
        BookingReference  string
    }
    ```
  - [x] 6.2 In `Create` handler, replace the category guard (which currently rejects non-activity/food categories) with full category support. Remove these lines:
    ```go
    // REMOVE this block entirely:
    if category != "" && category != string(domain.CategoryActivity) && category != string(domain.CategoryFood) {
        formErrors["category"] = "Only Activity and Food events are currently supported"
    }
    ```
    Instead, validate that category is one of the 5 valid types (call `domain.IsValidEventCategory`):
    ```go
    if category != "" && !domain.IsValidEventCategory(domain.EventCategory(category)) {
        formErrors["category"] = "Invalid event type"
    }
    ```
  - [x] 6.3 In `Create` handler, after parsing shared fields, parse flight-specific form values:
    ```go
    formData.Airline           = r.FormValue("airline")
    formData.FlightNumber      = r.FormValue("flight_number")
    formData.DepartureAirport  = r.FormValue("departure_airport")
    formData.ArrivalAirport    = r.FormValue("arrival_airport")
    formData.DepartureTerminal = r.FormValue("departure_terminal")
    formData.ArrivalTerminal   = r.FormValue("arrival_terminal")
    formData.DepartureGate     = r.FormValue("departure_gate")
    formData.ArrivalGate       = r.FormValue("arrival_gate")
    formData.BookingReference  = r.FormValue("booking_reference")
    ```
    Do this before the `formErrors` check so values are preserved on re-render.
  - [x] 6.4 In `Create` handler, build `FlightDetails` when category is flight and pass to service:
    ```go
    var flightDetails *domain.FlightDetails
    if category == string(domain.CategoryFlight) {
        flightDetails = &domain.FlightDetails{
            Airline:           formData.Airline,
            FlightNumber:      formData.FlightNumber,
            DepartureAirport:  formData.DepartureAirport,
            ArrivalAirport:    formData.ArrivalAirport,
            DepartureTerminal: formData.DepartureTerminal,
            ArrivalTerminal:   formData.ArrivalTerminal,
            DepartureGate:     formData.DepartureGate,
            ArrivalGate:       formData.ArrivalGate,
            BookingReference:  formData.BookingReference,
        }
    }

    input := &service.CreateEventInput{
        TripID: tripID, Title: title, Category: domain.EventCategory(category),
        Location: location, StartTime: startTime, EndTime: endTime,
        Notes: notes, Pinned: pinned,
        FlightDetails: flightDetails,
    }
    ```
  - [x] 6.5 In `Update` handler, parse flight-specific form values:
    ```go
    formData.Airline           = r.FormValue("airline")
    formData.FlightNumber      = r.FormValue("flight_number")
    // ... all 9 flight fields ...
    ```
    Add these right after parsing the shared form fields (dateStr, title, location, etc.).
  - [x] 6.6 In `Update` handler, build `FlightDetails` and pass to service:
    ```go
    var flightDetails *domain.FlightDetails
    if event.Category == domain.CategoryFlight {
        flightDetails = &domain.FlightDetails{
            Airline:           formData.Airline,
            FlightNumber:      formData.FlightNumber,
            // ... remaining 7 fields ...
        }
    }
    input := &service.UpdateEventInput{
        Title: &title, Category: &category, Location: &location,
        StartTime: &startTime, EndTime: &endTime, Notes: &notes, Pinned: &pinned,
        FlightDetails: flightDetails,
    }
    ```

### Template — Enable Flight in TypeSelector

- [x] Task 7: Enable Flight in the creation form (AC: #1)
  - [x] 7.1 In `internal/handler/event_form.templ`, update `enabledCategories`:
    ```go
    var enabledCategories = map[string]bool{
        "activity": true,
        "food":     true,
        "flight":   true, // ADD THIS
    }
    ```
  - [x] 7.2 Add flight-specific form fields section to `EventCreateForm` in `event_form.templ`. Place it AFTER the Notes block and BEFORE the Pinned toggle:
    ```templ
    <!-- Flight-specific fields (Alpine.js x-show) -->
    <div x-show="selected === 'flight'" class="mb-4 border-t-2 border-slate-100 pt-4">
        <p class="text-xs font-bold uppercase tracking-wide text-sky-700 mb-3">Flight Details</p>
        <div class="grid grid-cols-2 gap-3 mb-3">
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Airline</label>
                <input type="text" name="airline" value={ data.Airline }
                    class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand focus:shadow-[2px_2px_0px_0px_#008080]"/>
            </div>
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Flight No.</label>
                <input type="text" name="flight_number" value={ data.FlightNumber }
                    class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand focus:shadow-[2px_2px_0px_0px_#008080]"/>
            </div>
        </div>
        <div class="grid grid-cols-2 gap-3 mb-3">
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">From (airport)</label>
                <input type="text" name="departure_airport" value={ data.DepartureAirport }
                    placeholder="e.g. LHR"
                    class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
            </div>
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">To (airport)</label>
                <input type="text" name="arrival_airport" value={ data.ArrivalAirport }
                    placeholder="e.g. CDG"
                    class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
            </div>
        </div>
        <div class="grid grid-cols-2 gap-3 mb-3">
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Dep. Terminal</label>
                <input type="text" name="departure_terminal" value={ data.DepartureTerminal }
                    class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
            </div>
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Arr. Terminal</label>
                <input type="text" name="arrival_terminal" value={ data.ArrivalTerminal }
                    class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
            </div>
        </div>
        <div class="grid grid-cols-2 gap-3 mb-3">
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Dep. Gate</label>
                <input type="text" name="departure_gate" value={ data.DepartureGate }
                    class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
            </div>
            <div>
                <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Arr. Gate</label>
                <input type="text" name="arrival_gate" value={ data.ArrivalGate }
                    class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
            </div>
        </div>
        <div>
            <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1.5">Booking Reference</label>
            <input type="text" name="booking_reference" value={ data.BookingReference }
                class="w-full px-3 py-2 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
        </div>
    </div>
    ```
    **Important:** The `TypeSelector` component sets `x-data="{ selected: '...' }"` on its parent wrapper, but `EventCreateForm` wraps it. The `x-show="selected === 'flight'"` on the flight section works because Alpine.js crawls up the DOM to find the nearest ancestor with `selected` in scope. Verify this works — the `x-data` from `TypeSelector` is on its own `<div>`, so the flight section must be inside that same div, OR you move the `x-data` to the `<form>`. Check the actual rendered DOM. If it doesn't scope correctly, lift the `x-data="{ selected: '...' }"` to the `<form>` element and remove it from the TypeSelector's inner div.

### Template — FlightCardContent

- [x] Task 8: Create flight card display component (AC: #2)
  - [x] 8.1 Create `internal/handler/flight_card.templ`:
    ```templ
    package handler

    import "github.com/simopzz/traccia/internal/domain"

    // FlightCardContent renders the flight-specific detail section inside an expanded event card.
    // It is called from EventTimelineItem when event.Category == domain.CategoryFlight.
    templ FlightCardContent(fd *domain.FlightDetails) {
        if fd == nil {
            return
        }
        <div class="mt-3 pt-3 border-t border-slate-100 space-y-2">
            <!-- Route summary -->
            <div class="flex items-center gap-2 text-sm font-bold font-mono text-slate-800">
                <span>{ fd.DepartureAirport }</span>
                <span class="text-slate-400">→</span>
                <span>{ fd.ArrivalAirport }</span>
                if fd.FlightNumber != "" {
                    <span class="ml-auto text-xs font-mono text-sky-700 bg-sky-50 border border-sky-200 px-2 py-0.5">{ fd.Airline } { fd.FlightNumber }</span>
                }
            </div>
            <!-- Departure details -->
            if fd.DepartureTerminal != "" || fd.DepartureGate != "" {
                <div class="text-xs text-slate-500">
                    <span class="font-medium text-slate-600">Dep:</span>
                    if fd.DepartureTerminal != "" {
                        <span class="font-mono ml-1">Terminal { fd.DepartureTerminal }</span>
                    }
                    if fd.DepartureGate != "" {
                        <span class="font-mono ml-1">Gate { fd.DepartureGate }</span>
                    }
                </div>
            }
            <!-- Arrival details -->
            if fd.ArrivalTerminal != "" || fd.ArrivalGate != "" {
                <div class="text-xs text-slate-500">
                    <span class="font-medium text-slate-600">Arr:</span>
                    if fd.ArrivalTerminal != "" {
                        <span class="font-mono ml-1">Terminal { fd.ArrivalTerminal }</span>
                    }
                    if fd.ArrivalGate != "" {
                        <span class="font-mono ml-1">Gate { fd.ArrivalGate }</span>
                    }
                </div>
            }
            <!-- Booking reference -->
            if fd.BookingReference != "" {
                <div class="text-xs text-slate-500">
                    <span class="font-medium text-slate-600">Ref:</span>
                    <span class="font-mono ml-1 tracking-wider">{ fd.BookingReference }</span>
                </div>
            }
        </div>
    }
    ```
  - [x] 8.2 Run `just generate` after creating this file.

### Template — EventTimelineItem Updates for Flight

- [x] Task 9: Update EventTimelineItem to render flight details in view and edit modes (AC: #2, #3)
  - [x] 9.1 In `internal/handler/event.templ`, in the **view mode** section (`x-show="!editing"`), after the existing notes/pinned display, add:
    ```templ
    if event.Category == domain.CategoryFlight {
        @FlightCardContent(event.Flight)
    }
    ```
    Place this before the `<div class="flex gap-2">` that holds Edit/Delete buttons.
  - [x] 9.2 In the **edit mode** section (`x-show="editing"`), add flight-specific inline edit fields AFTER the pinned toggle and BEFORE the Actions div:
    ```templ
    if event.Category == domain.CategoryFlight {
        <div class="mb-3 pt-3 border-t-2 border-slate-100">
            <p class="text-xs font-bold uppercase tracking-wide text-sky-700 mb-3">Flight Details</p>
            <div class="grid grid-cols-2 gap-3 mb-3">
                <div>
                    <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1">Airline</label>
                    if props != nil {
                        <input type="text" name="airline" value={ props.FormValues.Airline }
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
                    } else {
                        <input type="text" name="airline" value={ flightFieldValue(event.Flight, "airline") }
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
                    }
                </div>
                <div>
                    <label class="block text-xs font-bold uppercase tracking-wide text-slate-500 mb-1">Flight No.</label>
                    if props != nil {
                        <input type="text" name="flight_number" value={ props.FormValues.FlightNumber }
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
                    } else {
                        <input type="text" name="flight_number" value={ flightFieldValue(event.Flight, "flight_number") }
                            class="w-full px-2 py-1.5 border-2 border-slate-300 bg-white text-sm font-mono focus:outline-none focus:border-brand"/>
                    }
                </div>
            </div>
            <!-- Repeat same pattern for departure_airport, arrival_airport, departure_terminal, arrival_terminal, departure_gate, arrival_gate, booking_reference -->
            <!-- Use 2-column grid for airports and terminals/gates, single column for booking_reference -->
        </div>
    }
    ```
  - [x] 9.3 Add `flightFieldValue` helper in `event.go` (not in a templ file) to safely access flight detail fields:
    ```go
    // flightFieldValue safely retrieves a named field from FlightDetails.
    // Returns empty string if fd is nil.
    func flightFieldValue(fd *domain.FlightDetails, field string) string {
        if fd == nil {
            return ""
        }
        switch field {
        case "airline":           return fd.Airline
        case "flight_number":     return fd.FlightNumber
        case "departure_airport": return fd.DepartureAirport
        case "arrival_airport":   return fd.ArrivalAirport
        case "departure_terminal":return fd.DepartureTerminal
        case "arrival_terminal":  return fd.ArrivalTerminal
        case "departure_gate":    return fd.DepartureGate
        case "arrival_gate":      return fd.ArrivalGate
        case "booking_reference": return fd.BookingReference
        }
        return ""
    }
    ```
    This avoids nil pointer panics when a flight event's details fail to load.

### DI Wiring

- [x] Task 10: Update `cmd/app/main.go` to wire FlightDetailsStore (AC: #2)
  - [x] 10.1 Add `flightDetailsStore` and pass to `NewEventStore`:
    ```go
    // Repositories
    tripStore := repository.NewTripStore(pool)
    flightDetailsStore := repository.NewFlightDetailsStore()
    eventStore := repository.NewEventStore(pool, flightDetailsStore)
    ```

### Testing

- [x] Task 11: Write service tests (AC: #1, #2, #3, #4)
  - [x] 11.1 Service test: `Create` flight event → `event.Flight` populated with correct field values
    - Mock `EventRepository.Create` to capture the `*domain.Event` passed; verify `event.Flight != nil` and fields match input
  - [x] 11.2 Service test: `Create` flight event with nil `FlightDetails` → defaults to empty `FlightDetails{}` (not nil panic)
  - [x] 11.3 Service test: `Update` flight event → `event.Flight` updated to new values
    - Input: existing event with `Flight.Airline = "BA"`, update with `FlightDetails{Airline: "LH"}`
    - Verify updater sets `event.Flight.Airline = "LH"`
  - [x] 11.4 Service test: `Update` non-flight event with nil `FlightDetails` → no change to `event.Flight`
  - [x] 11.5 Run `just test` — all passing, no races
  - [x] 11.6 Run `just lint` — zero violations
  - [x] 11.7 Run `just build` — binary compiles
  - [x] 11.8 Manual smoke tests:
    - Click "Add Event" → select Flight in TypeSelector → flight fields appear, other types don't show them
    - Fill in airline, flight number, airports → Save → event appears in timeline with sky-blue icon
    - Expand card → FlightCardContent shows route (LHR → CDG), flight number badge, gate/terminal info
    - Click Edit → flight fields appear pre-filled with existing values → modify → Save → card updates
    - Click Delete → event removed → Undo → event restored with flight details intact
    - Create Activity → no flight fields appear in TypeSelector or card

## Dev Notes

### Architecture: Base + Detail Tables Pattern

The `flight_details` table already exists in the schema (from `001_initial.up.sql`). **No new migration is needed for Story 1.4.** The schema has:
- `flight_details.event_id` → FK `REFERENCES events(id) ON DELETE CASCADE`
- Delete CASCADE means deleting a base event automatically deletes `flight_details` — no code needed for AC #4.

### Critical: Transactional Create and Update

Flight event creation MUST insert both `events` and `flight_details` in a single transaction. If the `flight_details` insert fails, the base event must NOT persist. Use `s.db.Begin(ctx)` → `sqlcgen.New(tx)` pattern. The `sqlcgen.New` accepts any `sqlcgen.DBTX` (interface satisfied by both `*pgxpool.Pool` and `pgx.Tx`).

The `EventStore` needs access to `*pgxpool.Pool` to begin transactions — store it as `db *pgxpool.Pool` alongside `queries`. Update `NewEventStore(db *pgxpool.Pool, flightStore *FlightDetailsStore)` accordingly.

### Critical: EventStore.Create Duplication

The Create method now has two code paths (transactional for flight, simple for others). Both call `CreateEvent` sqlc query. This duplication is intentional — don't try to unify into a helper at this stage.

### Critical: TypeSelector x-data Scope for x-show

`EventCreateForm` uses `@TypeSelector(data.Category)` which renders:
```html
<div x-data="{ selected: '...' }" role="radiogroup" ...>
  <!-- icon buttons -->
  <input type="hidden" name="category" x-ref="categoryInput"/>
</div>
```

The flight fields section uses `x-show="selected === 'flight'"`. For `x-show` to reference `selected`, the flight fields div MUST be inside the same `<div x-data=...>` wrapper, OR you must move `x-data` to the `<form>` level.

**Simplest fix:** Move the `x-data="{ selected: '...' }"` from TypeSelector's wrapper div to the `<form>` tag in `EventCreateForm`. Adjust `TypeSelector` to not carry its own `x-data` (receive it from parent context). OR, restructure so the entire form is inside TypeSelector's `<div x-data=...>`.

**Alternative (cleaner):** Keep TypeSelector's `x-data` on its own div, and use `x-data` on the flight section div too, but use Alpine.js `$store` or event dispatching to sync selection. This is more complex.

**Recommended:** Lift `x-data="{ selected: '{{data.Category}}' }"` to the `<form>` element in `EventCreateForm`, remove it from TypeSelector's div wrapper. The TypeSelector's buttons still use `x-on:click="selected = '..'"` and will find `selected` in the parent `<form>` scope.

### Critical: Inline Edit — Category Cannot Change

The inline edit form in `EventTimelineItem` does NOT include a TypeSelector. Category is fixed at creation time. The flight fields in the edit form are shown/hidden based on `event.Category` (a Go template condition, evaluated server-side), NOT Alpine.js `x-show`. This is correct because the card is re-rendered on every event update.

### Critical: Collapsed Card Header for Flight Events

The collapsed card header (always visible part) shows `event.Title`, time range, and location. For flights, the title (e.g., "London to Paris") is in `event.Title`. Airport codes in the flight section are only visible when expanded. The collapsed header is NOT modified — it uses the same template for all categories.

### Pattern: Monospace for Structured Data

Per UX spec: "Monospace treatment for structured data (flight numbers, booking references, addresses)." Use `font-mono` Tailwind class for:
- Airline + Flight number (e.g., `BA 234`)
- Airport codes (e.g., `LHR`, `CDG`)
- Terminal and gate values
- Booking reference

### Pattern: All Existing Patterns from Stories 1.1-1.3 Apply

- **pgtype helpers**: `toPgText()`, `toPgTimestamptz()`, etc. in `internal/repository/helpers.go` — use for all sqlcgen params
- **Error mapping**: `domain.ErrNotFound` → 404, `domain.ErrInvalidInput` → 422, else → 500
- **HTMX success**: Day-level swap via `HX-Retarget: #day-{date}` + `HX-Reswap: outerHTML` — NO CHANGE from prior stories
- **Sheet close**: `HX-Trigger: {"close-sheet": true}` on successful create — keep this
- **`just generate`**: Run after every `.templ` or `.sql` file change
- **Never edit generated files**: `sqlcgen/*.go`, `*_templ.go`
- **Form error pattern**: 422 + form HTML with rose borders + `aria-describedby` messages
- **Soft delete**: DELETE handler is unchanged — it soft-deletes base event; CASCADE removes flight_details from DB permanently (which is correct — you can't undo a hard deletion of flight_details, but the base event's soft-delete preserves undo)

**Wait — soft delete interaction with flight_details:**
The `events.deleted_at` column soft-deletes the base event. `flight_details` has `ON DELETE CASCADE` — this only fires on hard DELETE, NOT on UPDATE (which soft delete uses). So:
- Soft delete (set `deleted_at = NOW()`) → base event hidden, `flight_details` row stays in DB
- Restore (set `deleted_at = NULL`) → base event visible again, `flight_details` row still there
- Hard delete (trip deletion via CASCADE) → `flight_details` deleted by DB CASCADE
This is correct — no code change needed. The undo flow works: soft delete → restore → flight_details still present.

### Project Structure Notes

New files:
- `internal/repository/sql/flight_details.sql` (NEW)
- `internal/repository/sqlcgen/flight_details.sql.go` (regenerated by `just generate`)
- `internal/repository/flight_details_store.go` (NEW)
- `internal/handler/flight_card.templ` (NEW)
- `internal/handler/flight_card_templ.go` (regenerated by `just generate`)

Modified files:
- `internal/domain/models.go` — add `FlightDetails` struct, `Flight *FlightDetails` on `Event`
- `internal/repository/event_store.go` — add `db` + `flight` fields, update Create/GetByID/ListByTripAndDate/ListByTrip/Update methods
- `internal/service/event.go` — add `FlightDetails` to inputs, populate in Create/Update
- `internal/handler/event.go` — add flight fields to `EventFormData`, update Create/Update handlers, add `flightFieldValue` helper
- `internal/handler/event_form.templ` — enable flight in `enabledCategories`, add flight form fields with `x-show`
- `internal/handler/event.templ` — add `FlightCardContent` call in view mode, flight edit fields in edit mode
- `cmd/app/main.go` — wire `FlightDetailsStore`

Do NOT modify:
- `migrations/` — no new migration needed (schema already has `flight_details` table)
- `internal/repository/sqlcgen/*` — always regenerated
- `internal/handler/*_templ.go` — always regenerated

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story 1.4] — Acceptance criteria: airline, flight number, airports, terminals, gates, booking reference; transactional persist; FlightCardContent metadata display; CASCADE delete
- [Source: _bmad-output/planning-artifacts/architecture.md#Data Architecture] — Base + Detail Tables: `flight_details` with 1:1 event_id FK; CASCADE DELETE; detail stores handle write-path via DBTX
- [Source: _bmad-output/planning-artifacts/architecture.md#Structure Patterns] — Event type code organization: 8-step layer ordering; `FlightDetailsStore` adapter; `flight_card.templ` per architecture
- [Source: _bmad-output/planning-artifacts/architecture.md#Communication Patterns] — No JSON responses; server-rendered HTML; day-level HTMX swaps unchanged
- [Source: _bmad-output/planning-artifacts/architecture.md#Enforcement Guidelines] — Rule 7: add new event types as separate files, never by modifying existing type components
- [Source: _bmad-output/planning-artifacts/architecture.md#Frontend Architecture] — Form morphing via Alpine.js `x-show`; TypeSelector instant switch
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md] — Monospace for flight numbers, booking refs; Sky-blue (`bg-sky-100 text-sky-700`) for flight category (already in `categoryBgColors` map in `event.templ`)
- [Source: _bmad-output/implementation-artifacts/1-3-event-edit-delete-and-detail-view.md] — `EventTimelineItem(event, props)` pattern; `EventCardProps`; inline edit; `parseDateAndTime`; `HX-Retarget`; soft delete behavior

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- Implemented full Flight event support across all layers (SQL → repository → service → handler → templates).
- `FlightDetail` model was already in `sqlcgen/models.go` (generated from migration schema); only queries needed to be added.
- Lifted Alpine.js `x-data="{ selected: '...' }"` from TypeSelector's inner div to the `<form>` element in `EventCreateForm` so that `x-show="selected === 'flight'"` on flight fields resolves correctly.
- Flight edit fields in `EventTimelineItem` are shown via server-side Go `if event.Category == domain.CategoryFlight` (not Alpine.js), since category is fixed at creation time.
- `fieldalignment` linter triggered struct field reordering on `Event`, `CreateEventInput`, and `EventFormData` — applied automatically via `fieldalignment -fix` tool.
- Fixed pre-existing lint issues in `cmd/seed/main.go` (exitAfterDefer, unnamedResult) by extracting a `run()` function.
- All 5 new service tests pass, no regressions, no race conditions.

### File List

**New files:**
- `internal/repository/sql/flight_details.sql`
- `internal/repository/sqlcgen/flight_details.sql.go` (generated)
- `internal/repository/flight_details_store.go`
- `internal/handler/flight_card.templ`
- `internal/handler/flight_card_templ.go` (generated)

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
- `_bmad-output/implementation-artifacts/sprint-status.yaml`

## Change Log

- 2026-02-20: Story 1.4 implemented — Flight events with airline, flight number, airports, terminals, gates, booking reference. Transactional create/update via `FlightDetailsStore`. `FlightCardContent` renders expanded flight details. 5 new service tests added.
