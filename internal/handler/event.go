package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/service"
)

// EventCardProps carries edit-mode state for 422 re-renders.
// Nil means normal view-mode rendering (the common path from TimelineDay).
type EventCardProps struct {
	FormValues EventFormData
	Editing    bool
}

// EventFormData is used for both initial render and error re-render of the event creation form.
type EventFormData struct {
	Errors            map[string]string
	DepartureAirport  string
	FlightNumber      string
	Title             string
	Location          string
	StartTime         string
	EndTime           string
	Notes             string
	BookingReference  string
	Category          string
	ArrivalGate       string
	Airline           string
	Date              string
	ArrivalAirport    string
	DepartureTerminal string
	ArrivalTerminal   string
	DepartureGate     string
	CheckInTime       string // "2006-01-02T15:04" format, empty = not provided
	CheckOutTime      string
	TripID            int
	Pinned            bool
}

// renderEventFormError sends a 422 response with the appropriate form template.
// HTMX requests get the Sheet fragment; direct browser submissions get the full page.
func renderEventFormError(w http.ResponseWriter, r *http.Request, data *EventFormData) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	if r.Header.Get("HX-Request") == "true" {
		templ.Handler(EventCreateForm(data)).ServeHTTP(w, r)
	} else {
		templ.Handler(EventNewPage(data)).ServeHTTP(w, r)
	}
}

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

	dateStr := r.URL.Query().Get("date")
	category := r.URL.Query().Get("category")
	if category == "" {
		category = string(domain.CategoryActivity)
	}

	eventDate := parseDate(dateStr)
	if eventDate.IsZero() {
		eventDate = time.Now()
		dateStr = eventDate.Format("2006-01-02")
	} else if dateStr == "" {
		dateStr = eventDate.Format("2006-01-02")
	}

	defaults := h.eventService.SuggestDefaults(r.Context(), tripID, eventDate, domain.EventCategory(category))

	formData := &EventFormData{
		TripID:    tripID,
		Date:      dateStr,
		Category:  category,
		StartTime: defaults.StartTime.Format("15:04"),
		EndTime:   defaults.EndTime.Format("15:04"),
	}

	// Dual-path: HTMX request → Sheet fragment; otherwise → full page fallback
	if r.Header.Get("HX-Request") == "true" {
		templ.Handler(EventCreateForm(formData)).ServeHTTP(w, r)
		return
	}

	templ.Handler(EventNewPage(formData)).ServeHTTP(w, r)
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

	dateStr := r.FormValue("date")
	category := r.FormValue("category")
	title := r.FormValue("title")
	location := r.FormValue("location")
	startTimeStr := r.FormValue("start_time")
	endTimeStr := r.FormValue("end_time")
	notes := r.FormValue("notes")
	pinned := r.FormValue("pinned") == "on" || r.FormValue("pinned") == "true"

	flightDetails := parseFlightDetails(r)

	formData := &EventFormData{
		TripID:            tripID,
		Date:              dateStr,
		Category:          category,
		Title:             title,
		Location:          location,
		StartTime:         startTimeStr,
		EndTime:           endTimeStr,
		Notes:             notes,
		Pinned:            pinned,
		Airline:           flightDetails.Airline,
		FlightNumber:      flightDetails.FlightNumber,
		DepartureAirport:  flightDetails.DepartureAirport,
		ArrivalAirport:    flightDetails.ArrivalAirport,
		DepartureTerminal: flightDetails.DepartureTerminal,
		ArrivalTerminal:   flightDetails.ArrivalTerminal,
		DepartureGate:     flightDetails.DepartureGate,
		ArrivalGate:       flightDetails.ArrivalGate,
		BookingReference:  flightDetails.BookingReference,
		CheckInTime:       r.FormValue("check_in_time"),
		CheckOutTime:      r.FormValue("check_out_time"),
	}

	// Handler pre-validates required fields for field-level errors
	formErrors := make(map[string]string)
	if title == "" {
		formErrors["title"] = "Title is required"
	}
	if startTimeStr == "" {
		formErrors["start_time"] = "Start time is required"
	}
	if endTimeStr == "" {
		formErrors["end_time"] = "End time is required"
	}
	if dateStr == "" {
		formErrors["date"] = "Date is required"
	}
	if category != "" && !domain.IsValidEventCategory(domain.EventCategory(category)) {
		formErrors["category"] = "Invalid event type"
	}

	// Validate flight fields
	if category == string(domain.CategoryFlight) {
		if flightDetails.DepartureAirport == "" {
			formErrors["departure_airport"] = "Required"
		}
		if flightDetails.ArrivalAirport == "" {
			formErrors["arrival_airport"] = "Required"
		}
	}

	if len(formErrors) > 0 {
		formData.Errors = formErrors
		renderEventFormError(w, r, formData)
		return
	}

	startTime, err := parseDateAndTime(dateStr, startTimeStr)
	if err != nil {
		formErrors["start_time"] = "Invalid start time format"
		formData.Errors = formErrors
		renderEventFormError(w, r, formData)
		return
	}

	endTime, err := parseDateAndTime(dateStr, endTimeStr)
	if err != nil {
		formErrors["end_time"] = "Invalid end time format"
		formData.Errors = formErrors
		renderEventFormError(w, r, formData)
		return
	}

	var serviceFlightDetails *domain.FlightDetails
	if category == string(domain.CategoryFlight) {
		serviceFlightDetails = flightDetails
	}

	var lodgingDetails *domain.LodgingDetails
	if category == string(domain.CategoryLodging) {
		lodgingDetails = parseLodgingDetails(formData)
	}

	input := &service.CreateEventInput{
		TripID:         tripID,
		Title:          title,
		Category:       domain.EventCategory(category),
		Location:       location,
		StartTime:      startTime,
		EndTime:        endTime,
		Notes:          notes,
		Pinned:         pinned,
		FlightDetails:  serviceFlightDetails,
		LodgingDetails: lodgingDetails,
	}

	event, err := h.eventService.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			// Strip the domain error prefix to avoid leaking internals to the UI
			formErrors["general"] = strings.TrimPrefix(err.Error(), "invalid input: ")
			formData.Errors = formErrors
			renderEventFormError(w, r, formData)
			return
		}
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	// HTMX success path: return full day HTML with retarget headers
	if r.Header.Get("HX-Request") == "true" {
		eventDateStr := event.EventDate.Format("2006-01-02")

		// Fetch all events for this day to render the full day
		events, err := h.eventService.ListByTripAndDate(r.Context(), tripID, event.EventDate)
		if err != nil {
			http.Error(w, "Failed to load events", http.StatusInternalServerError)
			return
		}

		dayData := TimelineDayData{
			Date:   event.EventDate,
			Events: events,
		}

		// Set HTMX response headers for retarget to day container
		w.Header().Set("HX-Retarget", fmt.Sprintf("#day-%s", eventDateStr))
		w.Header().Set("HX-Reswap", "outerHTML")
		w.Header().Set("HX-Trigger", `{"close-sheet": true}`)

		templ.Handler(TimelineDay(tripID, dayData)).ServeHTTP(w, r)
		return
	}

	// Non-HTMX fallback: redirect
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
	tripID, err := strconv.Atoi(tripIDStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Fetch event first — needed for oldEventDate capture and 422 re-render
	event, err := h.eventService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load event", http.StatusInternalServerError)
		return
	}
	oldEventDate := event.EventDate

	if err = r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	dateStr := r.FormValue("date")
	startTimeStr := r.FormValue("start_time")
	endTimeStr := r.FormValue("end_time")
	title := r.FormValue("title")
	location := r.FormValue("location")
	notes := r.FormValue("notes")
	pinned := r.FormValue("pinned") == "on" || r.FormValue("pinned") == "true"

	flightDetails := parseFlightDetails(r)

	formData := EventFormData{
		TripID:            tripID,
		Date:              dateStr,
		Title:             title,
		Location:          location,
		StartTime:         startTimeStr,
		EndTime:           endTimeStr,
		Notes:             notes,
		Pinned:            pinned,
		Airline:           flightDetails.Airline,
		FlightNumber:      flightDetails.FlightNumber,
		DepartureAirport:  flightDetails.DepartureAirport,
		ArrivalAirport:    flightDetails.ArrivalAirport,
		DepartureTerminal: flightDetails.DepartureTerminal,
		ArrivalTerminal:   flightDetails.ArrivalTerminal,
		DepartureGate:     flightDetails.DepartureGate,
		ArrivalGate:       flightDetails.ArrivalGate,
		BookingReference:  flightDetails.BookingReference,
		CheckInTime:       r.FormValue("check_in_time"),
		CheckOutTime:      r.FormValue("check_out_time"),
	}

	formErrors := make(map[string]string)
	if title == "" {
		formErrors["title"] = "Title is required"
	}
	if startTimeStr == "" {
		formErrors["start_time"] = "Start time is required"
	}
	if endTimeStr == "" {
		formErrors["end_time"] = "End time is required"
	}
	if dateStr == "" {
		formErrors["date"] = "Date is required"
	}

	// Validate flight fields
	if event.Category == domain.CategoryFlight {
		if flightDetails.DepartureAirport == "" {
			formErrors["departure_airport"] = "Required"
		}
		if flightDetails.ArrivalAirport == "" {
			formErrors["arrival_airport"] = "Required"
		}
	}

	// renderCardError sends a 422 with the inline edit card (HTMX) or redirects to
	// the edit page (non-HTMX fallback) so the browser always gets a usable response.
	renderCardError := func(data EventFormData) {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Retarget", fmt.Sprintf("#event-%d", id))
			w.Header().Set("HX-Reswap", "outerHTML")
			w.WriteHeader(http.StatusUnprocessableEntity)
			templ.Handler(EventTimelineItem(*event, &EventCardProps{Editing: true, FormValues: data})).ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/trips/%s/events/%d/edit", tripIDStr, id), http.StatusSeeOther)
		}
	}

	if len(formErrors) > 0 {
		formData.Errors = formErrors
		renderCardError(formData)
		return
	}

	startTime, err := parseDateAndTime(dateStr, startTimeStr)
	if err != nil {
		formErrors["start_time"] = "Invalid start time format"
		formData.Errors = formErrors
		renderCardError(formData)
		return
	}

	endTime, err := parseDateAndTime(dateStr, endTimeStr)
	if err != nil {
		formErrors["end_time"] = "Invalid end time format"
		formData.Errors = formErrors
		renderCardError(formData)
		return
	}

	category := event.Category // preserve existing category
	var serviceFlightDetails *domain.FlightDetails
	if event.Category == domain.CategoryFlight {
		serviceFlightDetails = flightDetails
	}

	var lodgingDetails *domain.LodgingDetails
	if event.Category == domain.CategoryLodging {
		lodgingDetails = parseLodgingDetails(&formData)
	}

	input := &service.UpdateEventInput{
		Title:          &title,
		Category:       &category,
		Location:       &location,
		StartTime:      &startTime,
		EndTime:        &endTime,
		Notes:          &notes,
		Pinned:         &pinned,
		FlightDetails:  serviceFlightDetails,
		LodgingDetails: lodgingDetails,
	}

	updatedEvent, err := h.eventService.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			formErrors["general"] = strings.TrimPrefix(err.Error(), "invalid input: ")
			formData.Errors = formErrors
			renderCardError(formData)
			return
		}
		http.Error(w, "Failed to update event", http.StatusInternalServerError)
		return
	}

	// HTMX path
	if r.Header.Get("HX-Request") == "true" {
		// Cross-day: full-page redirect
		if !updatedEvent.EventDate.Equal(oldEventDate) {
			w.Header().Set("HX-Redirect", "/trips/"+tripIDStr)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Same day: render TimelineDay with retarget
		events, err := h.eventService.ListByTripAndDate(r.Context(), tripID, updatedEvent.EventDate)
		if err != nil {
			http.Error(w, "Failed to load events", http.StatusInternalServerError)
			return
		}
		newEventDateStr := updatedEvent.EventDate.Format("2006-01-02")
		dayData := TimelineDayData{
			Date:   updatedEvent.EventDate,
			Events: events,
		}
		w.Header().Set("HX-Retarget", fmt.Sprintf("#day-%s", newEventDateStr))
		w.Header().Set("HX-Reswap", "outerHTML")
		templ.Handler(TimelineDay(tripID, dayData)).ServeHTTP(w, r)
		return
	}

	http.Redirect(w, r, "/trips/"+tripIDStr, http.StatusSeeOther)
}

