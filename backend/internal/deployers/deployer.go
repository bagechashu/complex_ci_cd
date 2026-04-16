package deployers

import (
	"context"

	"built-and-deploy/internal/models"
)

// DeployStrategy defines the interface for deployment strategies
type DeployStrategy interface {
	// Deploy deploys the application
	Deploy(ctx context.Context, info *models.DeploymentInfo, image string) error

	// Validate validates the deployment configuration
	Validate(ctx context.Context, info *models.DeploymentInfo) error

	// Rollback rollbacks the deployment to previous version
	Rollback(ctx context.Context, info *models.DeploymentInfo, previousImage string) error

	// GetStatus returns the current deployment status
	GetStatus(ctx context.Context, info *models.DeploymentInfo) (string, error)

	// HealthCheck checks the health of the deployed application
	HealthCheck(ctx context.Context, info *models.DeploymentInfo) (bool, error)

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
