package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"

	z "github.com/Oudwins/zog"
	"github.com/Oudwins/zog/conf"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/service"
)

var htmxDateTimeCoercer = conf.TimeCoercerFactory(func(val string) (time.Time, error) {
	// HTMX datetime-local inputs use "2006-01-02T15:04"
	return time.Parse("2006-01-02T15:04", val)
})

var eventDiscriminatorSchema = z.Struct(z.Shape{
	"Category": z.String().Required(),
})

var baseEventShape = z.Shape{
	"Date":      z.String().Required(z.Message("date is required")),
	"Title":     z.String().Required(z.Message("title is required")),
	"Location":  z.String().Optional(),
	"StartTime": z.String().Required(z.Message("start time is required")),
	"EndTime":   z.String().Required(z.Message("end time is required")),
	"Notes":     z.String().Optional(),
	"Pinned":    z.Bool().Optional(),
	"Category":  z.String().Optional(),
}

var activityEventSchema = z.Struct(baseEventShape)

var flightEventSchema = z.Struct(baseEventShape).Extend(z.Shape{
	"Airline":           z.String().Optional(),
	"FlightNumber":      z.String().Optional(),
	"DepartureAirport":  z.String().Required(z.Message("departure airport is required")),
	"ArrivalAirport":    z.String().Required(z.Message("arrival airport is required")),
	"DepartureTerminal": z.String().Optional(),
	"ArrivalTerminal":   z.String().Optional(),
	"DepartureGate":     z.String().Optional(),
	"ArrivalGate":       z.String().Optional(),
	"BookingReference":  z.String().Optional(),
})

var lodgingEventSchema = z.Struct(baseEventShape).Extend(z.Shape{
	"CheckInTime":      z.Time(z.WithCoercer(htmxDateTimeCoercer)).Optional(),
	"CheckOutTime":     z.Time(z.WithCoercer(htmxDateTimeCoercer)).Optional(),
	"BookingReference": z.String().Optional(),
})

var transitEventSchema = z.Struct(baseEventShape).Extend(z.Shape{
	"Origin":        z.String().Optional(),
	"Destination":   z.String().Optional(),
	"TransportMode": z.String().Optional(),
})

// EventCardProps carries edit-mode state for 422 re-renders.
// Nil means normal view-mode rendering (the common path from TimelineDay).
type EventCardProps struct {
	FormValues EventFormData
	Editing    bool
}

// EventFormData is used for both initial render and error re-render of the event creation form.
type EventFormData struct {
	Errors            map[string]string    `zog:"-"`
	DepartureAirport  string               `zog:"departure_airport"`
	FlightNumber      string               `zog:"flight_number"`
	Title             string               `zog:"title"`
	Location          string               `zog:"location"`
	StartTime         string               `zog:"start_time"`
	EndTime           string               `zog:"end_time"`
	Notes             string               `zog:"notes"`
	BookingReference  string               `zog:"booking_reference"`
	Category          domain.EventCategory // Ignored, mapped through wrapper
	ArrivalGate       string               `zog:"arrival_gate"`
	Airline           string               `zog:"airline"`
	Date              string               `zog:"date"`
	ArrivalAirport    string               `zog:"arrival_airport"`
	DepartureTerminal string               `zog:"departure_terminal"`
	ArrivalTerminal   string               `zog:"arrival_terminal"`
	DepartureGate     string               `zog:"departure_gate"`
	CheckInTime       string               `zog:"check_in_time"`
	CheckOutTime      string               `zog:"check_out_time"`
	Origin            string               `zog:"origin"`
	Destination       string               `zog:"destination"`
	TransportMode     string               `zog:"transport_mode"`
	TripID            int                  `zog:"trip_id"`
	Pinned            bool                 `zog:"pinned"`
}

