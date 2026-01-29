package timeline_test

import (
	"context"
	"strings"
	"testing"
	"time"
	"traccia/internal/features/timeline"
)

func TestEventCardRendering(t *testing.T) {
	start, _ := time.Parse(time.RFC3339, "2026-01-01T10:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2026-01-01T11:30:00Z") // 1.5 hours
	cat := "Activity"

	event := timeline.Event{
		Title:     "Test Event",
		Category:  &cat,
		StartTime: &start,
		EndTime:   &end,
	}

	// Render to string
	var sb strings.Builder
	component := timeline.EventCard(event)
	err := component.Render(context.Background(), &sb)
	if err != nil {
		t.Fatalf("failed to render component: %v", err)
	}

	output := sb.String()

	// Check height: 1.5 * 64 = 96px
	// Look for style="height: 96px" or similar
	expectedStyle := "height: 96px"
	if !strings.Contains(output, expectedStyle) {
		t.Errorf("expected style %s in output, got %s", expectedStyle, output)
	}
}
