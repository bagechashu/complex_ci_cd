package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/op/release-control/internal/repository"
	"github.com/op/release-control/pkg/logger"
	"github.com/op/release-control/pkg/middleware"
)

// listApplicationsHandler handles GET /api/v1/applications
func listApplicationsHandler(appRepo *repository.ApplicationRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		offset := 0

		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}

		apps, err := appRepo.List(limit, offset)
		if err != nil {
			log.Error("Failed to list applications: %v", err)
			middleware.ErrorResponse(w, 500, "INTERNAL_ERROR", "Failed to list applications")
			return
		}

		middleware.JSONResponse(w, 200, map[string]interface{}{
			"data": apps,
			"pagination": map[string]int{
				"limit":  limit,
				"offset": offset,
			},
		})
	}
}

// getApplicationHandler handles GET /api/v1/applications/{id}
func getApplicationHandler(appRepo *repository.ApplicationRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			middleware.ErrorResponse(w, 400, "BAD_REQUEST", "Invalid application ID")
			return
		}

		app, err := appRepo.GetByID(id)
		if err != nil {
			log.Error("Failed to get application: %v", err)
			middleware.ErrorResponse(w, 404, "NOT_FOUND", "Application not found")
			return
		}

		middleware.JSONResponse(w, 200, app)
	}
}

// listEnvironmentsHandler handles GET /api/v1/environments
func listEnvironmentsHandler(envRepo *repository.EnvironmentRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		offset := 0

		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}

		envs, err := envRepo.List(limit, offset)
		if err != nil {
			log.Error("Failed to list environments: %v", err)
			middleware.ErrorResponse(w, 500, "INTERNAL_ERROR", "Failed to list environments")
			return
		}

		middleware.JSONResponse(w, 200, map[string]interface{}{
			"data": envs,
			"pagination": map[string]int{
				"limit":  limit,
				"offset": offset,
			},
		})
	}
}

// getEnvironmentHandler handles GET /api/v1/environments/{id}
func getEnvironmentHandler(envRepo *repository.EnvironmentRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			middleware.ErrorResponse(w, 400, "BAD_REQUEST", "Invalid environment ID")
			return
		}

		env, err := envRepo.GetByID(id)
		if err != nil {
			log.Error("Failed to get environment: %v", err)
			middleware.ErrorResponse(w, 404, "NOT_FOUND", "Environment not found")
			return
		}

		middleware.JSONResponse(w, 200, env)
	}
}

// listClustersHandler handles GET /api/v1/clusters
func listClustersHandler(clusterRepo *repository.ClusterRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		offset := 0

		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}

		clusters, err := clusterRepo.List(limit, offset)
		if err != nil {
			log.Error("Failed to list clusters: %v", err)
			middleware.ErrorResponse(w, 500, "INTERNAL_ERROR", "Failed to list clusters")
			return
		}

		middleware.JSONResponse(w, 200, map[string]interface{}{
			"data": clusters,
			"pagination": map[string]int{
				"limit":  limit,
				"offset": offset,
			},
		})
	}
}

// getClusterHandler handles GET /api/v1/clusters/{id}
func getClusterHandler(clusterRepo *repository.ClusterRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			middleware.ErrorResponse(w, 400, "BAD_REQUEST", "Invalid cluster ID")
			return
		}

		cluster, err := clusterRepo.GetByID(id)
		if err != nil {
			log.Error("Failed to get cluster: %v", err)
			middleware.ErrorResponse(w, 404, "NOT_FOUND", "Cluster not found")
			return
		}

		middleware.JSONResponse(w, 200, cluster)
	}
}

// listDeploymentTargetsHandler handles GET /api/v1/deployment-targets
func listDeploymentTargetsHandler(targetRepo *repository.DeploymentTargetRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 50
		offset := 0

		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.Atoi(l); err == nil {
				limit = parsed
			}
		}

		if o := r.URL.Query().Get("offset"); o != "" {
			if parsed, err := strconv.Atoi(o); err == nil {
				offset = parsed
			}
		}

		targets, err := targetRepo.List(limit, offset)
		if err != nil {
			log.Error("Failed to list deployment targets: %v", err)
			middleware.ErrorResponse(w, 500, "INTERNAL_ERROR", "Failed to list deployment targets")
			return
		}

		middleware.JSONResponse(w, 200, map[string]interface{}{
			"data": targets,
			"pagination": map[string]int{
				"limit":  limit,
				"offset": offset,
			},
		})
	}
}

// getDeploymentTargetsByAppAndEnvHandler handles GET /api/v1/deployment-targets/app/{appId}/env/{envId}
func getDeploymentTargetsByAppAndEnvHandler(targetRepo *repository.DeploymentTargetRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appIDStr := chi.URLParam(r, "appId")
		envIDStr := chi.URLParam(r, "envId")

		appID, err := strconv.Atoi(appIDStr)
		if err != nil {
			middleware.ErrorResponse(w, 400, "BAD_REQUEST", "Invalid application ID")
			return
		}

		envID, err := strconv.Atoi(envIDStr)
		if err != nil {
			middleware.ErrorResponse(w, 400, "BAD_REQUEST", "Invalid environment ID")
			return
		}

		targets, err := targetRepo.ListByAppAndEnv(appID, envID)
		if err != nil {
			log.Error("Failed to list deployment targets: %v", err)
			middleware.ErrorResponse(w, 500, "INTERNAL_ERROR", "Failed to list deployment targets")
			return
		}

		middleware.JSONResponse(w, 200, map[string]interface{}{
			"data": targets,
		})
	}
}

// CreateMetadataRoutes sets up metadata routes
func CreateMetadataRoutes(
	r chi.Router,
	appRepo *repository.ApplicationRepository,
	envRepo *repository.EnvironmentRepository,
	clusterRepo *repository.ClusterRepository,
	targetRepo *repository.DeploymentTargetRepository,
	log *logger.Logger,
) {
	// Applications
	r.Get("/applications", listApplicationsHandler(appRepo, log))
	r.Get("/applications/{id}", getApplicationHandler(appRepo, log))

	// Environments
	r.Get("/environments", listEnvironmentsHandler(envRepo, log))
	r.Get("/environments/{id}", getEnvironmentHandler(envRepo, log))

	// Clusters
	r.Get("/clusters", listClustersHandler(clusterRepo, log))
	r.Get("/clusters/{id}", getClusterHandler(clusterRepo, log))

	// Deployment Targets
	r.Get("/deployment-targets", listDeploymentTargetsHandler(targetRepo, log))
	r.Get("/deployment-targets/app/{appId}/env/{envId}", getDeploymentTargetsByAppAndEnvHandler(targetRepo, log))
}
