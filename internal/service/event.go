package service

import (
	"context"
	"fmt"
	"time"

	"github.com/simopzz/traccia/internal/domain"
)

type EventStore interface {
	domain.EventRepository
	GetLastEventByTrip(ctx context.Context, tripID int) (*domain.Event, error)
}

// Default durations for smart time defaults per event category.
const (
	DefaultActivityDuration = 2 * time.Hour
	DefaultFoodDuration     = 90 * time.Minute
)

type EventService struct {
	repo EventStore
}

func NewEventService(repo EventStore) *EventService {
	return &EventService{repo: repo}
}

type CreateEventInput struct {
	StartTime time.Time
	EndTime   time.Time
	Title     string
	Category  domain.EventCategory
	Latitude  *float64
	Longitude *float64
	Location  string
	Notes     string
	TripID    int
	Pinned    bool
}

func (s *EventService) Create(ctx context.Context, input *CreateEventInput) (*domain.Event, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrInvalidInput)
	}
	if input.TripID <= 0 {
		return nil, fmt.Errorf("%w: trip_id is required", domain.ErrInvalidInput)
	}
	if input.StartTime.IsZero() {
		return nil, fmt.Errorf("%w: start time is required", domain.ErrInvalidInput)
	}
	if input.EndTime.IsZero() {
		return nil, fmt.Errorf("%w: end time is required", domain.ErrInvalidInput)
	}
	if input.EndTime.Before(input.StartTime) {
		return nil, fmt.Errorf("%w: end time must be on or after start time", domain.ErrInvalidInput)
	}

	if input.Category == "" {
		input.Category = domain.CategoryActivity
	}
	if !domain.IsValidEventCategory(input.Category) {
		return nil, fmt.Errorf("%w: invalid category %q", domain.ErrInvalidInput, input.Category)
	}

	event := &domain.Event{
		TripID:    input.TripID,
		EventDate: time.Date(input.StartTime.Year(), input.StartTime.Month(), input.StartTime.Day(), 0, 0, 0, 0, input.StartTime.Location()),
		Title:     input.Title,
		Category:  input.Category,
		Location:  input.Location,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
		Pinned:    input.Pinned,
		Notes:     input.Notes,
	}

	if err := s.repo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *EventService) GetByID(ctx context.Context, id int) (*domain.Event, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *EventService) ListByTrip(ctx context.Context, tripID int) ([]domain.Event, error) {
	return s.repo.ListByTrip(ctx, tripID)
}

func (s *EventService) ListByTripAndDate(ctx context.Context, tripID int, date time.Time) ([]domain.Event, error) {
	return s.repo.ListByTripAndDate(ctx, tripID, date)
}

func (s *EventService) CountByTrip(ctx context.Context, tripID int) (int, error) {
	return s.repo.CountByTrip(ctx, tripID)
}

type UpdateEventInput struct {
	Title     *string
	Category  *domain.EventCategory
	Location  *string
	Latitude  *float64
	Longitude *float64
	StartTime *time.Time
	EndTime   *time.Time
	Pinned    *bool
	Position  *int
	Notes     *string
}

func (s *EventService) Update(ctx context.Context, id int, input *UpdateEventInput) (*domain.Event, error) {
	return s.repo.Update(ctx, id, func(event *domain.Event) *domain.Event {
		if input.Title != nil {
			event.Title = *input.Title
		}
		if input.Category != nil {
			event.Category = *input.Category
		}
		if input.Location != nil {
			event.Location = *input.Location
		}
		if input.Latitude != nil {
			event.Latitude = input.Latitude
		}
		if input.Longitude != nil {
			event.Longitude = input.Longitude
		}
		if input.StartTime != nil {
			event.StartTime = *input.StartTime
			event.EventDate = time.Date(input.StartTime.Year(), input.StartTime.Month(), input.StartTime.Day(), 0, 0, 0, 0, input.StartTime.Location())
		}
		if input.EndTime != nil {
			event.EndTime = *input.EndTime
		}
		if input.Pinned != nil {
			event.Pinned = *input.Pinned
		}
		if input.Position != nil {
			event.Position = *input.Position
		}
		if input.Notes != nil {
			event.Notes = *input.Notes
		}
		return event
	})
}

func (s *EventService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

// EventDefaults holds suggested start and end times for a new event.
type EventDefaults struct {
	StartTime time.Time
	EndTime   time.Time
}

// SuggestDefaults returns smart time defaults for a new event on a given day.
// If the day has existing events, start time = latest end time among them.
// If no events exist, start time = 9:00 AM on that date.
// End time = start time + category-based duration.
func (s *EventService) SuggestDefaults(ctx context.Context, tripID int, eventDate time.Time, category domain.EventCategory) EventDefaults {
	events, err := s.repo.ListByTripAndDate(ctx, tripID, eventDate)

	var startTime time.Time
	if err != nil || len(events) == 0 {
		startTime = time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(), 9, 0, 0, 0, eventDate.Location())
	} else {
		// Find the event with the latest EndTime (not last-by-position)
		latestEnd := events[0].EndTime
		for i := range events[1:] {
			if events[i+1].EndTime.After(latestEnd) {
				latestEnd = events[i+1].EndTime
			}
		}
		startTime = latestEnd
	}

	duration := durationForCategory(category)
	return EventDefaults{
		StartTime: startTime,
		EndTime:   startTime.Add(duration),
	}
}

func durationForCategory(category domain.EventCategory) time.Duration {
	switch category {
	case domain.CategoryFood:
		return DefaultFoodDuration
	default:
		return DefaultActivityDuration
	}
}
