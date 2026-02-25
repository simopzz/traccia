package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/service"
)

type TripHandler struct {
	tripService  *service.TripService
	eventService *service.EventService
}

func NewTripHandler(tripService *service.TripService, eventService *service.EventService) *TripHandler {
	return &TripHandler{
		tripService:  tripService,
		eventService: eventService,
	}
}

func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	trips, err := h.tripService.List(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to load trips", http.StatusInternalServerError)
		return
	}

	templ.Handler(TripListPage(trips)).ServeHTTP(w, r)
}

func (h *TripHandler) NewPage(w http.ResponseWriter, r *http.Request) {
	templ.Handler(TripNewPage(nil, nil)).ServeHTTP(w, r)
}

func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	input := &service.CreateTripInput{
		Name:        r.FormValue("name"),
		Destination: r.FormValue("destination"),
		StartDate:   parseDate(r.FormValue("start_date")),
		EndDate:     parseDate(r.FormValue("end_date")),
	}

	trip, err := h.tripService.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			templ.Handler(TripNewPage(input, newFormErrors(err))).ServeHTTP(w, r)
			return
		}
		http.Error(w, "Failed to create trip", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/trips/"+strconv.Itoa(trip.ID), http.StatusSeeOther)
}

// TimelineDayData holds a day's date, day number, and events for timeline rendering.
type TimelineDayData struct {
	Date      time.Time
	Events    []domain.Event
	DayNumber int
}

func (h *TripHandler) Detail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	trip, err := h.tripService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Trip not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load trip", http.StatusInternalServerError)
		return
	}

	events, err := h.eventService.ListByTrip(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to load events", http.StatusInternalServerError)
		return
	}

	// Build day-by-day timeline from trip date range
	days := buildTimelineDays(trip, events)

	templ.Handler(TripDetailPage(trip, days)).ServeHTTP(w, r)
}

func (h *TripHandler) EditPage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	trip, err := h.tripService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Trip not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load trip", http.StatusInternalServerError)
		return
	}

	eventCount, err := h.eventService.CountByTrip(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to count events", http.StatusInternalServerError)
		return
	}

	templ.Handler(TripEditPage(trip, eventCount, nil)).ServeHTTP(w, r)
}

func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	if err = r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	destination := r.FormValue("destination")
	startDate := parseDate(r.FormValue("start_date"))
	endDate := parseDate(r.FormValue("end_date"))

	input := service.UpdateTripInput{
		Name:        &name,
		Destination: &destination,
		StartDate:   &startDate,
		EndDate:     &endDate,
	}

	_, err = h.tripService.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Trip not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrInvalidInput) || errors.Is(err, domain.ErrDateRangeConflict) {
			trip, getErr := h.tripService.GetByID(r.Context(), id)
			if getErr != nil {
				http.Error(w, "Failed to load trip", http.StatusInternalServerError)
				return
			}
			// Overlay user's form input so the form preserves what they typed
			trip.Name = name
			trip.Destination = destination
			trip.StartDate = startDate
			trip.EndDate = endDate
			eventCount, countErr := h.eventService.CountByTrip(r.Context(), id)
			if countErr != nil {
				http.Error(w, "Failed to count events", http.StatusInternalServerError)
				return
			}
			templ.Handler(TripEditPage(trip, eventCount, newFormErrors(err))).ServeHTTP(w, r)
			return
		}
		http.Error(w, "Failed to update trip", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/trips/"+strconv.Itoa(id))
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/trips/"+strconv.Itoa(id), http.StatusSeeOther)
}

func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	if err := h.tripService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Trip not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete trip", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// buildTimelineDays generates a slice of TimelineDayData from trip's date range, distributing events by date.
func buildTimelineDays(trip *domain.Trip, events []domain.Event) []TimelineDayData {
	// Build event lookup by date
	eventsByDate := make(map[string][]domain.Event)
	for i := range events {
		key := events[i].EventDate.Format("2006-01-02")
		eventsByDate[key] = append(eventsByDate[key], events[i])
	}

	var days []TimelineDayData
	dayNum := 1
	for d := trip.StartDate; !d.After(trip.EndDate); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		days = append(days, TimelineDayData{
			Date:      d,
			DayNumber: dayNum,
			Events:    eventsByDate[key],
		})
		dayNum++
	}
	return days
}

// FormErrors holds form validation error messages.
type FormErrors struct {
	General string
}

func newFormErrors(err error) *FormErrors {
	return &FormErrors{General: err.Error()}
}
