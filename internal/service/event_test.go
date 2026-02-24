package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/service"
)

// mockEventRepo implements service.EventStore for testing.
type mockEventRepo struct {
	events    map[int]*domain.Event
	deletedAt map[int]bool
	lastEvent *domain.Event
	nextID    int
}

func newMockEventRepo() *mockEventRepo {
	return &mockEventRepo{
		events:    make(map[int]*domain.Event),
		deletedAt: make(map[int]bool),
		nextID:    1,
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
	for id, e := range m.events {
		if e.TripID == tripID && e.EventDate.Equal(date) && !m.deletedAt[id] {
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
	m.deletedAt[id] = true
	return nil
}

func (m *mockEventRepo) Restore(_ context.Context, id int) (*domain.Event, error) {
	e, ok := m.events[id]
	if !ok || !m.deletedAt[id] {
		return nil, domain.ErrNotFound
	}
	delete(m.deletedAt, id)
	cp := *e
	return &cp, nil
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
		wantErr error
		input   *service.CreateEventInput
		name    string
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
		wantErr error
		setup   func(*mockEventRepo)
		input   *service.UpdateEventInput
		name    string
		id      int
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
			name:          "first event of day defaults to 9:00 AM + transit duration",
			setup:         func(_ *mockEventRepo) {},
			category:      domain.CategoryTransit,
			wantStartHour: 9,
			wantStartMin:  0,
			wantDuration:  30 * time.Minute,
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
		wantErr error
		setup   func(*mockEventRepo)
		name    string
		id      int
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

// Tests for Story 1.3: event_date recalculation, soft delete, and restore.

func TestEventService_Update_EventDateRecalculation(t *testing.T) {
	eventDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 3, 10, 11, 0, 0, 0, time.UTC)

	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:        1,
		TripID:    1,
		EventDate: eventDate,
		StartTime: startTime,
		EndTime:   endTime,
		Title:     "Test",
		Category:  domain.CategoryActivity,
	}
	svc := service.NewEventService(repo)

	newStart := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	newEnd := time.Date(2026, 3, 11, 12, 0, 0, 0, time.UTC)
	updated, err := svc.Update(context.Background(), 1, &service.UpdateEventInput{
		StartTime: &newStart,
		EndTime:   &newEnd,
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}

	wantDate := time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC)
	if !updated.EventDate.Equal(wantDate) {
		t.Errorf("EventDate = %v, want %v", updated.EventDate, wantDate)
	}
}

func TestEventService_Update_EventDateUnchangedWhenOnlyTitleUpdated(t *testing.T) {
	eventDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	startTime := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC)

	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:        1,
		TripID:    1,
		EventDate: eventDate,
		StartTime: startTime,
		EndTime:   startTime.Add(2 * time.Hour),
		Title:     "Old Title",
		Category:  domain.CategoryActivity,
	}
	svc := service.NewEventService(repo)

	updated, err := svc.Update(context.Background(), 1, &service.UpdateEventInput{
		Title: strPtr("New Title"),
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if !updated.EventDate.Equal(eventDate) {
		t.Errorf("EventDate changed unexpectedly: got %v, want %v", updated.EventDate, eventDate)
	}
	if updated.Title != "New Title" {
		t.Errorf("Title = %q, want %q", updated.Title, "New Title")
	}
}

func TestEventService_Update_InvalidEndTimeBeforeStartTime(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	start := time.Date(2026, 3, 10, 11, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC) // before start

	_, err := svc.Update(context.Background(), 1, &service.UpdateEventInput{
		StartTime: &start,
		EndTime:   &end,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("Update() error = %v, want ErrInvalidInput", err)
	}
}

func TestEventService_Update_NotFound(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	title := "X"
	_, err := svc.Update(context.Background(), 999, &service.UpdateEventInput{Title: &title})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Update() error = %v, want ErrNotFound", err)
	}
}

func TestEventService_Update_OnlyStartTimeMovedPastEndTime(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:        1,
		TripID:    1,
		EventDate: time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		StartTime: time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 3, 10, 11, 0, 0, 0, time.UTC),
		Title:     "Test",
		Category:  domain.CategoryActivity,
	}
	svc := service.NewEventService(repo)

	// Move StartTime to 12:00, past existing EndTime of 11:00 — no EndTime provided.
	newStart := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	_, err := svc.Update(context.Background(), 1, &service.UpdateEventInput{
		StartTime: &newStart,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("Update() error = %v, want ErrInvalidInput", err)
	}
}

func TestEventService_DeleteAndRestoreRoundTrip(t *testing.T) {
	eventDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:        1,
		TripID:    1,
		EventDate: eventDate,
		Title:     "Test Event",
		Category:  domain.CategoryActivity,
	}
	svc := service.NewEventService(repo)

	// Delete (soft)
	if err := svc.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}

	// Verify not in ListByTripAndDate
	events, err := svc.ListByTripAndDate(context.Background(), 1, eventDate)
	if err != nil {
		t.Fatalf("ListByTripAndDate() unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events after delete, got %d", len(events))
	}

	// Restore
	restored, err := svc.Restore(context.Background(), 1)
	if err != nil {
		t.Fatalf("Restore() unexpected error: %v", err)
	}
	if restored.ID != 1 {
		t.Errorf("Restore() ID = %d, want 1", restored.ID)
	}

	// Verify back in ListByTripAndDate
	events, err = svc.ListByTripAndDate(context.Background(), 1, eventDate)
	if err != nil {
		t.Fatalf("ListByTripAndDate() after restore unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event after restore, got %d", len(events))
	}
}

