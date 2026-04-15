package services

import (
	"context"
	"fmt"
	"time"

	"github.com/op/release-control/internal/deployers"
	"github.com/op/release-control/internal/models"
	"github.com/op/release-control/internal/repository"
	"github.com/op/release-control/pkg/logger"
)

type ReleaseService struct {
	releaseRepo     *repository.ReleaseRecordRepository
	appRepo         *repository.ApplicationRepository
	envRepo         *repository.EnvironmentRepository
	clusterRepo     *repository.ClusterRepository
	targetRepo      *repository.DeploymentTargetRepository
	deployerFactory *deployers.DeployerFactory
	log             *logger.Logger
}

func NewReleaseService(
	releaseRepo *repository.ReleaseRecordRepository,
	appRepo *repository.ApplicationRepository,
	envRepo *repository.EnvironmentRepository,
	clusterRepo *repository.ClusterRepository,
	targetRepo *repository.DeploymentTargetRepository,
	factory *deployers.DeployerFactory,
	log *logger.Logger,
) *ReleaseService {
	return &ReleaseService{
		releaseRepo:     releaseRepo,
		appRepo:         appRepo,
		envRepo:         envRepo,
		clusterRepo:     clusterRepo,
		targetRepo:      targetRepo,
		deployerFactory: factory,
		log:             log,
	}
}

// Release performs the release process
func (s *ReleaseService) Release(ctx context.Context, req *models.ReleaseRequest) (*models.ReleaseRecord, error) {
	// Validate request
	if req.AppID <= 0 || req.EnvID <= 0 || req.Image == "" {
		return nil, fmt.Errorf("invalid release request")
	}

	// Get application
	app, err := s.appRepo.GetByID(req.AppID)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	// Get environment
	env, err := s.envRepo.GetByID(req.EnvID)
	if err != nil {
		return nil, fmt.Errorf("environment not found: %w", err)
	}

	s.log.Info("Starting release for app=%s env=%s image=%s", app.Name, env.Name, req.Image)

	// Create release record
	release := &models.ReleaseRecord{
		AppID:       req.AppID,
		EnvID:       req.EnvID,
		Image:       req.Image,
		Status:      models.StatusPending,
		TriggeredBy: req.User,
	}

	release, err = s.releaseRepo.Create(release)
	if err != nil {
		return nil, fmt.Errorf("failed to create release record: %w", err)
	}

	// Launch async deployment process
	go s.deployAsync(context.Background(), release, app, env)

	return release, nil
}

// deployAsync executes deployment asynchronously
func (s *ReleaseService) deployAsync(ctx context.Context, release *models.ReleaseRecord, app *models.Application, env *models.Environment) {
	s.log.Info("Starting async deployment for release %d", release.ID)

	// Record start event
	s.releaseRepo.CreateEvent(&models.ReleaseEvent{
		ReleaseID: release.ID,
		Type:      "started",
		Message:   fmt.Sprintf("Release %d started", release.ID),
		CreatedAt: time.Now(),
	})

	// Update status
	release.Status = models.StatusValidating
	now := time.Now()
	release.StartedAt = &now
	s.releaseRepo.Update(release)

	// Step 1: Get deployment targets for this app and environment
	targets, err := s.targetRepo.ListByAppAndEnv(app.ID, env.ID)
	if err != nil {
		s.log.Error("Failed to get deployment targets: %v", err)
		s.recordDeploymentFailure(release, "Failed to retrieve deployment targets")
		return
	}

	if len(targets) == 0 {
		s.log.Error("No deployment targets configured for app %s in env %s", app.Name, env.Name)
		s.recordDeploymentFailure(release, "No deployment targets configured")
		return
	}

	// Step 2: Validate all targets
	for _, target := range targets {
		cluster, err := s.clusterRepo.GetByID(target.ClusterID)
		if err != nil {
			s.log.Error("Failed to get cluster %d: %v", target.ClusterID, err)
			s.recordDeploymentFailure(release, fmt.Sprintf("Cluster %d not found", target.ClusterID))
			return
		}

		deployer, err := s.deployerFactory.CreateDeployer(cluster.Type)
		if err != nil {
			s.log.Error("Failed to create deployer for cluster type %s: %v", cluster.Type, err)
			s.recordDeploymentFailure(release, fmt.Sprintf("Unsupported cluster type: %s", cluster.Type))
			return
		}

		deploymentInfo := &models.DeploymentInfo{
			App:    app,
			Target: target,
			Cluster: cluster,
		}

		if err := deployer.Validate(ctx, deploymentInfo); err != nil {
			s.log.Error("Validation failed for target %d: %v", target.ID, err)
			s.recordDeploymentFailure(release, fmt.Sprintf("Validation failed: %v", err))
			return
		}

		s.releaseRepo.CreateEvent(&models.ReleaseEvent{
			ReleaseID: release.ID,
			Type:      "validated",
			Message:   fmt.Sprintf("Validated deployment to %s", cluster.Name),
		})
	}

	// Step 3: Execute deployment on all targets
	release.Status = models.StatusDeploying
	s.releaseRepo.Update(release)

	deploymentErrors := make([]string, 0)
	for _, target := range targets {
		cluster, _ := s.clusterRepo.GetByID(target.ClusterID)
		deployer, _ := s.deployerFactory.CreateDeployer(cluster.Type)

		deploymentInfo := &models.DeploymentInfo{
			App:    app,
			Target: target,
			Cluster: cluster,
		}

		if err := deployer.Deploy(ctx, deploymentInfo, release.Image); err != nil {
			s.log.Error("Deployment failed for target %d on cluster %s: %v", target.ID, cluster.Name, err)
			deploymentErrors = append(deploymentErrors, err.Error())
			continue
		}

		s.releaseRepo.CreateEvent(&models.ReleaseEvent{
			ReleaseID: release.ID,
			Type:      "deployed",
			Message:   fmt.Sprintf("Successfully deployed to %s", cluster.Name),
		})
	}

	// If there were errors, mark as failed
	if len(deploymentErrors) > 0 {
		s.recordDeploymentFailure(release, fmt.Sprintf("Deployment failed on some targets: %v", deploymentErrors))
		return
	}

	// Step 4: Health check (using validating status)
	release.Status = models.StatusValidating
	s.releaseRepo.Update(release)

	for _, target := range targets {
		cluster, _ := s.clusterRepo.GetByID(target.ClusterID)
		deployer, _ := s.deployerFactory.CreateDeployer(cluster.Type)

		deploymentInfo := &models.DeploymentInfo{
			App:    app,
			Target: target,
			Cluster: cluster,
		}

		healthy, err := deployer.HealthCheck(ctx, deploymentInfo)
		if err != nil || !healthy {
			s.log.Info("Health check failed for target %d on cluster %s", target.ID, cluster.Name)
			continue
		}

		s.releaseRepo.CreateEvent(&models.ReleaseEvent{
			ReleaseID: release.ID,
			Type:      "health_check",
			Message:   fmt.Sprintf("Health check passed on %s", cluster.Name),
		})
	}

	// Step 5: Mark as success
	release.Status = models.StatusSuccess
	completedAt := time.Now()
	release.CompletedAt = &completedAt
	s.releaseRepo.Update(release)

	s.releaseRepo.CreateEvent(&models.ReleaseEvent{
		ReleaseID: release.ID,
		Type:      "success",
		Message:   fmt.Sprintf("Release %d completed successfully", release.ID),
	})

	s.log.Info("Release %d completed successfully", release.ID)
}

