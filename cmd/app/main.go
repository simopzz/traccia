package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/simopzz/traccia/internal/handler"
	"github.com/simopzz/traccia/internal/infra/config"
	"github.com/simopzz/traccia/internal/infra/database"
	"github.com/simopzz/traccia/internal/infra/server"
	"github.com/simopzz/traccia/internal/repository"
	"github.com/simopzz/traccia/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	// Setup logging
	var logHandler slog.Handler
	if cfg.IsDevelopment() {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	// Database connection
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	logger.Info("connected to database")

	// Repositories
	tripStore := repository.NewTripStore(pool)
	flightDetailsStore := repository.NewFlightDetailsStore()
	eventStore := repository.NewEventStore(pool, flightDetailsStore)

	// Services
	tripService := service.NewTripService(tripStore)
	eventService := service.NewEventService(eventStore)

	// Handlers
	tripHandler := handler.NewTripHandler(tripService, eventService)
	eventHandler := handler.NewEventHandler(eventService)

	// Router
	router := handler.NewRouter(tripHandler, eventHandler)

	// Server
	srv := server.New(cfg.ServerAddress, router, logger)

	// Graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
		logger.Info("received shutdown signal")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
