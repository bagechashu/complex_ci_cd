package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/op/release-control/internal/services"
	"github.com/op/release-control/pkg/logger"
	"github.com/op/release-control/pkg/middleware"
)

// SetupRoutes sets up all API routes
func SetupRoutes(
	router *chi.Mux,
	releaseService *services.ReleaseService,
	log *logger.Logger,
) {
	// Add CORS middleware
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logging(log))

	// Health check
	router.Get("/health", healthCheck)

	// API v1 routes
	router.Route("/api/v1", func(r chi.Router) {
		// Release endpoints
		r.Post("/releases", releaseHandler(releaseService, log))
		r.Get("/releases/{id}", getReleaseHandler(releaseService, log))
		r.Get("/releases/{id}/events", getReleaseEventsHandler(releaseService, log))
		r.Get("/releases", listReleasesHandler(releaseService, log))
		r.Post("/releases/{id}/rollback", rollbackReleaseHandler(releaseService, log))

		// TODO: Add other endpoints
		// r.Route("/applications", applicationRoutes(appService))
		// r.Route("/environments", environmentRoutes(envService))
		// r.Route("/clusters", clusterRoutes(clusterService))
		// r.Route("/deployment-targets", targetRoutes(targetService))
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	middleware.JSONResponse(w, 200, map[string]string{"status": "ok"})
}