// recordDeploymentFailure records a deployment failure and updates the release status
func (s *ReleaseService) recordDeploymentFailure(release *models.ReleaseRecord, message string) {
	release.Status = models.StatusFailed
	completedAt := time.Now()
	release.CompletedAt = &completedAt
	s.releaseRepo.Update(release)

	s.releaseRepo.CreateEvent(&models.ReleaseEvent{
		ReleaseID: release.ID,
		Type:      "failed",
		Message:   message,
	})

	s.log.Error("Release %d failed: %s", release.ID, message)
}

// GetReleaseStatus returns the current release status
func (s *ReleaseService) GetReleaseStatus(releaseID int) (*models.ReleaseRecord, error) {
	return s.releaseRepo.GetByID(releaseID)
}

// GetReleaseEvents returns events for a release
func (s *ReleaseService) GetReleaseEvents(releaseID int) ([]*models.ReleaseEvent, error) {
	return s.releaseRepo.GetEvents(releaseID)
}

// ListReleases returns a list of releases
func (s *ReleaseService) ListReleases(limit int, offset int) ([]*models.ReleaseRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.releaseRepo.List(limit, offset)
}

// Rollback rolls back a release
func (s *ReleaseService) Rollback(ctx context.Context, releaseID int) (*models.ReleaseRecord, error) {
	// Get the current release
	currentRelease, err := s.releaseRepo.GetByID(releaseID)
	if err != nil {
		return nil, fmt.Errorf("release not found: %w", err)
	}

	s.log.Info("Rolling back release %d", releaseID)

	// Get the application and environment
	app, err := s.appRepo.GetByID(currentRelease.AppID)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	env, err := s.envRepo.GetByID(currentRelease.EnvID)
	if err != nil {
		return nil, fmt.Errorf("environment not found: %w", err)
	}

	// Find the previous successful release
	releases, err := s.releaseRepo.List(100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}

	var previousRelease *models.ReleaseRecord
	for _, r := range releases {
		// Skip current release and failed releases
		if r.ID == releaseID || r.Status != models.StatusSuccess {
			continue
		}
		// Find the most recent successful release before this one
		if r.AppID == currentRelease.AppID && r.EnvID == currentRelease.EnvID {
			if previousRelease == nil || r.CreatedAt.After(previousRelease.CreatedAt) {
				previousRelease = r
			}
		}
	}

	if previousRelease == nil {
		return nil, fmt.Errorf("no previous successful release found for rollback")
	}

	s.log.Info("Rolling back to release %d with image %s", previousRelease.ID, previousRelease.Image)

	// Create a new rollback release record
	rollbackRelease := &models.ReleaseRecord{
		AppID:         currentRelease.AppID,
		EnvID:         currentRelease.EnvID,
		Image:         previousRelease.Image,
		PreviousImage: currentRelease.Image,
		Status:        models.StatusPending,
		TriggeredBy:   "rollback",
	}

	rollbackRelease, err = s.releaseRepo.Create(rollbackRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to create rollback release: %w", err)
	}

	// Record rollback event on original release
	s.releaseRepo.CreateEvent(&models.ReleaseEvent{
		ReleaseID: currentRelease.ID,
		Type:      "rollback_initiated",
		Message:   fmt.Sprintf("Rollback initiated to release %d", rollbackRelease.ID),
	})

	// Launch async rollback deployment
	go s.deployAsync(context.Background(), rollbackRelease, app, env)

	return rollbackRelease, nil
}
