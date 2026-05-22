package services

import (
	"context"
	"fmt"
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
//	release, err := service.Release(ctx, appID, envID, clusterID, "app:v1.2.3")
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
//   - envID: The ID of the environment to release to
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
//	release, err := service.Release(ctx, 1, 2, 5, "api-service:v2.0.0")
//	if err != nil {
//	    log.Error("Failed to create release", "error", err)
//	    return
//	}
//	fmt.Printf("Release %d created\n", release.ID)
func (s *ReleaseService) Release(ctx context.Context, appID, envID, clusterID int, image string) (*models.ReleaseRecord, error) {
	now := time.Now()
	release := &models.ReleaseRecord{
		AppID:       appID,
		EnvID:       envID,
		ClusterID:   clusterID,
		Image:       image,
		Status:      "pending",
		TriggeredBy: "system",
		StartedAt:   &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.releaseRepo.Create(ctx, release)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to create release record", err)
	}

	s.log.Info("Release created", "appID", appID, "envID", envID, "clusterID", clusterID, "image", image, "releaseID", release.ID)

	// Execute deployment asynchronously
	go func() {
		s.executeRelease(context.Background(), release)
	}()

	return release, nil
}

// executeRelease performs the actual deployment to the cluster
func (s *ReleaseService) executeRelease(ctx context.Context, release *models.ReleaseRecord) {
	// Record deployment started event
	event := &models.ReleaseEvent{
		ReleaseID: release.ID,
		Type:      "deployment_started",
		Message:   "Deployment to cluster started",
		CreatedAt: time.Now(),
	}
	s.eventRepo.Create(ctx, event)

	// Get workload target to find the K8s deployment details
	targets, err := s.workloadRepo.GetByApp(release.AppID)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get workload targets: %v", err)
		s.log.Error(errMsg, "releaseID", release.ID, "error", err)
		s.recordErrorEvent(ctx, release.ID, "workload_fetch_error", errMsg)
		release.Status = "failed"
		release.ErrorMsg = &errMsg
		s.releaseRepo.Update(ctx, release)
		return
	}
	if len(targets) == 0 {
		errMsg := fmt.Sprintf("No workload targets found for application (app_id=%d)", release.AppID)
		s.log.Error(errMsg, "releaseID", release.ID, "appID", release.AppID)
		s.recordErrorEvent(ctx, release.ID, "workload_not_found", errMsg)
		release.Status = "failed"
		release.ErrorMsg = &errMsg
		s.releaseRepo.Update(ctx, release)
		return
	}

	// Find the target for this cluster
	var targetWorkload *models.WorkloadTarget
	for _, t := range targets {
		if t.ClusterID == release.ClusterID {
			targetWorkload = t
			break
		}
	}

	if targetWorkload == nil {
		errMsg := fmt.Sprintf("No workload target found for cluster (cluster_id=%d). Available clusters for this app: %v", 
			release.ClusterID, getClusterIDs(targets))
		s.log.Error(errMsg, "releaseID", release.ID, "clusterID", release.ClusterID)
		s.recordErrorEvent(ctx, release.ID, "cluster_mapping_not_found", errMsg)
		release.Status = "failed"
		release.ErrorMsg = &errMsg
		s.releaseRepo.Update(ctx, release)
		return
	}

	// Log workload target configuration
	s.log.Info("Workload target found", 
		"releaseID", release.ID, 
		"namespace", targetWorkload.K8sNamespace,
		"workload", targetWorkload.K8sWorkload,
		"workloadType", targetWorkload.WorkloadType,
		"containerName", targetWorkload.ContainerName)

	// Get cluster info for deployment
	cluster, err := s.clusterRepo.GetByID(ctx, release.ClusterID)
	if err != nil || cluster == nil {
		errMsg := fmt.Sprintf("Failed to get cluster info (cluster_id=%d): %v", release.ClusterID, err)
		s.log.Error(errMsg, "releaseID", release.ID, "clusterID", release.ClusterID)
		s.recordErrorEvent(ctx, release.ID, "cluster_not_found", errMsg)
		release.Status = "failed"
		release.ErrorMsg = &errMsg
		s.releaseRepo.Update(ctx, release)
		return
	}

	s.log.Info("Executing release", "releaseID", release.ID, "cluster", cluster.Name, "image", release.Image)
	s.recordEvent(ctx, release.ID, "cluster_info_retrieved", fmt.Sprintf("Cluster: %s, Type: %s", cluster.Name, cluster.Type))

	// Get appropriate deployer for the cluster type
	deployer, err := s.deployerFact.CreateDeployer(cluster.Type)
	if err != nil || deployer == nil {
		errMsg := fmt.Sprintf("No deployer available for cluster type: %s", cluster.Type)
		s.log.Error(errMsg, "releaseID", release.ID, "clusterType", cluster.Type)
		s.recordErrorEvent(ctx, release.ID, "deployer_not_found", errMsg)
		release.Status = "failed"
		release.ErrorMsg = &errMsg
		s.releaseRepo.Update(ctx, release)
		return
	}

	// Get app info
	app, err := s.appRepo.GetByID(ctx, release.AppID)
	if err != nil || app == nil {
		errMsg := fmt.Sprintf("Failed to get application info (app_id=%d): %v", release.AppID, err)
		s.log.Error(errMsg, "appID", release.AppID)
		s.recordErrorEvent(ctx, release.ID, "app_not_found", errMsg)
		release.Status = "failed"
		release.ErrorMsg = &errMsg
		s.releaseRepo.Update(ctx, release)
		return
	}

	// Build workload info for deployment
	workloadInfo := &models.WorkloadInfo{
		Target:  targetWorkload,
		App:     app,
		Cluster: cluster,
		Env:     nil, // Env is optional
	}

	s.recordEvent(ctx, release.ID, "deployment_starting", 
		fmt.Sprintf("Deploying %s to namespace %s", release.Image, targetWorkload.K8sNamespace))

	// Execute deployment
	err = deployer.Deploy(ctx, workloadInfo, release.Image)

	// Update release status based on deployment result
	now := time.Now()
	if err != nil {
		s.log.Error("Deployment failed", "releaseID", release.ID, "error", err)
		release.Status = "failed"
		errMsg := fmt.Sprintf("Deployment error: %v", err)
		release.ErrorMsg = &errMsg
		release.CompletedAt = &now
		s.recordErrorEvent(ctx, release.ID, "deployment_failed", errMsg)
	} else {
		s.log.Info("Deployment succeeded", "releaseID", release.ID)
		release.Status = "success"
		release.CompletedAt = &now

		// Record success event
		successEvent := &models.ReleaseEvent{
			ReleaseID: release.ID,
			Type:      "deployment_success",
			Message:   "Deployment to cluster completed successfully",
			CreatedAt: time.Now(),
		}
		s.eventRepo.Create(ctx, successEvent)
	}

	// Update release record with final status
	release.UpdatedAt = time.Now()
	err = s.releaseRepo.Update(ctx, release)
	if err != nil {
		s.log.Error("Failed to update release status", "releaseID", release.ID, "error", err)
	}
}

