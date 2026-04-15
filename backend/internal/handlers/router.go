package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/op/release-control/internal/repository"
	"github.com/op/release-control/pkg/logger"
	"github.com/op/release-control/pkg/middleware"
)

// SetupRoutes sets up all API routes - SIMPLIFIED VERSION
func SetupRoutes(
	router *chi.Mux,
	appRepo repository.ApplicationRepository,
	clusterRepo repository.ClusterRepository,
	releaseRepo repository.ReleaseRecordRepository,
	deploymentTargetRepo *repository.DeploymentTargetRepository,
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
		// Applications
		r.Get("/applications", ListApplicationsHandler(appRepo))
		r.Post("/applications", CreateApplicationHandler(appRepo))

		// Clusters
		r.Get("/clusters", ListClustersHandler(clusterRepo))
		r.Post("/clusters", CreateClusterHandler(clusterRepo))

		// Deployment Targets (App-Cluster Configs)
		r.Get("/app-cluster-configs", ListDeploymentTargetsHandler(deploymentTargetRepo))
		r.Get("/app-cluster-configs/by-app/{app_id}", ListDeploymentTargetsByAppHandler(deploymentTargetRepo))
		r.Post("/app-cluster-configs", CreateDeploymentTargetHandler(deploymentTargetRepo))
		r.Get("/app-cluster-configs/{id}", GetDeploymentTargetHandler(deploymentTargetRepo))
		r.Put("/app-cluster-configs/{id}", UpdateDeploymentTargetHandler(deploymentTargetRepo))
		r.Delete("/app-cluster-configs/{id}", DeleteDeploymentTargetHandler(deploymentTargetRepo))

		// Releases
		r.Get("/releases", ListReleasesHandler(releaseRepo))
		r.Post("/releases", CreateReleaseHandler(releaseRepo))

		// Environments (placeholder - returns empty list for now)
		r.Get("/environments", environmentsHandler)
		
		// Deployment Targets (placeholder - returns empty list for now)
		r.Get("/deployment-targets", deploymentTargetsHandler)
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	middleware.JSONResponse(w, 200, map[string]string{"status": "ok"})
}

func environmentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"total":0,"data":[]}`))
}

func deploymentTargetsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"total":0,"data":[]}`))
}
