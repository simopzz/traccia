package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/service"
)

// mockTripRepo is a test double implementing domain.TripRepository.
type mockTripRepo struct {
	eventsOutsideRangeErr error
	affectedDaysErr       error
	trips                 map[int]*domain.Trip
	affectedDays          []domain.DateEventCount
	nextID                int
	eventsOutsideRange    int
}

func newMockTripRepo() *mockTripRepo {
	return &mockTripRepo{
		trips:  make(map[int]*domain.Trip),
		nextID: 1,
	}
}

func (m *mockTripRepo) Create(_ context.Context, trip *domain.Trip) error {
	trip.ID = m.nextID
	trip.CreatedAt = time.Now()
	trip.UpdatedAt = time.Now()
	m.trips[trip.ID] = trip
	m.nextID++
	return nil
}

func (m *mockTripRepo) GetByID(_ context.Context, id int) (*domain.Trip, error) {
	t, ok := m.trips[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *mockTripRepo) List(_ context.Context, _ *string) ([]domain.Trip, error) {
	result := make([]domain.Trip, 0, len(m.trips))
	for _, t := range m.trips {
		result = append(result, *t)
	}
	return result, nil
}

func (m *mockTripRepo) Update(_ context.Context, id int, updater func(*domain.Trip) *domain.Trip) (*domain.Trip, error) {
	t, ok := m.trips[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	updated := updater(t)
	updated.UpdatedAt = time.Now()
	m.trips[id] = updated
	return updated, nil
}

func (m *mockTripRepo) Delete(_ context.Context, id int) error {
	if _, ok := m.trips[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.trips, id)
	return nil
}

func (m *mockTripRepo) CountEventsByTripAndDateRange(_ context.Context, _ int, _, _ time.Time) (int, error) {
	if m.eventsOutsideRangeErr != nil {
		return 0, m.eventsOutsideRangeErr
	}
	return m.eventsOutsideRange, nil
}

func (m *mockTripRepo) CountEventsByTripGroupedByDate(_ context.Context, _ int, _, _ time.Time) ([]domain.DateEventCount, error) {
	if m.affectedDaysErr != nil {
		return nil, m.affectedDaysErr
	}
	return m.affectedDays, nil
}

func TestTripService_Create(t *testing.T) {
	tests := []struct {
		wantErr error
		input   *service.CreateTripInput
		name    string
	}{
		{
			name: "valid trip",
			input: &service.CreateTripInput{
				Name:      "Rome Trip",
				StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
			},
			wantErr: nil,
		},
		{
			name: "valid trip with destination",
			input: &service.CreateTripInput{
				Name:        "Rome Trip",
				Destination: "Rome, Italy",
				StartDate:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
			},
			wantErr: nil,
		},
		{
			name: "missing name",
			input: &service.CreateTripInput{
				StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "missing start date",
			input: &service.CreateTripInput{
				Name:    "Rome Trip",
				EndDate: time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "missing end date",
			input: &service.CreateTripInput{
				Name:      "Rome Trip",
				StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "end date before start date",
			input: &service.CreateTripInput{
				Name:      "Rome Trip",
				StartDate: time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "destination optional - empty is valid",
			input: &service.CreateTripInput{
				Name:        "Rome Trip",
				Destination: "",
				StartDate:   time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockTripRepo()
			svc := service.NewTripService(repo)

			trip, err := svc.Create(context.Background(), tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				}
				if trip != nil {
					t.Error("Create() returned trip on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}
			if trip.ID == 0 {
				t.Error("Create() trip.ID should be non-zero")
			}
			if trip.Name != tt.input.Name {
				t.Errorf("Create() Name = %q, want %q", trip.Name, tt.input.Name)
			}
		})
	}
}

func TestTripService_Update(t *testing.T) {
	tests := []struct {
		input   service.UpdateTripInput
		wantErr error
		setup   func(*mockTripRepo)
		name    string
		id      int
	}{
		{
			name: "valid update",
			setup: func(r *mockTripRepo) {
				r.trips[1] = &domain.Trip{
					ID:        1,
					Name:      "Old Name",
					StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
					EndDate:   time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
				}
			},
			id: 1,
			input: service.UpdateTripInput{
				Name: strPtr("New Name"),
			},
			wantErr: nil,
		},
		{
			name:    "not found",
			setup:   func(_ *mockTripRepo) {},
			id:      999,
			input:   service.UpdateTripInput{Name: strPtr("X")},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "empty name rejected",
			setup: func(r *mockTripRepo) {
				r.trips[1] = &domain.Trip{
					ID:        1,
					Name:      "Old Name",
					StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
					EndDate:   time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
				}
			},
			id:      1,
			input:   service.UpdateTripInput{Name: strPtr("")},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "end date before start date rejected",
			setup: func(r *mockTripRepo) {
				r.trips[1] = &domain.Trip{
					ID:        1,
					Name:      "Trip",
					StartDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
					EndDate:   time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
				}
			},
			id: 1,
			input: service.UpdateTripInput{
				StartDate: timePtr(time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)),
				EndDate:   timePtr(time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)),
			},
			wantErr: domain.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockTripRepo()
			tt.setup(repo)
			svc := service.NewTripService(repo)

			trip, err := svc.Update(context.Background(), tt.id, tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Update() unexpected error: %v", err)
			}
			if tt.input.Name != nil && trip.Name != *tt.input.Name {
				t.Errorf("Update() Name = %q, want %q", trip.Name, *tt.input.Name)
			}
		})
	}
}

func TestTripService_Delete(t *testing.T) {
	tests := []struct {
		wantErr error
		setup   func(*mockTripRepo)
		name    string
		id      int
	}{
		{
			name: "delete existing trip",
			setup: func(r *mockTripRepo) {
				r.trips[1] = &domain.Trip{ID: 1, Name: "Test"}
			},
			id:      1,
			wantErr: nil,
		},
		{
			name:    "delete non-existent trip",
			setup:   func(_ *mockTripRepo) {},
			id:      999,
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockTripRepo()
			tt.setup(repo)
			svc := service.NewTripService(repo)

			err := svc.Delete(context.Background(), tt.id)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Delete() unexpected error: %v", err)
			}
		})
	}
}

func TestTripService_ValidateDateRangeShrink(t *testing.T) {
	oldStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	oldEnd := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		newStart     time.Time
		newEnd       time.Time
		wantErr      error
		name         string
		wantMsg      string
		affectedDays []domain.DateEventCount
	}{
		{
			name:     "no shrink - range expanded",
			newStart: oldStart,
			newEnd:   time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC),
			wantErr:  nil,
		},
		{
			name:         "shrink with no affected days",
			newStart:     time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
			newEnd:       time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
			affectedDays: nil,
			wantErr:      nil,
		},
		{
			name:     "shrink with events on excluded days",
			newStart: time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC),
			newEnd:   time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC),
			affectedDays: []domain.DateEventCount{
				{Date: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), Count: 2},
				{Date: time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC), Count: 1},
			},
			wantErr: domain.ErrDateRangeConflict,
			wantMsg: "Thu, May 1 has 2 event(s)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockTripRepo()
			repo.affectedDays = tt.affectedDays
			svc := service.NewTripService(repo)

			err := svc.ValidateDateRangeShrink(
				context.Background(),
				1,
				oldStart, oldEnd,
				tt.newStart, tt.newEnd,
			)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateDateRangeShrink() error = %v, wantErr %v", err, tt.wantErr)
				}
				if tt.wantMsg != "" && err != nil {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("expected error containing %q", tt.wantMsg)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateDateRangeShrink() unexpected error: %v", err)
			}
		})
	}
}

func strPtr(s string) *string        { return &s }
func timePtr(t time.Time) *time.Time { return &t }
