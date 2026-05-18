//go:build unit

package services

import (
	"context"
	"testing"

	"built-and-deploy/internal/models"
	"built-and-deploy/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockReleaseRecordRepositoryForService 模拟发布记录存储库
type MockReleaseRecordRepositoryForService struct {
	mock.Mock
}

func (m *MockReleaseRecordRepositoryForService) Create(ctx context.Context, record *models.ReleaseRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockReleaseRecordRepositoryForService) GetByID(ctx context.Context, id int) (*models.ReleaseRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ReleaseRecord), args.Error(1)
}

func (m *MockReleaseRecordRepositoryForService) Update(ctx context.Context, record *models.ReleaseRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockReleaseRecordRepositoryForService) List(ctx context.Context, offset, limit int) ([]*models.ReleaseRecord, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ReleaseRecord), args.Int(1), args.Error(2)
}

func (m *MockReleaseRecordRepositoryForService) GetByApplicationAndCluster(ctx context.Context, appID, clusterID int) ([]*models.ReleaseRecord, error) {
	args := m.Called(ctx, appID, clusterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ReleaseRecord), args.Error(1)
}

func (m *MockReleaseRecordRepositoryForService) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockReleaseEventRepositoryForService 模拟发布事件存储库
type MockReleaseEventRepositoryForService struct {
	mock.Mock
}

func (m *MockReleaseEventRepositoryForService) Create(ctx context.Context, event *models.ReleaseEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockReleaseEventRepositoryForService) ListByRelease(ctx context.Context, releaseID int) ([]*models.ReleaseEvent, error) {
	args := m.Called(ctx, releaseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ReleaseEvent), args.Error(1)
}

// TestReleaseService_Release tests release operation
func TestReleaseService_Release(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRecordRepositoryForService)
	mockEventRepo := new(MockReleaseEventRepositoryForService)
	log := logger.NewLogger()

	mockReleaseRepo.On("Create", mock.Anything, mock.MatchedBy(func(r *models.ReleaseRecord) bool {
		return r.AppID > 0
	})).Run(func(args mock.Arguments) {
		r := args.Get(1).(*models.ReleaseRecord)
		r.ID = 1
	}).Return(nil)

	mockEventRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	service := NewReleaseService(
		mockReleaseRepo,
		nil,
		nil,
		nil,
		mockEventRepo,
		nil,
		log,
		nil,
	)

	// Execute
	release, err := service.Release(context.Background(), 1, 1, "test:1.0.0")

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, release)
	mockReleaseRepo.AssertExpectations(t)
}

// TestReleaseService_GetReleaseStatus tests getting release status
func TestReleaseService_GetReleaseStatus(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRecordRepositoryForService)
	log := logger.NewLogger()

	expectedRelease := &models.ReleaseRecord{
		ID:     1,
		AppID:  1,
		Image:  "test:1.0.0",
		Status: "completed",
	}

	mockReleaseRepo.On("GetByID", mock.Anything, 1).Return(expectedRelease, nil)

	service := NewReleaseService(
		mockReleaseRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		log,
		nil,
	)

	// Execute
	release, err := service.GetReleaseStatus(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedRelease.ID, release.ID)
	mockReleaseRepo.AssertExpectations(t)
}

// TestReleaseService_GetReleaseHistory tests listing release history
func TestReleaseService_GetReleaseHistory(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRecordRepositoryForService)
	log := logger.NewLogger()

	releases := []*models.ReleaseRecord{
		{ID: 1, AppID: 1, Image: "test:1.0.0", Status: "completed"},
		{ID: 2, AppID: 1, Image: "test:1.0.1", Status: "completed"},
	}

	mockReleaseRepo.On("List", mock.Anything, 0, 50).Return(releases, 2, nil)

	service := NewReleaseService(
		mockReleaseRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		log,
		nil,
	)

	// Execute
	result, err := service.GetReleaseHistory(context.Background(), 0, 50)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 2)
	mockReleaseRepo.AssertExpectations(t)
}

// TestReleaseService_Rollback tests rollback operation
func TestReleaseService_Rollback(t *testing.T) {
	mockReleaseRepo := new(MockReleaseRecordRepositoryForService)
	mockEventRepo := new(MockReleaseEventRepositoryForService)
	log := logger.NewLogger()

	originalRelease := &models.ReleaseRecord{
		ID:     1,
		AppID:  1,
		Image:  "test:1.0.0",
		Status: "completed",
	}

	mockReleaseRepo.On("GetByID", mock.Anything, 1).Return(originalRelease, nil)
	mockReleaseRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
	mockEventRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	service := NewReleaseService(
		mockReleaseRepo,
		nil,
		nil,
		nil,
		mockEventRepo,
		nil,
		log,
		nil,
	)

	// Execute
	release, err := service.Rollback(context.Background(), 1)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, release)
	assert.Equal(t, "rolled_back", release.Status)
	mockReleaseRepo.AssertExpectations(t)
}
