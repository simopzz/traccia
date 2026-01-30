package timeline

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.handleHome)
	r.Post("/trips", h.handleCreateTrip)
	r.Get("/trips/{id}", h.handleGetTrip)
	r.Post("/trips/{id}/reset", h.handleResetTrip)
	r.Post("/trips/{id}/events", h.handleCreateEvent)
	r.Post("/trips/{id}/events/reorder", h.handleReorderEvents)
}

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request) {
	templ.Handler(Home()).ServeHTTP(w, r)
}

func (h *Handler) handleCreateTrip(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	destination := r.FormValue("destination")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")

	var startDate, endDate *time.Time

	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start date format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		startDate = &parsed
	}
	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end date format (expected YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		endDate = &parsed
	}

	trip, err := h.service.CreateTrip(r.Context(), CreateTripParams{
		Name:        name,
		Destination: destination,
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create trip: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/trips/%s", trip.ID), http.StatusSeeOther)
}

func (h *Handler) handleGetTrip(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	trip, err := h.service.GetTrip(r.Context(), id)
	if err != nil {
		if err == ErrTripNotFound {
			http.Error(w, "Trip not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get trip: %v", err), http.StatusInternalServerError)
		return
	}

	events, err := h.service.GetEvents(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get events: %v", err), http.StatusInternalServerError)
		return
	}

	templ.Handler(View(trip, events)).ServeHTTP(w, r)
}

func (h *Handler) handleResetTrip(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	err = h.service.ResetTrip(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to reset trip: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/trips/%s", id), http.StatusSeeOther)
}

func (h *Handler) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	category := r.FormValue("category")
	location := r.FormValue("location")
	geoLatStr := r.FormValue("geo_lat")
	geoLngStr := r.FormValue("geo_lng")
	startTimeStr := r.FormValue("start_time")
	endTimeStr := r.FormValue("end_time")

	var catPtr *string
	if category != "" {
		catPtr = &category
	}
	var locPtr *string
	if location != "" {
		locPtr = &location
	}

	var geoLat *float64
	if geoLatStr != "" {
		val, err := strconv.ParseFloat(geoLatStr, 64)
		if err == nil {
			geoLat = &val
		}
	}
	var geoLng *float64
	if geoLngStr != "" {
		val, err := strconv.ParseFloat(geoLngStr, 64)
		if err == nil {
			geoLng = &val
		}
	}

	var startTime *time.Time
	if startTimeStr != "" {
		parsed, err := time.Parse("2006-01-02T15:04", startTimeStr)
		if err != nil {
			parsed, err = time.Parse("2006-01-02T15:04:05", startTimeStr)
			if err != nil {
				http.Error(w, "Invalid start time format", http.StatusBadRequest)
				return
			}
		}
		startTime = &parsed
	}

	var endTime *time.Time
	if endTimeStr != "" {
		parsed, err := time.Parse("2006-01-02T15:04", endTimeStr)
		if err != nil {
			parsed, err = time.Parse("2006-01-02T15:04:05", endTimeStr)
			if err != nil {
				http.Error(w, "Invalid end time format", http.StatusBadRequest)
				return
			}
		}
		endTime = &parsed
	}

	_, err = h.service.CreateEvent(r.Context(), CreateEventParams{
		TripID:    id,
		Title:     title,
		Category:  catPtr,
		Location:  locPtr,
		GeoLat:    geoLat,
		GeoLng:    geoLng,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf("<div class='text-red-500 mb-4'>Error: %v</div>", err)))
		return
	}

	// Fetch all events to re-render the list sorted
	events, err := h.service.GetEvents(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get events: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the whole list, not just the card
	// We need a component for the list wrapper content to swap just the inner HTML of #event-list
	// But since View defines the list container, let's create a small helper or just loop here if we can't export a component easily.
	// Better: Create an EventList component in components.templ
	templ.Handler(EventList(id, events)).ServeHTTP(w, r)
}

func (h *Handler) handleReorderEvents(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	tripID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	eventIDStrs := r.Form["event_id"]
	var eventIDs []uuid.UUID
	for _, idStr := range eventIDStrs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid event ID", http.StatusBadRequest)
			return
		}
		eventIDs = append(eventIDs, id)
	}

	events, err := h.service.ReorderEvents(r.Context(), tripID, eventIDs)
	if err != nil {
		// Log the actual error (in a real app, use a logger)
		fmt.Printf("ReorderEvents error: %v\n", err)
		http.Error(w, "Failed to reorder events", http.StatusInternalServerError)
		return
	}

	templ.Handler(EventList(tripID, events)).ServeHTTP(w, r)
}
