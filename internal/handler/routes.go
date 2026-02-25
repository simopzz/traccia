package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(tripHandler *TripHandler, eventHandler *EventHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	r.Group(func(r chi.Router) {
		// TODO: enable r.Use(authMiddleware) when Supabase auth is ready

		// Trip routes
		r.Get("/", tripHandler.List)
		r.Get("/trips/new", tripHandler.NewPage)
		r.Post("/trips", tripHandler.Create)
		r.Get("/trips/{id}", tripHandler.Detail)
		r.Get("/trips/{id}/edit", tripHandler.EditPage)
		r.Put("/trips/{id}", tripHandler.Update)
		r.Delete("/trips/{id}", tripHandler.Delete)

		// Event routes
		r.Get("/trips/{tripID}/events/new", eventHandler.NewPage)
		r.Post("/trips/{tripID}/events", eventHandler.Create)
		r.Get("/trips/{tripID}/events/{id}/edit", eventHandler.EditPage)
		r.Put("/trips/{tripID}/events/{id}", eventHandler.Update)
		r.Delete("/trips/{tripID}/events/{id}", eventHandler.Delete)
		r.Post("/trips/{tripID}/events/{id}/restore", eventHandler.Restore)
	})

	return r
}
