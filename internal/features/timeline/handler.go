package timeline

import (
	"fmt"
	"net/http"
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

	templ.Handler(View(trip)).ServeHTTP(w, r)
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
