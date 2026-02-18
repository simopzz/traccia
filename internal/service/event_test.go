package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/service"
)

// mockEventRepo implements service.EventStore for testing.
type mockEventRepo struct {
	events    map[int]*domain.Event
	nextID    int
	lastEvent *domain.Event
}

func newMockEventRepo() *mockEventRepo {
	return &mockEventRepo{
		events: make(map[int]*domain.Event),
		nextID: 1,
	}
}

func (m *mockEventRepo) Create(_ context.Context, event *domain.Event) error {
	event.ID = m.nextID
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()
	if event.Position == 0 {
		event.Position = m.nextID * 1000
	}
	m.events[event.ID] = event
	m.nextID++
	return nil
}

func (m *mockEventRepo) GetByID(_ context.Context, id int) (*domain.Event, error) {
	e, ok := m.events[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *e
	return &cp, nil
}

func (m *mockEventRepo) ListByTrip(_ context.Context, tripID int) ([]domain.Event, error) {
	var result []domain.Event
	for _, e := range m.events {
		if e.TripID == tripID {
			result = append(result, *e)
		}
	}
	return result, nil
}

func (m *mockEventRepo) ListByTripAndDate(_ context.Context, tripID int, date time.Time) ([]domain.Event, error) {
	var result []domain.Event
	for _, e := range m.events {
		if e.TripID == tripID && e.EventDate.Equal(date) {
			result = append(result, *e)
		}
	}
	return result, nil
}

func (m *mockEventRepo) Update(_ context.Context, id int, updater func(*domain.Event) *domain.Event) (*domain.Event, error) {
	e, ok := m.events[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	updated := updater(e)
	updated.UpdatedAt = time.Now()
	m.events[id] = updated
	return updated, nil
}

func (m *mockEventRepo) Delete(_ context.Context, id int) error {
	if _, ok := m.events[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.events, id)
	return nil
}

func (m *mockEventRepo) CountByTrip(_ context.Context, tripID int) (int, error) {
	count := 0
	for _, e := range m.events {
		if e.TripID == tripID {
			count++
		}
	}
	return count, nil
}

func (m *mockEventRepo) GetLastEventByTrip(_ context.Context, _ int) (*domain.Event, error) {
	if m.lastEvent != nil {
		return m.lastEvent, nil
	}
	return nil, domain.ErrNotFound
}

func TestEventService_Create(t *testing.T) {
	baseStart := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	baseEnd := time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		input   *service.CreateEventInput
		wantErr error
	}{
		{
			name: "valid event",
			input: &service.CreateEventInput{
				TripID:    1,
				Title:     "Visit Colosseum",
				Category:  domain.CategoryActivity,
				StartTime: baseStart,
				EndTime:   baseEnd,
			},
			wantErr: nil,
		},
		{
			name: "missing title",
			input: &service.CreateEventInput{
				TripID:    1,
				Category:  domain.CategoryActivity,
				StartTime: baseStart,
				EndTime:   baseEnd,
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "missing trip ID",
			input: &service.CreateEventInput{
				Title:     "Event",
				Category:  domain.CategoryActivity,
				StartTime: baseStart,
				EndTime:   baseEnd,
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "missing start time",
			input: &service.CreateEventInput{
				TripID:   1,
				Title:    "Event",
				Category: domain.CategoryActivity,
				EndTime:  baseEnd,
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "missing end time",
			input: &service.CreateEventInput{
				TripID:    1,
				Title:     "Event",
				Category:  domain.CategoryActivity,
				StartTime: baseStart,
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "end time before start time",
			input: &service.CreateEventInput{
				TripID:    1,
				Title:     "Event",
				Category:  domain.CategoryActivity,
				StartTime: baseEnd,
				EndTime:   baseStart,
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "invalid category",
			input: &service.CreateEventInput{
				TripID:    1,
				Title:     "Event",
				Category:  "invalid",
				StartTime: baseStart,
				EndTime:   baseEnd,
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "empty category defaults to activity",
			input: &service.CreateEventInput{
				TripID:    1,
				Title:     "Event",
				StartTime: baseStart,
				EndTime:   baseEnd,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockEventRepo()
			svc := service.NewEventService(repo)

			event, err := svc.Create(context.Background(), tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				}
				if event != nil {
					t.Error("Create() returned event on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}
			if event.ID == 0 {
				t.Error("Create() event.ID should be non-zero")
			}
			if event.Title != tt.input.Title {
				t.Errorf("Create() Title = %q, want %q", event.Title, tt.input.Title)
			}
		})
	}
}

func TestEventService_Create_ActivityCategory(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	input := &service.CreateEventInput{
		TripID:    1,
		Title:     "Visit Museum",
		Category:  domain.CategoryActivity,
		Location:  "National Museum",
		StartTime: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
		Notes:     "Bring camera",
		Pinned:    false,
	}

	event, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	if event.Category != domain.CategoryActivity {
		t.Errorf("Category = %q, want %q", event.Category, domain.CategoryActivity)
	}
	if event.Location != "National Museum" {
		t.Errorf("Location = %q, want %q", event.Location, "National Museum")
	}
	// Verify EventDate is derived from StartTime
	wantDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	if !event.EventDate.Equal(wantDate) {
		t.Errorf("EventDate = %v, want %v", event.EventDate, wantDate)
	}
}

func TestEventService_Create_FoodCategory(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	input := &service.CreateEventInput{
		TripID:    1,
		Title:     "Lunch at Trattoria",
		Category:  domain.CategoryFood,
		Location:  "Trastevere",
		StartTime: time.Date(2026, 5, 1, 12, 30, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 5, 1, 14, 0, 0, 0, time.UTC),
	}

	event, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	if event.Category != domain.CategoryFood {
		t.Errorf("Category = %q, want %q", event.Category, domain.CategoryFood)
	}
	// Verify EventDate is derived from StartTime
	wantDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	if !event.EventDate.Equal(wantDate) {
		t.Errorf("EventDate = %v, want %v", event.EventDate, wantDate)
	}
}

// TestEventService_Create_MultipleEvents verifies that multiple events can be created
// for the same trip without error and each receives a unique non-zero ID.
// Position assignment is a repository concern and is tested at the integration level.
func TestEventService_Create_MultipleEvents(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	inputs := []*service.CreateEventInput{
		{
			TripID:    1,
			Title:     "First Event",
			Category:  domain.CategoryActivity,
			StartTime: time.Date(2026, 5, 1, 9, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC),
		},
		{
			TripID:    1,
			Title:     "Second Event",
			Category:  domain.CategoryFood,
			StartTime: time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 5, 1, 13, 30, 0, 0, time.UTC),
		},
	}

	ids := make(map[int]bool)
	for i, input := range inputs {
		event, err := svc.Create(context.Background(), input)
		if err != nil {
			t.Fatalf("Create event %d: %v", i+1, err)
		}
		if event.ID == 0 {
			t.Errorf("Event %d: ID should be non-zero", i+1)
		}
		if ids[event.ID] {
			t.Errorf("Event %d: duplicate ID %d", i+1, event.ID)
		}
		ids[event.ID] = true
	}
}

func TestEventService_Update(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mockEventRepo)
		id      int
		input   *service.UpdateEventInput
		wantErr error
	}{
		{
			name: "valid update",
			setup: func(r *mockEventRepo) {
				r.events[1] = &domain.Event{
					ID:        1,
					TripID:    1,
					Title:     "Old Title",
					Category:  domain.CategoryActivity,
					StartTime: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2026, 5, 1, 11, 0, 0, 0, time.UTC),
				}
			},
			id:      1,
			input:   &service.UpdateEventInput{Title: strPtr("New Title")},
			wantErr: nil,
		},
		{
			name:    "not found",
			setup:   func(_ *mockEventRepo) {},
			id:      999,
			input:   &service.UpdateEventInput{Title: strPtr("X")},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockEventRepo()
			tt.setup(repo)
			svc := service.NewEventService(repo)

			event, err := svc.Update(context.Background(), tt.id, tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Update() unexpected error: %v", err)
			}
			if tt.input.Title != nil && event.Title != *tt.input.Title {
				t.Errorf("Update() Title = %q, want %q", event.Title, *tt.input.Title)
			}
		})
	}
}

func TestEventService_SuggestDefaults(t *testing.T) {
	eventDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		setup         func(*mockEventRepo)
		category      domain.EventCategory
		wantStartHour int
		wantStartMin  int
		wantDuration  time.Duration
	}{
		{
			name:          "first event of day defaults to 9:00 AM + activity duration",
			setup:         func(_ *mockEventRepo) {},
			category:      domain.CategoryActivity,
			wantStartHour: 9,
			wantStartMin:  0,
			wantDuration:  2 * time.Hour,
		},
		{
			name:          "first event of day defaults to 9:00 AM + food duration",
			setup:         func(_ *mockEventRepo) {},
			category:      domain.CategoryFood,
			wantStartHour: 9,
			wantStartMin:  0,
			wantDuration:  90 * time.Minute,
		},
		{
			name: "subsequent event uses latest end time as start",
			setup: func(r *mockEventRepo) {
				r.events[1] = &domain.Event{
					ID:        1,
					TripID:    1,
					EventDate: eventDate,
					StartTime: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
				}
			},
			category:      domain.CategoryActivity,
			wantStartHour: 12,
			wantStartMin:  0,
			wantDuration:  2 * time.Hour,
		},
		{
			name: "picks latest end time across multiple events",
			setup: func(r *mockEventRepo) {
				r.events[1] = &domain.Event{
					ID:        1,
					TripID:    1,
					EventDate: eventDate,
					StartTime: time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
					Position:  1000,
				}
				r.events[2] = &domain.Event{
					ID:        2,
					TripID:    1,
					EventDate: eventDate,
					StartTime: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2026, 5, 1, 14, 0, 0, 0, time.UTC),
					Position:  2000,
				}
			},
			category:      domain.CategoryFood,
			wantStartHour: 14,
			wantStartMin:  0,
			wantDuration:  90 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockEventRepo()
			tt.setup(repo)
			svc := service.NewEventService(repo)

			defaults := svc.SuggestDefaults(context.Background(), 1, eventDate, tt.category)

			if defaults.StartTime.Hour() != tt.wantStartHour || defaults.StartTime.Minute() != tt.wantStartMin {
				t.Errorf("StartTime = %s, want %02d:%02d", defaults.StartTime.Format("15:04"), tt.wantStartHour, tt.wantStartMin)
			}

			gotDuration := defaults.EndTime.Sub(defaults.StartTime)
			if gotDuration != tt.wantDuration {
				t.Errorf("Duration = %v, want %v", gotDuration, tt.wantDuration)
			}
		})
	}
}

func TestEventService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mockEventRepo)
		id      int
		wantErr error
	}{
		{
			name: "delete existing event",
			setup: func(r *mockEventRepo) {
				r.events[1] = &domain.Event{ID: 1, TripID: 1, Title: "Test"}
			},
			id:      1,
			wantErr: nil,
		},
		{
			name:    "delete non-existent event",
			setup:   func(_ *mockEventRepo) {},
			id:      999,
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockEventRepo()
			tt.setup(repo)
			svc := service.NewEventService(repo)

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
