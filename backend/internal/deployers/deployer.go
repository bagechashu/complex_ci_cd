package deployers

import (
	"built-and-deploy/internal/models"
	"context"
)

// DeployStrategy defines the interface for workload strategies
type DeployStrategy interface {
	// Deploy deploys the application
	Deploy(ctx context.Context, info *models.WorkloadInfo, image string) error

	// Validate validates the workload configuration
	Validate(ctx context.Context, info *models.WorkloadInfo) error

	// Rollback rollbacks the workload to previous version
	Rollback(ctx context.Context, info *models.WorkloadInfo, previousImage string) error

	// GetStatus returns the current workload status
	GetStatus(ctx context.Context, info *models.WorkloadInfo) (string, error)

	// HealthCheck checks the health of the deployed application
	HealthCheck(ctx context.Context, info *models.WorkloadInfo) (bool, error)

	// Type returns the deployer type
	Type() string
}

// BaseDeployer provides common functionality
type BaseDeployer struct {
	name string
}

func (b *BaseDeployer) Type() string {
	return b.name
}
