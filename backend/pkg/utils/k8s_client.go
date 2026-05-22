package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClientManager manages K8s client connections with caching and connection validation
type K8sClientManager struct {
	clientCache   map[string]kubernetes.Interface
	clientCacheMu sync.RWMutex
}

// NewK8sClientManager creates a new K8s client manager
func NewK8sClientManager() *K8sClientManager {
	return &K8sClientManager{
		clientCache: make(map[string]kubernetes.Interface),
	}
}

// BuildRESTConfig creates a REST config from kubeconfig content and optional K8s version
// Parameters:
//   - kubeconfigContent: Plain text kubeconfig (must be decrypted before calling)
//   - k8sVersion: Optional K8s version for timeout adjustment
//
// Returns:
//   - *rest.Config: The REST configuration ready for K8s client usage
//   - error: Any error during config creation
func BuildRESTConfig(kubeconfigContent string, k8sVersion *string) (*rest.Config, error) {
	if kubeconfigContent == "" {
		return nil, fmt.Errorf("kubeconfig content is empty")
	}

	// Load REST config from kubeconfig content
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	// Set timeout based on Kubernetes version
	timeout := getTimeoutForK8sVersion(k8sVersion)
	restConfig.Timeout = timeout

	return restConfig, nil
}

// ValidateK8sConnection attempts to connect to K8s API server and verify connectivity
// Parameters:
//   - kubeconfigContent: Plain text kubeconfig (must be decrypted before calling)
//   - k8sVersion: Optional K8s version for timeout adjustment
//
// Returns:
//   - error: Non-nil if connection fails, nil if connection is successful
//
// This method:
//   - Builds REST config from kubeconfig
//   - Creates a discovery client (lightweight)
//   - Queries server version to verify authentication and connectivity
func ValidateK8sConnection(kubeconfigContent string, k8sVersion *string) error {
	// Build REST config
	restConfig, err := BuildRESTConfig(kubeconfigContent, k8sVersion)
	if err != nil {
		return fmt.Errorf("failed to build REST config: %w", err)
	}

	// Create discovery client (lightweight and doesn't require full clientset)
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create K8s client: %w", err)
	}

	// Query server version to verify connection (includes authentication)
	_, err = discoveryClient.ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to query K8s API: %w", err)
	}

	return nil
}

// GetOrCreateK8sClient retrieves a cached K8s client or creates a new one
// Parameters:
//   - cacheKey: Unique key for caching (e.g., "clusterID_clusterName")
//   - kubeconfigContent: Plain text kubeconfig (must be decrypted before calling)
//   - k8sVersion: Optional K8s version for timeout adjustment
//
// Returns:
//   - kubernetes.Interface: The K8s clientset
//   - error: Non-nil if client creation fails
//
// Notes:
//   - Clients are cached to avoid recreating them repeatedly
//   - Use this method when you need the full K8s clientset for operations like Deployment updates
func (m *K8sClientManager) GetOrCreateK8sClient(cacheKey string, kubeconfigContent string, k8sVersion *string) (kubernetes.Interface, error) {
	if cacheKey == "" {
		return nil, fmt.Errorf("cache key cannot be empty")
	}

	// Check cache first
	m.clientCacheMu.RLock()
	if client, exists := m.clientCache[cacheKey]; exists {
		m.clientCacheMu.RUnlock()
		return client, nil
	}
	m.clientCacheMu.RUnlock()

	// Create new client
	restConfig, err := BuildRESTConfig(kubeconfigContent, k8sVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to build REST config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s clientset: %w", err)
	}

	// Cache the client
	m.clientCacheMu.Lock()
	m.clientCache[cacheKey] = clientset
	m.clientCacheMu.Unlock()

	return clientset, nil
}

// ClearClient removes a cached client by key
func (m *K8sClientManager) ClearClient(cacheKey string) {
	m.clientCacheMu.Lock()
	delete(m.clientCache, cacheKey)
	m.clientCacheMu.Unlock()
}

// ClearAllClients removes all cached clients
func (m *K8sClientManager) ClearAllClients() {
	m.clientCacheMu.Lock()
	m.clientCache = make(map[string]kubernetes.Interface)
	m.clientCacheMu.Unlock()
}

// getTimeoutForK8sVersion returns the appropriate timeout based on K8s version
// Different K8s versions may have different API response times
func getTimeoutForK8sVersion(k8sVersion *string) time.Duration {
	defaultTimeout := 15 * time.Second

	if k8sVersion == nil || *k8sVersion == "" {
		return defaultTimeout
	}

	// For older K8s versions, use longer timeout
	if strings.HasPrefix(*k8sVersion, "1.19") || strings.HasPrefix(*k8sVersion, "1.20") {
		return 25 * time.Second
	} else if strings.HasPrefix(*k8sVersion, "1.21") || strings.HasPrefix(*k8sVersion, "1.22") {
		return 20 * time.Second
	}

	return defaultTimeout
}
