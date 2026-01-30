package timeline_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"traccia/internal/features/timeline"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func mustStartPostgresContainer(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	ctx := context.Background()
	dbName := "testdb"
	dbUser := "testuser"
	dbPwd := "testpassword"

	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:latest",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPwd),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open db connection: %v", err)
	}

	// Run migrations
	// Read the migration file from absolute path to be safe
	migrationSQL, err := os.ReadFile("/home/simopzz/dev/personal/traccia/migrations/000001_create_trips_events_tables.up.sql")
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	_, err = db.ExecContext(ctx, string(migrationSQL))
	if err != nil {
		t.Fatalf("failed to apply migration: %v", err)
	}

	// Run migration 000002
	migrationSQL2, err := os.ReadFile("/home/simopzz/dev/personal/traccia/migrations/000002_add_event_details.up.sql")
	if err != nil {
		t.Fatalf("failed to read migration file 000002: %v", err)
	}

	_, err = db.ExecContext(ctx, string(migrationSQL2))
	if err != nil {
		t.Fatalf("failed to apply migration 000002: %v", err)
	}

	// Run migration 000003
	migrationSQL3, err := os.ReadFile("/home/simopzz/dev/personal/traccia/migrations/000003_add_is_pinned_to_events.up.sql")
	if err != nil {
		t.Fatalf("failed to read migration file 000003: %v", err)
	}

	_, err = db.ExecContext(ctx, string(migrationSQL3))
	if err != nil {
		t.Fatalf("failed to apply migration 000003: %v", err)
	}

	return db, func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	}
}

