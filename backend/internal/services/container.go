package services

import (
	"database/sql"
	"fmt"

	"built-and-deploy/internal/deployers"
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"
)

type ServiceContainer struct {
	// Service instances
	applicationService *ApplicationService
	clusterService     *ClusterService
	releaseService     *ReleaseService  // ← 添加 ReleaseService
	shellService       *ShellService

	// Repository instances
	releaseRepo                repository.ReleaseRecordRepository
	appRepo                    repository.ApplicationRepository
	clusterRepo                repository.ClusterRepository
	workloadRepo               repository.WorkloadTargetRepository
	envRepo                    *repository.EnvironmentRepository
	eventRepo                  repository.ReleaseEventRepository
	shellServerRepo            repository.ShellServerRepository
	shellCommandRepo           repository.ShellCommandRepository
	shellCommandExecutionRepo     repository.ShellCommandExecutionRepository

	deployerFact *deployers.DeployerFactory
	log          *logger.Logger
	db           *sql.DB
}

// Option is a functional option for ServiceContainer
type Option func(*ServiceContainer) error

// NewServiceContainer creates a new ServiceContainer with functional options
func NewServiceContainer(
	db *sql.DB,
	log *logger.Logger,
	opts ...Option,
) (*ServiceContainer, error) {
	c := &ServiceContainer{
		log: log,
		db:  db,
		deployerFact: deployers.NewDeployerFactory(log),
	}

	// Apply all options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
		}
	}

	// Validate that all required repositories are set
	if err := c.validate(); err != nil {
		return nil, err
	}

	// Initialize services
	// ApplicationService
	if c.appRepo != nil && c.releaseRepo != nil {
		c.applicationService = NewApplicationService(c.appRepo, c.releaseRepo, log)
	}

	// ClusterService
	if c.clusterRepo != nil {
		c.clusterService = NewClusterService(c.clusterRepo, c.deployerFact, "default-key", log)
	}

	// ReleaseService - 新增
	if c.releaseRepo != nil && c.workloadRepo != nil && c.clusterRepo != nil && c.appRepo != nil && c.eventRepo != nil {
		c.releaseService = NewReleaseService(
			c.releaseRepo, c.workloadRepo, c.clusterRepo, c.appRepo, c.eventRepo,
			c.deployerFact, log, db,
		)
	}

	// ShellService
	if c.shellServerRepo != nil && c.shellCommandRepo != nil && c.shellCommandExecutionRepo != nil {
		encryptionKey := "default-key" // This should be passed as option
		c.shellService = NewShellService(c.shellServerRepo, c.shellCommandRepo, c.shellCommandExecutionRepo, encryptionKey, log)
	}

	return c, nil
}

// validate checks that all required repositories are set
func (c *ServiceContainer) validate() error {
	if c.appRepo == nil {
		return fmt.Errorf("application repository is required")
	}
	if c.clusterRepo == nil {
		return fmt.Errorf("cluster repository is required")
	}
	if c.releaseRepo == nil {
		return fmt.Errorf("release repository is required")
	}
	if c.workloadRepo == nil {
		return fmt.Errorf("workload repository is required")
	}
	if c.eventRepo == nil {
		return fmt.Errorf("event repository is required")
	}
	if c.shellServerRepo == nil {
		return fmt.Errorf("shell server repository is required")
	}
	if c.shellCommandRepo == nil {
		return fmt.Errorf("shell command repository is required")
	}
	if c.shellCommandExecutionRepo == nil {
		return fmt.Errorf("shell command execution repository is required")
	}
	return nil
}

// Functional options for each repository

// WithApplicationRepository sets the application repository
func WithApplicationRepository(repo repository.ApplicationRepository) Option {
	return func(c *ServiceContainer) error {
		if repo == nil {
			return fmt.Errorf("application repository cannot be nil")
		}
		c.appRepo = repo
		return nil
	}
}

// WithClusterRepository sets the cluster repository
func WithClusterRepository(repo repository.ClusterRepository) Option {
	return func(c *ServiceContainer) error {
		if repo == nil {
			return fmt.Errorf("cluster repository cannot be nil")
		}
		c.clusterRepo = repo
		return nil
	}
}

