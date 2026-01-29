package timeline_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgresForSchemaTest(t *testing.T) (*sql.DB, func()) {
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

	// Run migration 000001 ONLY
	migrationSQL, err := os.ReadFile("../../../migrations/000001_create_trips_events_tables.up.sql")
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	_, err = db.ExecContext(ctx, string(migrationSQL))
	if err != nil {
		t.Fatalf("failed to apply migration: %v", err)
	}

	// Run migration 000002
	migrationSQL2, err := os.ReadFile("../../../migrations/000002_add_event_details.up.sql")
	if err != nil {
		t.Fatalf("failed to read migration file 000002: %v", err)
	}

	_, err = db.ExecContext(ctx, string(migrationSQL2))
	if err != nil {
		t.Fatalf("failed to apply migration 000002: %v", err)
	}

	return db, func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	}
}

func TestSchemaEventsColumns(t *testing.T) {
	db, teardown := startPostgresForSchemaTest(t)
	defer teardown()

	// This section checks if columns exist.
	// At this stage (Red), we expect them to be MISSING if we only applied migration 000001.
	// But the "Test" is "Do they exist?". So the test expects them to exist.
	// If they don't, the test fails -> Red.

	query := `
        SELECT column_name 
        FROM information_schema.columns 
        WHERE table_name = 'events' AND column_name IN ('category', 'geo_lat', 'geo_lng');
    `
	rows, err := db.Query(query)
	if err != nil {
		t.Fatalf("Failed to query columns: %v", err)
	}
	defer rows.Close()

	found := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("Failed to scan: %v", err)
		}
		found[name] = true
	}

	expected := []string{"category", "geo_lat", "geo_lng"}
	for _, col := range expected {
		if !found[col] {
			t.Errorf("Column %s missing in events table", col)
		}
	}
}
