package timeline

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const DefaultEventDuration = 1 * time.Hour

var ErrTripNotFound = fmt.Errorf("trip not found")

type Service interface {
	CreateTrip(ctx context.Context, params CreateTripParams) (*Trip, error)
	GetTrip(ctx context.Context, id uuid.UUID) (*Trip, error)
	ResetTrip(ctx context.Context, id uuid.UUID) error
	CreateEvent(ctx context.Context, params CreateEventParams) (*Event, error)
	GetEvents(ctx context.Context, tripID uuid.UUID) ([]Event, error)
	ReorderEvents(ctx context.Context, tripID uuid.UUID, eventIDs []uuid.UUID) ([]Event, error)
	TogglePin(ctx context.Context, id uuid.UUID) (*Event, error)
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
		RETURNING id, trip_id, title, location, category, geo_lat, geo_lng, start_time, end_time, is_pinned, created_at, updated_at
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
		&event.IsPinned,
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
		SELECT id, trip_id, title, location, category, geo_lat, geo_lng, start_time, end_time, is_pinned, created_at, updated_at
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
			&e.ID, &e.TripID, &e.Title, &e.Location, &e.Category, &e.GeoLat, &e.GeoLng, &e.StartTime, &e.EndTime, &e.IsPinned, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}

func (s *service) ReorderEvents(ctx context.Context, tripID uuid.UUID, eventIDs []uuid.UUID) ([]Event, error) {
	// 1. Start Transaction & Lock
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Fetch all events for the trip with locking to prevent race conditions
	query := `
		SELECT id, trip_id, title, location, category, geo_lat, geo_lng, start_time, end_time, is_pinned, created_at, updated_at
		FROM events
		WHERE trip_id = $1
		ORDER BY start_time ASC
		FOR UPDATE
	`
	rows, err := tx.QueryContext(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for locking: %w", err)
	}
	defer rows.Close()

	var existingEvents []Event
	for rows.Next() {
		var e Event
		if err := rows.Scan(
			&e.ID, &e.TripID, &e.Title, &e.Location, &e.Category, &e.GeoLat, &e.GeoLng, &e.StartTime, &e.EndTime, &e.IsPinned, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		existingEvents = append(existingEvents, e)
	}

	if len(existingEvents) == 0 {
		return []Event{}, nil
	}

	// 2. Map for lookup & Validation
	eventMap := make(map[uuid.UUID]*Event)
	for i := range existingEvents {
		eventMap[existingEvents[i].ID] = &existingEvents[i]
	}

	if len(eventIDs) != len(existingEvents) {
		return nil, fmt.Errorf("event count mismatch: expected %d, got %d", len(existingEvents), len(eventIDs))
	}

	// Validation: Ensure no duplicates in input
	seen := make(map[uuid.UUID]bool)
	for _, id := range eventIDs {
		if seen[id] {
			return nil, fmt.Errorf("duplicate event ID in reorder list: %s", id)
		}
		if _, exists := eventMap[id]; !exists {
			return nil, fmt.Errorf("event %s not found in trip", id)
		}
		seen[id] = true
	}

	// 3. Determine Anchor Start Time
	var currentTime time.Time
	if existingEvents[0].StartTime != nil {
		currentTime = *existingEvents[0].StartTime
	} else {
		// Fallback if the first event has no time: Use Now
		currentTime = time.Now().Truncate(time.Minute)
	}

	// 4. Update Loop
	var reorderedEvents []Event

	for _, id := range eventIDs {
		evt, ok := eventMap[id]
		if !ok {
			return nil, fmt.Errorf("event %s not found in trip", id)
		}

		// Calculate Duration
		var duration time.Duration
		if evt.StartTime != nil && evt.EndTime != nil {
			duration = evt.EndTime.Sub(*evt.StartTime)
		} else {
			duration = DefaultEventDuration
		}

		// Update Times
		var newStart time.Time
		if evt.IsPinned && evt.StartTime != nil {
			newStart = *evt.StartTime
		} else {
			newStart = currentTime
		}

		newEnd := newStart.Add(duration)

		// Create new pointers for time values
		sTime := newStart
		eTime := newEnd
		evt.StartTime = &sTime
		evt.EndTime = &eTime

		// Update in DB (using the transaction)
		updateQuery := `UPDATE events SET start_time = $1, end_time = $2, updated_at = NOW() WHERE id = $3`
		_, err := tx.ExecContext(ctx, updateQuery, newStart, newEnd, evt.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to update event %s: %w", evt.ID, err)
		}

		reorderedEvents = append(reorderedEvents, *evt)

		// Advance time
		currentTime = newEnd
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return reorderedEvents, nil
}

func (s *service) TogglePin(ctx context.Context, id uuid.UUID) (*Event, error) {
	query := `
		UPDATE events 
		SET is_pinned = NOT is_pinned, updated_at = NOW() 
		WHERE id = $1 
		RETURNING id, trip_id, title, location, category, geo_lat, geo_lng, start_time, end_time, is_pinned, created_at, updated_at
	`

	var event Event
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.TripID,
		&event.Title,
		&event.Location,
		&event.Category,
		&event.GeoLat,
		&event.GeoLng,
		&event.StartTime,
		&event.EndTime,
		&event.IsPinned,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to toggle pin: %w", err)
	}

	return &event, nil
}
