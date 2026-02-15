package service

import (
	"context"
	"fmt"
	"time"

	"github.com/simopzz/traccia/internal/domain"
)

type TripService struct {
	repo domain.TripRepository
}

func NewTripService(repo domain.TripRepository) *TripService {
	return &TripService{repo: repo}
}

type CreateTripInput struct {
	StartDate   time.Time
	EndDate     time.Time
	Name        string
	Destination string
}

func (s *TripService) Create(ctx context.Context, input *CreateTripInput) (*domain.Trip, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	if input.StartDate.IsZero() {
		return nil, fmt.Errorf("%w: start date is required", domain.ErrInvalidInput)
	}
	if input.EndDate.IsZero() {
		return nil, fmt.Errorf("%w: end date is required", domain.ErrInvalidInput)
	}
	if input.EndDate.Before(input.StartDate) {
		return nil, fmt.Errorf("%w: end date must be on or after start date", domain.ErrInvalidInput)
	}

	trip := &domain.Trip{
		Name:        input.Name,
		Destination: input.Destination,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
	}

	if err := s.repo.Create(ctx, trip); err != nil {
		return nil, err
	}

	return trip, nil
}

func (s *TripService) GetByID(ctx context.Context, id int) (*domain.Trip, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TripService) List(ctx context.Context, userID *string) ([]domain.Trip, error) {
	return s.repo.List(ctx, userID)
}

type UpdateTripInput struct {
	Name        *string
	Destination *string
	StartDate   *time.Time
	EndDate     *time.Time
}

func (s *TripService) Update(ctx context.Context, id int, input UpdateTripInput) (*domain.Trip, error) {
	if input.Name != nil && *input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	if input.StartDate != nil && input.StartDate.IsZero() {
		return nil, fmt.Errorf("%w: start date is required", domain.ErrInvalidInput)
	}
	if input.EndDate != nil && input.EndDate.IsZero() {
		return nil, fmt.Errorf("%w: end date is required", domain.ErrInvalidInput)
	}
	if input.StartDate != nil && input.EndDate != nil && input.EndDate.Before(*input.StartDate) {
		return nil, fmt.Errorf("%w: end date must be on or after start date", domain.ErrInvalidInput)
	}

	return s.repo.Update(ctx, id, func(trip *domain.Trip) *domain.Trip {
		if input.Name != nil {
			trip.Name = *input.Name
		}
		if input.Destination != nil {
			trip.Destination = *input.Destination
		}
		if input.StartDate != nil {
			trip.StartDate = *input.StartDate
		}
		if input.EndDate != nil {
			trip.EndDate = *input.EndDate
		}
		return trip
	})
}

// ValidateDateRangeShrink checks if shrinking a trip's date range would exclude days with events.
// It only queries the database when the range actually shrinks (new start after old start or new end before old end).
func (s *TripService) ValidateDateRangeShrink(ctx context.Context, tripID int, oldStart, oldEnd, newStart, newEnd time.Time) error {
	// Only validate when range actually shrinks
	if !newStart.After(oldStart) && !newEnd.Before(oldEnd) {
		return nil
	}

	affectedDays, err := s.repo.CountEventsByTripGroupedByDate(ctx, tripID, newStart, newEnd)
	if err != nil {
		return fmt.Errorf("checking events outside date range: %w", err)
	}
	if len(affectedDays) == 0 {
		return nil
	}

	msg := "cannot shorten trip: "
	for i, day := range affectedDays {
		if i > 0 {
			msg += "; "
		}
		msg += fmt.Sprintf("%s has %d event(s)", day.Date.Format("Mon, Jan 2"), day.Count)
	}
	msg += ". Remove or move them first"
	return fmt.Errorf("%w: %s", domain.ErrDateRangeConflict, msg)
}

func (s *TripService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
