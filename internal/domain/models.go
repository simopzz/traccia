package domain

import "time"

type Trip struct {
	ID          int
	Name        string
	Destination string
	StartDate   time.Time
	EndDate     time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type EventCategory string

const (
	CategoryActivity EventCategory = "activity"
	CategoryFood     EventCategory = "food"
	CategoryLodging  EventCategory = "lodging"
	CategoryTransit  EventCategory = "transit"
)

type Event struct {
	ID        int
	TripID    int
	Title     string
	Category  EventCategory
	Location  string
	Latitude  *float64 // nullable for optional coordinates
	Longitude *float64
	StartTime time.Time
	EndTime   time.Time
	Pinned    bool
	Position  int // for ordering within trip
	CreatedAt time.Time
	UpdatedAt time.Time
}
