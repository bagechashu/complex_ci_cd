package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/repository"
)

// Application Handlers
func ListApplicationsHandler(repo repository.ApplicationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		apps, total, err := repo.List(ctx, 0, 100)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": total,
			"data":  apps,
		})
	}
}

func CreateApplicationHandler(repo repository.ApplicationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.Application
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		if req.ImageName == "" {
			http.Error(w, "image_name is required", http.StatusBadRequest)
			return
		}

		app := &models.Application{
			Name:        req.Name,
			ImageName:   req.ImageName,
			GitRepo:     req.GitRepo,
			BuildType:   req.BuildType,
			Description: req.Description,
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = repo.Create(ctx, app)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create application: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(app)
	}
}

// Cluster Handlers
func ListClustersHandler(repo repository.ClusterRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		clusters, total, err := repo.List(ctx, 0, 100)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": total,
			"data":  clusters,
		})
	}
}

func CreateClusterHandler(repo repository.ClusterRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.Cluster
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		if req.Type == "" {
			http.Error(w, "type is required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = repo.Create(ctx, &req)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create cluster: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(req)
	}
}

// App-Cluster Config Handlers
func ListConfigsHandler(repo repository.ApplicationClusterConfigRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		configs, total, err := repo.List(ctx, 0, 100)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": total,
			"data":  configs,
		})
	}
}

// Release Handlers
func ListReleasesHandler(repo repository.ReleaseRecordRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		releases, total, err := repo.List(ctx, 0, 100)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": total,
			"data":  releases,
		})
	}
}

func CreateReleaseHandler(repo repository.ReleaseRecordRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var release models.ReleaseRecord
		err := json.NewDecoder(r.Body).Decode(&release)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if release.AppID == 0 {
			http.Error(w, "app_id is required", http.StatusBadRequest)
			return
		}
		if release.EnvID == 0 {
			http.Error(w, "env_id is required", http.StatusBadRequest)
			return
		}
		if release.ClusterID == 0 {
			http.Error(w, "cluster_id is required", http.StatusBadRequest)
			return
		}
		if release.Image == "" {
			http.Error(w, "image is required", http.StatusBadRequest)
			return
		}

		// Set default values
		if release.Status == "" {
			release.Status = "pending"
		}
		now := time.Now()
		release.StartedAt = &now

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = repo.Create(ctx, &release)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create release: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(release)
	}
}

// Deployment Target Handlers (App-Cluster Configs)
func ListDeploymentTargetsHandler(repo *repository.DeploymentTargetRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 100
		offset := 0

		targets, err := repo.List(limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": len(targets),
			"data":  targets,
		})
	}
}

func ListDeploymentTargetsByAppHandler(repo *repository.DeploymentTargetRepository, clusterRepo repository.ClusterRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appID := r.PathValue("app_id")
		if appID == "" {
			http.Error(w, "app_id is required", http.StatusBadRequest)
			return
		}

		// Parse app_id as integer (chi provides string by default)
		var appIDInt int
		_, err := fmt.Sscanf(appID, "%d", &appIDInt)
		if err != nil {
			http.Error(w, "invalid app_id", http.StatusBadRequest)
			return
		}

		// Get deployment targets by app using the GetByApp method
		targets, err := repo.GetByApp(appIDInt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Enrich targets with cluster information
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		for _, target := range targets {
			cluster, err := clusterRepo.GetByID(ctx, target.ClusterID)
			if err == nil && cluster != nil {
				target.ClusterName = cluster.Name
				target.Environment = cluster.Environment
				target.RegistryPrefix = cluster.RegistryPrefix
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": len(targets),
			"data":  targets,
		})
	}
}

func GetDeploymentTargetHandler(repo *repository.DeploymentTargetRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		var id int
		_, err := fmt.Sscanf(idStr, "%d", &id)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		target, err := repo.GetByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(target)
	}
}

