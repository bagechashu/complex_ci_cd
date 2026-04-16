package handlers

import (
	"net/http"

	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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
		r.Get("/clusters/{id}", GetClusterHandler(clusterRepo))
		r.Put("/clusters/{id}", UpdateClusterHandler(clusterRepo))
		r.Delete("/clusters/{id}", DeleteClusterHandler(clusterRepo))

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
		r.Get("/releases/{id}/events", ReleaseEventsHandler())

		// Environments (placeholder - returns empty list for now)
		r.Get("/environments", environmentsHandler)
		
		// Deployment Targets (placeholder - returns empty list for now)
		r.Get("/deployment-targets", deploymentTargetsHandler)

		// Shell Servers
		r.Get("/shell-servers", GetShellServersHandler())
		r.Post("/shell-servers", CreateShellServerHandler())
		r.Get("/shell-servers/{id}", GetShellServerDetailHandler())
		r.Put("/shell-servers/{id}", UpdateShellServerHandler())
		r.Delete("/shell-servers/{id}", DeleteShellServerHandler())

		// Shell Tasks
		r.Get("/shell-tasks", GetShellTasksHandler())
		r.Post("/shell-tasks", CreateShellTaskHandler())
		r.Get("/shell-tasks/{id}", GetShellTaskDetailHandler())
		r.Put("/shell-tasks/{id}", UpdateShellTaskHandler())
		r.Delete("/shell-tasks/{id}", DeleteShellTaskHandler())
		r.Post("/shell-tasks/{id}/execute", ExecuteShellTaskHandler())

		// Shell Execution History
		r.Get("/shell-execution-history", GetShellExecutionHistoryHandler())

		// Command Approvals
		r.Get("/command-approvals", GetCommandApprovalsHandler())
		r.Post("/command-approvals", CreateCommandApprovalHandler())
		r.Put("/command-approvals/{id}/approve", ApproveCommandHandler())
		r.Put("/command-approvals/{id}/reject", RejectCommandHandler())
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