func TestEventService_Restore_NotFound(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	_, err := svc.Restore(context.Background(), 999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Restore() error = %v, want ErrNotFound", err)
	}
}

func TestEventService_ListByTripAndDate_ExcludesSoftDeleted(t *testing.T) {
	eventDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{ID: 1, TripID: 1, EventDate: eventDate, Title: "Keep"}
	repo.events[2] = &domain.Event{ID: 2, TripID: 1, EventDate: eventDate, Title: "Deleted"}
	repo.deletedAt[2] = true
	svc := service.NewEventService(repo)

	events, err := svc.ListByTripAndDate(context.Background(), 1, eventDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Title != "Keep" {
		t.Errorf("got event %q, want %q", events[0].Title, "Keep")
	}
}

// Tests for Story 1.4: Flight event creation, nil-safety, and update.

func TestEventService_Create_FlightEvent_PopulatesFlightDetails(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	input := &service.CreateEventInput{
		TripID:    1,
		Title:     "London to Paris",
		Category:  domain.CategoryFlight,
		StartTime: time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 6, 1, 10, 30, 0, 0, time.UTC),
		FlightDetails: &domain.FlightDetails{
			Airline:          "BA",
			FlightNumber:     "234",
			DepartureAirport: "LHR",
			ArrivalAirport:   "CDG",
			DepartureGate:    "A12",
		},
	}

	event, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if event.Flight == nil {
		t.Fatal("Create() event.Flight is nil, expected non-nil for flight event")
	}
	if event.Flight.Airline != "BA" {
		t.Errorf("Flight.Airline = %q, want %q", event.Flight.Airline, "BA")
	}
	if event.Flight.DepartureAirport != "LHR" {
		t.Errorf("Flight.DepartureAirport = %q, want %q", event.Flight.DepartureAirport, "LHR")
	}
	if event.Flight.ArrivalAirport != "CDG" {
		t.Errorf("Flight.ArrivalAirport = %q, want %q", event.Flight.ArrivalAirport, "CDG")
	}
}

