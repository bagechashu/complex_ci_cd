//go:build unit

package services

import (
	"context"
	"testing"

	"built-and-deploy/internal/models"
	"built-and-deploy/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockShellServerRepository 模拟 shell 服务器存储库
type MockShellServerRepository struct {
	mock.Mock
}

func (m *MockShellServerRepository) Create(ctx context.Context, server *models.ShellServer) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockShellServerRepository) GetByID(ctx context.Context, id int) (*models.ShellServer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShellServer), args.Error(1)
}

func (m *MockShellServerRepository) List(ctx context.Context, offset, limit int) ([]*models.ShellServer, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ShellServer), args.Int(1), args.Error(2)
}

func (m *MockShellServerRepository) Update(ctx context.Context, server *models.ShellServer) error {
	args := m.Called(ctx, server)
	return args.Error(0)
}

func (m *MockShellServerRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockShellServerRepository) GetByName(ctx context.Context, name string) (*models.ShellServer, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShellServer), args.Error(1)
}

// MockShellCommandRepository 模拟 shell 命令存储库
type MockShellCommandRepository struct {
	mock.Mock
}

func (m *MockShellCommandRepository) Create(ctx context.Context, cmd *models.ShellCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockShellCommandRepository) GetByID(ctx context.Context, id int) (*models.ShellCommand, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShellCommand), args.Error(1)
}

func (m *MockShellCommandRepository) List(ctx context.Context, offset, limit int) ([]*models.ShellCommand, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ShellCommand), args.Int(1), args.Error(2)
}

func (m *MockShellCommandRepository) Update(ctx context.Context, cmd *models.ShellCommand) error {
	args := m.Called(ctx, cmd)
	return args.Error(0)
}

func (m *MockShellCommandRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockShellCommandRepository) ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellCommand, int, error) {
	args := m.Called(ctx, serverID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ShellCommand), args.Int(1), args.Error(2)
}

func (m *MockShellCommandRepository) Publish(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockShellCommandRepository) Unpublish(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockShellTaskRepository 模拟 shell 任务存储库
type MockShellTaskRepository struct {
	mock.Mock
}

func (m *MockShellTaskRepository) Create(ctx context.Context, task *models.ShellTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockShellTaskRepository) GetByID(ctx context.Context, id int) (*models.ShellTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShellTask), args.Error(1)
}

func (m *MockShellTaskRepository) List(ctx context.Context, offset, limit int) ([]*models.ShellTask, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ShellTask), args.Int(1), args.Error(2)
}

func (m *MockShellTaskRepository) Update(ctx context.Context, task *models.ShellTask) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockShellTaskRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockShellTaskRepository) ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellTask, int, error) {
	args := m.Called(ctx, serverID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ShellTask), args.Int(1), args.Error(2)
}

// MockShellTaskExecutionRepository 模拟 shell 执行任务存储库
type MockShellTaskExecutionRepository struct {
	mock.Mock
}

func (m *MockShellTaskExecutionRepository) Create(ctx context.Context, task *models.ShellTaskExecution) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockShellTaskExecutionRepository) GetByID(ctx context.Context, id int) (*models.ShellTaskExecution, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShellTaskExecution), args.Error(1)
}

func (m *MockShellTaskExecutionRepository) List(ctx context.Context, offset, limit int) ([]*models.ShellTaskExecution, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ShellTaskExecution), args.Int(1), args.Error(2)
}

func (m *MockShellTaskExecutionRepository) Update(ctx context.Context, task *models.ShellTaskExecution) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockShellTaskExecutionRepository) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockShellTaskExecutionRepository) ListByTask(ctx context.Context, taskID int, offset, limit int) ([]*models.ShellTaskExecution, int, error) {
	args := m.Called(ctx, taskID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ShellTaskExecution), args.Int(1), args.Error(2)
}

func (m *MockShellTaskExecutionRepository) ListByServer(ctx context.Context, serverID int, offset, limit int) ([]*models.ShellTaskExecution, int, error) {
	args := m.Called(ctx, serverID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.ShellTaskExecution), args.Int(1), args.Error(2)
}

func (m *MockShellTaskExecutionRepository) GetLatestByTaskAndServer(ctx context.Context, taskID, serverID int) (*models.ShellTaskExecution, error) {
	args := m.Called(ctx, taskID, serverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ShellTaskExecution), args.Error(1)
}

// TestShellService_Creation tests shell service initialization
func TestShellService_Creation(t *testing.T) {
	// Setup
	mockTaskRepo := new(MockShellTaskRepository)
	mockServerRepo := new(MockShellServerRepository)
	mockCommandRepo := new(MockShellCommandRepository)
	mockExecRepo := new(MockShellTaskExecutionRepository)
	log := logger.NewLogger()

	// Create service
	service := NewShellService(
		mockTaskRepo,
		mockServerRepo,
		mockCommandRepo,
		mockExecRepo,
		"test-key",
		log,
	)

	// Assert service is not nil
	assert.NotNil(t, service)
}

// TestShellService_ShellServerRepo tests server repository accessor
func TestShellService_ShellServerRepo(t *testing.T) {
	mockTaskRepo := new(MockShellTaskRepository)
	mockServerRepo := new(MockShellServerRepository)
	mockCommandRepo := new(MockShellCommandRepository)
	mockExecRepo := new(MockShellTaskExecutionRepository)
	log := logger.NewLogger()

	service := NewShellService(
		mockTaskRepo,
		mockServerRepo,
		mockCommandRepo,
		mockExecRepo,
		"test-key",
		log,
	)

	// Verify repository accessor
	assert.Equal(t, mockServerRepo, service.ShellServerRepo())
}

// TestShellService_ShellCommandRepo tests command repository accessor
func TestShellService_ShellCommandRepo(t *testing.T) {
	mockTaskRepo := new(MockShellTaskRepository)
	mockServerRepo := new(MockShellServerRepository)
	mockCommandRepo := new(MockShellCommandRepository)
	mockExecRepo := new(MockShellTaskExecutionRepository)
	log := logger.NewLogger()

	service := NewShellService(
		mockTaskRepo,
		mockServerRepo,
		mockCommandRepo,
		mockExecRepo,
		"test-key",
		log,
	)

	// Verify repository accessor
	assert.Equal(t, mockCommandRepo, service.ShellCommandRepo())
}

// TestShellService_ShellTaskRepo tests exec repository accessor
func TestShellService_ShellTaskRepo(t *testing.T) {
	mockTaskRepo := new(MockShellTaskRepository)
	mockServerRepo := new(MockShellServerRepository)
	mockCommandRepo := new(MockShellCommandRepository)
	mockExecRepo := new(MockShellTaskExecutionRepository)
	log := logger.NewLogger()

	service := NewShellService(
		mockTaskRepo,
		mockServerRepo,
		mockCommandRepo,
		mockExecRepo,
		"test-key",
		log,
	)

	// Verify repository accessor
	assert.Equal(t, mockExecRepo, service.ShellTaskRepo())
}
