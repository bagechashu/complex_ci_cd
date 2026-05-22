package workloads

import (
	"encoding/json"
	"net/http"
	"strconv"

	"built-and-deploy/internal/repository"
	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/logger"

	"github.com/go-chi/chi/v5"
)

type WorkloadHandler struct {
	workloadRepo     repository.WorkloadTargetRepository
	clusterRepo      repository.ClusterRepository
	inspectorService *services.WorkloadInspectorService
	log              *logger.Logger
}

// NewWorkloadHandler creates a new WorkloadHandler
func NewWorkloadHandler(
	workloadRepo repository.WorkloadTargetRepository,
	clusterRepo repository.ClusterRepository,
	encryptionKey string,
	log *logger.Logger,
) *WorkloadHandler {
	return &WorkloadHandler{
		workloadRepo:     workloadRepo,
		clusterRepo:      clusterRepo,
		inspectorService: services.NewWorkloadInspectorService(clusterRepo, encryptionKey, log),
		log:              log,
	}
}

// GetWorkloadPodsHandler is a handler factory function that returns an http.HandlerFunc
// This follows the same pattern as other handlers in the codebase
func GetWorkloadPodsHandler(
	workloadRepo repository.WorkloadTargetRepository,
	clusterRepo repository.ClusterRepository,
	encryptionKey string,
	log *logger.Logger,
) http.HandlerFunc {
	handler := NewWorkloadHandler(workloadRepo, clusterRepo, encryptionKey, log)
	return handler.GetWorkloadPods
}

// GetWorkloadPods handles GET /v1/workloads/{id}/pods
// Returns the list of pods for a specific workload
func (h *WorkloadHandler) GetWorkloadPods(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeResponse(w, http.StatusOK, 3001, "Invalid workload ID", nil)
		return
	}

	// Get workload target from repository
	workloadTarget, err := h.workloadRepo.GetByID(id)
	if err != nil || workloadTarget == nil {
		h.log.Warn("Workload not found", "id", id)
		h.writeResponse(w, http.StatusOK, 1001, "Workload not found", nil)
		return
	}

	// Get cluster (with encrypted kubeconfig)
	cluster, err := h.clusterRepo.GetByID(r.Context(), workloadTarget.ClusterID)
	if err != nil || cluster == nil {
		h.log.Warn("Cluster not found", "clusterId", workloadTarget.ClusterID)
		h.writeResponse(w, http.StatusOK, 1002, "Cluster not found", nil)
		return
	}

	// Get pods
	pods, err := h.inspectorService.GetWorkloadPods(r.Context(), cluster, workloadTarget)
	if err != nil {
		h.log.Error("Failed to get workload pods", "error", err)
		h.writeResponse(w, http.StatusOK, 9999, err.Error(), nil)
		return
	}

	h.writeResponse(w, http.StatusOK, 0, "success", pods)
}

// ==================== Helper Methods ====================

func (h *WorkloadHandler) writeResponse(w http.ResponseWriter, statusCode int, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    code,
		"message": message,
		"data":    data,
	})
}
