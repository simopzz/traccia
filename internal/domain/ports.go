package domain

import (
	"context"
	"time"
)

// DateEventCount holds an event count for a specific date.
type DateEventCount struct {
	Date  time.Time
	Count int
}

type TripRepository interface {
	Create(ctx context.Context, trip *Trip) error
	GetByID(ctx context.Context, id int) (*Trip, error)
	List(ctx context.Context, userID *string) ([]Trip, error)
	Update(ctx context.Context, id int, updater func(*Trip) *Trip) (*Trip, error)
	Delete(ctx context.Context, id int) error
	CountEventsByTripAndDateRange(ctx context.Context, tripID int, newStart, newEnd time.Time) (int, error)
	CountEventsByTripGroupedByDate(ctx context.Context, tripID int, newStart, newEnd time.Time) ([]DateEventCount, error)
}

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id int) (*Event, error)
	ListByTrip(ctx context.Context, tripID int) ([]Event, error)
	ListByTripAndDate(ctx context.Context, tripID int, date time.Time) ([]Event, error)
	Update(ctx context.Context, id int, updater func(*Event) *Event) (*Event, error)
	Delete(ctx context.Context, id int) error
	Restore(ctx context.Context, id int) (*Event, error)
	CountByTrip(ctx context.Context, tripID int) (int, error)
}
