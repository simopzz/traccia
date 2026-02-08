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

	if input.Category == "" {
		input.Category = domain.CategoryActivity
	}

	event := &domain.Event{
		TripID:    input.TripID,
		Title:     input.Title,
		Category:  input.Category,
		Location:  input.Location,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
		Pinned:    input.Pinned,
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
}

func (s *EventService) Update(ctx context.Context, id int, input UpdateEventInput) (*domain.Event, error) {
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
		return event
	})
}

func (s *EventService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *EventService) SuggestStartTime(ctx context.Context, tripID int) time.Time {
	lastEvent, err := s.repo.GetLastEventByTrip(ctx, tripID)
	if err != nil {
		return time.Now().Truncate(time.Hour).Add(time.Hour)
	}
	return lastEvent.EndTime
}
