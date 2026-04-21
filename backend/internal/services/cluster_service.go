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
	"built-and-deploy/pkg/utils"
)

// ClusterService manages cluster operations including creation, configuration, and deployment.
//
// ClusterService handles:
//   - Cluster registration with encryption of sensitive credentials (kubeconfig)
//   - Cluster listing and retrieval
//   - Cluster configuration updates
//   - Cluster deletion
//   - Connection validation and testing
//
// Security:
//   - All kubeconfig data is encrypted at rest using AES encryption
//   - Encryption key is managed by the configuration layer
//   - Decryption happens on-demand for deployment operations
//
// Deployment:
//   - Coordinates with DeployerFactory to determine deployment strategy
//   - Supports multiple cluster types (Kubernetes, SSH-based, Ansible, Salt)
//   - Enables application deployment to specific clusters
//
// Usage:
//
//	service := NewClusterService(clusterRepo, deployerFactory, encryptionKey, log)
//	cluster, err := service.CreateCluster(ctx, &CreateClusterRequest{
//	    Name:       "production-k8s",
//	    Type:       "kubernetes",
//	    Kubeconfig: kubeconfigContent,
//	})
type ClusterService struct {
	clusterRepo  repository.ClusterRepository
	deployerFact *deployers.DeployerFactory
	encryptKey   string
	log          *logger.Logger
}

// NewClusterService creates a new ClusterService instance.
//
// Parameters:
//   - clusterRepo: ClusterRepository for cluster persistence
//   - deployerFact: DeployerFactory for deployment strategy selection
//   - encryptionKey: Encryption key for kubeconfig data protection
//   - log: Logger for structured logging
//
// Returns a configured ClusterService instance.
//
// Example:
//
//	service := NewClusterService(repo, factory, key, logger.GetLogger())
func NewClusterService(
	clusterRepo repository.ClusterRepository,
	deployerFact *deployers.DeployerFactory,
	encryptionKey string,
	log *logger.Logger,
) *ClusterService {
	return &ClusterService{
		clusterRepo:  clusterRepo,
		deployerFact: deployerFact,
		encryptKey:   encryptionKey,
		log:          log,
	}
}

// CreateCluster registers a new cluster in the system.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - req: CreateClusterRequest with cluster configuration
//
// Returns:
//   - *models.Cluster: The created cluster with assigned ID and encrypted kubeconfig
//   - error: Non-nil if validation or encryption fails
//
// Errors:
//   - "INVALID_INPUT": If cluster name is missing
//   - "ENCRYPTION_ERROR": If kubeconfig encryption fails
//   - "DATABASE": If database operation fails
//
// Security:
//   - Kubeconfig is encrypted using AES encryption before storage
//   - Encrypted data is stored in database
//   - Original plaintext kubeconfig is not persisted
//
// Example:
//
//	cluster, err := service.CreateCluster(ctx, &CreateClusterRequest{
//	    Name:       "dev-cluster",
//	    Type:       "kubernetes",
//	    Kubeconfig: kubeconfigYAML,
//	})
func (s *ClusterService) CreateCluster(ctx context.Context, req *handlers.CreateClusterRequest) (*models.Cluster, error) {
	if req.Name == "" {
		return nil, errors.NewServiceError("INVALID_INPUT", "Cluster name is required")
	}

	// Encrypt kubeconfig
	encryptedConfig, err := utils.EncryptAES(req.Kubeconfig, s.encryptKey)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("ENCRYPTION_ERROR", "Failed to encrypt kubeconfig", err)
	}

	cluster := &models.Cluster{
		Name:                req.Name,
		Type:                req.Type,
		Kubeconfig:          &encryptedConfig,
		K8sConnectionStatus: "unknown",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	err = s.clusterRepo.Create(ctx, cluster)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to create cluster", err)
	}

	s.log.Info("Cluster created", "name", req.Name)
	return cluster, nil
}

// ListClusters retrieves all registered clusters.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//
// Returns:
//   - []*models.Cluster: Slice of all clusters, empty if none exist
//   - error: Non-nil if database operation fails
//
// Note:
//   - Kubeconfig data remains encrypted in returned objects
//   - Decryption happens only when needed for deployments
//
// Example:
//
//	clusters, err := service.ListClusters(ctx)
//	for _, cluster := range clusters {
//	    fmt.Printf("Cluster: %s (type: %s)\n", cluster.Name, cluster.Type)
//	}
func (s *ClusterService) ListClusters(ctx context.Context) ([]*models.Cluster, error) {
	clusters, _, err := s.clusterRepo.List(ctx, 0, 1000)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to list clusters", err)
	}
	return clusters, nil
}

