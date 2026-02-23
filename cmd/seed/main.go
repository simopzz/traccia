package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/simopzz/traccia/internal/domain"
	"github.com/simopzz/traccia/internal/infra/config"
	"github.com/simopzz/traccia/internal/infra/database"
	"github.com/simopzz/traccia/internal/repository"
	"github.com/simopzz/traccia/internal/service"
)

const SeedPrefix = "[SEED]"

var (
	airlines = []string{"Delta", "United", "Lufthansa", "Emirates", "British Airways", "Air France", "Ryanair", "EasyJet"}
	airports = []string{"JFK", "LHR", "HND", "CDG", "DXB", "FRA", "SIN", "AMS", "FCO", "MXP", "LIN"}
)

func main() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	if cfg.Environment == "production" {
		return fmt.Errorf("CRITICAL: cannot run seeder in production environment")
	}

	clean := flag.Bool("clean", false, "Clean up existing seed data before running")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	tripStore := repository.NewTripStore(pool)
	flightDetailsStore := repository.NewFlightDetailsStore()
	lodgingDetailsStore := repository.NewLodgingDetailsStore()
	eventStore := repository.NewEventStore(pool, flightDetailsStore, lodgingDetailsStore)

	tripService := service.NewTripService(tripStore)
	eventService := service.NewEventService(eventStore)

	if *clean {
		if err := cleanup(ctx, pool); err != nil {
			return fmt.Errorf("cleanup: %w", err)
		}
	}

	if err := seed(ctx, tripService, eventService); err != nil {
		return fmt.Errorf("seeding: %w", err)
	}

	slog.Info("Seeding completed successfully")
	return nil
}

func cleanup(ctx context.Context, pool *pgxpool.Pool) error {
	slog.Info("Cleaning up existing seed data...")

	tag, err := pool.Exec(ctx, "DELETE FROM trips WHERE name LIKE $1 || '%'", SeedPrefix)
	if err != nil {
		return fmt.Errorf("failed to delete seed trips: %w", err)
	}

	slog.Info("Cleanup complete", "deleted_trips", tag.RowsAffected())
	return nil
}

func seed(ctx context.Context, tripService *service.TripService, eventService *service.EventService) error {
	slog.Info("Starting database seed...")

	totalEvents := 0
	failedEvents := 0

	for i := 0; i < 5; i++ {
		trip, err := seedTrip(ctx, tripService)
		if err != nil {
			return fmt.Errorf("failed to seed trip %d: %w", i, err)
		}

		s, f, err := seedEvents(ctx, eventService, trip)
		if err != nil {
			return fmt.Errorf("failed to seed events for trip %d: %w", trip.ID, err)
		}
		totalEvents += s
		failedEvents += f
	}

	slog.Info("Seeding summary", "total_events_created", totalEvents, "failed_events", failedEvents)

	if totalEvents == 0 && failedEvents > 0 {
		return fmt.Errorf("failed to seed any events")
	}

	return nil
}

func seedTrip(ctx context.Context, tripService *service.TripService) (*domain.Trip, error) {
	destinations := []string{"Paris", "Tokyo", "New York", "Milano", "London", "Berlin", "Sydney"}

	daysFromNow := rand.Intn(365)
	startDate := time.Now().AddDate(0, 0, daysFromNow)

	duration := rand.Intn(12) + 3
	endDate := startDate.AddDate(0, 0, duration)

	name := fmt.Sprintf("%s Trip to %s", SeedPrefix, destinations[rand.Intn(len(destinations))])

	input := &service.CreateTripInput{
		Name:        name,
		Destination: destinations[rand.Intn(len(destinations))],
		StartDate:   startDate,
		EndDate:     endDate,
	}

	trip, err := tripService.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	slog.Info("Created trip", "id", trip.ID, "name", trip.Name, "start", trip.StartDate.Format(time.DateOnly), "end", trip.EndDate.Format(time.DateOnly))
	return trip, nil
}

