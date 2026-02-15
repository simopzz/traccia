package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/repository/sqlcgen"
)

var _ domain.TripRepository = (*TripStore)(nil)

type TripStore struct {
	queries *sqlcgen.Queries
}

func NewTripStore(db *pgxpool.Pool) *TripStore {
	return &TripStore{
		queries: sqlcgen.New(db),
	}
}

func (s *TripStore) Create(ctx context.Context, trip *domain.Trip) error {
	row, err := s.queries.CreateTrip(ctx, sqlcgen.CreateTripParams{
		Name:        trip.Name,
		Destination: toPgText(trip.Destination),
		StartDate:   toPgDate(trip.StartDate),
		EndDate:     toPgDate(trip.EndDate),
		UserID:      pgtype.UUID{},
	})
	if err != nil {
		return err
	}
	*trip = tripRowToDomain(&row)
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
	trip := tripRowToDomain(&row)
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
	for i := range rows {
		trips[i] = tripRowToDomain(&rows[i])
	}
	return trips, nil
}

func (s *TripStore) Update(ctx context.Context, id int, updater func(*domain.Trip) *domain.Trip) (*domain.Trip, error) {
	trip, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updated := updater(trip)

	row, err := s.queries.UpdateTrip(ctx, sqlcgen.UpdateTripParams{
		ID:          int32(id),
		Name:        updated.Name,
		Destination: toPgText(updated.Destination),
		StartDate:   toPgDate(updated.StartDate),
		EndDate:     toPgDate(updated.EndDate),
	})
	if err != nil {
		return nil, err
	}

	result := tripRowToDomain(&row)
	return &result, nil
}

func (s *TripStore) Delete(ctx context.Context, id int) error {
	return s.queries.DeleteTrip(ctx, int32(id))
}

func (s *TripStore) CountEventsByTripAndDateRange(ctx context.Context, tripID int, newStart, newEnd time.Time) (int, error) {
	count, err := s.queries.CountEventsByTripAndDateRange(ctx, sqlcgen.CountEventsByTripAndDateRangeParams{
		TripID:      int32(tripID),
		EventDate:   toPgDate(newStart),
		EventDate_2: toPgDate(newEnd),
	})
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (s *TripStore) CountEventsByTripGroupedByDate(ctx context.Context, tripID int, newStart, newEnd time.Time) ([]domain.DateEventCount, error) {
	rows, err := s.queries.CountEventsByTripGroupedByDate(ctx, sqlcgen.CountEventsByTripGroupedByDateParams{
		TripID:      int32(tripID),
		EventDate:   toPgDate(newStart),
		EventDate_2: toPgDate(newEnd),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.DateEventCount, len(rows))
	for i, row := range rows {
		result[i] = domain.DateEventCount{
			Date:  row.EventDate.Time,
			Count: int(row.EventCount),
		}
	}
	return result, nil
}

func tripRowToDomain(row *sqlcgen.Trip) domain.Trip {
	return domain.Trip{
		ID:          int(row.ID),
		Name:        row.Name,
		Destination: row.Destination.String,
		StartDate:   row.StartDate.Time,
		EndDate:     row.EndDate.Time,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