// EventFormPayload avoids type cast panics extracting string payloads over domain enum aliases
type EventFormPayload struct {
	Category string `zog:"category"`
	EventFormData
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
	tripID, _, err := parseTripID(r)
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
		Category:  domain.EventCategory(category),
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
	tripID, tripIDStr, err := parseTripID(r)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	formPayload, formErrors, err := parseEventForm(r, domain.CategoryActivity)
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	formData := formPayload.EventFormData
	formData.Category = domain.EventCategory(formPayload.Category)
	formData.TripID = tripID

	if len(formErrors) > 0 {
		formData.Errors = formErrors
		renderEventFormError(w, r, &formData)
		return
	}

	details, err := buildCategoryDetails(formData.Category, &formData)
	if err != nil {
		formData.Errors = map[string]string{"general": "Invalid lodging time format"}
		renderEventFormError(w, r, &formData)
		return
	}
	if len(details.FormErrors) > 0 {
		formData.Errors = details.FormErrors
		renderEventFormError(w, r, &formData)
		return
	}

	startTime, err := parseDateAndTime(formData.Date, formData.StartTime)
	if err != nil {
		formData.Errors = map[string]string{"start_time": "Invalid start time format"}
		renderEventFormError(w, r, &formData)
		return
	}

	endTime, err := parseDateAndTime(formData.Date, formData.EndTime)
	if err != nil {
		formData.Errors = map[string]string{"end_time": "Invalid end time format"}
		renderEventFormError(w, r, &formData)
		return
	}

	input := &service.CreateEventInput{
		TripID:         tripID,
		Title:          formData.Title,
		Category:       formData.Category,
		Location:       formData.Location,
		StartTime:      startTime,
		EndTime:        endTime,
		Notes:          formData.Notes,
		Pinned:         formData.Pinned,
		FlightDetails:  details.FlightDetails,
		LodgingDetails: details.LodgingDetails,
		TransitDetails: details.TransitDetails,
	}

	if errs := service.CreateEventSchema.Validate(input); len(errs) > 0 {
		formData.Errors = mapServiceSchemaErrors(errs)
		renderEventFormError(w, r, &formData)
		return
	}

	event, err := h.eventService.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			formData.Errors = map[string]string{"general": strings.TrimPrefix(err.Error(), "invalid input: ")}
			renderEventFormError(w, r, &formData)
			return
		}
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		if err := h.renderDayResponse(w, r, tripID, event.EventDate, map[string]string{"HX-Trigger": `{"close-sheet": true}`}); err != nil {
			http.Error(w, "Failed to load events", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/trips/"+tripIDStr, http.StatusSeeOther)
}

func (h *EventHandler) EditPage(w http.ResponseWriter, r *http.Request) {
	id, _, err := parseEventID(r)
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
	tripID, tripIDStr, err := parseTripID(r)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	id, _, err := parseEventID(r)
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

	formPayload, formErrors, err := parseEventForm(r, event.Category)
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	formData := formPayload.EventFormData
	formData.Category = event.Category // preserve existing category, not from form
	formData.TripID = tripID

	if len(formErrors) > 0 {
		formData.Errors = formErrors
		h.renderCardError(w, r, tripIDStr, id, event, &formData)
		return
	}

	details, err := buildCategoryDetails(event.Category, &formData)
	if err != nil {
		formData.Errors = map[string]string{"general": "Invalid lodging time format"}
		h.renderCardError(w, r, tripIDStr, id, event, &formData)
		return
	}
	if len(details.FormErrors) > 0 {
		formData.Errors = details.FormErrors
		h.renderCardError(w, r, tripIDStr, id, event, &formData)
		return
	}

	startTime, err := parseDateAndTime(formData.Date, formData.StartTime)
	if err != nil {
		formData.Errors = map[string]string{"start_time": "Invalid start time format"}
		h.renderCardError(w, r, tripIDStr, id, event, &formData)
		return
	}

	endTime, err := parseDateAndTime(formData.Date, formData.EndTime)
	if err != nil {
		formData.Errors = map[string]string{"end_time": "Invalid end time format"}
		h.renderCardError(w, r, tripIDStr, id, event, &formData)
		return
	}

	category := event.Category // preserve existing category
	input := &service.UpdateEventInput{
		Title:          &formData.Title,
		Category:       &category,
		Location:       &formData.Location,
		StartTime:      &startTime,
		EndTime:        &endTime,
		Notes:          &formData.Notes,
		Pinned:         &formData.Pinned,
		FlightDetails:  details.FlightDetails,
		LodgingDetails: details.LodgingDetails,
		TransitDetails: details.TransitDetails,
	}

	if errs := service.UpdateEventSchema.Validate(input); len(errs) > 0 {
		formData.Errors = mapServiceSchemaErrors(errs)
		h.renderCardError(w, r, tripIDStr, id, event, &formData)
		return
	}

	updatedEvent, err := h.eventService.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrInvalidInput) {
			formData.Errors = map[string]string{"general": strings.TrimPrefix(err.Error(), "invalid input: ")}
			h.renderCardError(w, r, tripIDStr, id, event, &formData)
			return
		}
		http.Error(w, "Failed to update event", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		// Cross-day: full-page redirect
		if !updatedEvent.EventDate.Equal(oldEventDate) {
			w.Header().Set("HX-Redirect", "/trips/"+tripIDStr)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Same day: render TimelineDay with retarget
		if err := h.renderDayResponse(w, r, tripID, updatedEvent.EventDate, nil); err != nil {
			http.Error(w, "Failed to load events", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/trips/"+tripIDStr, http.StatusSeeOther)
}

func (h *EventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tripID, tripIDStr, err := parseTripID(r)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	id, _, err := parseEventID(r)
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

	if r.Header.Get("HX-Request") == "true" {
		if err := h.renderDayResponse(w, r, tripID, eventDate, nil); err != nil {
			http.Error(w, "Failed to load events", http.StatusInternalServerError)
			return
		}
		// Cannot use HX-Trigger header here because the triggering element (delete button)
		// is removed from the DOM by the swap, so the event wouldn't bubble to window.
		// Instead, we append a script to dispatch the event directly on window.
		eventDateStr := eventDate.Format("2006-01-02")
		fmt.Fprintf(w, `<script>window.dispatchEvent(new CustomEvent('showundotoast', {detail: {"eventId": %d, "tripId": %d, "eventDate": "%s"}}));</script>`, id, tripID, eventDateStr)
		return
	}

	http.Redirect(w, r, "/trips/"+tripIDStr, http.StatusSeeOther)
}

func (h *EventHandler) Restore(w http.ResponseWriter, r *http.Request) {
	tripID, tripIDStr, err := parseTripID(r)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	id, _, err := parseEventID(r)
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

	if err := h.renderDayResponse(w, r, tripID, event.EventDate, map[string]string{"HX-Trigger": `{"hideundotoast": true}`}); err != nil {
		http.Error(w, "Failed to load events", http.StatusInternalServerError)
	}
}

func mapZogErrorsToForm(errs z.ZogIssueList) map[string]string {
	formErrors := make(map[string]string)
	for _, e := range errs {
		path := strings.Join(e.Path, ".")
		// Zog lowercase's keys from struct tags, but sometimes uses TitleCase if no tag is provided for TopLevel.
		// For our HTML forms we generally want lowercase underscore maps matching HTML name= attributes.
		key := strings.ToLower(path)

		// Overrides matching exact struct names mapped back to snake_case for UI.
		// E.g 'DepartureAirport' -> 'departure_airport'
		switch path {
		case "DepartureAirport":
			key = "departure_airport"
		case "ArrivalAirport":
			key = "arrival_airport"
		case "StartTime":
			key = "start_time"
		case "EndTime":
			key = "end_time"
		case "Date":
			key = "date"
		case "TripID":
			key = "general"
		}

		formErrors[key] = e.Message
	}
	return formErrors
}

func parseFlightDetails(formData *EventFormData) (fd *domain.FlightDetails, formErrors map[string]string) {
	fd = &domain.FlightDetails{
		Airline:           formData.Airline,
		FlightNumber:      formData.FlightNumber,
		DepartureAirport:  formData.DepartureAirport,
		ArrivalAirport:    formData.ArrivalAirport,
		DepartureTerminal: formData.DepartureTerminal,
		ArrivalTerminal:   formData.ArrivalTerminal,
		DepartureGate:     formData.DepartureGate,
		ArrivalGate:       formData.ArrivalGate,
		BookingReference:  formData.BookingReference,
	}
	formErrors = make(map[string]string)
	if fd.DepartureAirport == "" {
		formErrors["departure_airport"] = "Required"
	}
	if fd.ArrivalAirport == "" {
		formErrors["arrival_airport"] = "Required"
	}
	return fd, formErrors
}

func parseLodgingDetails(formData *EventFormData) (*domain.LodgingDetails, error) {
	ld := &domain.LodgingDetails{
		BookingReference: formData.BookingReference,
	}
	if formData.CheckInTime != "" {
		t, err := time.ParseInLocation("2006-01-02T15:04", formData.CheckInTime, time.UTC)
		if err != nil {
			return nil, fmt.Errorf("parsing check-in time: %w", err)
		}
		ld.CheckInTime = &t
	}
	if formData.CheckOutTime != "" {
		t, err := time.ParseInLocation("2006-01-02T15:04", formData.CheckOutTime, time.UTC)
		if err != nil {
			return nil, fmt.Errorf("parsing check-out time: %w", err)
		}
		ld.CheckOutTime = &t
	}
	return ld, nil
}

func parseTransitDetails(formData *EventFormData) *domain.TransitDetails {
	return &domain.TransitDetails{
		Origin:        formData.Origin,
		Destination:   formData.Destination,
		TransportMode: formData.TransportMode,
	}
}
