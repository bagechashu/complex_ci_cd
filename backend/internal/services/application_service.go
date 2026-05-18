// Package services provides business logic layer for the Release Control System.
//
// This package contains service implementations that orchestrate domain logic,
// manage transactions, and coordinate between repositories. Each service handles
// a specific domain entity (Application, Cluster, Release, Shell, Workload).
//
// Services are responsible for:
//   - Input validation
//   - Authorization checks
//   - Business rule enforcement
//   - Transaction management
//   - Error handling and logging
//   - Repository coordination
//
// All service methods accept context.Context for cancellation and timeout support.
//
// See:
//   - ApplicationService for application management
//   - ClusterService for cluster operations
//   - ReleaseService for deployment orchestration
//   - ShellService for command execution
//   - WorkloadService for workload-based deployments
package services

import (
	"context"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/errors"
	"built-and-deploy/pkg/handlers"
	"built-and-deploy/pkg/logger"
)

// ApplicationService manages application lifecycle operations.
//
// ApplicationService coordinates creation, retrieval, updating, and deletion of
// applications. It validates input, enforces business rules, logs operations,
// and coordinates with ApplicationRepository and ReleaseRecordRepository.
//
// Usage:
//
//	service := NewApplicationService(appRepo, releaseRepo, log)
//	app, err := service.Create(ctx, &CreateApplicationRequest{
//	    Name:       "api-service",
//	    Repository: "docker.io/api:v1.0.0",
//	})
//	if err != nil {
//	    log.Error("Failed to create application", "error", err)
//	}
type ApplicationService struct {
	appRepo     repository.ApplicationRepository
	releaseRepo repository.ReleaseRecordRepository
	log         *logger.Logger
}

// NewApplicationService creates a new ApplicationService instance.
//
// Parameters:
//   - appRepo: ApplicationRepository for data persistence
//   - releaseRepo: ReleaseRecordRepository for release tracking
//   - log: Logger for structured logging
//
// Returns a configured ApplicationService ready for use.
//
// Example:
//
//	service := NewApplicationService(appRepo, releaseRepo, logger.GetLogger())
func NewApplicationService(
	appRepo repository.ApplicationRepository,
	releaseRepo repository.ReleaseRecordRepository,
	log *logger.Logger,
) *ApplicationService {
	return &ApplicationService{
		appRepo:     appRepo,
		releaseRepo: releaseRepo,
		log:         log,
	}
}

// Create creates a new application in the system.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - req: CreateApplicationRequest containing application details
//
// Returns:
//   - *models.Application: The created application with assigned ID
//   - error: Non-nil if validation fails or database operation fails
//
// Errors:
//   - "INVALID_INPUT": If required fields are missing
//   - "DATABASE": If database operation fails
//
// Side effects:
//   - Creates new Application record in database
//   - Logs operation with application name
//
// Example:
//
//	app, err := service.Create(ctx, &CreateApplicationRequest{
//	    Name:       "payment-service",
//	    Repository: "docker.io/payment:v1.0.0",
//	})
func (s *ApplicationService) Create(ctx context.Context, req *handlers.CreateApplicationRequest) (*models.Application, error) {
	// 验证输入
	if req.Name == "" {
		return nil, errors.NewServiceError("INVALID_INPUT", "Application name is required")
	}

	// 创建应用
	app := &models.Application{
		Name:        req.Name,
		ImageName:   req.Repository, // temporary
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.appRepo.Create(ctx, app)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to create application", err)
	}

	s.log.Info("Application created", "name", req.Name)
	return app, nil
}

// GetApplication retrieves a single application by ID.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - appID: The numeric ID of the application to retrieve
//
// Returns:
//   - *models.Application: The application if found
//   - error: Non-nil if application not found or database operation fails
//
// Errors:
//   - "NOT_FOUND": If application with given ID does not exist
//   - "DATABASE": If database operation fails
//
// Example:
//
//	app, err := service.GetApplication(ctx, 42)
//	if err != nil {
//	    if isNotFound(err) {
//	        // Handle missing application
//	    }
//	}
func (s *ApplicationService) GetApplication(ctx context.Context, appID int) (*models.Application, error) {
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to get application", err)
	}
	if app == nil {
		return nil, errors.NewServiceError("NOT_FOUND", "Application not found")
	}
	return app, nil
}

