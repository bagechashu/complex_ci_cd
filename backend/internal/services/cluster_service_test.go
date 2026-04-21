//go:build unit

package services

import (
	"context"
	"testing"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/pkg/handlers"
	"built-and-deploy/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockClusterRepo 模拟集群存储库
type MockClusterRepo struct {
	mock.Mock
}

func (m *MockClusterRepo) Create(ctx context.Context, cluster *models.Cluster) error {
	args := m.Called(ctx, cluster)
	return args.Error(0)
}

func (m *MockClusterRepo) GetByID(ctx context.Context, id int) (*models.Cluster, error) {
	args := m.Called(ctx, id)
	if cluster := args.Get(0); cluster != nil {
		return cluster.(*models.Cluster), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockClusterRepo) List(ctx context.Context, offset, limit int) ([]*models.Cluster, int, error) {
	args := m.Called(ctx, offset, limit)
	if clusters := args.Get(0); clusters != nil {
		return clusters.([]*models.Cluster), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockClusterRepo) Update(ctx context.Context, cluster *models.Cluster) error {
	args := m.Called(ctx, cluster)
	return args.Error(0)
}

func (m *MockClusterRepo) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockEnvRepo 模拟环境存储库
type MockEnvRepo struct {
	mock.Mock
}

func (m *MockEnvRepo) Create(ctx context.Context, env *models.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *MockEnvRepo) GetByID(ctx context.Context, id int) (*models.Environment, error) {
	args := m.Called(ctx, id)
	if env := args.Get(0); env != nil {
		return env.(*models.Environment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockEnvRepo) List(ctx context.Context, offset, limit int) ([]*models.Environment, int, error) {
	args := m.Called(ctx, offset, limit)
	if envs := args.Get(0); envs != nil {
		return envs.([]*models.Environment), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockEnvRepo) Update(ctx context.Context, env *models.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *MockEnvRepo) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestClusterService_CreateCluster_Success tests cluster creation
func TestClusterService_CreateCluster_Success(t *testing.T) {
	// Setup
	mockClusterRepo := new(MockClusterRepo)
	log := logger.NewLogger()

	req := &handlers.CreateClusterRequest{
		Name:       "prod-k8s",
		Type:       "kubernetes",
		Kubeconfig: "apiVersion: v1\nkind: Config\n...",
	}

	mockClusterRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *models.Cluster) bool {
		return c.Name == "prod-k8s" && c.Type == "kubernetes"
	})).Run(func(args mock.Arguments) {
		c := args.Get(1).(*models.Cluster)
		c.ID = 1
	}).Return(nil)

	service := NewClusterService(mockClusterRepo, nil, "test-encryption-key", log)

	// Execute
	cluster, err := service.CreateCluster(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, "prod-k8s", cluster.Name)
	assert.NotNil(t, cluster.Kubeconfig)
	mockClusterRepo.AssertExpectations(t)
}

// TestClusterService_CreateCluster_Validation tests cluster validation
func TestClusterService_CreateCluster_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *handlers.CreateClusterRequest
		wantErr bool
	}{
		{
			name: "missing cluster name",
			req: &handlers.CreateClusterRequest{
				Type:       "kubernetes",
				Kubeconfig: "...",
			},
			wantErr: true,
		},
		{
			name: "valid cluster",
			req: &handlers.CreateClusterRequest{
				Name:       "test-cluster",
				Type:       "kubernetes",
				Kubeconfig: "...",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClusterRepo := new(MockClusterRepo)
			log := logger.NewLogger()

			if tt.wantErr == false {
				mockClusterRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			}

			service := NewClusterService(mockClusterRepo, nil, "test-key", log)

			// Execute
			_, err := service.CreateCluster(context.Background(), tt.req)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestClusterService_GetCluster_Success tests getting cluster by ID
func TestClusterService_GetCluster_Success(t *testing.T) {
	// Setup
	mockClusterRepo := new(MockClusterRepo)
	log := logger.NewLogger()

	expectedCluster := &models.Cluster{
		ID:        1,
		Name:      "prod-cluster",
		Type:      "kubernetes",
		CreatedAt: time.Now(),
	}

	mockClusterRepo.On("GetByID", mock.Anything, 1).Return(expectedCluster, nil)

	service := NewClusterService(mockClusterRepo, nil, "test-key", log)

	// Execute
	cluster, err := service.GetCluster(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedCluster.ID, cluster.ID)
	assert.Equal(t, "prod-cluster", cluster.Name)
	mockClusterRepo.AssertExpectations(t)
}

// TestClusterService_ListClusters_Success tests listing clusters
func TestClusterService_ListClusters_Success(t *testing.T) {
	// Setup
	mockClusterRepo := new(MockClusterRepo)
	log := logger.NewLogger()

	clusters := []*models.Cluster{
		{
			ID:        1,
			Name:      "cluster-1",
			Type:      "kubernetes",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Name:      "cluster-2",
			Type:      "ssh",
			CreatedAt: time.Now(),
		},
	}

	mockClusterRepo.On("List", mock.Anything, 0, 1000).Return(clusters, 2, nil)

	service := NewClusterService(mockClusterRepo, nil, "test-key", log)

	// Execute
	result, err := service.ListClusters(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 2)
	mockClusterRepo.AssertExpectations(t)
}

// TestClusterService_UpdateCluster_Success tests cluster update
func TestClusterService_UpdateCluster_Success(t *testing.T) {
	// Setup
	mockClusterRepo := new(MockClusterRepo)
	log := logger.NewLogger()

	mockClusterRepo.On("GetByID", mock.Anything, 1).Return(&models.Cluster{
		ID:        1,
		Name:      "old-name",
		Type:      "kubernetes",
		UpdatedAt: time.Now(),
	}, nil)

	mockClusterRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *models.Cluster) bool {
		return c.Name == "updated-cluster"
	})).Return(nil)

	service := NewClusterService(mockClusterRepo, nil, "test-key", log)

	// Execute
	req := &handlers.UpdateClusterRequest{
		Name:       "updated-cluster",
		Kubeconfig: "...",
	}
	_, err := service.UpdateCluster(context.Background(), 1, req)

	// Assert
	require.NoError(t, err)
	mockClusterRepo.AssertExpectations(t)
}

// TestClusterService_DeleteCluster_Success tests cluster deletion
func TestClusterService_DeleteCluster_Success(t *testing.T) {
	// Setup
	mockClusterRepo := new(MockClusterRepo)
	log := logger.NewLogger()

	mockClusterRepo.On("Delete", mock.Anything, 1).Return(nil)

	service := NewClusterService(mockClusterRepo, nil, "test-key", log)

	// Execute
	err := service.DeleteCluster(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	mockClusterRepo.AssertExpectations(t)
}
