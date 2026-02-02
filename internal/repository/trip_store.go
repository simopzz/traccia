package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/simopzz/traccia/internal/domain"
)

var _ domain.TripRepository = (*TripStore)(nil)

type TripStore struct {
	queries *Queries
}

func NewTripStore(db *pgxpool.Pool) *TripStore {
	return &TripStore{
		queries: New(db),
	}
}

func (s *TripStore) Create(ctx context.Context, trip *domain.Trip) error {
	row, err := s.queries.CreateTrip(ctx, CreateTripParams{
		Name:        trip.Name,
		Destination: trip.Destination,
		StartDate:   toPgTimestamptz(trip.StartDate),
		EndDate:     toPgTimestamptz(trip.EndDate),
		UserID:      pgtype.UUID{},
	})
	if err != nil {
		return err
	}
	*trip = tripRowToDomain(row)
	return nil
}

func (s *TripStore) GetByID(ctx context.Context, id int) (*domain.Trip, error) {
	row, err := s.queries.GetTripByID(ctx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	trip := tripRowToDomain(row)
	return &trip, nil
}

func (s *TripStore) List(ctx context.Context, userID *string) ([]domain.Trip, error) {
	var uid pgtype.UUID
	if userID != nil {
		if err := uid.Scan(*userID); err == nil {
			uid.Valid = true
		}
	}

	rows, err := s.queries.ListTrips(ctx, uid)
	if err != nil {
		return nil, err
	}

	trips := make([]domain.Trip, len(rows))
	for i, row := range rows {
		trips[i] = tripRowToDomain(row)
	}
	return trips, nil
}

func (s *TripStore) Update(ctx context.Context, id int, updater func(*domain.Trip) *domain.Trip) (*domain.Trip, error) {
	trip, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updated := updater(trip)

	row, err := s.queries.UpdateTrip(ctx, UpdateTripParams{
		ID:          int32(id),
		Name:        updated.Name,
		Destination: updated.Destination,
		StartDate:   toPgTimestamptz(updated.StartDate),
		EndDate:     toPgTimestamptz(updated.EndDate),
	})
	if err != nil {
		return nil, err
	}

	result := tripRowToDomain(row)
	return &result, nil
}

func (s *TripStore) Delete(ctx context.Context, id int) error {
	return s.queries.DeleteTrip(ctx, int32(id))
}

func tripRowToDomain(row Trip) domain.Trip {
	return domain.Trip{
		ID:          int(row.ID),
		Name:        row.Name,
		Destination: row.Destination,
		StartDate:   row.StartDate.Time,
		EndDate:     row.EndDate.Time,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}

func toPgTimestamptz(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}
