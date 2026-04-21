//go:build integration

package shell_tasks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"built-and-deploy/internal/database"
	"built-and-deploy/internal/repository"
	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupShellTestDB(t *testing.T) *services.ShellService {
	dbPath := "test_shell_integration.db"
	db, err := database.Init(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		database.Close(db)
		os.Remove(dbPath)
	})

	return services.NewShellService(
		repository.NewSQLiteShellServerRepository(db),
		repository.NewSQLiteShellCommandRepository(db),
		repository.NewSQLiteShellTaskExecutionRepository(db),
		"test-key",
		logger.NewLogger(),
	)
}

// TestIntegration_ListShellTasks tests listing shell tasks
func TestIntegration_ListShellTasks(t *testing.T) {
	shellService := setupShellTestDB(t)
	log := logger.NewLogger()

	// Create request
	req := httptest.NewRequest("GET", "/shell-tasks", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler := List(shellService, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var result []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
}

// TestIntegration_CreateShellTask tests creating a shell task
func TestIntegration_CreateShellTask(t *testing.T) {
	shellService := setupShellTestDB(t)
	log := logger.NewLogger()

	// Create request
	reqBody := map[string]interface{}{
		"command":     "echo test",
		"description": "Test command",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/shell-tasks", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	// Execute handler
	handler := Create(shellService, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, "echo test", result["command"])
}