func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tripIDStr := chi.URLParam(r, "tripID")
	tripID, err := strconv.Atoi(tripIDStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Fetch before deleting to get EventDate for response
	event, err := h.eventService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load event", http.StatusInternalServerError)
		return
	}
	eventDate := event.EventDate

	if err := h.eventService.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}

	// HTMX path: return updated day HTML + trigger undo toast
	if r.Header.Get("HX-Request") == "true" {
		events, err := h.eventService.ListByTripAndDate(r.Context(), tripID, eventDate)
		if err != nil {
			http.Error(w, "Failed to load events", http.StatusInternalServerError)
			return
		}
		dayData := TimelineDayData{
			Date:   eventDate,
			Events: events,
		}
		eventDateStr := eventDate.Format("2006-01-02")
		// Cannot use HX-Trigger header here because the triggering element (delete button)
		// is removed from the DOM by the swap, so the event wouldn't bubble to window.
		// Instead, we append a script to dispatch the event directly on window.
		templ.Handler(TimelineDay(tripID, dayData)).ServeHTTP(w, r)
		fmt.Fprintf(w, `<script>window.dispatchEvent(new CustomEvent('showundotoast', {detail: {"eventId": %d, "tripId": %d, "eventDate": "%s"}}));</script>`, id, tripID, eventDateStr)
		return
	}

	http.Redirect(w, r, "/trips/"+tripIDStr, http.StatusSeeOther)
}

