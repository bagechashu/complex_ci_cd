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

	var labels *string
	if req.Labels != "" {
		labels = &req.Labels
	}

	var kubernetesVersion *string
	if req.KubernetesVersion != "" {
		kubernetesVersion = &req.KubernetesVersion
	}

	cluster := &models.Cluster{
		Name:                req.Name,
		Type:                req.Type,
		Environment:         req.Environment,
		RegistryPrefix:      req.RegistryPrefix,
		Labels:              labels,
		KubernetesVersion:   kubernetesVersion,
		Kubeconfig:          &encryptedConfig,
		K8sConnectionStatus: "unknown",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	err = s.clusterRepo.Create(ctx, cluster)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to create cluster", err)
	}

	s.log.Info("Cluster created", "name", req.Name, "id", cluster.ID)
	
	// Asynchronously test the connection
	s.UpdateConnectionStatus(cluster.ID)
	
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

	kubeConfigUpdated := false
	if req.Name != "" {
		cluster.Name = req.Name
	}
	if req.Kubeconfig != "" {
		encrypted, err := utils.EncryptAES(req.Kubeconfig, s.encryptKey)
		if err != nil {
			return nil, errors.NewServiceErrorWithCause("ENCRYPTION_ERROR", "Failed to encrypt kubeconfig", err)
		}
		cluster.Kubeconfig = &encrypted
		kubeConfigUpdated = true
		// Reset connection status when kubeconfig is updated
		cluster.K8sConnectionStatus = "unknown"
	}
	if req.KubernetesVersion != "" {
		cluster.KubernetesVersion = &req.KubernetesVersion
	}
	cluster.UpdatedAt = time.Now()

	err = s.clusterRepo.Update(ctx, cluster)
	if err != nil {
		return nil, errors.NewServiceErrorWithCause("DATABASE", "Failed to update cluster", err)
	}

	s.log.Info("Cluster updated", "id", id)
	
	// Asynchronously test the connection if kubeconfig was updated
	if kubeConfigUpdated {
		s.UpdateConnectionStatus(id)
	}
	
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

// TestConnectionResult 连接测试结果，包含状态和错误详情
type TestConnectionResult struct {
	Status  string `json:"status"`  // "connected" or "disconnected"
	Message string `json:"message"` // 错误详情或成功消息
}

// TestConnection validates connectivity to the cluster by attempting to connect to K8s API.
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - id: The ID of the cluster to test
//
// Returns:
//   - *TestConnectionResult: Connection status with detailed message
//   - error: Non-nil if cluster not found or other errors
//
// This method:
//   - Retrieves the cluster from the repository
//   - Decrypts the kubeconfig
//   - Attempts to create a K8s client and connect to the API server
//   - Returns detailed error information on failure
//
// Example:
//
//	result, err := service.TestConnection(ctx, 1)
//	if err == nil && result.Status == "connected" {
//	    fmt.Println("Cluster is reachable")
//	}
func (s *ClusterService) TestConnection(ctx context.Context, id int) (*TestConnectionResult, error) {
	cluster, err := s.clusterRepo.GetByID(ctx, id)
	if err != nil || cluster == nil {
		return &TestConnectionResult{
			Status:  "disconnected",
			Message: "Cluster not found",
		}, errors.NewServiceError("NOT_FOUND", "Cluster not found")
	}

	if cluster.Kubeconfig == nil || *cluster.Kubeconfig == "" {
		return &TestConnectionResult{
			Status:  "disconnected",
			Message: "Cluster kubeconfig not configured",
		}, errors.NewServiceError("INVALID_STATE", "Cluster kubeconfig not configured")
	}

	// Decrypt kubeconfig
	decryptedKubeconfig, err := utils.DecryptAES(*cluster.Kubeconfig, s.encryptKey)
	if err != nil {
		s.log.Warn("Failed to decrypt kubeconfig", "error", err.Error())
		return &TestConnectionResult{
			Status:  "disconnected",
			Message: "Failed to decrypt kubeconfig",
		}, nil
	}

	// Try to connect to K8s API using shared utils function
	err = utils.ValidateK8sConnection(decryptedKubeconfig, cluster.KubernetesVersion)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to validate K8s connection: %v", err)
		s.log.Warn(errMsg)
		return &TestConnectionResult{
			Status:  "disconnected",
			Message: errMsg,
		}, nil
	}

	s.log.Info("K8s API connection successful")
	return &TestConnectionResult{
		Status:  "connected",
		Message: "Connection successful",
	}, nil
}

// UpdateConnectionStatus updates the connection status of a cluster asynchronously.
func (s *ClusterService) UpdateConnectionStatus(clusterId int) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, _ := s.TestConnection(ctx, clusterId)
		if result == nil {
			return
		}
		cluster, err := s.clusterRepo.GetByID(ctx, clusterId)
		if err != nil || cluster == nil {
			return
		}

		cluster.K8sConnectionStatus = result.Status
		cluster.UpdatedAt = time.Now()
		err = s.clusterRepo.Update(ctx, cluster)
		if err != nil {
			s.log.Error("Failed to update cluster connection status", "error", err.Error())
		} else {
			s.log.Info("Updated cluster connection status", "clusterId", clusterId, "status", result.Status, "message", result.Message)
		}
	}()
}