// recordEvent records an informational event for a release
func (s *ReleaseService) recordEvent(ctx context.Context, releaseID int, eventType, message string) {
	event := &models.ReleaseEvent{
		ReleaseID: releaseID,
		Type:      eventType,
		Message:   message,
		CreatedAt: time.Now(),
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.log.Error("Failed to record event", "releaseID", releaseID, "eventType", eventType, "error", err)
	}
}

// recordErrorEvent records an error event for a release
func (s *ReleaseService) recordErrorEvent(ctx context.Context, releaseID int, eventType, message string) {
	event := &models.ReleaseEvent{
		ReleaseID: releaseID,
		Type:      eventType,
		Message:   fmt.Sprintf("[ERROR] %s", message),
		CreatedAt: time.Now(),
	}
	if err := s.eventRepo.Create(ctx, event); err != nil {
		s.log.Error("Failed to record error event", "releaseID", releaseID, "eventType", eventType, "error", err)
	}
}

// getClusterIDs extracts cluster IDs from workload targets for logging
func getClusterIDs(targets []*models.WorkloadTarget) []int {
	var clusterIDs []int
	for _, t := range targets {
		clusterIDs = append(clusterIDs, t.ClusterID)
	}
	return clusterIDs
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

// GetReleaseHistoryByApp retrieves a paginated list of release records for a specific application.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - appID: Application ID to filter releases
//   - offset: Starting position in result set (zero-based)
//   - limit: Maximum number of records to return
//
// Returns:
//   - []*models.ReleaseRecord: Slice of release records for the application
//   - error: Non-nil if database operation fails
//
// Example:
//
//	releases, err := service.GetReleaseHistoryByApp(ctx, 1, 0, 20)
//	for _, rel := range releases {
//	    fmt.Printf("Release %d: %s\n", rel.ID, rel.Status)
//	}
func (s *ReleaseService) GetReleaseHistoryByApp(ctx context.Context, appID, offset, limit int) ([]*models.ReleaseRecord, error) {
	releases, _, err := s.releaseRepo.ListByApp(ctx, appID, offset, limit)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to list releases for application", err)
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
//   - Converts request IDs to the Release method format
//   - Delegates to Release() for actual release creation
//
// Example:
//
//	release, err := service.ReleaseWithRequest(ctx, &CreateReleaseRequest{
//	    AppID:     1,
//	    EnvID:     1,
//	    ClusterID: 5,
//	    Image:     "app:v1.2.3",
//	    User:      "admin",
//	})
func (s *ReleaseService) ReleaseWithRequest(ctx context.Context, req *handlers.CreateReleaseRequest) (*models.ReleaseRecord, error) {
	// Use the provided app, env, and cluster IDs
	return s.Release(ctx, req.AppID, req.EnvID, req.ClusterID, req.Image)
}
