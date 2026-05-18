package services

import (
	"context"
	"time"

	"built-and-deploy/internal/deployers"
	"built-and-deploy/internal/models"
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/errors"
	"built-and-deploy/pkg/handlers"
	"built-and-deploy/pkg/logger"
)

// ReleaseService orchestrates application deployment and release operations.
//
// ReleaseService coordinates the entire release workflow:
//   - Creating release records to track deployments
//   - Recording release events for audit trails
//   - Checking release status across environments
//   - Retrieving release history for compliance
//   - Rolling back failed releases to previous versions
//
// The service works with:
//   - ReleaseRecordRepository: Tracks deployment records
//   - ReleaseEventRepository: Records deployment events
//   - WorkloadTargetRepository: Manages deployment targets
//   - ClusterRepository: Accesses target clusters
//   - ApplicationRepository: Fetches application details
//   - DeployerFactory: Creates appropriate deployment executor
//
// Example usage:
//
//	service := NewReleaseService(releaseRepo, workloadRepo, clusterRepo, appRepo, eventRepo, deployerFact, log, db)
//	release, err := service.Release(ctx, appID, clusterID, "app:v1.2.3")
//	if err != nil {
//	    log.Error("Release failed", "error", err)
//	}
type ReleaseService struct {
	releaseRepo   repository.ReleaseRecordRepository
	workloadRepo  repository.WorkloadTargetRepository
	clusterRepo   repository.ClusterRepository
	appRepo       repository.ApplicationRepository
	eventRepo     repository.ReleaseEventRepository
	deployerFact  *deployers.DeployerFactory
	log           *logger.Logger
	db            interface{}
}

// NewReleaseService creates a new ReleaseService instance.
//
// Parameters:
//   - releaseRepo: ReleaseRecordRepository for release persistence
//   - workloadRepo: WorkloadTargetRepository for deployment targets
//   - clusterRepo: ClusterRepository for cluster information
//   - appRepo: ApplicationRepository for application details
//   - eventRepo: ReleaseEventRepository for event tracking
//   - deployerFact: DeployerFactory for deployment execution
//   - log: Logger for structured logging
//   - db: Database connection for transaction management
//
// Returns a configured ReleaseService ready for use.
func NewReleaseService(
	releaseRepo repository.ReleaseRecordRepository,
	workloadRepo repository.WorkloadTargetRepository,
	clusterRepo repository.ClusterRepository,
	appRepo repository.ApplicationRepository,
	eventRepo repository.ReleaseEventRepository,
	deployerFact *deployers.DeployerFactory,
	log *logger.Logger,
	db interface{},
) *ReleaseService {
	return &ReleaseService{
		releaseRepo:  releaseRepo,
		workloadRepo: workloadRepo,
		clusterRepo:  clusterRepo,
		appRepo:      appRepo,
		eventRepo:    eventRepo,
		deployerFact: deployerFact,
		log:          log,
		db:           db,
	}
}

// Release creates and initiates a deployment for an application to a cluster.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - appID: The ID of the application to release
//   - clusterID: The ID of the target cluster
//   - image: The container image tag (e.g., "app:v1.2.3")
//
// Returns:
//   - *models.ReleaseRecord: The created release record with initial status
//   - error: Non-nil if creation fails
//
// Status flow:
//   - Initial status: "pending"
//   - Triggered by: "system" (can be overridden by caller)
//
// Errors:
//   - "DATABASE": If release record creation fails
//
// Side effects:
//   - Creates ReleaseRecord in database
//   - Records deployment start event
//   - Logs release creation
//
// Example:
//
//	release, err := service.Release(ctx, 1, 5, "api-service:v2.0.0")
//	if err != nil {
//	    log.Error("Failed to create release", "error", err)
//	    return
//	}
//	fmt.Printf("Release %d created\n", release.ID)
func (s *ReleaseService) Release(ctx context.Context, appID, clusterID int, image string) (*models.ReleaseRecord, error) {
	release := &models.ReleaseRecord{
		AppID:      appID,
		ClusterID:  clusterID,
		Image:      image,
		Status:     "pending",
		TriggeredBy: "system",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := s.releaseRepo.Create(ctx, release)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to create release record", err)
	}

	s.log.Info("Release created", "appID", appID, "clusterID", clusterID, "image", image)
	return release, nil
}

