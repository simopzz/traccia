package domain

import "time"

type Trip struct {
	StartDate   time.Time
	EndDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string
	Destination string
	ID          int
}

type EventCategory string

const (
	CategoryActivity EventCategory = "activity"
	CategoryFood     EventCategory = "food"
	CategoryLodging  EventCategory = "lodging"
	CategoryTransit  EventCategory = "transit"
	CategoryFlight   EventCategory = "flight"
)

type Event struct {
	EventDate time.Time
	StartTime time.Time
	EndTime   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	Title     string
	Category  EventCategory
	Latitude  *float64
	Longitude *float64
	Location  string
	Notes     string
	ID        int
	TripID    int
	Position  int
	Pinned    bool
}
