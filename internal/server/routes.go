// SPDX-License-Identifier: AGPL-3.0-only
package server

import (
	"net/http"

	"traccia/internal/features/health"
	"traccia/internal/features/timeline"
	"traccia/web"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	timelineService := timeline.NewService(s.db.DB())
	timelineHandler := timeline.NewHandler(timelineService)
	timelineHandler.RegisterRoutes(r)

	healthHandler := health.NewHandler(s.db)
	r.Get("/health", healthHandler.HealthHandler)

	fileServer := http.FileServer(http.FS(web.Files))
	r.Handle("/assets/*", fileServer)
	r.Get("/web", templ.Handler(web.HelloForm()).ServeHTTP)
	r.Post("/hello", web.HelloWebHandler)

	return r
}
