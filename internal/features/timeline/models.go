package timeline

import (
	"time"

	"github.com/google/uuid"
)

type Trip struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Destination string     `json:"destination"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Event struct {
	ID        uuid.UUID  `json:"id"`
	TripID    uuid.UUID  `json:"tripId"`
	Title     string     `json:"title"`
	Location  *string    `json:"location"`
	Category  *string    `json:"category"`
	GeoLat    *float64   `json:"geoLat"`
	GeoLng    *float64   `json:"geoLng"`
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
	IsPinned  bool       `json:"isPinned"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}
