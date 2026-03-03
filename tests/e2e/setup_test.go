//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/playwright-community/playwright-go"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/handler"
	"github.com/simopzz/traccia/internal/repository"
	"github.com/simopzz/traccia/internal/service"
)

var (
	testPort   string
	testServer *http.Server
	dbPool     *pgxpool.Pool
	tripRepo   domain.TripRepository
	eventRepo  domain.EventRepository
)

func TestMain(m *testing.M) {
	// Change to project root so static files can be served
	if err := os.Chdir("../.."); err != nil {
		log.Fatalf("could not change to project root: %v", err)
	}

	// 1. Setup Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://traccia:traccia@localhost:5432/traccia_test?sslmode=disable"
	}

	ctx := context.Background()
	var err error
	dbPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	// 2. Setup server
	tripRepo = repository.NewTripStore(dbPool)

	// Setting up event store with specific detail stores
	flightStore := repository.NewFlightDetailsStore()
	lodgingStore := repository.NewLodgingDetailsStore()
	transitStore := repository.NewTransitDetailsStore()
	eventRepo = repository.NewEventStore(dbPool, flightStore, lodgingStore, transitStore)

	tripService := service.NewTripService(tripRepo)
	eventService := service.NewEventService(eventRepo.(*repository.EventStore))
	tripHandler := handler.NewTripHandler(tripService, eventService)
	eventHandler := handler.NewEventHandler(eventService)

	// Context Key type for UserID (similar to actual implementation)
	type contextKey string
	const userContextKey contextKey = "user_id"

	// Test Middleware to inject a mock UserID = 1 into context
	mockUserMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), userContextKey, 1)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	router := handler.NewRouter(tripHandler, eventHandler)

	// Wrap the router with our mock user middleware
	mux := chi.NewRouter()
	mux.Use(mockUserMiddleware)
	mux.Mount("/", router)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to bind port: %v\n", err)
	}
	testPort = fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)

	testServer = &http.Server{Handler: mux}
	go func() {
		if err := testServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	// 3. Playwright install
	err = playwright.Install()
	if err != nil {
		log.Fatalf("could not install playwright drivers: %v", err)
	}

	code := m.Run()

	testServer.Shutdown(ctx)
	os.Exit(code)
}

func setupBrowser(t *testing.T) (*playwright.Playwright, playwright.Browser, playwright.BrowserContext, playwright.Page) {
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Args: []string{"--no-sandbox", "--disable-setuid-sandbox"},
	})
	if err != nil {
		t.Fatalf("could not launch browser: %v", err)
	}

	context, err := browser.NewContext()
	if err != nil {
		t.Fatalf("could not create context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}

	return pw, browser, context, page
}

func teardownBrowser(pw *playwright.Playwright, browser playwright.Browser, context playwright.BrowserContext) {
	context.Close()
	browser.Close()
	pw.Stop()
}