// ListApplications retrieves a paginated list of applications.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - offset: Starting position (zero-based, defaults to 0 if negative)
//   - limit: Maximum number of results (defaults to 10 if invalid)
//
// Returns:
//   - []*models.Application: Slice of applications, empty if none exist
//   - error: Non-nil if database operation fails
//
// Pagination:
//   - Invalid limit values are clamped to [1, 100]
//   - Invalid offset values are clamped to >= 0
//
// Example:
//
//	apps, err := service.ListApplications(ctx, 0, 20)
//	for _, app := range apps {
//	    fmt.Printf("App %d: %s\n", app.ID, app.Name)
//	}
func (s *ApplicationService) ListApplications(ctx context.Context, offset, limit int) ([]*models.Application, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	apps, _, err := s.appRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to list applications", err)
	}
	return apps, nil
}

// ListApplicationsWithPagination lists applications with pagination info
func (s *ApplicationService) ListApplicationsWithPagination(ctx context.Context, offset, limit int) ([]*models.Application, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	apps, total, err := s.appRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, errors.NewServiceErrorWithCause("DATABASE", "Failed to list applications", err)
	}
	return apps, total, nil
}

// UpdateApplication updates an existing application.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - appID: The ID of the application to update
//   - req: UpdateApplicationRequest with new field values
//
// Returns:
//   - *models.Application: The updated application
//   - error: Non-nil if application not found or database operation fails
//
// Errors:
//   - "NOT_FOUND": If application with given ID does not exist
//   - "DATABASE": If database operation fails
//
// Note:
//   - UpdatedAt is automatically set to current time
//   - Empty fields in request are ignored (not updated)
//
// Example:
//
//	updated, err := service.UpdateApplication(ctx, 42, &UpdateApplicationRequest{
//	    Name: "new-name",
//	})
func (s *ApplicationService) UpdateApplication(ctx context.Context, appID int, req *handlers.UpdateApplicationRequest) (*models.Application, error) {
	// 获取现有应用
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil || app == nil {
		return nil, errors.NewServiceError("NOT_FOUND", "Application not found")
	}

	// 更新字段
	if req.Name != "" {
		app.Name = req.Name
	}
	app.UpdatedAt = time.Now()

	// 持久化
	err = s.appRepo.Update(ctx, app)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to update application", err)
	}

	s.log.Info("Application updated", "id", appID)
	return app, nil
}

// DeleteApplication removes an application from the system.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - appID: The ID of the application to delete
//
// Returns:
//   - error: Non-nil if application not found or database operation fails
//
// Errors:
//   - "NOT_FOUND": If application with given ID does not exist
//   - "DATABASE": If database operation fails
//
// Note:
//   - This is a hard delete - data cannot be recovered
//   - Consider audit logging before deletion
//   - Related release records should be cleaned up separately
//
// Example:
//
//	err := service.DeleteApplication(ctx, 42)
//	if err != nil {
//	    log.Error("Failed to delete application", "error", err)
//	}
func (s *ApplicationService) DeleteApplication(ctx context.Context, appID int) error {
	// 检查是否存在
	app, err := s.appRepo.GetByID(ctx, appID)
	if err != nil || app == nil {
		return errors.NewServiceError("NOT_FOUND", "Application not found")
	}

	// 删除
	err = s.appRepo.Delete(ctx, appID)
	if err != nil {
		return errors.NewServiceErrorWithCause("DATABASE", "Failed to delete application", err)
	}

	s.log.Info("Application deleted", "id", appID)
	return nil
}

// === 辅助方法 ===

// validateCreateRequest 验证创建请求
func (s *ApplicationService) validateCreateRequest(req *handlers.CreateApplicationRequest) error {
	if req.Name == "" || len(req.Name) > 100 {
		return errors.NewServiceError("INVALID_INPUT", "Application name must be 1-100 characters")
	}
	return nil
}
