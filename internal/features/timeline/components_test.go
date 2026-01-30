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

	// Check for time range display "10:00 - 11:30"
	expectedTimeRange := "10:00 - 11:30"
	if !strings.Contains(output, expectedTimeRange) {
		t.Errorf("expected time range %s in output, got %s", expectedTimeRange, output)
	}

	// Check height: 1.5 * 64 = 96px
	// Look for style="height: 96px" or similar
	expectedStyle := "height: 96px"
	if !strings.Contains(output, expectedStyle) {
		t.Errorf("expected style %s in output, got %s", expectedStyle, output)
	}

	// Check for min-h classes
	if !strings.Contains(output, "min-h-[40px]") {
		t.Errorf("expected class min-h-[40px] in output for normal event, got %s", output)
	}
	if !strings.Contains(output, "max-h-[300px]") {
		t.Errorf("expected class max-h-[300px] in output, got %s", output)
	}
	// For 1.5 hours (96px), it should be overflow-hidden, NOT overflow-y-auto
	if !strings.Contains(output, "overflow-hidden") {
		t.Errorf("expected class overflow-hidden for normal event in output, got %s", output)
	}
}

func TestEventCardHeightConstraints(t *testing.T) {
	start, _ := time.Parse(time.RFC3339, "2026-01-01T10:00:00Z")

	// Case 1: Short event (15 mins)
	endShort, _ := time.Parse(time.RFC3339, "2026-01-01T10:15:00Z")
	eventShort := timeline.Event{
		Title:     "Short Event",
		StartTime: &start,
		EndTime:   &endShort,
	}

	var sbShort strings.Builder
	timeline.EventCard(eventShort).Render(context.Background(), &sbShort)
	outputShort := sbShort.String()

	// Should have min-h-[16px] for compact mode
	if !strings.Contains(outputShort, "min-h-[16px]") {
		t.Errorf("expected min-h-[16px] for short event")
	}
	if !strings.Contains(outputShort, "overflow-hidden") {
		t.Errorf("expected overflow-hidden for short event")
	}

	// Case 2: Long event (6 hours)
	endLong, _ := time.Parse(time.RFC3339, "2026-01-01T16:00:00Z")
	eventLong := timeline.Event{
		Title:     "Long Event",
		StartTime: &start,
		EndTime:   &endLong,
	}

	var sbLong strings.Builder
	timeline.EventCard(eventLong).Render(context.Background(), &sbLong)
	outputLong := sbLong.String()

	// Should have max-h-[300px] and overflow-y-auto
	if !strings.Contains(outputLong, "max-h-[300px]") {
		t.Errorf("expected max-h-[300px] for long event")
	}
	if !strings.Contains(outputLong, "overflow-y-auto") {
		t.Errorf("expected overflow-y-auto for long event")
	}
}
