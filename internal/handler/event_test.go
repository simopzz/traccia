package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/service"
)

// mockEventRepo for handler testing
type mockEventRepo struct {
	event         *domain.Event
	capturedEvent *domain.Event
}

func (m *mockEventRepo) Create(ctx context.Context, event *domain.Event) error {
	event.ID = 1
	event.EventDate = event.StartTime.Truncate(24 * time.Hour)
	m.capturedEvent = event
	return nil
}
func (m *mockEventRepo) GetByID(ctx context.Context, id int) (*domain.Event, error) {
	if m.event != nil && m.event.ID == id {
		return m.event, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mockEventRepo) ListByTrip(ctx context.Context, tripID int) ([]domain.Event, error) {
	return nil, nil
}
func (m *mockEventRepo) ListByTripAndDate(ctx context.Context, tripID int, date time.Time) ([]domain.Event, error) {
	if m.event != nil && m.event.TripID == tripID && m.event.EventDate.Equal(date) {
		// return empty list to simulate deletion for the list view
		return []domain.Event{}, nil
	}
	return []domain.Event{}, nil
}
func (m *mockEventRepo) Update(ctx context.Context, id int, updater func(*domain.Event) *domain.Event) (*domain.Event, error) {
	return nil, nil
}
func (m *mockEventRepo) Delete(ctx context.Context, id int) error {
	return nil
}
func (m *mockEventRepo) Restore(ctx context.Context, id int) (*domain.Event, error) {
	return nil, nil
}
func (m *mockEventRepo) CountByTrip(ctx context.Context, tripID int) (int, error) {
	return 0, nil
}
func (m *mockEventRepo) GetLastEventByTrip(ctx context.Context, tripID int) (*domain.Event, error) {
	return nil, nil
}

func TestEventHandler_Delete_ScriptInjection(t *testing.T) {
	eventDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	event := &domain.Event{
		ID:        1,
		TripID:    1,
		EventDate: eventDate,
		Title:     "Test Event",
	}

	repo := &mockEventRepo{event: event}
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	// Create request
	r := httptest.NewRequest("DELETE", "/trips/1/events/1", nil)
	r.Header.Set("HX-Request", "true")

	// Setup Chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("tripID", "1")
	rctx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Delete() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()

	// Check for script tag
	expectedScript := `<script>window.dispatchEvent(new CustomEvent('showundotoast', {detail: {"eventId": 1, "tripId": 1, "eventDate": "2026-05-01"}}));</script>`
	if !strings.Contains(body, expectedScript) {
		t.Errorf("Delete() body missing script tag.\nGot: %s\nWant substring: %s", body, expectedScript)
	}

	// Check for HX-Trigger header (should NOT be present)
	if w.Header().Get("HX-Trigger") != "" {
		t.Errorf("Delete() unexpected HX-Trigger header: %s", w.Header().Get("HX-Trigger"))
	}
}

func TestEventHandler_Create_Flight(t *testing.T) {
	repo := &mockEventRepo{}
	svc := service.NewEventService(repo)
	h := NewEventHandler(svc)

	form := strings.NewReader("title=Flight to Paris&date=2026-06-01&start_time=10:00&end_time=12:00&category=flight&airline=BA&flight_number=123&departure_airport=LHR&arrival_airport=CDG")
	r := httptest.NewRequest("POST", "/trips/1/events", form)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("HX-Request", "true")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("tripID", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.Create(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Create() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify flight details were parsed and passed through to the repository
	if repo.capturedEvent == nil {
		t.Fatal("Create() did not call repo.Create")
	}
	if repo.capturedEvent.Flight == nil {
		t.Fatal("Create() event.Flight is nil — flight details not passed to repo")
	}
	if repo.capturedEvent.Flight.Airline != "BA" {
		t.Errorf("Flight.Airline = %q, want %q", repo.capturedEvent.Flight.Airline, "BA")
	}
	if repo.capturedEvent.Flight.FlightNumber != "123" {
		t.Errorf("Flight.FlightNumber = %q, want %q", repo.capturedEvent.Flight.FlightNumber, "123")
	}
	if repo.capturedEvent.Flight.DepartureAirport != "LHR" {
		t.Errorf("Flight.DepartureAirport = %q, want %q", repo.capturedEvent.Flight.DepartureAirport, "LHR")
	}
	if repo.capturedEvent.Flight.ArrivalAirport != "CDG" {
		t.Errorf("Flight.ArrivalAirport = %q, want %q", repo.capturedEvent.Flight.ArrivalAirport, "CDG")
	}
}
