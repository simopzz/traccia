package timeline_test

import (
	"context"
	"testing"
	"time"

	"traccia/internal/features/timeline"

	"github.com/google/uuid"
)

func TestReorderEvents_OverlapPinned(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	trip, err := svc.CreateTrip(ctx, timeline.CreateTripParams{Name: "Overlap Trip", Destination: "Overlap Dest"})
	if err != nil {
		t.Fatalf("failed to create trip: %v", err)
	}

	// Helper to create event
	createEvent := func(title string, start time.Time) *timeline.Event {
		end := start.Add(1 * time.Hour)
		e, err := svc.CreateEvent(ctx, timeline.CreateEventParams{
			TripID:    trip.ID,
			Title:     title,
			StartTime: &start,
			EndTime:   &end,
		})
		if err != nil {
			t.Fatalf("failed to create event %s: %v", title, err)
		}
		return e
	}

	// Base time: 10:00
	baseTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	// A: 10:00 - 12:30 (Long event, 2.5h)
	// Create it manually to set long duration
	endA := baseTime.Add(2*time.Hour + 30*time.Minute)
	evtA, err := svc.CreateEvent(ctx, timeline.CreateEventParams{
		TripID:    trip.ID,
		Title:     "Event A",
		StartTime: &baseTime,
		EndTime:   &endA,
	})
	if err != nil {
		t.Fatalf("failed to create event A: %v", err)
	}

	// P: 12:00 - 13:00 (Pinned)
	// Note: P starts BEFORE A ends (12:00 < 12:30). This is the conflict.
	startP := baseTime.Add(2 * time.Hour)
	evtP := createEvent("Event P", startP)

	// Pin P
	_, err = db.ExecContext(ctx, "UPDATE events SET is_pinned = true WHERE id = $1", evtP.ID)
	if err != nil {
		t.Fatalf("failed to pin event P: %v", err)
	}

	// B: 13:00 - 14:00 (Starts after P)
	startB := startP.Add(1 * time.Hour)
	evtB := createEvent("Event B", startB)

	// Reorder: A, P, B (Just to trigger logic, logic should keep them here)
	newOrder := []uuid.UUID{evtA.ID, evtP.ID, evtB.ID}

	events, err := svc.ReorderEvents(ctx, trip.ID, newOrder)
	if err != nil {
		t.Fatalf("ReorderEvents failed: %v", err)
	}

	// Verify A
	// A should start at 10:00 (since it's first, or anchor).
	// A ends at 12:30.
	if !events[0].StartTime.Equal(baseTime) {
		t.Errorf("A moved? Expected %v, got %v", baseTime, events[0].StartTime)
	}

	// Verify P
	// P is Pinned at 12:00.
	// It should NOT move to 12:30 (A's end).
	// It should stay at 12:00.
	if !events[1].StartTime.Equal(startP) {
		t.Errorf("P moved or wasn't pinned! Expected %v, got %v", startP, events[1].StartTime)
	}

	// Verify B
	// B should start after P ends (13:00).
	// B should NOT start after A ends (12:30).
	// Current Time logic:
	// Loop A -> current = 12:30.
	// Loop P -> Pinned at 12:00. New Start = 12:00. New End = 13:00. current = 13:00.
	// Loop B -> New Start = 13:00.
	if !events[2].StartTime.Equal(startP.Add(1 * time.Hour)) {
		t.Errorf("B logic failed! Expected %v, got %v", startP.Add(1*time.Hour), events[2].StartTime)
	}
}