func (h *EventHandler) Restore(w http.ResponseWriter, r *http.Request) {
	tripIDStr := chi.URLParam(r, "tripID")
	tripID, err := strconv.Atoi(tripIDStr)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	event, err := h.eventService.Restore(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to restore event", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/trips/"+tripIDStr, http.StatusSeeOther)
		return
	}

	events, err := h.eventService.ListByTripAndDate(r.Context(), tripID, event.EventDate)
	if err != nil {
		http.Error(w, "Failed to load events", http.StatusInternalServerError)
		return
	}
	dayData := TimelineDayData{
		Date:   event.EventDate,
		Events: events,
	}
	eventDateStr := event.EventDate.Format("2006-01-02")
	w.Header().Set("HX-Retarget", fmt.Sprintf("#day-%s", eventDateStr))
	w.Header().Set("HX-Reswap", "outerHTML")
	w.Header().Set("HX-Trigger", `{"hideUndoToast": true}`)
	templ.Handler(TimelineDay(tripID, dayData)).ServeHTTP(w, r)
}

func parseLodgingDetails(formData *EventFormData) *domain.LodgingDetails {
	ld := &domain.LodgingDetails{
		BookingReference: formData.BookingReference,
	}
	if formData.CheckInTime != "" {
		t, err := time.ParseInLocation("2006-01-02T15:04", formData.CheckInTime, time.UTC)
		if err == nil {
			ld.CheckInTime = &t
		}
	}
	if formData.CheckOutTime != "" {
		t, err := time.ParseInLocation("2006-01-02T15:04", formData.CheckOutTime, time.UTC)
		if err == nil {
			ld.CheckOutTime = &t
		}
	}
	return ld
}

func parseFlightDetails(r *http.Request) *domain.FlightDetails {
	return &domain.FlightDetails{
		Airline:           r.FormValue("airline"),
		FlightNumber:      r.FormValue("flight_number"),
		DepartureAirport:  r.FormValue("departure_airport"),
		ArrivalAirport:    r.FormValue("arrival_airport"),
		DepartureTerminal: r.FormValue("departure_terminal"),
		ArrivalTerminal:   r.FormValue("arrival_terminal"),
		DepartureGate:     r.FormValue("departure_gate"),
		ArrivalGate:       r.FormValue("arrival_gate"),
		BookingReference:  r.FormValue("booking_reference"),
	}
}