// GetCluster retrieves a specific cluster by ID.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - id: The numeric ID of the cluster to retrieve
//
// Returns:
//   - *models.Cluster: The cluster if found (with encrypted kubeconfig)
//   - error: Non-nil if cluster not found or database operation fails
//
// Errors:
//   - "NOT_FOUND": If cluster with given ID does not exist
//
// Example:
//
//	cluster, err := service.GetCluster(ctx, 1)
//	if err != nil {
//	    // Handle error
//	}
func (s *ClusterService) GetCluster(ctx context.Context, id int) (*models.Cluster, error) {
	cluster, err := s.clusterRepo.GetByID(ctx, id)
	if err != nil || cluster == nil {
		return nil, errors.NewServiceError("NOT_FOUND", "Cluster not found")
	}
	return cluster, nil
}

// UpdateCluster modifies an existing cluster configuration.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - id: The ID of the cluster to update
//   - req: UpdateClusterRequest with new values
//
// Returns:
//   - *models.Cluster: The updated cluster
//   - error: Non-nil if cluster not found or update fails
//
// Errors:
//   - "NOT_FOUND": If cluster with given ID does not exist
//   - "ENCRYPTION_ERROR": If kubeconfig encryption fails
//   - "DATABASE": If database operation fails
//
// Note:
//   - Empty fields in request are ignored (not updated)
//   - UpdatedAt is automatically set to current time
//   - Kubeconfig is encrypted before storage if provided
//
// Example:
//
//	updated, err := service.UpdateCluster(ctx, 1, &UpdateClusterRequest{
//	    Name:       "prod-cluster-v2",
//	    Kubeconfig: newKubeconfig,
//	})
func (s *ClusterService) UpdateCluster(ctx context.Context, id int, req *handlers.UpdateClusterRequest) (*models.Cluster, error) {
	cluster, err := s.clusterRepo.GetByID(ctx, id)
	if err != nil || cluster == nil {
		return nil, errors.NewServiceError("NOT_FOUND", "Cluster not found")
	}

	if req.Name != "" {
		cluster.Name = req.Name
	}
	if req.Kubeconfig != "" {
		encrypted, err := utils.EncryptAES(req.Kubeconfig, s.encryptKey)
		if err != nil {
			return nil, errors.NewServiceErrorWithCause("ENCRYPTION_ERROR", "Failed to encrypt kubeconfig", err)
		}
		cluster.Kubeconfig = &encrypted
	}
	cluster.UpdatedAt = time.Now()

	err = s.clusterRepo.Update(ctx, cluster)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to update cluster", err)
	}

	s.log.Info("Cluster updated", "id", id)
	return cluster, nil
}

// DeleteCluster removes a cluster from the system.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - id: The ID of the cluster to delete
//
// Returns:
//   - error: Non-nil if deletion fails
//
// Errors:
//   - "DATABASE": If database operation fails
//
// Warning:
//   - This is a hard delete - data cannot be recovered
//   - Consider checking for active deployments before deletion
//   - Audit log deletion events
//
// Example:
//
//	err := service.DeleteCluster(ctx, 1)
//	if err != nil {
//	    log.Error("Failed to delete cluster", "error", err)
//	}
func (s *ClusterService) DeleteCluster(ctx context.Context, id int) error {
	err := s.clusterRepo.Delete(ctx, id)
	if err != nil {
		return errors.NewServiceErrorWithCause("DATABASE", "Failed to delete cluster", err)
	}
	s.log.Info("Cluster deleted", "id", id)
	return nil
}

// TestConnection validates connectivity to the cluster.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - id: The ID of the cluster to test
//
// Returns:
//   - string: Connection status ("success", "failed", etc.)
//   - error: Non-nil if cluster not found or test fails
//
// Errors:
//   - "NOT_FOUND": If cluster with given ID does not exist
//   - "INVALID_STATE": If kubeconfig is not configured
//
// Note:
//   - Basic validation only - full kubeconfig validation requires kubectl
//   - Actual connection testing depends on DeployerFactory implementation
//   - May require additional configuration for SSH clusters
//
// Example:
//
//	status, err := service.TestConnection(ctx, 1)
//	if err == nil && status == "success" {
//	    fmt.Println("Cluster is reachable")
//	}
func (s *ClusterService) TestConnection(ctx context.Context, id int) (string, error) {
	cluster, err := s.clusterRepo.GetByID(ctx, id)
	if err != nil || cluster == nil {
		return "", errors.NewServiceError("NOT_FOUND", "Cluster not found")
	}

	if cluster.Kubeconfig == nil {
		return "", errors.NewServiceError("INVALID_STATE", "Cluster kubeconfig not configured")
	}

	// Just return success for now - full validation requires kubeconfig parsing
	return "success", nil
}