func TestEventService_Create_FlightEvent_NilDetailsDefaultsToEmpty(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	input := &service.CreateEventInput{
		TripID:        1,
		Title:         "Mystery Flight",
		Category:      domain.CategoryFlight,
		StartTime:     time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC),
		EndTime:       time.Date(2026, 6, 1, 11, 0, 0, 0, time.UTC),
		FlightDetails: nil, // explicitly nil
	}

	event, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if event.Flight == nil {
		t.Fatal("Create() event.Flight is nil for flight category, expected empty FlightDetails{}")
	}
	// Fields should be empty strings, not a panic
	if event.Flight.Airline != "" {
		t.Errorf("Flight.Airline = %q, want empty", event.Flight.Airline)
	}
}

func TestEventService_Update_FlightEvent_UpdatesFlightDetails(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:        1,
		TripID:    1,
		Title:     "London to Paris",
		Category:  domain.CategoryFlight,
		StartTime: time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 6, 1, 10, 30, 0, 0, time.UTC),
		Flight: &domain.FlightDetails{
			Airline:      "BA",
			FlightNumber: "234",
		},
	}
	svc := service.NewEventService(repo)

	updatedDetails := &domain.FlightDetails{
		Airline:      "LH",
		FlightNumber: "100",
	}
	event, err := svc.Update(context.Background(), 1, &service.UpdateEventInput{
		FlightDetails: updatedDetails,
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if event.Flight == nil {
		t.Fatal("Update() event.Flight is nil")
	}
	if event.Flight.Airline != "LH" {
		t.Errorf("Flight.Airline = %q, want %q", event.Flight.Airline, "LH")
	}
	if event.Flight.FlightNumber != "100" {
		t.Errorf("Flight.FlightNumber = %q, want %q", event.Flight.FlightNumber, "100")
	}
}

func TestEventService_Update_NonFlightEvent_NilDetailsUnchanged(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:        1,
		TripID:    1,
		Title:     "Walk in Park",
		Category:  domain.CategoryActivity,
		StartTime: time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		Flight:    nil,
	}
	svc := service.NewEventService(repo)

	// Provide nil FlightDetails — should not change anything
	event, err := svc.Update(context.Background(), 1, &service.UpdateEventInput{
		Title:         strPtr("Updated Walk"),
		FlightDetails: nil,
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if event.Flight != nil {
		t.Error("Update() event.Flight should remain nil for non-flight event")
	}
	if event.Title != "Updated Walk" {
		t.Errorf("Title = %q, want %q", event.Title, "Updated Walk")
	}
}

// Tests for Story 1.5: Lodging event creation, nil-safety, and update.

func TestEventService_Create_LodgingEvent_PopulatesLodgingDetails(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	checkIn := time.Date(2026, 6, 1, 15, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 6, 5, 11, 0, 0, 0, time.UTC)

	input := &service.CreateEventInput{
		TripID:    1,
		Title:     "Grand Hotel",
		Category:  domain.CategoryLodging,
		StartTime: checkIn,
		EndTime:   checkOut,
		LodgingDetails: &domain.LodgingDetails{
			CheckInTime:      &checkIn,
			CheckOutTime:     &checkOut,
			BookingReference: "HTL12345",
		},
	}

	event, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if event.Lodging == nil {
		t.Fatal("Create() event.Lodging is nil, expected non-nil for lodging event")
	}
	if event.Lodging.BookingReference != "HTL12345" {
		t.Errorf("Lodging.BookingReference = %q, want %q", event.Lodging.BookingReference, "HTL12345")
	}
}

func TestEventService_Create_LodgingEvent_NilDetailsDefaultsToEmpty(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	checkIn := time.Date(2026, 6, 1, 15, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 6, 5, 11, 0, 0, 0, time.UTC)

	input := &service.CreateEventInput{
		TripID:         1,
		Title:          "Mystery Hotel",
		Category:       domain.CategoryLodging,
		StartTime:      checkIn,
		EndTime:        checkOut,
		LodgingDetails: nil, // explicitly nil
	}

	event, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if event.Lodging == nil {
		t.Fatal("Create() event.Lodging is nil for lodging category, expected empty LodgingDetails{}")
	}
	if event.Lodging.BookingReference != "" {
		t.Errorf("Lodging.BookingReference = %q, want empty", event.Lodging.BookingReference)
	}
}

func TestEventService_Update_LodgingEvent_UpdatesLodgingDetails(t *testing.T) {
	oldCheckIn := time.Date(2026, 6, 1, 15, 0, 0, 0, time.UTC)
	oldCheckOut := time.Date(2026, 6, 5, 11, 0, 0, 0, time.UTC)

	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:        1,
		TripID:    1,
		Title:     "Grand Hotel",
		Category:  domain.CategoryLodging,
		StartTime: oldCheckIn,
		EndTime:   oldCheckOut,
		Lodging: &domain.LodgingDetails{
			BookingReference: "ABC123",
			CheckInTime:      &oldCheckIn,
			CheckOutTime:     &oldCheckOut,
		},
	}
	svc := service.NewEventService(repo)

	updatedDetails := &domain.LodgingDetails{
		BookingReference: "XYZ789",
	}
	event, err := svc.Update(context.Background(), 1, &service.UpdateEventInput{
		LodgingDetails: updatedDetails,
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if event.Lodging == nil {
		t.Fatal("Update() event.Lodging is nil")
	}
	if event.Lodging.BookingReference != "XYZ789" {
		t.Errorf("Lodging.BookingReference = %q, want %q", event.Lodging.BookingReference, "XYZ789")
	}
}

func TestEventService_Update_NonLodgingEvent_NilLodgingDetailsUnchanged(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:        1,
		TripID:    1,
		Title:     "Walk in Park",
		Category:  domain.CategoryActivity,
		StartTime: time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		Lodging:   nil,
	}
	svc := service.NewEventService(repo)

	event, err := svc.Update(context.Background(), 1, &service.UpdateEventInput{
		Title:          strPtr("Updated Walk"),
		LodgingDetails: nil,
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if event.Lodging != nil {
		t.Error("Update() event.Lodging should remain nil for non-lodging event")
	}
	if event.Title != "Updated Walk" {
		t.Errorf("Title = %q, want %q", event.Title, "Updated Walk")
	}
}

func TestEventService_SuggestDefaults_FlightDuration(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	eventDate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	defaults := svc.SuggestDefaults(context.Background(), 1, eventDate, domain.CategoryFlight)

	gotDuration := defaults.EndTime.Sub(defaults.StartTime)
	wantDuration := 3 * time.Hour
	if gotDuration != wantDuration {
		t.Errorf("Flight duration = %v, want %v", gotDuration, wantDuration)
	}
}

func TestEventService_Create_Lodging_Validation(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	checkIn := time.Date(2026, 6, 1, 15, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 6, 1, 11, 0, 0, 0, time.UTC) // Before check-in

	input := &service.CreateEventInput{
		TripID:    1,
		Title:     "Invalid Times",
		Category:  domain.CategoryLodging,
		StartTime: checkIn,
		EndTime:   checkIn.Add(2 * time.Hour), // Valid base times
		LodgingDetails: &domain.LodgingDetails{
			CheckInTime:  &checkIn,
			CheckOutTime: &checkOut,
		},
	}

	_, err := svc.Create(context.Background(), input)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("Create() error = %v, want ErrInvalidInput", err)
	}
	if err != nil && !strings.Contains(err.Error(), "check-out time must be after check-in time") {
		t.Errorf("Create() error message mismatch: %v", err)
	}
}

func TestEventService_Update_Lodging_Validation(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:       1,
		TripID:   1,
		Category: domain.CategoryLodging,
		Title:    "Hotel",
	}
	svc := service.NewEventService(repo)

	checkIn := time.Date(2026, 6, 1, 15, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 6, 1, 11, 0, 0, 0, time.UTC) // Before check-in

	input := &service.UpdateEventInput{
		LodgingDetails: &domain.LodgingDetails{
			CheckInTime:  &checkIn,
			CheckOutTime: &checkOut,
		},
	}

	_, err := svc.Update(context.Background(), 1, input)
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("Update() error = %v, want ErrInvalidInput", err)
	}
}

// Tests for Story 1.6: Transit event creation and update.

func TestEventService_Create_TransitEvent_PopulatesTransitDetails(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	input := &service.CreateEventInput{
		TripID:    1,
		Title:     "Shibuya to Asakusa",
		Category:  domain.CategoryTransit,
		StartTime: time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 6, 1, 9, 30, 0, 0, time.UTC),
		TransitDetails: &domain.TransitDetails{
			Origin:        "Shibuya Station",
			Destination:   "Asakusa Station",
			TransportMode: "Metro",
		},
	}

	event, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if event.Transit == nil {
		t.Fatal("Create() event.Transit is nil, expected non-nil for transit event")
	}
	if event.Transit.Origin != "Shibuya Station" {
		t.Errorf("Transit.Origin = %q, want %q", event.Transit.Origin, "Shibuya Station")
	}
	if event.Transit.Destination != "Asakusa Station" {
		t.Errorf("Transit.Destination = %q, want %q", event.Transit.Destination, "Asakusa Station")
	}
	if event.Transit.TransportMode != "Metro" {
		t.Errorf("Transit.TransportMode = %q, want %q", event.Transit.TransportMode, "Metro")
	}
}

