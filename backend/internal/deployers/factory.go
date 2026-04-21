package deployers

import (
	"fmt"

	"built-and-deploy/pkg/logger"
)

// DeployerFactory creates deployers based on type
type DeployerFactory struct {
	log *logger.Logger
}

func NewDeployerFactory(log *logger.Logger) *DeployerFactory {
	return &DeployerFactory{log: log}
}

// CreateDeployer creates a deployer based on cluster type
// Currently supports Kubernetes deployer; other deployment methods (Salt, Ansible, etc.)
// are handled through SSH shell execution via ShellService
func (f *DeployerFactory) CreateDeployer(clusterType string) (DeployStrategy, error) {
	switch clusterType {
	case "kubernetes":
		return NewK8sDeployer(f.log), nil
	default:
		return nil, fmt.Errorf("unsupported cluster type: %s (only 'kubernetes' is supported; use shell execution for other methods)", clusterType)
	}
}
