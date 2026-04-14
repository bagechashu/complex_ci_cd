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

	// TODO: Continue with validation and deployment steps
	// For now, just mark as success
	release.Status = models.StatusSuccess
	completedAt := time.Now()
	release.CompletedAt = &completedAt
	s.releaseRepo.Update(release)

	s.releaseRepo.CreateEvent(&models.ReleaseEvent{
		ReleaseID: release.ID,
		Type:      "success",
		Message:   fmt.Sprintf("Release %d completed successfully", release.ID),
	})
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
	return s.releaseRepo.List(limit, offset)
}

// Rollback rolls back a release
func (s *ReleaseService) Rollback(ctx context.Context, releaseID int) error {
	_, err := s.releaseRepo.GetByID(releaseID)
	if err != nil {
		return err
	}

	s.log.Info("Rolling back release %d", releaseID)

	// TODO: Implement rollback logic
	return fmt.Errorf("rollback not implemented yet")
}
