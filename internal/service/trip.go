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
func (s *TripService) ValidateDateRangeShrink(ctx context.Context, tripID int, newStart, newEnd time.Time) error {
	count, err := s.repo.CountEventsByTripAndDateRange(ctx, tripID, newStart, newEnd)
	if err != nil {
		return fmt.Errorf("checking events outside date range: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("%w: cannot shorten trip: %d event(s) exist outside the new date range. Remove or move them first", domain.ErrDateRangeConflict, count)
	}
	return nil
}

func (s *TripService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
