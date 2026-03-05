package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"

	z "github.com/Oudwins/zog"
	zh "github.com/Oudwins/zog/zhttp"

	"github.com/simopzz/traccia/internal/domain"
)

// parseTripID extracts and converts the "tripID" URL param.
func parseTripID(r *http.Request) (id int, raw string, err error) {
	raw = chi.URLParam(r, "tripID")
	id, err = strconv.Atoi(raw)
	return
}

// parseEventID extracts and converts the "id" URL param.
func parseEventID(r *http.Request) (id int, raw string, err error) {
	raw = chi.URLParam(r, "id")
	id, err = strconv.Atoi(raw)
	return
}

// parseEventForm centralises ParseForm, the discriminator check, and the category-specific schema parse.
// defaultCategory is used when the form field is missing or invalid.
// Returns (payload, formErrors, err); err is non-nil only on r.ParseForm failure.
func parseEventForm(r *http.Request, defaultCategory domain.EventCategory) (EventFormPayload, map[string]string, error) {
	if err := r.ParseForm(); err != nil {
		return EventFormPayload{}, nil, err
	}

	var baseForm struct {
		Category string `zog:"category"`
	}
	if err := eventDiscriminatorSchema.Parse(zh.Request(r), &baseForm); err != nil {
		baseForm.Category = string(defaultCategory)
	}

	var payload EventFormPayload
	var formErrors map[string]string

	switch domain.EventCategory(baseForm.Category) {
	case domain.CategoryFlight:
		if errs := flightEventSchema.Parse(zh.Request(r), &payload); errs != nil {
			formErrors = mapZogErrorsToForm(errs)
		}
	case domain.CategoryLodging:
		if errs := lodgingEventSchema.Parse(zh.Request(r), &payload); errs != nil {
			formErrors = mapZogErrorsToForm(errs)
		}
	case domain.CategoryTransit:
		if errs := transitEventSchema.Parse(zh.Request(r), &payload); errs != nil {
			formErrors = mapZogErrorsToForm(errs)
		}
	default:
		if errs := activityEventSchema.Parse(zh.Request(r), &payload); errs != nil {
			formErrors = mapZogErrorsToForm(errs)
		}
	}

	payload.Category = baseForm.Category
	return payload, formErrors, nil
}

// CategoryDetailsResult holds the category-specific detail structs and any form validation
// errors produced during parsing. Adding a new category only requires a new field here.
type CategoryDetailsResult struct {
	FlightDetails  *domain.FlightDetails
	LodgingDetails *domain.LodgingDetails
	TransitDetails *domain.TransitDetails
	FormErrors     map[string]string
}

// buildCategoryDetails builds category-specific detail structs from parsed form data.
// category is authoritative (avoids Create/Update ambiguity).
// FormErrors is populated for required-field violations; err is non-nil on time-parse failure.
func buildCategoryDetails(category domain.EventCategory, d *EventFormData) (CategoryDetailsResult, error) {
	switch category {
	case domain.CategoryFlight:
		fd, formErrors := parseFlightDetails(d)
		return CategoryDetailsResult{FlightDetails: fd, FormErrors: formErrors}, nil

	case domain.CategoryLodging:
		ld, err := parseLodgingDetails(d)
		if err != nil {
			return CategoryDetailsResult{}, fmt.Errorf("parsing lodging details: %w", err)
		}
		return CategoryDetailsResult{LodgingDetails: ld}, nil

	case domain.CategoryTransit:
		return CategoryDetailsResult{TransitDetails: parseTransitDetails(d)}, nil

	default:
		return CategoryDetailsResult{}, nil
	}
}

// mapServiceSchemaErrors maps zog issue paths from service-level schema validation
// to HTML form field keys (distinct from mapZogErrorsToForm which handles handler-level).
func mapServiceSchemaErrors(errs z.ZogIssueList) map[string]string {
	formErrors := make(map[string]string)
	for _, e := range errs {
		path := strings.Join(e.Path, ".")
		switch path {
		case "Title":
			formErrors["title"] = e.Message
		case "StartTime":
			formErrors["start_time"] = e.Message
		case "EndTime":
			formErrors["end_time"] = e.Message
		case "TripID":
			formErrors["general"] = e.Message
		default:
			formErrors["general"] = e.Message
		}
	}
	return formErrors
}

// renderDayResponse fetches events for a day and renders TimelineDay with retarget headers.
// extraHeaders contains optional additional HX-* headers (e.g. HX-Trigger).
func (h *EventHandler) renderDayResponse(w http.ResponseWriter, r *http.Request, tripID int, date time.Time, extraHeaders map[string]string) error {
	events, err := h.eventService.ListByTripAndDate(r.Context(), tripID, date)
	if err != nil {
		return fmt.Errorf("listing events for day: %w", err)
	}

	dateStr := date.Format("2006-01-02")
	dayData := TimelineDayData{
		Date:   date,
		Events: events,
	}

	w.Header().Set("HX-Retarget", fmt.Sprintf("#day-%s", dateStr))
	w.Header().Set("HX-Reswap", "outerHTML")
	for k, v := range extraHeaders {
		w.Header().Set(k, v)
	}
	templ.Handler(TimelineDay(tripID, dayData)).ServeHTTP(w, r)
	return nil
}

// renderCardError sends a 422 with the inline edit card (HTMX) or redirects to
// the edit page (non-HTMX fallback).
func (h *EventHandler) renderCardError(w http.ResponseWriter, r *http.Request, tripIDStr string, id int, event *domain.Event, data *EventFormData) {
	if r.Header.Get("HX-Request") == "true" {
		if strings.Contains(r.Header.Get("HX-Current-URL"), "/edit") {
			w.Header().Set("HX-Redirect", fmt.Sprintf("/trips/%s/events/%d/edit", tripIDStr, id))
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("HX-Retarget", fmt.Sprintf("#event-%d", id))
		w.Header().Set("HX-Reswap", "outerHTML")
		w.WriteHeader(http.StatusUnprocessableEntity)
		templ.Handler(EventTimelineItem(*event, &EventCardProps{Editing: true, FormValues: *data})).ServeHTTP(w, r)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/trips/%s/events/%d/edit", tripIDStr, id), http.StatusSeeOther)
	}
}
