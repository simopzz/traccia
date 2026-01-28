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
