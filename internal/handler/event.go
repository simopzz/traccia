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

type EventHandler struct {
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

func (h *EventHandler) NewPage(w http.ResponseWriter, r *http.Request) {
	tripIDStr := chi.URLParam(r, "tripID")
	tripID, err := strconv.Atoi(tripIDStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	suggestedStart := h.eventService.SuggestStartTime(r.Context(), tripID)
	templ.Handler(EventNewPage(tripID, suggestedStart)).ServeHTTP(w, r)
}

func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	tripIDStr := chi.URLParam(r, "tripID")
	tripID, err := strconv.Atoi(tripIDStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	if err = r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	input := &service.CreateEventInput{
		TripID:    tripID,
		Title:     r.FormValue("title"),
		Category:  domain.EventCategory(r.FormValue("category")),
		Location:  r.FormValue("location"),
		StartTime: parseDateTime(r.FormValue("start_time")),
		EndTime:   parseDateTime(r.FormValue("end_time")),
		Notes:     r.FormValue("notes"),
		Pinned:    r.FormValue("pinned") == "true",
	}

	_, err = h.eventService.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/trips/"+tripIDStr, http.StatusSeeOther)
}

func (h *EventHandler) EditPage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	event, err := h.eventService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load event", http.StatusInternalServerError)
		return
	}

	templ.Handler(EventEditPage(event)).ServeHTTP(w, r)
}

func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	tripIDStr := chi.URLParam(r, "tripID")
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	if err = r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	category := domain.EventCategory(r.FormValue("category"))
	location := r.FormValue("location")
	startTime := parseDateTime(r.FormValue("start_time"))
	endTime := parseDateTime(r.FormValue("end_time"))
	notes := r.FormValue("notes")
	pinned := r.FormValue("pinned") == "true"

	input := &service.UpdateEventInput{
		Title:     &title,
		Category:  &category,
		Location:  &location,
		StartTime: &startTime,
		EndTime:   &endTime,
		Notes:     &notes,
		Pinned:    &pinned,
	}

	_, err = h.eventService.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update event", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/trips/"+tripIDStr, http.StatusSeeOther)
}

func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tripIDStr := chi.URLParam(r, "tripID")
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	if err := h.eventService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/trips/"+tripIDStr, http.StatusSeeOther)
}
