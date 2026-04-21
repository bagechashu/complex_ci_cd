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

// MockAppRepository 模拟应用存储库
type MockAppRepository struct {
	mock.Mock
}

func (m *MockAppRepository) Create(ctx context.Context, app *models.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockAppRepository) GetByID(ctx context.Context, id int) (*models.Application, error) {
	args := m.Called(ctx, id)
	if app := args.Get(0); app != nil {
		return app.(*models.Application), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAppRepository) List(ctx context.Context, offset, limit int) ([]*models.Application, int, error) {
	args := m.Called(ctx, offset, limit)
	if apps := args.Get(0); apps != nil {
		return apps.([]*models.Application), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockAppRepository) ListWithSearch(ctx context.Context, offset, limit int, search string) ([]*models.Application, int, error) {
	args := m.Called(ctx, offset, limit, search)
	if apps := args.Get(0); apps != nil {
		return apps.([]*models.Application), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockAppRepository) Update(ctx context.Context, app *models.Application) error {
	args := m.Called(ctx, app)
	return args.Error(0)
}

func (m *MockAppRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockReleaseRepository 模拟发布存储库
type MockReleaseRepository struct {
	mock.Mock
}

func (m *MockReleaseRepository) Create(ctx context.Context, record *models.ReleaseRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockReleaseRepository) GetByID(ctx context.Context, id int) (*models.ReleaseRecord, error) {
	args := m.Called(ctx, id)
	if record := args.Get(0); record != nil {
		return record.(*models.ReleaseRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockReleaseRepository) List(ctx context.Context, offset, limit int) ([]*models.ReleaseRecord, int, error) {
	args := m.Called(ctx, offset, limit)
	if records := args.Get(0); records != nil {
		return records.([]*models.ReleaseRecord), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockReleaseRepository) ListByApplication(ctx context.Context, appID int, offset, limit int) ([]*models.ReleaseRecord, int, error) {
	args := m.Called(ctx, appID, offset, limit)
	if records := args.Get(0); records != nil {
		return records.([]*models.ReleaseRecord), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockReleaseRepository) Update(ctx context.Context, record *models.ReleaseRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockReleaseRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockReleaseRepository) GetByApplicationAndCluster(ctx context.Context, appID, clusterID int) ([]*models.ReleaseRecord, error) {
	args := m.Called(ctx, appID, clusterID)
	if records := args.Get(0); records != nil {
		return records.([]*models.ReleaseRecord), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestApplicationService_Create_Success(t *testing.T) {
	// Setup
	mockAppRepo := new(MockAppRepository)
	mockReleaseRepo := new(MockReleaseRepository)
	log := logger.NewLogger()

	req := &handlers.CreateApplicationRequest{
		Name:       "test-app",
		Repository: "docker.io/test:latest",
	}

	mockAppRepo.On("Create", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), mock.MatchedBy(func(app *models.Application) bool {
		return app.Name == "test-app"
	})).Run(func(args mock.Arguments) {
		app := args.Get(1).(*models.Application)
		app.ID = 1
	}).Return(nil)

	service := NewApplicationService(mockAppRepo, mockReleaseRepo, log)

	// Execute
	app, err := service.Create(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, app)
	assert.Equal(t, "test-app", app.Name)
	mockAppRepo.AssertExpectations(t)
}

func TestApplicationService_Create_InvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		req     *handlers.CreateApplicationRequest
		wantErr bool
	}{
		{
			name: "empty name",
			req: &handlers.CreateApplicationRequest{
				Name:       "",
				Repository: "docker.io/test:latest",
			},
			wantErr: true,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAppRepo := new(MockAppRepository)
			mockReleaseRepo := new(MockReleaseRepository)
			log := logger.NewLogger()

			service := NewApplicationService(mockAppRepo, mockReleaseRepo, log)

			// Execute
			app, err := service.Create(context.Background(), tt.req)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, app)
			}
		})
	}
}

func TestApplicationService_GetApplication_Success(t *testing.T) {
	// Setup
	mockAppRepo := new(MockAppRepository)
	mockReleaseRepo := new(MockReleaseRepository)
	log := logger.NewLogger()

	expectedApp := &models.Application{
		ID:        1,
		Name:      "test-app",
		ImageName: "docker.io/test:latest",
		CreatedAt: time.Now(),
	}

	mockAppRepo.On("GetByID", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), 1).Return(expectedApp, nil)

	service := NewApplicationService(mockAppRepo, mockReleaseRepo, log)

	// Execute
	app, err := service.GetApplication(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedApp.ID, app.ID)
	assert.Equal(t, expectedApp.Name, app.Name)
	mockAppRepo.AssertExpectations(t)
}

func TestApplicationService_GetApplication_NotFound(t *testing.T) {
	// Setup
	mockAppRepo := new(MockAppRepository)
	mockReleaseRepo := new(MockReleaseRepository)
	log := logger.NewLogger()

	mockAppRepo.On("GetByID", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), 1).Return(nil, nil)

	service := NewApplicationService(mockAppRepo, mockReleaseRepo, log)

	// Execute
	app, err := service.GetApplication(context.Background(), 1)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, app)
	mockAppRepo.AssertExpectations(t)
}

func TestApplicationService_ListApplications_Success(t *testing.T) {
	// Setup
	mockAppRepo := new(MockAppRepository)
	mockReleaseRepo := new(MockReleaseRepository)
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

	mockAppRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), 0, 10).Return(apps, 2, nil)

	service := NewApplicationService(mockAppRepo, mockReleaseRepo, log)

	// Execute
	result, err := service.ListApplications(context.Background(), 0, 10)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "app1", result[0].Name)
	mockAppRepo.AssertExpectations(t)
}

func TestApplicationService_ListApplications_DefaultLimit(t *testing.T) {
	// Setup
	mockAppRepo := new(MockAppRepository)
	mockReleaseRepo := new(MockReleaseRepository)
	log := logger.NewLogger()

	mockAppRepo.On("List", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	}), 0, 10).Return([]*models.Application{}, 0, nil)

	service := NewApplicationService(mockAppRepo, mockReleaseRepo, log)

	// Execute - pass invalid limit, should use default (10)
	result, err := service.ListApplications(context.Background(), 0, -1)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	mockAppRepo.AssertExpectations(t)
}
