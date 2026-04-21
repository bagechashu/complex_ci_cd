//go:build unit

package applications

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockApplicationRepository 模拟应用存储库
type MockApplicationRepository struct {
	mock.Mock
}

func (m *MockApplicationRepository) Create(ctx context.Context, app *models.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationRepository) GetByID(ctx context.Context, id int) (*models.Application, error) {
	args := m.Called(ctx, id)
	if app := args.Get(0); app != nil {
		return app.(*models.Application), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockApplicationRepository) List(ctx context.Context, offset, limit int) ([]*models.Application, int, error) {
	args := m.Called(ctx, offset, limit)
	if apps := args.Get(0); apps != nil {
		return apps.([]*models.Application), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockApplicationRepository) ListWithSearch(ctx context.Context, offset, limit int, search string) ([]*models.Application, int, error) {
	args := m.Called(ctx, offset, limit, search)
	if apps := args.Get(0); apps != nil {
		return apps.([]*models.Application), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockApplicationRepository) Update(ctx context.Context, app *models.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockApplicationRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreate_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockApplicationRepository)
	log := logger.NewLogger()

	// Expected application (will be set by Create)
	expectedApp := &models.Application{
		Name:      "test-app",
		ImageName: "docker.io/test:latest",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock the repository
	mockRepo.On("Create", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), mock.MatchedBy(func(app *models.Application) bool {
		return app.Name == "test-app" && app.ImageName == "docker.io/test:latest"
	})).Run(func(args mock.Arguments) {
		app := args.Get(1).(*models.Application)
		app.ID = 1
		app.CreatedAt = expectedApp.CreatedAt
		app.UpdatedAt = expectedApp.UpdatedAt
	}).Return(nil)

	// Create request
	reqBody := map[string]string{
		"name":       "test-app",
		"repository": "docker.io/test:latest",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/applications", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	// Execute handler
	handler := Create(mockRepo, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestCreate_InvalidRequest(t *testing.T) {
	tests := []struct {
		name       string
		bodyFunc   func() []byte
		statusCode int
		shouldFail bool
	}{
		{
			name: "missing name",
			bodyFunc: func() []byte {
				return []byte(`{"repository":"docker.io/test:latest"}`)
			},
			statusCode: http.StatusBadRequest,
			shouldFail: true,
		},
		{
			name: "invalid json",
			bodyFunc: func() []byte {
				return []byte(`{invalid json}`)
			},
			statusCode: http.StatusBadRequest,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockApplicationRepository)
			log := logger.NewLogger()

			req := httptest.NewRequest("POST", "/applications", bytes.NewReader(tt.bodyFunc()))
			w := httptest.NewRecorder()

			handler := Create(mockRepo, log)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestList_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockApplicationRepository)
	log := logger.NewLogger()

	apps := []*models.Application{
		{
			ID:        1,
			Name:      "app1",
			ImageName: "docker.io/app1:latest",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Name:      "app2",
			ImageName: "docker.io/app2:latest",
			CreatedAt: time.Now(),
		},
	}

	mockRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), 0, 100).Return(apps, 2, nil)

	// Create request
	req := httptest.NewRequest("GET", "/applications", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler := List(mockRepo, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var result []*models.Application
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "app1", result[0].Name)

	mockRepo.AssertExpectations(t)
}

func TestList_Empty(t *testing.T) {
	// Setup
	mockRepo := new(MockApplicationRepository)
	log := logger.NewLogger()

	mockRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), 0, 100).Return([]*models.Application{}, 0, nil)

	// Create request
	req := httptest.NewRequest("GET", "/applications", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler := List(mockRepo, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}