func TestEventService_Create_TransitEvent_NilDetailsDefaultsToEmpty(t *testing.T) {
	repo := newMockEventRepo()
	svc := service.NewEventService(repo)

	input := &service.CreateEventInput{
		TripID:         1,
		Title:          "Mystery Transit",
		Category:       domain.CategoryTransit,
		StartTime:      time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC),
		EndTime:        time.Date(2026, 6, 1, 9, 30, 0, 0, time.UTC),
		TransitDetails: nil,
	}

	event, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if event.Transit == nil {
		t.Fatal("Create() event.Transit is nil for transit category, expected empty TransitDetails{}")
	}
	if event.Transit.Origin != "" {
		t.Errorf("Transit.Origin = %q, want empty", event.Transit.Origin)
	}
}

func TestEventService_Update_TransitEvent_UpdatesTransitDetails(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:        1,
		TripID:    1,
		Title:     "Shibuya to Asakusa",
		Category:  domain.CategoryTransit,
		StartTime: time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 6, 1, 9, 30, 0, 0, time.UTC),
		Transit: &domain.TransitDetails{
			Origin:        "A",
			Destination:   "B",
			TransportMode: "Metro",
		},
	}
	svc := service.NewEventService(repo)

	event, err := svc.Update(context.Background(), 1, &service.UpdateEventInput{
		TransitDetails: &domain.TransitDetails{
			Origin:        "B",
			Destination:   "C",
			TransportMode: "Bus",
		},
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if event.Transit == nil {
		t.Fatal("Update() event.Transit is nil")
	}
	if event.Transit.Origin != "B" {
		t.Errorf("Transit.Origin = %q, want %q", event.Transit.Origin, "B")
	}
	if event.Transit.TransportMode != "Bus" {
		t.Errorf("Transit.TransportMode = %q, want %q", event.Transit.TransportMode, "Bus")
	}
}

func TestEventService_Update_NonTransitEvent_NilTransitDetailsUnchanged(t *testing.T) {
	repo := newMockEventRepo()
	repo.events[1] = &domain.Event{
		ID:        1,
		TripID:    1,
		Title:     "Walk in Park",
		Category:  domain.CategoryActivity,
		StartTime: time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC),
		Transit:   nil,
	}
	svc := service.NewEventService(repo)

	event, err := svc.Update(context.Background(), 1, &service.UpdateEventInput{
		Title:          strPtr("Updated Walk"),
		TransitDetails: nil,
	})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if event.Transit != nil {
		t.Error("Update() event.Transit should remain nil for non-transit event")
	}
	if event.Title != "Updated Walk" {
		t.Errorf("Title = %q, want %q", event.Title, "Updated Walk")
	}
}