// WithReleaseRepository sets the release record repository
func WithReleaseRepository(repo repository.ReleaseRecordRepository) Option {
	return func(c *ServiceContainer) error {
		if repo == nil {
			return fmt.Errorf("release record repository cannot be nil")
		}
		c.releaseRepo = repo
		return nil
	}
}

// WithWorkloadRepository sets the workload target repository
func WithWorkloadRepository(repo repository.WorkloadTargetRepository) Option {
	return func(c *ServiceContainer) error {
		if repo == nil {
			return fmt.Errorf("workload target repository cannot be nil")
		}
		c.workloadRepo = repo
		return nil
	}
}

// WithEventRepository sets the release event repository
func WithEventRepository(repo repository.ReleaseEventRepository) Option {
	return func(c *ServiceContainer) error {
		if repo == nil {
			return fmt.Errorf("event repository cannot be nil")
		}
		c.eventRepo = repo
		return nil
	}
}

// WithShellServerRepository sets the shell server repository
func WithShellServerRepository(repo repository.ShellServerRepository) Option {
	return func(c *ServiceContainer) error {
		if repo == nil {
			return fmt.Errorf("shell server repository cannot be nil")
		}
		c.shellServerRepo = repo
		return nil
	}
}

// WithShellCommandRepository sets the shell command repository
func WithShellCommandRepository(repo repository.ShellCommandRepository) Option {
	return func(c *ServiceContainer) error {
		if repo == nil {
			return fmt.Errorf("shell command repository cannot be nil")
		}
		c.shellCommandRepo = repo
		return nil
	}
}

// WithShellCommandExecutionRepository sets the shell command execution repository
func WithShellCommandExecutionRepository(repo repository.ShellCommandExecutionRepository) Option {
	return func(c *ServiceContainer) error {
		if repo == nil {
			return fmt.Errorf("shell command execution repository cannot be nil")
		}
		c.shellCommandExecutionRepo = repo
		return nil
	}
}

// WithEnvironmentRepository sets the environment repository
func WithEnvironmentRepository(repo *repository.EnvironmentRepository) Option {
	return func(c *ServiceContainer) error {
		if repo == nil {
			return fmt.Errorf("environment repository cannot be nil")
		}
		c.envRepo = repo
		return nil
	}
}

// Getter methods

func (c *ServiceContainer) Application() *ApplicationService { return c.applicationService }
func (c *ServiceContainer) Cluster() *ClusterService { return c.clusterService }
func (c *ServiceContainer) Release() *ReleaseService { return c.releaseService }  // 新增
func (c *ServiceContainer) Shell() *ShellService { return c.shellService }
func (c *ServiceContainer) Workload() *WorkloadService { 
	return NewWorkloadService(c.workloadRepo, c.appRepo, c.envRepo, c.clusterRepo, c.log)
}
func (c *ServiceContainer) ApplicationRepo() repository.ApplicationRepository { return c.appRepo }
func (c *ServiceContainer) ClusterRepo() repository.ClusterRepository { return c.clusterRepo }
func (c *ServiceContainer) ReleaseRepo() repository.ReleaseRecordRepository { return c.releaseRepo }
func (c *ServiceContainer) WorkloadRepo() repository.WorkloadTargetRepository { return c.workloadRepo }
func (c *ServiceContainer) EnvironmentRepo() *repository.EnvironmentRepository { return c.envRepo }
func (c *ServiceContainer) EventRepo() repository.ReleaseEventRepository { return c.eventRepo }
func (c *ServiceContainer) ShellServerRepo() repository.ShellServerRepository { return c.shellServerRepo }
func (c *ServiceContainer) ShellCommandRepo() repository.ShellCommandRepository { return c.shellCommandRepo }
func (c *ServiceContainer) ShellCommandExecutionRepo() repository.ShellCommandExecutionRepository { return c.shellCommandExecutionRepo }
func (c *ServiceContainer) Logger() *logger.Logger { return c.log }
func (c *ServiceContainer) DB() *sql.DB { return c.db }
func (c *ServiceContainer) DeployerFactory() *deployers.DeployerFactory { return c.deployerFact }
