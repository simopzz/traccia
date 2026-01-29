package timeline

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrTripNotFound = fmt.Errorf("trip not found")

type Service interface {
	CreateTrip(ctx context.Context, params CreateTripParams) (*Trip, error)
	GetTrip(ctx context.Context, id uuid.UUID) (*Trip, error)
	ResetTrip(ctx context.Context, id uuid.UUID) error
}

type service struct {
	db *sql.DB
}

func NewService(db *sql.DB) Service {
	return &service{db: db}
}

type CreateTripParams struct {
	Name        string
	Destination string
	StartDate   *time.Time
	EndDate     *time.Time
}

func (s *service) CreateTrip(ctx context.Context, params CreateTripParams) (*Trip, error) {
	query := `
		INSERT INTO trips (name, destination, start_date, end_date)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, destination, start_date, end_date, created_at, updated_at
	`
	row := s.db.QueryRowContext(ctx, query, params.Name, params.Destination, params.StartDate, params.EndDate)

	var trip Trip
	err := row.Scan(
		&trip.ID,
		&trip.Name,
		&trip.Destination,
		&trip.StartDate,
		&trip.EndDate,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trip: %w", err)
	}

	return &trip, nil
}

func (s *service) GetTrip(ctx context.Context, id uuid.UUID) (*Trip, error) {
	query := `
		SELECT id, name, destination, start_date, end_date, created_at, updated_at
		FROM trips
		WHERE id = $1
	`
	row := s.db.QueryRowContext(ctx, query, id)

	var trip Trip
	err := row.Scan(
		&trip.ID,
		&trip.Name,
		&trip.Destination,
		&trip.StartDate,
		&trip.EndDate,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTripNotFound
		}
		// We might want to wrap this to avoid leaking DB details, but for now this is fine.
		return nil, fmt.Errorf("failed to get trip: %w", err)
	}

	return &trip, nil
}

func (s *service) ResetTrip(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events WHERE trip_id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to reset trip: %w", err)
	}
	return nil
}
