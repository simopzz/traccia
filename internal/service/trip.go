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
	Name        string
	Destination string
	StartDate   time.Time
	EndDate     time.Time
}

func (s *TripService) Create(ctx context.Context, input CreateTripInput) (*domain.Trip, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", domain.ErrInvalidInput)
	}
	if input.Destination == "" {
		return nil, fmt.Errorf("%w: destination is required", domain.ErrInvalidInput)
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

func (s *TripService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
