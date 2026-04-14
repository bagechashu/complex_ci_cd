package deployers

import (
	"fmt"

	"github.com/op/release-control/pkg/logger"
)

// DeployerFactory creates deployers based on type
type DeployerFactory struct {
	log *logger.Logger
}

func NewDeployerFactory(log *logger.Logger) *DeployerFactory {
	return &DeployerFactory{log: log}
}

// CreateDeployer creates a deployer based on cluster type
func (f *DeployerFactory) CreateDeployer(clusterType string) (DeployStrategy, error) {
	switch clusterType {
	case "kubernetes":
		return NewK8sDeployer(f.log), nil
	case "salt":
		// return NewSaltDeployer(f.log), nil
		return nil, fmt.Errorf("salt deployer not implemented yet")
	case "ansible":
		// return NewAnsibleDeployer(f.log), nil
		return nil, fmt.Errorf("ansible deployer not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported cluster type: %s", clusterType)
	}
}
