//go:build unit

package shell_tasks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockShellServerRepository 模拟 shell 服务器存储库
type MockShellServerRepository struct {
	mock.Mock
}

func (m *MockShellServerRepository) Create(ctx interface{}, server interface{}) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockShellServerRepository) GetByID(ctx interface{}, id int) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockShellServerRepository) List(ctx interface{}, offset, limit int) (interface{}, int, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0), args.Int(1), args.Error(2)
}

func (m *MockShellServerRepository) Update(ctx interface{}, server interface{}) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockShellServerRepository) Delete(ctx interface{}, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockShellCommandRepository 模拟 shell 命令存储库
type MockShellCommandRepository struct {
	mock.Mock
}

func (m *MockShellCommandRepository) Create(ctx interface{}, cmd interface{}) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockShellCommandRepository) GetByID(ctx interface{}, id int) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockShellCommandRepository) List(ctx interface{}, offset, limit int) (interface{}, int, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0), args.Int(1), args.Error(2)
}

func (m *MockShellCommandRepository) Update(ctx interface{}, cmd interface{}) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockShellCommandRepository) Delete(ctx interface{}, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockShellTaskRepository 模拟 shell 执行存储库
type MockShellTaskRepository struct {
	mock.Mock
}

func (m *MockShellTaskRepository) Create(ctx interface{}, task interface{}) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockShellTaskRepository) GetByID(ctx interface{}, id int) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

func (m *MockShellTaskRepository) List(ctx interface{}, offset, limit int) (interface{}, int, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0), args.Int(1), args.Error(2)
}

func (m *MockShellTaskRepository) Update(ctx interface{}, task interface{}) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockShellTaskRepository) Delete(ctx interface{}, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestList_EmptyResult tests that list handler returns empty array
func TestList_EmptyResult(t *testing.T) {
	// Setup
	mockShellService := &services.ShellService{}
	log := logger.NewLogger()

	// Create request
	req := httptest.NewRequest("GET", "/shell-tasks", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler := List(mockShellService, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var result []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
}

// TestCreate_ValidRequest tests shell task creation with valid request
func TestCreate_ValidRequest(t *testing.T) {
	// Setup
	mockShellService := &services.ShellService{}
	log := logger.NewLogger()

	// Create request
	reqBody := map[string]interface{}{
		"command":     "ls -la",
		"description": "List files",
		"server_ids":  []int{1, 2},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/shell-tasks", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	// Execute handler
	handler := Create(mockShellService, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "ls -la", result["command"])
}

// TestCreate_InvalidJSON tests shell task creation with invalid JSON
func TestCreate_InvalidJSON(t *testing.T) {
	// Setup
	mockShellService := &services.ShellService{}
	log := logger.NewLogger()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/shell-tasks", bytes.NewReader([]byte("{invalid json}")))
	w := httptest.NewRecorder()

	// Execute handler
	handler := Create(mockShellService, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
