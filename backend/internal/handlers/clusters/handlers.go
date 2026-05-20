// Package clusters provides HTTP handlers for cluster management.
//
// This package handles all API endpoints related to clusters:
//   - List all registered clusters
//   - Create new clusters with encryption
//   - Retrieve cluster details
//   - Update cluster configuration
//   - Delete clusters
//   - Test cluster connectivity
//
// All handlers coordinate with ClusterService for business logic implementation.
package clusters

import (
	"encoding/json"
	"net/http"
	"strconv"

	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/handlers"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/responses"
)

// ListClustersHandler handles GET /clusters request to retrieve all clusters.
//
// @Summary List All Clusters
// @Description Retrieves a list of all registered clusters in the system
// @Tags Clusters
// @Produce json
// @Success 200 {array} models.Cluster "List of clusters"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /clusters [get]
//
// Returns:
//   - Array of Cluster objects with metadata (name, type, status, etc.)
//   - Empty array if no clusters exist
//   - Note: Kubeconfig data is encrypted in returned objects
//
// Example response (200):
//
//	[
//	  {
//	    "id": 1,
//	    "name": "production-k8s",
//	    "type": "kubernetes",
//	    "environment": "prod",
//	    "connection_status": "connected",
//	    "created_at": "2026-04-21T10:00:00Z"
//	  },
//	  {
//	    "id": 2,
//	    "name": "staging-ssh",
//	    "type": "ssh",
//	    "environment": "staging",
//	    "connection_status": "unknown",
//	    "created_at": "2026-04-21T10:05:00Z"
//	  }
//	]
func ListClustersHandler(clusterService *services.ClusterService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusters, err := clusterService.ListClusters(r.Context())
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}
		responses.SuccessResponse(w, clusters)
	}
}

// CreateClusterHandler handles POST /clusters request to create a new cluster.
//
// @Summary Create Cluster
// @Description Creates a new cluster with encrypted kubeconfig storage
// @Tags Clusters
// @Accept json
// @Produce json
// @Param request body CreateClusterRequest true "Create Cluster Request"
// @Success 201 {object} models.Cluster "Cluster created successfully"
// @Failure 400 {object} responses.ErrorResponse "Invalid request body or missing required fields"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /clusters [post]
//
// Security considerations:
//   - Kubeconfig is encrypted before storage using AES encryption
//   - Sensitive data never stored in plaintext
//   - Encryption key is provided via environment configuration
//
// Request body:
//
//	{
//	  "name": "production-k8s",
//	  "type": "kubernetes",
//	  "environment": "prod",
//	  "registry_prefix": "docker.io",
//	  "labels": "tier=production,critical=true",
//	  "kubeconfig": "apiVersion: v1\nkind: Config\n..."
//	}
//
// Response (201):
//
//	{
//	  "id": 1,
//	  "name": "production-k8s",
//	  "type": "kubernetes",
//	  "environment": "prod",
//	  "created_at": "2026-04-21T10:00:00Z"
//	}
func CreateClusterHandler(clusterService *services.ClusterService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req handlers.CreateClusterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "invalid request body")
			return
		}

		cluster, err := clusterService.CreateCluster(r.Context(), &req)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		responses.CreatedResponse(w, cluster)
	}
}

// GetClusterHandler 获取集群详情
func GetClusterHandler(clusterService *services.ClusterService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "invalid cluster id")
			return
		}

		cluster, err := clusterService.GetCluster(r.Context(), id)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		responses.SuccessResponse(w, cluster)
	}
}

// UpdateClusterHandler 更新集群
func UpdateClusterHandler(clusterService *services.ClusterService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "invalid cluster id")
			return
		}

		var req handlers.UpdateClusterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.BadRequestResponse(w, "invalid request body")
			return
		}

		cluster, err := clusterService.UpdateCluster(r.Context(), id, &req)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		responses.SuccessResponse(w, cluster)
	}
}

// DeleteClusterHandler 删除集群
func DeleteClusterHandler(clusterService *services.ClusterService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "invalid cluster id")
			return
		}

		err = clusterService.DeleteCluster(r.Context(), id)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		responses.SuccessResponse(w, map[string]interface{}{"id": id})
	}
}

// TestClusterConnectionHandler 测试集群连接状态
func TestClusterConnectionHandler(clusterService *services.ClusterService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			responses.BadRequestResponse(w, "invalid cluster id")
			return
		}

		result, err := clusterService.TestConnection(r.Context(), id)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		responses.SuccessResponse(w, map[string]interface{}{
			"id":      id,
			"status":  result.Status,
			"message": result.Message,
		})
	}
}