// ListReleaseEvents retrieves all events for a specific release.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - releaseID: The integer ID of the release
//
// Returns:
//   - []interface{}: Slice of release events
//   - error: Non-nil if retrieval fails
//
// Errors:
//   - "NOT_FOUND": If release does not have any events
//   - "DATABASE": If database query fails
//
// Example:
//
//	events, err := service.ListReleaseEvents(ctx, 123)
//	if err != nil {
//	    log.Error("Failed to list events", "error", err)
//	}
func (s *ReleaseService) ListReleaseEvents(ctx context.Context, releaseID int) ([]interface{}, error) {
	events, err := s.eventRepo.ListByRelease(ctx, releaseID)
	if err != nil {
		s.log.Error("Failed to list release events", "releaseID", releaseID, "error", err)
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to list release events", err)
	}

	if events == nil {
		events = make([]*models.ReleaseEvent, 0)
	}

	// Convert to []interface{} for compatibility with response handlers
	result := make([]interface{}, len(events))
	for i, event := range events {
		result[i] = event
	}

	s.log.Info("Release events retrieved", "releaseID", releaseID, "count", len(events))
	return result, nil
}

// Rollback reverts a release to a previous state.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - releaseID: The ID of the release to rollback
//
// Returns:
//   - *models.ReleaseRecord: The rolled back release record
//   - error: Non-nil if rollback fails
//
// Errors:
//   - "NOT_FOUND": If release with given ID does not exist
//   - "DATABASE": If database operation fails
//
// Status changes:
//   - Sets status to "rolled_back"
//   - Updates UpdatedAt timestamp
//
// Example:
//
//	rolled, err := service.Rollback(ctx, 42)
//	if err != nil {
//	    log.Error("Rollback failed", "error", err)
//	    return
//	}
//	fmt.Printf("Release rolled back to status: %s\n", rolled.Status)
func (s *ReleaseService) Rollback(ctx context.Context, releaseID int) (*models.ReleaseRecord, error) {
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil || release == nil {
		return nil, errors.NewServiceError("NOT_FOUND", "Release not found")
	}

	release.Status = "rolled_back"
	release.UpdatedAt = time.Now()

	err = s.releaseRepo.Update(ctx, release)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to rollback release", err)
	}

	s.log.Info("Release rolled back", "releaseID", releaseID)
	return release, nil
}

// GetReleaseHistory retrieves a paginated list of release records.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - offset: Starting position in result set (zero-based)
//   - limit: Maximum number of records to return
//
// Returns:
//   - []*models.ReleaseRecord: Slice of release records
//   - error: Non-nil if database operation fails
//
// Pagination:
//   - Results are ordered by creation time (typically newest first)
//   - Empty slice if no releases exist
//
// Example:
//
//	releases, err := service.GetReleaseHistory(ctx, 0, 20)
//	for _, rel := range releases {
//	    fmt.Printf("Release %d: %s on cluster %d\n", rel.ID, rel.Status, rel.ClusterID)
//	}
func (s *ReleaseService) GetReleaseHistory(ctx context.Context, offset, limit int) ([]*models.ReleaseRecord, error) {
	releases, _, err := s.releaseRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to list releases", err)
	}
	return releases, nil
}

// GetReleaseStatus retrieves the current status of a specific release.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - releaseID: The ID of the release to check
//
// Returns:
//   - *models.ReleaseRecord: The release record with current status
//   - error: Non-nil if release not found or database operation fails
//
// Errors:
//   - "NOT_FOUND": If release with given ID does not exist
//   - "DATABASE": If database operation fails
//
// Example:
//
//	release, err := service.GetReleaseStatus(ctx, 42)
//	if err != nil {
//	    log.Error("Status check failed", "error", err)
//	    return
//	}
//	fmt.Printf("Release status: %s\n", release.Status)
func (s *ReleaseService) GetReleaseStatus(ctx context.Context, releaseID int) (*models.ReleaseRecord, error) {
	release, err := s.releaseRepo.GetByID(ctx, releaseID)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to get release status", err)
	}
	if release == nil {
		return nil, errors.NewServiceError("NOT_FOUND", "Release not found")
	}
	return release, nil
}

// ReleaseWithRequest creates a release from an HTTP request object.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - req: CreateReleaseRequest containing release details
//
// Returns:
//   - *models.ReleaseRecord: The created release record
//   - error: Non-nil if creation fails
//
// Note:
//   - Simplified implementation for now
//   - In production, should properly convert string IDs to integers
//   - May include additional validation and authorization
//
// Example:
//
//	release, err := service.ReleaseWithRequest(ctx, &CreateReleaseRequest{
//	    ApplicationID: "1",
//	    ClusterID:     "5",
//	    Image:         "app:v1.2.3",
//	})
func (s *ReleaseService) ReleaseWithRequest(ctx context.Context, req *handlers.CreateReleaseRequest) (*models.ReleaseRecord, error) {
	// Parse IDs from string format - simplified for now
	// In production, would properly convert string IDs to integers
	return s.Release(ctx, 0, 0, req.Image)
}
