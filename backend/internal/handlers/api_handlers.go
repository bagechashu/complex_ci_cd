package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/op/release-control/internal/models"
	"github.com/op/release-control/internal/repository"
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

func ListDeploymentTargetsByAppHandler(repo *repository.DeploymentTargetRepository) http.HandlerFunc {
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

		// For now, return all deployment targets
		// TODO: Implement GetByApp method in repository
		targets, err := repo.List(100, 0)
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

		created, err := repo.Create(&target)
		if err != nil {
			// Check if it's a unique constraint error
			if err != nil {
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
		if updates.ContainerName != "" {
			existing.ContainerName = updates.ContainerName
		}
		if updates.RegistryDomain != "" {
			existing.RegistryDomain = updates.RegistryDomain
		}
		if updates.ImageRepo != "" {
			existing.ImageRepo = updates.ImageRepo
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
