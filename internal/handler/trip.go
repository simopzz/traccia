package handler

import (
	"errors"
	"net/http"
	"strconv"

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
	templ.Handler(TripNewPage()).ServeHTTP(w, r)
}

func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	input := service.CreateTripInput{
		Name:        r.FormValue("name"),
		Destination: r.FormValue("destination"),
		StartDate:   parseDate(r.FormValue("start_date")),
		EndDate:     parseDate(r.FormValue("end_date")),
	}

	trip, err := h.tripService.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to create trip", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/trips/"+strconv.Itoa(trip.ID), http.StatusSeeOther)
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

	templ.Handler(TripDetailPage(trip, events)).ServeHTTP(w, r)
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

	templ.Handler(TripEditPage(trip)).ServeHTTP(w, r)
}

func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
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
		http.Error(w, "Failed to update trip", http.StatusInternalServerError)
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

	w.WriteHeader(http.StatusOK)
}
