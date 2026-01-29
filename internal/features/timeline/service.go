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
	CreateEvent(ctx context.Context, params CreateEventParams) (*Event, error)
	GetEvents(ctx context.Context, tripID uuid.UUID) ([]Event, error)
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

type CreateEventParams struct {
	TripID    uuid.UUID
	Title     string
	Location  *string
	Category  *string
	GeoLat    *float64
	GeoLng    *float64
	StartTime *time.Time
	EndTime   *time.Time
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

func (s *service) CreateEvent(ctx context.Context, params CreateEventParams) (*Event, error) {
	if params.StartTime != nil && params.EndTime != nil {
		if params.EndTime.Before(*params.StartTime) {
			return nil, fmt.Errorf("end time must be after start time")
		}
	}

	// Ensure UTC
	if params.StartTime != nil {
		t := params.StartTime.UTC()
		params.StartTime = &t
	}
	if params.EndTime != nil {
		t := params.EndTime.UTC()
		params.EndTime = &t
	}

	query := `
		INSERT INTO events (trip_id, title, location, category, geo_lat, geo_lng, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, trip_id, title, location, category, geo_lat, geo_lng, start_time, end_time, created_at, updated_at
	`
	row := s.db.QueryRowContext(ctx, query,
		params.TripID,
		params.Title,
		params.Location,
		params.Category,
		params.GeoLat,
		params.GeoLng,
		params.StartTime,
		params.EndTime,
	)

	var event Event
	err := row.Scan(
		&event.ID,
		&event.TripID,
		&event.Title,
		&event.Location,
		&event.Category,
		&event.GeoLat,
		&event.GeoLng,
		&event.StartTime,
		&event.EndTime,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return &event, nil
}

func (s *service) GetEvents(ctx context.Context, tripID uuid.UUID) ([]Event, error) {
	query := `
		SELECT id, trip_id, title, location, category, geo_lat, geo_lng, start_time, end_time, created_at, updated_at
		FROM events
		WHERE trip_id = $1
		ORDER BY start_time ASC
	`
	rows, err := s.db.QueryContext(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(
			&e.ID, &e.TripID, &e.Title, &e.Location, &e.Category, &e.GeoLat, &e.GeoLng, &e.StartTime, &e.EndTime, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}
