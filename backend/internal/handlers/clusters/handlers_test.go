//go:build unit

package clusters

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

// MockClusterRepository 模拟集群存储库
type MockClusterRepository struct {
	mock.Mock
}

func (m *MockClusterRepository) Create(ctx context.Context, cluster *models.Cluster) error {
	args := m.Called(ctx, cluster)
	return args.Error(0)
}

func (m *MockClusterRepository) GetByID(ctx context.Context, id int) (*models.Cluster, error) {
	args := m.Called(ctx, id)
	if cluster := args.Get(0); cluster != nil {
		return cluster.(*models.Cluster), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockClusterRepository) List(ctx context.Context, offset, limit int) ([]*models.Cluster, int, error) {
	args := m.Called(ctx, offset, limit)
	if clusters := args.Get(0); clusters != nil {
		return clusters.([]*models.Cluster), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockClusterRepository) Update(ctx context.Context, cluster *models.Cluster) error {
	args := m.Called(ctx, cluster)
	return args.Error(0)
}

func (m *MockClusterRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestListClustersHandler_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockClusterRepository)
	log := logger.NewLogger()

	labels := "env=prod,tier=critical"
	kubeconfig := "apiVersion: v1\nkind: Config"

	clusters := []*models.Cluster{
		{
			ID:         1,
			Name:       "prod-cluster",
			Type:       "kubernetes",
			Labels:     &labels,
			Kubeconfig: &kubeconfig,
			CreatedAt:  time.Now(),
		},
	}

	mockRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), 0, 100).Return(clusters, 1, nil)

	// Create request
	req := httptest.NewRequest("GET", "/clusters", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler := ListClustersHandler(mockRepo, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var result []*models.Cluster
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "prod-cluster", result[0].Name)

	mockRepo.AssertExpectations(t)
}

func TestCreateClusterHandler_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockClusterRepository)
	log := logger.NewLogger()

	mockRepo.On("Create", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), mock.MatchedBy(func(cluster *models.Cluster) bool {
		return cluster.Name == "test-cluster" && cluster.Type == "kubernetes"
	})).Run(func(args mock.Arguments) {
		cluster := args.Get(1).(*models.Cluster)
		cluster.ID = 1
		cluster.CreatedAt = time.Now()
		cluster.UpdatedAt = time.Now()
	}).Return(nil)

	// Create request
	reqBody := map[string]string{
		"name":       "test-cluster",
		"type":       "kubernetes",
		"labels":     "env=dev",
		"kubeconfig": "apiVersion: v1",
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/clusters", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	// Execute handler
	handler := CreateClusterHandler(mockRepo, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestCreateClusterHandler_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		reqBody  map[string]string
		wantCode int
	}{
		{
			name: "missing name",
			reqBody: map[string]string{
				"type": "kubernetes",
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing type",
			reqBody: map[string]string{
				"name": "test-cluster",
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockClusterRepository)
			log := logger.NewLogger()

			bodyBytes, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/clusters", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			handler := CreateClusterHandler(mockRepo, log)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestGetClusterHandler_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockClusterRepository)
	log := logger.NewLogger()

	labels := "env=prod"
	kubeconfig := "config"

	cluster := &models.Cluster{
		ID:         1,
		Name:       "test-cluster",
		Type:       "kubernetes",
		Labels:     &labels,
		Kubeconfig: &kubeconfig,
		CreatedAt:  time.Now(),
	}

	mockRepo.On("GetByID", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), 1).Return(cluster, nil)

	// Create request
	req := httptest.NewRequest("GET", "/clusters/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	// Execute handler
	handler := GetClusterHandler(mockRepo, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestGetClusterHandler_NotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockClusterRepository)
	log := logger.NewLogger()

	mockRepo.On("GetByID", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), 1).Return(nil, nil)

	// Create request
	req := httptest.NewRequest("GET", "/clusters/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	// Execute handler
	handler := GetClusterHandler(mockRepo, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestGetClusterHandler_InvalidID(t *testing.T) {
	// Setup
	mockRepo := new(MockClusterRepository)
	log := logger.NewLogger()

	// Create request with invalid ID
	req := httptest.NewRequest("GET", "/clusters/invalid", nil)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()

	// Execute handler
	handler := GetClusterHandler(mockRepo, log)
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
