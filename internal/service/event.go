package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/conf"

	"github.com/simopzz/traccia/internal/domain"
)

var eventDateCoercer = conf.TimeCoercerFactory(func(val string) (time.Time, error) {
	// HTMX datetime-local inputs use "2006-01-02T15:04"
	return time.Parse("2006-01-02T15:04", val)
})

var CreateEventSchema = z.Struct(z.Shape{
	"Title":     z.String().Required(z.Message("title is required")),
	"TripID":    z.Int().Required(z.Message("trip_id is required")).GT(0, z.Message("trip_id is required")),
	"StartTime": z.Time(z.WithCoercer(eventDateCoercer)).Required(z.Message("start time is required")),
	"EndTime":   z.Time(z.WithCoercer(eventDateCoercer)).Required(z.Message("end time is required")),
	"Category":  z.StringLike[domain.EventCategory]().Optional(),
})

var UpdateEventSchema = z.Struct(z.Shape{
	"Title":     z.Ptr(z.String().Required(z.Message("title cannot be empty"))),
	"StartTime": z.Ptr(z.Time(z.WithCoercer(eventDateCoercer)).Required(z.Message("start time cannot be zero"))),
	"EndTime":   z.Ptr(z.Time(z.WithCoercer(eventDateCoercer)).Required(z.Message("end time cannot be zero"))),
	"Category":  z.Ptr(z.StringLike[domain.EventCategory]()),
})

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
	StartTime      time.Time
	EndTime        time.Time
	Latitude       *float64
	Longitude      *float64
	FlightDetails  *domain.FlightDetails
	LodgingDetails *domain.LodgingDetails
	TransitDetails *domain.TransitDetails
	Title          string
	Category       domain.EventCategory
	Location       string
	Notes          string
	TripID         int
	Pinned         bool
}

func (s *EventService) Create(ctx context.Context, input *CreateEventInput) (*domain.Event, error) {
	if errs := CreateEventSchema.Validate(input); len(errs) > 0 {
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalidInput, errs[0].Message)
	}
	if input.EndTime.Before(input.StartTime) {
		return nil, fmt.Errorf("%w: end time must be on or after start time", domain.ErrInvalidInput)
	}

	if input.Category == "" {
		input.Category = domain.CategoryActivity
	}
	if !input.Category.IsValid() {
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

	if input.Category == domain.CategoryFlight {
		event.Flight = input.FlightDetails
		if event.Flight == nil {
			event.Flight = &domain.FlightDetails{}
		}
	}

	if input.Category == domain.CategoryLodging {
		event.Lodging = input.LodgingDetails
		if event.Lodging == nil {
			event.Lodging = &domain.LodgingDetails{}
		} else if event.Lodging.CheckInTime != nil && event.Lodging.CheckOutTime != nil && !event.Lodging.CheckOutTime.After(*event.Lodging.CheckInTime) {
			return nil, fmt.Errorf("%w: check-out time must be after check-in time", domain.ErrInvalidInput)
		}
	}

	if input.Category == domain.CategoryTransit {
		event.Transit = input.TransitDetails
		if event.Transit == nil {
			event.Transit = &domain.TransitDetails{}
		}
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
	Title          *string
	Category       *domain.EventCategory
	Location       *string
	Latitude       *float64
	Longitude      *float64
	StartTime      *time.Time
	EndTime        *time.Time
	Pinned         *bool
	Position       *int
	Notes          *string
	FlightDetails  *domain.FlightDetails  // nil means "don't change flight details"
	LodgingDetails *domain.LodgingDetails // nil means "don't change lodging details"
	TransitDetails *domain.TransitDetails // nil means "don't change transit details"
}

func (s *EventService) Update(ctx context.Context, id int, input *UpdateEventInput) (*domain.Event, error) {
	if errs := UpdateEventSchema.Validate(input); len(errs) > 0 {
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalidInput, errs[0].Message)
	}

	// When both times are in the input, validate immediately (no DB fetch needed).
	if input.StartTime != nil && input.EndTime != nil && !input.EndTime.After(*input.StartTime) {
		return nil, fmt.Errorf("%w: end time must be after start time", domain.ErrInvalidInput)
	}
	// When only one time is provided, fetch the event to get the other existing
	// time and validate the combined result before writing.
	if (input.StartTime != nil) != (input.EndTime != nil) {
		existing, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		effectiveStart := existing.StartTime
		if input.StartTime != nil {
			effectiveStart = *input.StartTime
		}
		effectiveEnd := existing.EndTime
		if input.EndTime != nil {
			effectiveEnd = *input.EndTime
		}
		if !effectiveEnd.After(effectiveStart) {
			return nil, fmt.Errorf("%w: end time must be after start time", domain.ErrInvalidInput)
		}
	}

	// When both lodging times are in the input, validate before starting the update.
	if input.LodgingDetails != nil && input.LodgingDetails.CheckInTime != nil && input.LodgingDetails.CheckOutTime != nil &&
		!input.LodgingDetails.CheckOutTime.After(*input.LodgingDetails.CheckInTime) {
		return nil, fmt.Errorf("%w: lodging check-out time must be after check-in time", domain.ErrInvalidInput)
	}

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
		if input.FlightDetails != nil {
			event.Flight = input.FlightDetails
		}
		if input.LodgingDetails != nil {
			event.Lodging = input.LodgingDetails
		}
		if input.TransitDetails != nil {
			event.Transit = input.TransitDetails
		}
		return event
	})
}

func (s *EventService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *EventService) Restore(ctx context.Context, id int) (*domain.Event, error) {
	event, err := s.repo.Restore(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("restoring event %d: %w", id, err)
	}
	return event, nil
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
	switch {
	case err != nil:
		slog.WarnContext(ctx, "SuggestDefaults: failed to list events, using 9:00 AM default",
			"trip_id", tripID, "error", err)
		startTime = time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(), 9, 0, 0, 0, eventDate.Location())
	case len(events) == 0:
		startTime = time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(), 9, 0, 0, 0, eventDate.Location())
	default:
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
	case domain.CategoryFlight:
		return 3 * time.Hour
	case domain.CategoryTransit:
		return 30 * time.Minute
	default:
		return DefaultActivityDuration
	}
}
