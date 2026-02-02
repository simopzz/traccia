package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simopzz/traccia/internal/domain"
)

var _ domain.EventRepository = (*EventStore)(nil)

type EventStore struct {
	queries *Queries
}

func NewEventStore(db *pgxpool.Pool) *EventStore {
	return &EventStore{
		queries: New(db),
	}
}

func (s *EventStore) Create(ctx context.Context, event *domain.Event) error {
	maxPos, err := s.queries.GetMaxPositionByTrip(ctx, int32(event.TripID))
	if err != nil {
		return err
	}
	position := maxPos + 1
	if event.Position > 0 {
		position = int32(event.Position)
	}

	row, err := s.queries.CreateEvent(ctx, CreateEventParams{
		TripID:    int32(event.TripID),
		Title:     event.Title,
		Category:  string(event.Category),
		Location:  toPgText(event.Location),
		Latitude:  toPgFloat8(event.Latitude),
		Longitude: toPgFloat8(event.Longitude),
		StartTime: toPgTimestamptz(event.StartTime),
		EndTime:   toPgTimestamptz(event.EndTime),
		Pinned:    toPgBool(event.Pinned),
		Position:  position,
	})
	if err != nil {
		return err
	}
	*event = eventRowToDomain(row)
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
	event := eventRowToDomain(row)
	return &event, nil
}

func (s *EventStore) ListByTrip(ctx context.Context, tripID int) ([]domain.Event, error) {
	rows, err := s.queries.ListEventsByTrip(ctx, int32(tripID))
	if err != nil {
		return nil, err
	}

	events := make([]domain.Event, len(rows))
	for i, row := range rows {
		events[i] = eventRowToDomain(row)
	}
	return events, nil
}

func (s *EventStore) Update(ctx context.Context, id int, updater func(*domain.Event) *domain.Event) (*domain.Event, error) {
	event, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updated := updater(event)

	row, err := s.queries.UpdateEvent(ctx, UpdateEventParams{
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
	})
	if err != nil {
		return nil, err
	}

	result := eventRowToDomain(row)
	return &result, nil
}

func (s *EventStore) Delete(ctx context.Context, id int) error {
	return s.queries.DeleteEvent(ctx, int32(id))
}

func (s *EventStore) GetLastEventByTrip(ctx context.Context, tripID int) (*domain.Event, error) {
	row, err := s.queries.GetLastEventByTrip(ctx, int32(tripID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	event := eventRowToDomain(row)
	return &event, nil
}

func eventRowToDomain(row Event) domain.Event {
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
		Title:     row.Title,
		Category:  domain.EventCategory(row.Category),
		Location:  row.Location.String,
		Latitude:  lat,
		Longitude: lng,
		StartTime: row.StartTime.Time,
		EndTime:   row.EndTime.Time,
		Pinned:    row.Pinned.Bool,
		Position:  int(row.Position),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func toPgFloat8(f *float64) pgtype.Float8 {
	if f == nil {
		return pgtype.Float8{}
	}
	return pgtype.Float8{Float64: *f, Valid: true}
}

func toPgBool(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}