func TestCreateTrip(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	start := time.Now().Truncate(time.Second) // Truncate because DB might lose precision or timezone diffs
	end := start.Add(24 * time.Hour)

	params := timeline.CreateTripParams{
		Name:        "Japan Trip",
		Destination: "Tokyo",
		StartDate:   &start,
		EndDate:     &end,
	}

	trip, err := svc.CreateTrip(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if trip.ID == uuid.Nil {
		t.Error("expected valid UUID")
	}
	if trip.Name != params.Name {
		t.Errorf("expected name %s, got %s", params.Name, trip.Name)
	}
}

func TestGetTrip(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	start := time.Now().Truncate(time.Second)
	end := start.Add(48 * time.Hour)
	createdTrip, err := svc.CreateTrip(ctx, timeline.CreateTripParams{
		Name:        "Get Me",
		Destination: "There",
		StartDate:   &start,
		EndDate:     &end,
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Test GetTrip
	trip, err := svc.GetTrip(ctx, createdTrip.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if trip.ID != createdTrip.ID {
		t.Errorf("expected ID %v, got %v", createdTrip.ID, trip.ID)
	}
	if trip.Name != "Get Me" {
		t.Errorf("expected name 'Get Me', got %s", trip.Name)
	}
}

func TestResetTrip(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	// Create a trip
	trip, err := svc.CreateTrip(ctx, timeline.CreateTripParams{Name: "Reset Me", Destination: "Here"})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Insert an event
	_, err = db.ExecContext(ctx, "INSERT INTO events (trip_id, title) VALUES ($1, 'Test Event')", trip.ID)
	if err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Reset Trip
	err = svc.ResetTrip(ctx, trip.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify events count is 0
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM events WHERE trip_id = $1", trip.ID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count events: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 events, got %d", count)
	}
}

func TestCreateEvent(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	// Create a trip first
	trip, err := svc.CreateTrip(ctx, timeline.CreateTripParams{Name: "Trip 1", Destination: "Dest"})
	if err != nil {
		t.Fatalf("failed to create trip: %v", err)
	}

	start := time.Now().UTC()
	end := start.Add(1 * time.Hour)
	cat := "Activity"
	loc := "Museum St"
	lat := 10.0
	lng := 20.0

	params := timeline.CreateEventParams{
		TripID:    trip.ID,
		Title:     "Visit Museum",
		Category:  &cat,
		Location:  &loc,
		GeoLat:    &lat,
		GeoLng:    &lng,
		StartTime: &start,
		EndTime:   &end,
	}

	event, err := svc.CreateEvent(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if event.Title != params.Title {
		t.Errorf("expected title %s, got %s", params.Title, event.Title)
	}
	if event.Category == nil || *event.Category != cat {
		t.Errorf("expected category %s, got %v", cat, event.Category)
	}
	if event.GeoLat == nil || *event.GeoLat != lat {
		t.Errorf("expected lat %f, got %v", lat, event.GeoLat)
	}
}

func TestCreateEventValidation(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	trip, err := svc.CreateTrip(ctx, timeline.CreateTripParams{Name: "Trip 1", Destination: "Dest"})
	if err != nil {
		t.Fatalf("failed to create trip: %v", err)
	}

	start := time.Now().UTC()
	end := start.Add(-1 * time.Hour) // End before Start

	params := timeline.CreateEventParams{
		TripID:    trip.ID,
		Title:     "Bad Event",
		StartTime: &start,
		EndTime:   &end,
	}

	_, err = svc.CreateEvent(ctx, params)
	if err == nil {
		t.Error("expected error for EndTime < StartTime, got nil")
	}
}

func TestGetEvents(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	trip, err := svc.CreateTrip(ctx, timeline.CreateTripParams{Name: "Trip 1", Destination: "Dest"})
	if err != nil {
		t.Fatalf("failed to create trip: %v", err)
	}

	params := timeline.CreateEventParams{
		TripID: trip.ID,
		Title:  "Event 1",
	}
	_, err = svc.CreateEvent(ctx, params)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	events, err := svc.GetEvents(ctx, trip.ID)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if events[0].Title != "Event 1" {
		t.Errorf("expected title 'Event 1', got %s", events[0].Title)
	}
}

func TestReorderEvents_EdgeCases(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	trip, err := svc.CreateTrip(ctx, timeline.CreateTripParams{Name: "Edge Trip", Destination: "Edges"})
	if err != nil {
		t.Fatalf("failed to create trip: %v", err)
	}

	// Helper to create event
	createEvent := func(title string) *timeline.Event {
		// Create event with no times (nil)
		e, err := svc.CreateEvent(ctx, timeline.CreateEventParams{
			TripID: trip.ID,
			Title:  title,
		})
		if err != nil {
			t.Fatalf("failed to create event %s: %v", title, err)
		}
		return e
	}

	evtA := createEvent("Event A")
	evtB := createEvent("Event B")

	// 1. Test Duplicate IDs
	_, err = svc.ReorderEvents(ctx, trip.ID, []uuid.UUID{evtA.ID, evtA.ID})
	if err == nil {
		t.Error("expected error for duplicate IDs, got nil")
	}

	// 2. Test Invalid ID (Random UUID)
	_, err = svc.ReorderEvents(ctx, trip.ID, []uuid.UUID{evtA.ID, uuid.New()})
	if err == nil {
		t.Error("expected error for invalid ID, got nil")
	}

	// 3. Test Mismatch Count
	_, err = svc.ReorderEvents(ctx, trip.ID, []uuid.UUID{evtA.ID})
	if err == nil {
		t.Error("expected error for count mismatch, got nil")
	}

	// 4. Test Nil Start Time Fallback
	// Both events have nil start time. Reorder should set them based on current time.
	reordered, err := svc.ReorderEvents(ctx, trip.ID, []uuid.UUID{evtA.ID, evtB.ID})
	if err != nil {
		t.Fatalf("unexpected error reordering nil times: %v", err)
	}

	if reordered[0].StartTime == nil {
		t.Error("expected start time to be set after reorder")
	}
	// Verify duration is DefaultEventDuration (1 hour)
	duration := reordered[0].EndTime.Sub(*reordered[0].StartTime)
	if duration != timeline.DefaultEventDuration {
		t.Errorf("expected default duration %v, got %v", timeline.DefaultEventDuration, duration)
	}
}

func TestReorderEvents(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	trip, err := svc.CreateTrip(ctx, timeline.CreateTripParams{Name: "Trip 1", Destination: "Dest"})
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

	// 10:00, 11:00, 12:00
	baseTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	evtA := createEvent("Event A", baseTime)
	evtB := createEvent("Event B", baseTime.Add(1*time.Hour))
	evtC := createEvent("Event C", baseTime.Add(2*time.Hour))

	// Reorder: C, A, B
	newOrder := []uuid.UUID{evtC.ID, evtA.ID, evtB.ID}

	events, err := svc.ReorderEvents(ctx, trip.ID, newOrder)
	if err != nil {
		t.Fatalf("ReorderEvents failed: %v", err)
	}

	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}

	// Index 0: C -> Start 10:00
	if events[0].ID != evtC.ID {
		t.Errorf("expected first event to be C")
	}
	if !events[0].StartTime.Equal(baseTime) {
		t.Errorf("expected C start time %v, got %v", baseTime, events[0].StartTime)
	}

	// Index 1: A -> Start 11:00
	expectedAStart := baseTime.Add(1 * time.Hour)
	if events[1].ID != evtA.ID {
		t.Errorf("expected second event to be A")
	}
	if !events[1].StartTime.Equal(expectedAStart) {
		t.Errorf("expected A start time %v, got %v", expectedAStart, events[1].StartTime)
	}

	// Index 2: B -> Start 12:00
	expectedBStart := baseTime.Add(2 * time.Hour)
	if events[2].ID != evtB.ID {
		t.Errorf("expected third event to be B")
	}
	if !events[2].StartTime.Equal(expectedBStart) {
		t.Errorf("expected B start time %v, got %v", expectedBStart, events[2].StartTime)
	}
}

func TestReorderEvents_Pinned(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	trip, err := svc.CreateTrip(ctx, timeline.CreateTripParams{Name: "Pinned Trip", Destination: "Pinned Dest"})
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

	// A: 10:00 - 11:00
	baseTime := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	evtA := createEvent("Event A", baseTime)

	// B: 11:30 - 12:30 (Gap of 30 mins)
	// We want to Pin B so it doesn't snap to 11:00
	evtB := createEvent("Event B", baseTime.Add(90*time.Minute))

	// Manually pin B
	_, err = db.ExecContext(ctx, "UPDATE events SET is_pinned = true WHERE id = $1", evtB.ID)
	if err != nil {
		t.Fatalf("failed to pin event B: %v", err)
	}

	// Reorder [A, B]
	// Logic should be:
	// A starts at 10:00 (Anchor). Ends 11:00.
	// B is Pinned. Should Stay at 11:30.
	// If unpinned, B would move to 11:00.
	newOrder := []uuid.UUID{evtA.ID, evtB.ID}

	events, err := svc.ReorderEvents(ctx, trip.ID, newOrder)
	if err != nil {
		t.Fatalf("ReorderEvents failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events")
	}

	// Check A
	if !events[0].StartTime.Equal(baseTime) {
		t.Errorf("A moved? Expected %v, got %v", baseTime, events[0].StartTime)
	}

	// Check B
	expectedBStart := baseTime.Add(90 * time.Minute) // 11:30
	if !events[1].StartTime.Equal(expectedBStart) {
		t.Errorf("B moved despite being pinned! Expected %v, got %v", expectedBStart, events[1].StartTime)
	}
}

func TestTogglePin(t *testing.T) {
	db, teardown := mustStartPostgresContainer(t)
	defer teardown()

	svc := timeline.NewService(db)
	ctx := context.Background()

	trip, err := svc.CreateTrip(ctx, timeline.CreateTripParams{Name: "Pin Trip", Destination: "Dest"})
	if err != nil {
		t.Fatalf("failed to create trip: %v", err)
	}

	start := time.Now().UTC()
	end := start.Add(1 * time.Hour)
	event, err := svc.CreateEvent(ctx, timeline.CreateEventParams{
		TripID:    trip.ID,
		Title:     "Pin Event",
		StartTime: &start,
		EndTime:   &end,
	})
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	if event.IsPinned {
		t.Error("expected event to be unpinned initially")
	}

	// Toggle Pin -> ON
	updatedEvent, err := svc.TogglePin(ctx, event.ID)
	if err != nil {
		t.Fatalf("failed to toggle pin: %v", err)
	}
	if !updatedEvent.IsPinned {
		t.Error("expected event to be pinned")
	}

	// Toggle Pin -> OFF
	updatedEvent2, err := svc.TogglePin(ctx, event.ID)
	if err != nil {
		t.Fatalf("failed to toggle pin: %v", err)
	}
	if updatedEvent2.IsPinned {
		t.Error("expected event to be unpinned")
	}
}
