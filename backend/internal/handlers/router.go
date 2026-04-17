package handlers

import (
	"net/http"

	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

// SetupRoutes sets up all API routes
func SetupRoutes(
	router *chi.Mux,
	appRepo repository.ApplicationRepository,
	clusterRepo repository.ClusterRepository,
	releaseRepo repository.ReleaseRecordRepository,
	workloadTargetRepo *repository.WorkloadTargetRepository,
	shellServerRepo repository.ShellServerRepository,
	shellCommandRepo repository.ShellCommandRepository,
	shellTaskRepo repository.ShellTaskRepository,
	shellExecutionRepo repository.ShellTaskExecutionRepository,
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

		// workload Targets (App-Cluster Configs)
		r.Get("/app-cluster-configs", ListWorkloadTargetsHandler(workloadTargetRepo))
		r.Get("/app-cluster-configs/by-app/{app_id}", ListWorkloadTargetsByAppHandler(workloadTargetRepo, clusterRepo))
		r.Post("/app-cluster-configs", CreateWorkloadTargetHandler(workloadTargetRepo))
		r.Get("/app-cluster-configs/{id}", GetWorkloadTargetHandler(workloadTargetRepo))
		r.Put("/app-cluster-configs/{id}", UpdateWorkloadTargetHandler(workloadTargetRepo))
		r.Delete("/app-cluster-configs/{id}", DeleteWorkloadTargetHandler(workloadTargetRepo))

		// Releases
		r.Get("/releases", ListReleasesHandler(releaseRepo))
		r.Post("/releases", CreateReleaseHandler(releaseRepo))
		r.Get("/releases/{id}/events", ReleaseEventsHandler())

		// Environments (placeholder - returns empty list for now)
		r.Get("/environments", environmentsHandler)
		
		// workload Targets
		r.Get("/workload-targets", ListWorkloadTargetsHandler(workloadTargetRepo))

		// Shell Servers
		r.Get("/shell-servers", ListShellServersHandler(shellServerRepo))
		r.Post("/shell-servers", CreateShellServerHandler(shellServerRepo))
		r.Get("/shell-servers/{id}", GetShellServerHandler(shellServerRepo))
		r.Put("/shell-servers/{id}", UpdateShellServerHandler(shellServerRepo))
		r.Delete("/shell-servers/{id}", DeleteShellServerHandler(shellServerRepo))

		// Shell Commands
		r.Get("/shell-commands", ListShellCommandsHandler(shellCommandRepo))
		r.Post("/shell-commands", CreateShellCommandHandler(shellCommandRepo))
		r.Post("/shell-commands/{id}/publish", PublishShellCommandHandler(shellCommandRepo))
		r.Post("/shell-commands/{id}/unpublish", UnpublishShellCommandHandler(shellCommandRepo))
		r.Delete("/shell-commands/{id}", DeleteShellCommandHandler(shellCommandRepo))

		// Shell Tasks
		r.Get("/shell-tasks", ListShellTasksHandler(shellTaskRepo))
		r.Post("/shell-tasks", CreateShellTaskHandler(shellTaskRepo))
		r.Get("/shell-tasks/{id}", GetShellTaskHandler(shellTaskRepo))
		r.Put("/shell-tasks/{id}", UpdateShellTaskHandler(shellTaskRepo))
		r.Delete("/shell-tasks/{id}", DeleteShellTaskHandler(shellTaskRepo))

		// Shell Execution History
		r.Get("/shell-executions", ListShellExecutionsHandler(shellExecutionRepo))
		r.Get("/shell-executions/{id}", GetShellExecutionHandler(shellExecutionRepo))
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
