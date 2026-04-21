//go:build integration

package applications

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

func setupTestDB(t *testing.T) repository.ApplicationRepository {
	dbPath := "test_integration.db"
	db, err := database.Init(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		database.Close(db)
		os.Remove(dbPath)
	})
	return repository.NewSQLiteApplicationRepository(db)
}

// TestIntegration_CreateApplication_CompleteFlow tests end-to-end application creation
func TestIntegration_CreateApplication_CompleteFlow(t *testing.T) {
	repo := setupTestDB(t)
	log := logger.NewLogger()

	// Create request
	reqBody := map[string]string{
		"name":       "integration-test-app",
		"repository": "docker.io/integration:latest",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/applications", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	// Execute handler
	handler := Create(repo, log)
	handler.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	// Parse response and verify persistence
	var createdApp models.Application
	err := json.Unmarshal(w.Body.Bytes(), &createdApp)
	require.NoError(t, err)
	assert.Equal(t, "integration-test-app", createdApp.Name)
	assert.NotZero(t, createdApp.ID)

	// Verify data persisted in database
	storedApp, err := repo.GetByID(context.Background(), createdApp.ID)
	require.NoError(t, err)
	assert.Equal(t, createdApp.ID, storedApp.ID)
	assert.Equal(t, "integration-test-app", storedApp.Name)
}

// TestIntegration_ListApplications_MultipleCreated tests list endpoint with multiple apps
func TestIntegration_ListApplications_MultipleCreated(t *testing.T) {
	repo := setupTestDB(t)
	log := logger.NewLogger()

	// Create multiple applications
	apps := []map[string]string{
		{
			"name":       "app-1",
			"repository": "docker.io/app1:latest",
		},
		{
			"name":       "app-2",
			"repository": "docker.io/app2:latest",
		},
		{
			"name":       "app-3",
			"repository": "docker.io/app3:latest",
		},
	}

	for _, appData := range apps {
		bodyBytes, _ := json.Marshal(appData)
		req := httptest.NewRequest("POST", "/applications", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		handler := Create(repo, log)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// List applications
	req := httptest.NewRequest("GET", "/applications", nil)
	w := httptest.NewRecorder()

	handler := List(repo, log)
	handler.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var result []*models.Application
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 3)
}

// TestIntegration_CreateApplication_ValidationFailure tests validation in handler
func TestIntegration_CreateApplication_ValidationFailure(t *testing.T) {
	repo := setupTestDB(t)
	log := logger.NewLogger()

	// Create request with missing required fields
	reqBody := map[string]string{
		"name": "", // Empty name should fail
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/applications", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	// Execute handler
	handler := Create(repo, log)
	handler.ServeHTTP(w, req)

	// Assert validation error
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestIntegration_CreateAndListCycle tests create and list in sequence
func TestIntegration_CreateAndListCycle(t *testing.T) {
	repo := setupTestDB(t)
	log := logger.NewLogger()

	// Create first app
	req1Body := map[string]string{
		"name":       "cycle-test-1",
		"repository": "docker.io/cycle1:latest",
	}
	bodyBytes, _ := json.Marshal(req1Body)
	req := httptest.NewRequest("POST", "/applications", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler := Create(repo, log)
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Create second app
	req2Body := map[string]string{
		"name":       "cycle-test-2",
		"repository": "docker.io/cycle2:latest",
	}
	bodyBytes, _ = json.Marshal(req2Body)
	req = httptest.NewRequest("POST", "/applications", bytes.NewReader(bodyBytes))
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// List all applications
	req = httptest.NewRequest("GET", "/applications", nil)
	w = httptest.NewRecorder()

	listHandler := List(repo, log)
	listHandler.ServeHTTP(w, req)

	var result []*models.Application
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.True(t, len(result) >= 2, "Should have at least 2 applications")
}
