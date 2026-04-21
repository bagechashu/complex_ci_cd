//go:build integration

package clusters

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"built-and-deploy/internal/database"
	"built-and-deploy/internal/models"
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupClusterTestDB(t *testing.T) *repository.SQLiteClusterRepository {
	dbPath := "test_cluster_integration.db"
	db, err := database.Init(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		database.Close(db)
		os.Remove(dbPath)
	})
	return repository.NewSQLiteClusterRepository(db)
}

// TestIntegration_CreateCluster_CompleteFlow tests end-to-end cluster creation
func TestIntegration_CreateCluster_CompleteFlow(t *testing.T) {
	repo := setupClusterTestDB(t)
	log := logger.NewLogger()

	// Create request
	reqBody := map[string]string{
		"name":       "integration-k8s-cluster",
		"type":       "kubernetes",
		"labels":     "env=prod,tier=critical",
		"kubeconfig": "apiVersion: v1\nkind: Config",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/clusters", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	// Execute handler
	handler := CreateClusterHandler(repo, log)
	handler.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Parse response and verify persistence
	var createdCluster models.Cluster
	err := json.Unmarshal(w.Body.Bytes(), &createdCluster)
	require.NoError(t, err)
	assert.Equal(t, "integration-k8s-cluster", createdCluster.Name)
	assert.Equal(t, "kubernetes", createdCluster.Type)
	assert.NotZero(t, createdCluster.ID)

	// Verify data persisted in database
	storedCluster, err := repo.GetByID(context.Background(), createdCluster.ID)
	require.NoError(t, err)
	assert.Equal(t, createdCluster.ID, storedCluster.ID)
	assert.Equal(t, "integration-k8s-cluster", storedCluster.Name)
}

// TestIntegration_ListClusters_MultipleCreated tests list endpoint with multiple clusters
func TestIntegration_ListClusters_MultipleCreated(t *testing.T) {
	repo := setupClusterTestDB(t)
	log := logger.NewLogger()

	// Create multiple clusters
	clusterTypes := []string{"kubernetes", "ssh", "ansible"}

	for i, clusterType := range clusterTypes {
		reqBody := map[string]interface{}{
			"name":       "cluster-" + string(rune(i)),
			"type":       clusterType,
			"labels":     "env=test",
			"kubeconfig": "config",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/clusters", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		handler := CreateClusterHandler(repo, log)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// List clusters
	req := httptest.NewRequest("GET", "/clusters", nil)
	w := httptest.NewRecorder()

	handler := ListClustersHandler(repo, log)
	handler.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var result []*models.Cluster
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

// TestIntegration_GetCluster_ByID tests retrieval of single cluster
func TestIntegration_GetCluster_ByID(t *testing.T) {
	repo := setupClusterTestDB(t)
	log := logger.NewLogger()

	// Create a cluster
	reqBody := map[string]string{
		"name":       "get-test-cluster",
		"type":       "kubernetes",
		"labels":     "env=dev",
		"kubeconfig": "config-yaml",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/clusters", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	createHandler := CreateClusterHandler(repo, log)
	createHandler.ServeHTTP(w, req)

	var createdCluster models.Cluster
	json.Unmarshal(w.Body.Bytes(), &createdCluster)

	// Get the cluster by ID
	req = httptest.NewRequest("GET", "/clusters/1", nil)
	req.SetPathValue("id", "1")
	w = httptest.NewRecorder()

	getHandler := GetClusterHandler(repo, log)
	getHandler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var retrievedCluster models.Cluster
	json.Unmarshal(w.Body.Bytes(), &retrievedCluster)
	assert.Equal(t, createdCluster.ID, retrievedCluster.ID)
	assert.Equal(t, "get-test-cluster", retrievedCluster.Name)
}

// TestIntegration_UpdateCluster tests cluster update functionality
func TestIntegration_UpdateCluster_Success(t *testing.T) {
	repo := setupClusterTestDB(t)
	log := logger.NewLogger()

	// Create initial cluster
	reqBody := map[string]string{
		"name":       "update-test-cluster",
		"type":       "kubernetes",
		"labels":     "env=staging",
		"kubeconfig": "initial-config",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/clusters", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	createHandler := CreateClusterHandler(repo, log)
	createHandler.ServeHTTP(w, req)

	var createdCluster models.Cluster
	json.Unmarshal(w.Body.Bytes(), &createdCluster)

	// Update the cluster
	updateBody := map[string]string{
		"name":       "updated-cluster-name",
		"type":       "kubernetes",
		"labels":     "env=prod",
		"kubeconfig": "updated-config",
	}
	bodyBytes, _ = json.Marshal(updateBody)
	req = httptest.NewRequest("PUT", "/clusters/1", bytes.NewReader(bodyBytes))
	req.SetPathValue("id", "1")
	w = httptest.NewRecorder()

	updateHandler := UpdateClusterHandler(repo, log)
	updateHandler.ServeHTTP(w, req)

	// Assert update succeeded
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify updated data
	var updatedCluster models.Cluster
	json.Unmarshal(w.Body.Bytes(), &updatedCluster)
	assert.Equal(t, "updated-cluster-name", updatedCluster.Name)
}

// TestIntegration_DeleteCluster tests cluster deletion
func TestIntegration_DeleteCluster_Success(t *testing.T) {
	repo := setupClusterTestDB(t)
	log := logger.NewLogger()

	// Create a cluster
	reqBody := map[string]string{
		"name":       "delete-test-cluster",
		"type":       "kubernetes",
		"labels":     "env=test",
		"kubeconfig": "config",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/clusters", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	createHandler := CreateClusterHandler(repo, log)
	createHandler.ServeHTTP(w, req)

	// Delete the cluster
	req = httptest.NewRequest("DELETE", "/clusters/1", nil)
	req.SetPathValue("id", "1")
	w = httptest.NewRecorder()

	deleteHandler := DeleteClusterHandler(repo, log)
	deleteHandler.ServeHTTP(w, req)

	// Assert deletion succeeded
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify cluster is gone
	req = httptest.NewRequest("GET", "/clusters/1", nil)
	req.SetPathValue("id", "1")
	w = httptest.NewRecorder()

	getHandler := GetClusterHandler(repo, log)
	getHandler.ServeHTTP(w, req)

	// Should return not found
	assert.Equal(t, http.StatusNotFound, w.Code)
}
