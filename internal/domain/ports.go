package domain

import "context"

type TripRepository interface {
	Create(ctx context.Context, trip *Trip) error
	GetByID(ctx context.Context, id int) (*Trip, error)
	List(ctx context.Context, userID *string) ([]Trip, error)
	Update(ctx context.Context, id int, updater func(*Trip) *Trip) (*Trip, error)
	Delete(ctx context.Context, id int) error
}

type EventRepository interface {
	Create(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id int) (*Event, error)
	ListByTrip(ctx context.Context, tripID int) ([]Event, error)
	Update(ctx context.Context, id int, updater func(*Event) *Event) (*Event, error)
	Delete(ctx context.Context, id int) error
}
