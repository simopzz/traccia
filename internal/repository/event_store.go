package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/repository/sqlcgen"
)

var _ domain.EventRepository = (*EventStore)(nil)

type EventStore struct {
	queries *sqlcgen.Queries
}

func NewEventStore(db *pgxpool.Pool) *EventStore {
	return &EventStore{
		queries: sqlcgen.New(db),
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
	return events, nil
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
	return events, nil
}

func (s *EventStore) Update(ctx context.Context, id int, updater func(*domain.Event) *domain.Event) (*domain.Event, error) {
	event, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updated := updater(event)

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

func (s *EventStore) Delete(ctx context.Context, id int) error {
	rows, err := s.queries.DeleteEvent(ctx, int32(id))
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
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
