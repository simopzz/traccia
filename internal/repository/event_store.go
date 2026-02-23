package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/repository/sqlcgen"
)

var _ domain.EventRepository = (*EventStore)(nil)

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

		// Capture flight details before overwriting *event, as eventRowToDomain returns nil Flight
		flightDetails := event.Flight
		*event = eventRowToDomain(&row)

		fd, txErr := s.flight.Create(ctx, txq, event.ID, flightDetails)
		if txErr != nil {
			return txErr
		}
		event.Flight = fd

		return tx.Commit(ctx)
	}

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

	// Non-transactional path for Activity, Food (no detail table)
	row, err := s.queries.CreateEvent(ctx, sqlcgen.CreateEventParams{
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
	if err != nil {
		return err
	}
	*event = eventRowToDomain(&row)
	return nil
}

func (s *EventStore) GetByID(ctx context.Context, id int) (*domain.Event, error) {
	row, err := s.queries.GetEventByID(ctx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	event := eventRowToDomain(&row)
	if event.Category == domain.CategoryFlight {
		events := s.loadFlightDetails(ctx, []domain.Event{event})
		event = events[0]
	}
	if event.Category == domain.CategoryLodging {
		events := s.loadLodgingDetails(ctx, []domain.Event{event})
		event = events[0]
	}
	return &event, nil
}

func (s *EventStore) ListByTrip(ctx context.Context, tripID int) ([]domain.Event, error) {
	rows, err := s.queries.ListEventsByTrip(ctx, int32(tripID))
	if err != nil {
		return nil, err
	}

	events := make([]domain.Event, len(rows))
	for i := range rows {
		events[i] = eventRowToDomain(&rows[i])
	}
	events = s.loadFlightDetails(ctx, events)
	return s.loadLodgingDetails(ctx, events), nil
}

func (s *EventStore) ListByTripAndDate(ctx context.Context, tripID int, date time.Time) ([]domain.Event, error) {
	rows, err := s.queries.ListEventsByTripAndDate(ctx, sqlcgen.ListEventsByTripAndDateParams{
		TripID:    int32(tripID),
		EventDate: toPgDate(date),
	})
	if err != nil {
		return nil, err
	}

	events := make([]domain.Event, len(rows))
	for i := range rows {
		events[i] = eventRowToDomain(&rows[i])
	}
	events = s.loadFlightDetails(ctx, events)
	return s.loadLodgingDetails(ctx, events), nil
}

func (s *EventStore) Update(ctx context.Context, id int, updater func(*domain.Event) *domain.Event) (*domain.Event, error) {
	event, err := s.GetByID(ctx, id) // now loads Flight details for flight events
	if err != nil {
		return nil, err
	}

	updated := updater(event)

	if updated.Category == domain.CategoryFlight && updated.Flight != nil {
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

		fd, txErr := s.flight.Update(ctx, txq, id, updated.Flight)
		if txErr != nil {
			return nil, txErr
		}
		result.Flight = fd

		if txErr = tx.Commit(ctx); txErr != nil {
			return nil, fmt.Errorf("committing transaction: %w", txErr)
		}
		return &result, nil
	}

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

	// Non-transactional for Activity, Food
	row, err := s.queries.UpdateEvent(ctx, sqlcgen.UpdateEventParams{
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
	if err != nil {
		return nil, err
	}
	result := eventRowToDomain(&row)
	return &result, nil
}

// Delete soft-deletes the event (sets deleted_at). Events are permanently removed
// when their parent trip is deleted via ON DELETE CASCADE.
func (s *EventStore) Delete(ctx context.Context, id int) error {
	return s.queries.SoftDeleteEvent(ctx, int32(id))
}

func (s *EventStore) Restore(ctx context.Context, id int) (*domain.Event, error) {
	row, err := s.queries.RestoreEvent(ctx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	event := eventRowToDomain(&row)
	return &event, nil
}

func (s *EventStore) GetLastEventByTrip(ctx context.Context, tripID int) (*domain.Event, error) {
	row, err := s.queries.GetLastEventByTrip(ctx, int32(tripID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	event := eventRowToDomain(&row)
	return &event, nil
}

func (s *EventStore) CountByTrip(ctx context.Context, tripID int) (int, error) {
	count, err := s.queries.CountEventsByTrip(ctx, int32(tripID))
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// loadFlightDetails enriches flight events with their detail row.
// No-op for non-flight events. Errors are logged but not fatal.
func (s *EventStore) loadFlightDetails(ctx context.Context, events []domain.Event) []domain.Event {
	var flightEventIDs []int
	for i := range events {
		if events[i].Category == domain.CategoryFlight {
			flightEventIDs = append(flightEventIDs, events[i].ID)
		}
	}

	if len(flightEventIDs) == 0 {
		return events
	}

	detailsMap, err := s.flight.GetByEventIDs(ctx, s.queries, flightEventIDs)
	if err != nil {
		slog.WarnContext(ctx, "failed to load flight_details batch", "error", err)
		return events
	}

	for i := range events {
		if events[i].Category == domain.CategoryFlight {
			if fd, ok := detailsMap[events[i].ID]; ok {
				events[i].Flight = fd
			}
		}
	}
	return events
}

// loadLodgingDetails enriches lodging events with their detail row.
// No-op for non-lodging events. Errors are logged but not fatal.
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

func eventRowToDomain(row *sqlcgen.Event) domain.Event {
	var lat, lng *float64
	if row.Latitude.Valid {
		lat = &row.Latitude.Float64
	}
	if row.Longitude.Valid {
		lng = &row.Longitude.Float64
	}

	return domain.Event{
		ID:        int(row.ID),
		TripID:    int(row.TripID),
		EventDate: row.EventDate.Time,
		Title:     row.Title,
		Category:  domain.EventCategory(row.Category),
		Location:  row.Location.String,
		Latitude:  lat,
		Longitude: lng,
		StartTime: row.StartTime.Time,
		EndTime:   row.EndTime.Time,
		Pinned:    row.Pinned.Bool,
		Position:  int(row.Position),
		Notes:     row.Notes.String,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