func seedEvents(ctx context.Context, eventService *service.EventService, trip *domain.Trip) (successCount, failureCount int, err error) {
	currentDate := trip.StartDate
	for !currentDate.After(trip.EndDate) {
		// Add flights on start and end dates
		if isSameDay(currentDate, trip.StartDate) {
			if err := createFlightEvent(ctx, eventService, trip.ID, currentDate, true); err != nil {
				slog.Warn("Failed to create arrival flight", "error", err, "trip_id", trip.ID)
				failureCount++
			} else {
				slog.Debug("Created arrival flight", "date", currentDate.Format(time.DateOnly))
				successCount++
			}
		}

		if isSameDay(currentDate, trip.EndDate) {
			if err := createFlightEvent(ctx, eventService, trip.ID, currentDate, false); err != nil {
				slog.Warn("Failed to create departure flight", "error", err, "trip_id", trip.ID)
				failureCount++
			} else {
				slog.Debug("Created departure flight", "date", currentDate.Format(time.DateOnly))
				successCount++
			}
		}

		// 2-4 events per day
		eventsCount := rand.Intn(3) + 2

		for i := 0; i < eventsCount; i++ {
			startHour := rand.Intn(11) + 8 // 8 to 18
			startMin := rand.Intn(4) * 15  // 0, 15, 30, 45

			startTime := time.Date(
				currentDate.Year(), currentDate.Month(), currentDate.Day(),
				startHour, startMin, 0, 0, currentDate.Location(),
			)

			durationMinutes := (rand.Intn(3) + 1) * 60
			endTime := startTime.Add(time.Duration(durationMinutes) * time.Minute)

			categories := []domain.EventCategory{
				domain.CategoryActivity,
				domain.CategoryFood,
				domain.CategoryTransit,
				// CategoryLodging removed — created explicitly below via createLodgingEvent
			}
			category := categories[rand.Intn(len(categories))]

			title := fmt.Sprintf("Visit %s %d", category, i+1)
			if category == domain.CategoryFood {
				title = fmt.Sprintf("Meal at Local Spot %d", i+1)
			}

			input := &service.CreateEventInput{
				TripID:    trip.ID,
				Title:     title,
				Category:  category,
				StartTime: startTime,
				EndTime:   endTime,
				Location:  "Random Location",
				Notes:     "Generated by seeder",
			}

			_, createErr := eventService.Create(ctx, input)
			if createErr != nil {
				slog.Warn("Failed to create event", "error", createErr, "trip_id", trip.ID)
				failureCount++
			} else {
				slog.Debug("Created event", "title", title, "date", startTime.Format(time.DateOnly))
				successCount++
			}
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	// Add one lodging event per trip spanning the full trip duration
	checkIn := time.Date(
		trip.StartDate.Year(), trip.StartDate.Month(), trip.StartDate.Day(),
		15, 0, 0, 0, trip.StartDate.Location(),
	)
	checkOut := time.Date(
		trip.EndDate.Year(), trip.EndDate.Month(), trip.EndDate.Day(),
		11, 0, 0, 0, trip.EndDate.Location(),
	)
	if err := createLodgingEvent(ctx, eventService, trip.ID, checkIn, checkOut); err != nil {
		slog.Warn("Failed to create lodging event", "error", err, "trip_id", trip.ID)
		failureCount++
	} else {
		slog.Debug("Created lodging event", "trip_id", trip.ID)
		successCount++
	}

	return successCount, failureCount, nil
}

func createLodgingEvent(ctx context.Context, eventService *service.EventService, tripID int, checkIn, checkOut time.Time) error {
	hotelNames := []string{"Grand Hotel", "City Inn", "Palace Suites", "Central Hotel", "Boutique Stay"}
	refLetters := []string{"HTL", "BKG", "RSV", "CNF"}
	ref := fmt.Sprintf("%s%d", refLetters[rand.Intn(len(refLetters))], rand.Intn(99999))

	details := &domain.LodgingDetails{
		CheckInTime:      &checkIn,
		CheckOutTime:     &checkOut,
		BookingReference: ref,
	}

	input := &service.CreateEventInput{
		TripID:         tripID,
		Title:          hotelNames[rand.Intn(len(hotelNames))],
		Category:       domain.CategoryLodging,
		StartTime:      checkIn,
		EndTime:        checkOut,
		Location:       "Hotel Address",
		Notes:          "Generated by seeder",
		LodgingDetails: details,
	}

	_, err := eventService.Create(ctx, input)
	return err
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func createFlightEvent(ctx context.Context, eventService *service.EventService, tripID int, date time.Time, isArrival bool) error {
	airline := airlines[rand.Intn(len(airlines))]
	flightNum := fmt.Sprintf("%s%d", airline[:2], rand.Intn(9000)+100)

	dep := airports[rand.Intn(len(airports))]
	arr := airports[rand.Intn(len(airports))]
	for dep == arr {
		arr = airports[rand.Intn(len(airports))]
	}

	details := &domain.FlightDetails{
		Airline:           airline,
		FlightNumber:      flightNum,
		DepartureAirport:  dep,
		ArrivalAirport:    arr,
		DepartureTerminal: fmt.Sprintf("%d", rand.Intn(5)+1),
		ArrivalTerminal:   fmt.Sprintf("%d", rand.Intn(5)+1),
		DepartureGate:     fmt.Sprintf("G%d", rand.Intn(20)+1),
		ArrivalGate:       fmt.Sprintf("G%d", rand.Intn(20)+1),
		BookingReference:  fmt.Sprintf("BK%s%d", airline[:2], rand.Intn(9999)),
	}

	var title string
	var startTime, endTime time.Time

	if isArrival {
		title = fmt.Sprintf("Flight to %s", arr)
		// Arrive morning
		startTime = time.Date(date.Year(), date.Month(), date.Day(), 10, 0, 0, 0, date.Location())
	} else {
		title = fmt.Sprintf("Return Flight to %s", arr)
		// Depart evening
		startTime = time.Date(date.Year(), date.Month(), date.Day(), 16, 0, 0, 0, date.Location())
	}
	endTime = startTime.Add(time.Duration(rand.Intn(120)+60) * time.Minute)

	input := &service.CreateEventInput{
		TripID:        tripID,
		Title:         title,
		Category:      domain.CategoryFlight,
		StartTime:     startTime,
		EndTime:       endTime,
		Location:      fmt.Sprintf("%s Airport", dep),
		Notes:         "Generated by seeder",
		FlightDetails: details,
	}

	_, err := eventService.Create(ctx, input)
	return err
}
