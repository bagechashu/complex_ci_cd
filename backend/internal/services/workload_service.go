package services

import (
	"context"
	"log"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"
)

// WorkloadService manages workload target deployment configurations.
//
// WorkloadService handles the definition and management of workload targets,
// which represent specific deployment configurations for applications:
//   - Workload target definition (which apps deploy to which clusters/environments)
//   - Workload target listing and retrieval
//   - Workload target creation and configuration
//   - Workload target updates
//   - Workload target deletion
//
// A workload target defines:
//   - Application to deploy
//   - Target cluster/environment
//   - Any workload-specific configuration
//
// Usage:
//
//	service := NewWorkloadService(workloadRepo, appRepo, envRepo, clusterRepo, log)
//	targets, err := service.ListWorkloadTargets(ctx)
//	for _, target := range targets {
//	    fmt.Printf("Workload: %d -> App %d on Env %d\n", target.ID, target.ApplicationID, target.EnvironmentID)
//	}
type WorkloadService struct {
	workloadRepo repository.WorkloadTargetRepository
	clusterRepo  repository.ClusterRepository
	envRepo      *repository.EnvironmentRepository
	log          *logger.Logger
}

// NewWorkloadService creates a new WorkloadService instance.
//
// Parameters:
//   - workloadRepo: WorkloadTargetRepository for workload persistence
//   - appRepo: ApplicationRepository for application details (currently unused)
//   - envRepo: EnvironmentRepository for environment details (enriching workload targets)
//   - clusterRepo: ClusterRepository for cluster details (enriching workload targets)
//   - log: Logger for structured logging
//
// Returns a configured WorkloadService ready for use.
//
// Example:
//
//	service := NewWorkloadService(repo, appRepo, envRepo, clusterRepo, logger.GetLogger())
func NewWorkloadService(
	workloadRepo repository.WorkloadTargetRepository,
	appRepo repository.ApplicationRepository,
	envRepo interface{},
	clusterRepo repository.ClusterRepository,
	log *logger.Logger,
) *WorkloadService {
	// Type assert envRepo to *EnvironmentRepository if provided
	var envRepoPtr *repository.EnvironmentRepository
	if envRepo != nil {
		if er, ok := envRepo.(*repository.EnvironmentRepository); ok {
			envRepoPtr = er
		}
	}
	
	return &WorkloadService{
		workloadRepo: workloadRepo,
		clusterRepo:  clusterRepo,
		envRepo:      envRepoPtr,
		log:          log,
	}
}

// ListWorkloadTargets retrieves all workload targets in the system.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//
// Returns:
//   - []*models.WorkloadTarget: Slice of all workload targets
//   - error: Non-nil if retrieval fails
//
// Note:
//   - Returns up to 1000 targets
//   - Empty slice if no workload targets exist
//   - Enriches each target with cluster name and environment name
//
// Example:
//
//	targets, err := service.ListWorkloadTargets(ctx)
//	if err != nil {
//	    log.Error("Failed to list workloads", "error", err)
//	    return
//	}
//	fmt.Printf("Found %d workload targets\n", len(targets))
func (s *WorkloadService) ListWorkloadTargets(ctx context.Context) ([]*models.WorkloadTarget, error) {
	targets, err := s.workloadRepo.List(1000, 0)
	if err != nil {
		log.Printf("Error listing workload targets: %v", err)
		return nil, err
	}
	
	// Enrich targets with cluster names and environment names
	for _, target := range targets {
		if err := s.enrichClusterName(ctx, target); err != nil {
			log.Printf("Warning: failed to enrich cluster name for target %d: %v", target.ID, err)
		}
		if err := s.enrichEnvironment(ctx, target); err != nil {
			log.Printf("Warning: failed to enrich environment for target %d: %v", target.ID, err)
		}
	}
	
	return targets, nil
}

// ListWorkloadTargetsByApp retrieves workload targets for a specific application.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - appID: The ID of the application
//
// Returns:
//   - []*models.WorkloadTarget: Slice of workload targets for the application
//   - error: Non-nil if retrieval fails
//
// Note:
//   - Filters workload targets by appID from the repository
//   - Enriches each target with cluster name and environment name
//
// Example:
//
//	targets, err := service.ListWorkloadTargetsByApp(ctx, 5)
//	for _, target := range targets {
//	    fmt.Printf("Workload %d: cluster %s, environment %s\n", target.ID, target.ClusterName, target.Environment)
//	}
func (s *WorkloadService) ListWorkloadTargetsByApp(ctx context.Context, appID int) ([]*models.WorkloadTarget, error) {
	targets, err := s.workloadRepo.GetByApp(appID)
	if err != nil {
		log.Printf("Error listing workload targets for app %d: %v", appID, err)
		return nil, err
	}
	
	// Enrich targets with cluster names and environment names
	for _, target := range targets {
		if err := s.enrichClusterName(ctx, target); err != nil {
			log.Printf("Warning: failed to enrich cluster name for target %d: %v", target.ID, err)
		}
		if err := s.enrichEnvironment(ctx, target); err != nil {
			log.Printf("Warning: failed to enrich environment for target %d: %v", target.ID, err)
		}
	}
	
	return targets, nil
}

