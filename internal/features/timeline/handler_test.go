package timeline_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"traccia/internal/features/timeline"

	"github.com/go-chi/chi/v5"
)

func TestHandleCreateTrip(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	h := timeline.NewHandler(svc)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	form := url.Values{}
	form.Add("name", "Handler Test Trip")
	form.Add("destination", "Handler Dest")
	form.Add("start_date", "2026-01-01")
	form.Add("end_date", "2026-01-10")

	req := httptest.NewRequest("POST", "/trips", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.HasPrefix(location, "/trips/") {
		t.Errorf("expected redirection to /trips/{id}, got %s", location)
	}
}

func TestHandleResetTrip(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	h := timeline.NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	// Create trip via service
	trip, _ := svc.CreateTrip(context.Background(), timeline.CreateTripParams{Name: "Reset", Destination: "X"})

	// Call Reset
	req := httptest.NewRequest("POST", "/trips/"+trip.ID.String()+"/reset", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", w.Code)
	}
}

func TestHandleCreateEvent(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	h := timeline.NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	// Create trip
	trip, _ := svc.CreateTrip(context.Background(), timeline.CreateTripParams{Name: "Trip", Destination: "Dest"})

	form := url.Values{}
	form.Add("title", "My Event")
	form.Add("location", "Somewhere")
	form.Add("category", "Food")
	form.Add("geo_lat", "1.23")
	form.Add("geo_lng", "4.56")
	form.Add("start_time", "2026-01-01T10:00")
	form.Add("end_time", "2026-01-01T12:00")

	url := "/trips/" + trip.ID.String() + "/events"
	req := httptest.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), "My Event") {
		t.Errorf("expected body to contain 'My Event', got %s", w.Body.String())
	}
}
