package timeline

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

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