func CreateDeploymentTargetHandler(repo *repository.DeploymentTargetRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var target models.DeploymentTarget
		err := json.NewDecoder(r.Body).Decode(&target)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if target.AppID == 0 {
			http.Error(w, "app_id is required", http.StatusBadRequest)
			return
		}
		if target.EnvID == 0 {
			http.Error(w, "env_id is required", http.StatusBadRequest)
			return
		}
		if target.ClusterID == 0 {
			http.Error(w, "cluster_id is required", http.StatusBadRequest)
			return
		}
		if target.K8sNamespace == "" {
			http.Error(w, "k8s_namespace is required", http.StatusBadRequest)
			return
		}
		if target.K8sDeployment == "" {
			http.Error(w, "k8s_deployment is required", http.StatusBadRequest)
			return
		}
		if target.WorkloadType == "" {
			http.Error(w, "workload_type is required", http.StatusBadRequest)
			return
		}
		if target.WorkloadName == "" {
			http.Error(w, "workload_name is required", http.StatusBadRequest)
			return
		}

		created, err := repo.Create(&target)
		if err != nil {
			// Return 409 Conflict for duplicate unique constraint errors
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				http.Error(w, fmt.Sprintf("duplicate deployment target: %v", err), http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}
}

func UpdateDeploymentTargetHandler(repo *repository.DeploymentTargetRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		var id int
		_, err := fmt.Sscanf(idStr, "%d", &id)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		// Get existing target
		existing, err := repo.GetByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Decode request body
		var updates models.DeploymentTarget
		err = json.NewDecoder(r.Body).Decode(&updates)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Update fields
		if updates.K8sNamespace != "" {
			existing.K8sNamespace = updates.K8sNamespace
		}
		if updates.K8sDeployment != "" {
			existing.K8sDeployment = updates.K8sDeployment
		}
		if updates.ContainerName != nil && *updates.ContainerName != "" {
			existing.ContainerName = updates.ContainerName
		}
		if updates.RegistryDomain != nil && *updates.RegistryDomain != "" {
			existing.RegistryDomain = updates.RegistryDomain
		}
		if updates.ImageRepo != nil && *updates.ImageRepo != "" {
			existing.ImageRepo = updates.ImageRepo
		}
		if updates.WorkloadType != "" {
			existing.WorkloadType = updates.WorkloadType
		}
		if updates.WorkloadName != "" {
			existing.WorkloadName = updates.WorkloadName
		}

		err = repo.Update(existing)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(existing)
	}
}

func DeleteDeploymentTargetHandler(repo *repository.DeploymentTargetRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		var id int
		_, err := fmt.Sscanf(idStr, "%d", &id)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		err = repo.Delete(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
}

// Cluster Handlers - Additional
func GetClusterHandler(repo repository.ClusterRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		var id int
		_, err := fmt.Sscanf(idStr, "%d", &id)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		cluster, err := repo.GetByID(ctx, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cluster)
	}
}

func UpdateClusterHandler(repo repository.ClusterRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		var id int
		_, err := fmt.Sscanf(idStr, "%d", &id)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Get existing cluster
		existing, err := repo.GetByID(ctx, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Decode request body
		var updates models.Cluster
		err = json.NewDecoder(r.Body).Decode(&updates)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Update fields
		if updates.Name != "" {
			existing.Name = updates.Name
		}
		if updates.Type != "" {
			existing.Type = updates.Type
		}
		if updates.Environment != "" {
			existing.Environment = updates.Environment
		}
		if updates.RegistryPrefix != "" {
			existing.RegistryPrefix = updates.RegistryPrefix
		}
		if updates.Labels != nil && *updates.Labels != "" {
			existing.Labels = updates.Labels
		}
		if updates.KubeconfigPath != nil && *updates.KubeconfigPath != "" {
			existing.KubeconfigPath = updates.KubeconfigPath
		}
		if updates.Kubeconfig != nil && *updates.Kubeconfig != "" {
			existing.Kubeconfig = updates.Kubeconfig
		}
		if updates.KubeconfigEncrypted != nil && *updates.KubeconfigEncrypted != "" {
			existing.KubeconfigEncrypted = updates.KubeconfigEncrypted
		}
		if updates.AnsibleHosts != nil && *updates.AnsibleHosts != "" {
			existing.AnsibleHosts = updates.AnsibleHosts
		}

		err = repo.Update(ctx, existing)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(existing)
	}
}

func DeleteClusterHandler(repo repository.ClusterRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		var id int
		_, err := fmt.Sscanf(idStr, "%d", &id)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = repo.Delete(ctx, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
}

// Shell Server Handlers - Placeholder implementations
func GetShellServersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": 0,
			"data":  []interface{}{},
		})
	}
}

func CreateShellServerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var server map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&server)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(server)
	}
}

func UpdateShellServerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var server map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&server)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(server)
	}
}

func DeleteShellServerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
}

// Shell Task Handlers - Placeholder implementations
func GetShellTasksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": 0,
			"data":  []interface{}{},
		})
	}
}

func CreateShellTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}
}

func UpdateShellTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}

func DeleteShellTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
}

func ExecuteShellTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"message": "Task execution initiated",
		})
	}
}

// Command Approval Handlers - Placeholder implementations
func GetCommandApprovalsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": 0,
			"data":  []interface{}{},
		})
	}
}

func CreateCommandApprovalHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var approval map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&approval)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(approval)
	}
}

func ApproveCommandHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "approved",
		})
	}
}

func RejectCommandHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "rejected",
		})
	}
}

// Execution History Handler - Placeholder implementation
func GetShellExecutionHistoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": 0,
			"data":  []interface{}{},
		})
	}
}

// Release Events Handler - Placeholder
func ReleaseEventsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
	}
}

// Shell Server Detail Handler - Placeholder
func GetShellServerDetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}
}

// Shell Task Detail Handler - Placeholder
func GetShellTaskDetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}
}
