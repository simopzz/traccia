package timeline

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestView_SmartDefaults(t *testing.T) {
	startDate, _ := time.Parse("2006-01-02", "2026-05-01")
	trip := &Trip{
		ID:        uuid.New(),
		Name:      "Test Trip",
		StartDate: &startDate,
	}
	events := []Event{}

	component := View(trip, events)
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render view: %v", err)
	}
	output := buf.String()

	// Check that the input has the value set or x-init logic
	expectedDate := "2026-05-01"

	// We expect the form to have Alpine logic or value attribute using this date
	if !strings.Contains(output, expectedDate) {
		t.Errorf("expected trip start date %s in output", expectedDate)
	}
}

func TestView_SmartDefaults_WithPreviousEvents(t *testing.T) {
	trip := &Trip{ID: uuid.New(), Name: "Trip"}

	lastEventTime, _ := time.Parse("2006-01-02T15:04", "2026-06-15T14:30")
	events := []Event{
		{Title: "Event 1", StartTime: &lastEventTime},
	}

	component := View(trip, events)
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render view: %v", err)
	}
	output := buf.String()

	// Should default to last event time
	expectedTime := "2026-06-15T14:30"
	if !strings.Contains(output, "value=\""+expectedTime) {
		t.Errorf("expected default time %s based on last event, got output containing it? %v", expectedTime, strings.Contains(output, expectedTime))
	}

	// Check for x-model binding
	if !strings.Contains(output, "x-model=\"startTime\"") {
		t.Errorf("expected x-model=\"startTime\"")
	}
	if !strings.Contains(output, "x-model=\"endTime\"") {
		t.Errorf("expected x-model=\"endTime\"")
	}
}

func TestView_IncludesSortableAndReorderingLogic(t *testing.T) {
	trip := &Trip{
		ID:   uuid.New(),
		Name: "Test Trip",
	}
	events := []Event{
		{
			ID:    uuid.New(),
			Title: "Event 1",
		},
	}

	// Render the View component
	component := View(trip, events)
	var buf bytes.Buffer
	if err := component.Render(context.Background(), &buf); err != nil {
		t.Fatalf("failed to render view: %v", err)
	}
	output := buf.String()

	// Check for Sortable.js script
	if !strings.Contains(output, "sortable.min.js") {
		t.Errorf("expected view to include sortable.min.js")
	}

	// Check for x-data Sortable initialization
	if !strings.Contains(output, "new Sortable") {
		t.Errorf("expected view to include Sortable initialization")
	}

	// Check for hx-post to reorder endpoint
	expectedHxPost := "/trips/" + trip.ID.String() + "/events/reorder"
	if !strings.Contains(output, expectedHxPost) {
		t.Errorf("expected view to include hx-post to %s", expectedHxPost)
	}

	// Check for drag handle in EventCard
	if !strings.Contains(output, "drag-handle") {
		t.Errorf("expected view to include drag-handle class")
	}
}