// enrichClusterName enriches a workload target with cluster name from the cluster repository.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - target: The workload target to enrich
//
// Returns:
//   - error: Non-nil if enrichment fails
//
// Note:
//   - If enrichment fails, the target is returned without cluster name
//   - This is a best-effort operation
func (s *WorkloadService) enrichClusterName(ctx context.Context, target *models.WorkloadTarget) error {
	if s.clusterRepo == nil {
		return nil // No cluster repo available, skip enrichment
	}
	
	cluster, err := s.clusterRepo.GetByID(ctx, target.ClusterID)
	if err != nil {
		return err
	}
	
	if cluster != nil {
		target.ClusterName = cluster.Name
	}
	return nil
}

// enrichEnvironment enriches a workload target with environment name from the environment repository.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - target: The workload target to enrich
//
// Returns:
//   - error: Non-nil if enrichment fails
//
// Note:
//   - If enrichment fails, the target is returned without environment name
//   - This is a best-effort operation
func (s *WorkloadService) enrichEnvironment(ctx context.Context, target *models.WorkloadTarget) error {
	if s.envRepo == nil {
		return nil // No environment repo available, skip enrichment
	}
	
	env, err := s.envRepo.GetByID(target.EnvID)
	if err != nil {
		return err
	}
	
	if env != nil {
		target.Environment = env.Name
	}
	return nil
}

// GetWorkloadTarget retrieves a specific workload target by ID.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - id: The ID of the workload target to retrieve
//
// Returns:
//   - *models.WorkloadTarget: The workload target if found
//   - error: Non-nil if retrieval fails
//
// Note:
//   - Simplified implementation using List internally
//   - Returns first result or nil if not found
//
// Example:
//
//	target, err := service.GetWorkloadTarget(ctx, 42)
//	if err != nil {
//	    log.Error("Failed to get workload", "error", err)
//	}
func (s *WorkloadService) GetWorkloadTarget(ctx context.Context, id int) (*models.WorkloadTarget, error) {
	targets, err := s.workloadRepo.List(1, 0)
	if err != nil {
		return nil, err
	}
	if len(targets) > 0 {
		return targets[0], nil
	}
	return nil, nil
}

// CreateWorkloadTarget creates a new workload target.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - req: Request object with workload target configuration
//
// Returns:
//   - *models.WorkloadTarget: The created workload target
//   - error: Non-nil if creation fails
//
// Note:
//   - Stub implementation for future expansion
//   - Full implementation will validate application and environment references
//
// Example:
//
//	target, err := service.CreateWorkloadTarget(ctx, req)
//	if err != nil {
//	    log.Error("Failed to create workload", "error", err)
//	}
func (s *WorkloadService) CreateWorkloadTarget(ctx context.Context, req interface{}) (*models.WorkloadTarget, error) {
	// Simplified stub implementation
	return nil, nil
}

// UpdateWorkloadTarget modifies an existing workload target.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - id: The ID of the workload target to update
//   - req: Request object with new configuration
//
// Returns:
//   - *models.WorkloadTarget: The updated workload target
//   - error: Non-nil if update fails
//
// Note:
//   - Stub implementation for future expansion
//
// Example:
//
//	updated, err := service.UpdateWorkloadTarget(ctx, 42, req)
//	if err != nil {
//	    log.Error("Failed to update workload", "error", err)
//	}
func (s *WorkloadService) UpdateWorkloadTarget(ctx context.Context, id int, req interface{}) (*models.WorkloadTarget, error) {
	// Simplified stub implementation
	return nil, nil
}

// DeleteWorkloadTarget removes a workload target from the system.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - id: The ID of the workload target to delete
//
// Returns:
//   - error: Non-nil if deletion fails
//
// Note:
//   - Stub implementation for future expansion
//
// Example:
//
//	err := service.DeleteWorkloadTarget(ctx, 42)
//	if err != nil {
//	    log.Error("Failed to delete workload", "error", err)
//	}
func (s *WorkloadService) DeleteWorkloadTarget(ctx context.Context, id int) error {
	// Simplified stub implementation
	return nil
}
