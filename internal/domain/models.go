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

func (c EventCategory) IsValid() bool {
	validCategories := []EventCategory{CategoryActivity, CategoryFood, CategoryLodging, CategoryTransit, CategoryFlight}
	for _, valid := range validCategories {
		if c == valid {
			return true
		}
	}
	return false
}

type LodgingDetails struct {
	CheckInTime      *time.Time
	CheckOutTime     *time.Time
	BookingReference string
	ID               int
	EventID          int
}

type FlightDetails struct {
	Airline           string
	FlightNumber      string
	DepartureAirport  string
	ArrivalAirport    string
	DepartureTerminal string
	ArrivalTerminal   string
	DepartureGate     string
	ArrivalGate       string
	BookingReference  string
	EventID           int
	ID                int
}

type TransitDetails struct {
	Origin        string
	Destination   string
	TransportMode string
	ID            int
	EventID       int
}

type Event struct {
	EventDate time.Time
	StartTime time.Time
	EndTime   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	Longitude *float64
	Latitude  *float64
	Flight    *FlightDetails
	Lodging   *LodgingDetails
	Transit   *TransitDetails
	Category  EventCategory
	Title     string
	Location  string
	Notes     string
	ID        int
	TripID    int
	Position  int
	Pinned    bool
}
