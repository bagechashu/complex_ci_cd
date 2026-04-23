package workloads

import (
	"encoding/json"
	"net/http"
	"strconv"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/repository"
	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/responses"

	"github.com/go-chi/chi/v5"
)

// ListWorkloadTargetsHandler handles GET /workload-targets request.
func ListWorkloadTargetsHandler(workloadService *services.WorkloadService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targets, err := workloadService.ListWorkloadTargets(r.Context())
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}
		responses.SuccessResponse(w, targets)
	}
}

// GetAppClusterConfigsHandler handles GET /app-cluster-configs request.
func GetAppClusterConfigsHandler(workloadService *services.WorkloadService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targets, err := workloadService.ListWorkloadTargets(r.Context())
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}
		responses.SuccessResponse(w, targets)
	}
}

// GetAppClusterConfigsByAppHandler handles GET /app-cluster-configs/by-app/:id request.
func GetAppClusterConfigsByAppHandler(workloadService *services.WorkloadService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("GetAppClusterConfigsByAppHandler called", "path", r.URL.Path)
		appIDStr := chi.URLParam(r, "id")
		log.Info("Extracted appID", "appIDStr", appIDStr)
		appID, err := strconv.Atoi(appIDStr)
		if err != nil {
			log.Error("Invalid app id", "error", err, "appIDStr", appIDStr)
			responses.BadRequestResponse(w, "invalid app id")
			return
		}

		log.Info("Fetching workload targets for app", "appID", appID)
		targets, err := workloadService.ListWorkloadTargetsByApp(r.Context(), appID)
		if err != nil {
			log.Error("Failed to list workload targets", "error", err, "appID", appID)
			responses.InternalErrorResponse(w, err.Error())
			return
		}
		log.Info("Successfully fetched workload targets", "appID", appID, "count", len(targets))
		responses.SuccessResponse(w, targets)
	}
}

// CreateWorkloadTargetHandler handles POST /workload-targets request.
func CreateWorkloadTargetHandler(workloadService *services.WorkloadService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "invalid request body")
			return
		}

		// For now, just return success (full implementation would validate and save)
		responses.CreatedResponse(w, map[string]string{"message": "created"})
	}
}

// CreateAppClusterConfigHandler handles POST /app-cluster-configs request.
func CreateAppClusterConfigHandler(repo repository.WorkloadTargetRepository, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("CreateAppClusterConfigHandler called", "path", r.URL.Path)
		
		var req struct {
			AppID         int    `json:"app_id"`
			EnvID         int    `json:"env_id"`
			ClusterID     int    `json:"cluster_id"`
			K8sNamespace  string `json:"k8s_namespace"`
			WorkloadType  string `json:"workload_type"`
			WorkloadName  string `json:"workload_name"`
			ContainerName *string `json:"container_name"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("Failed to decode request body", "error", err)
			responses.BadRequestResponse(w, "invalid request body")
			return
		}

		// Validate required fields
		if req.AppID == 0 {
			responses.BadRequestResponse(w, "app_id is required")
			return
		}
		if req.EnvID == 0 {
			responses.BadRequestResponse(w, "env_id is required")
			return
		}
		if req.ClusterID == 0 {
			responses.BadRequestResponse(w, "cluster_id is required")
			return
		}
		if req.K8sNamespace == "" {
			responses.BadRequestResponse(w, "k8s_namespace is required")
			return
		}
		if req.WorkloadType == "" {
			responses.BadRequestResponse(w, "workload_type is required")
			return
		}
		if req.WorkloadName == "" {
			responses.BadRequestResponse(w, "workload_name is required")
			return
		}

		// Create WorkloadTarget
		target := &models.WorkloadTarget{
			AppID:        req.AppID,
			EnvID:        req.EnvID,
			ClusterID:    req.ClusterID,
			K8sNamespace: req.K8sNamespace,
			K8sWorkload:  req.WorkloadName + "-" + req.WorkloadType,  // Generate from workload_name and type
			WorkloadType: req.WorkloadType,
			WorkloadName: req.WorkloadName,
			ContainerName: req.ContainerName,
		}

		result, err := repo.Create(target)
		if err != nil {
			log.Error("Failed to create workload target", "error", err, "appID", req.AppID, "clusterID", req.ClusterID)
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		log.Info("Successfully created workload target", "id", result.ID, "appID", req.AppID, "clusterID", req.ClusterID)
		responses.CreatedResponse(w, result)
	}
}
