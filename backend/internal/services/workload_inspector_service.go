package services

import (
	"context"
	"fmt"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodInfo represents a Kubernetes Pod's information
type PodInfo struct {
	Name             string    `json:"name"`
	Namespace        string    `json:"namespace"`
	Status           string    `json:"status"`           // Running, Pending, Failed, etc.
	RestartCount     int32     `json:"restart_count"`
	CreatedAt        time.Time `json:"created_at"`
	ReadyCondition   *string   `json:"ready_condition,omitempty"`   // "True" or "False"
	ContainerCount   int       `json:"container_count"`
	ReadyContainers  int       `json:"ready_containers"`
	Image            string    `json:"image"`
	NodeName         string    `json:"node_name,omitempty"`
}

// WorkloadInspectorService provides methods to inspect and query Kubernetes workloads
// This service focuses on READ operations (querying), not WRITE operations (deploying)
//
// Usage:
//
//	service := NewWorkloadInspectorService(clusterRepo, encryptKey, log)
//	pods, err := service.GetWorkloadPods(ctx, workloadTarget)
type WorkloadInspectorService struct {
	clusterRepo   repository.ClusterRepository
	encryptKey    string
	clientManager *utils.K8sClientManager
	log           *logger.Logger
}

// NewWorkloadInspectorService creates a new WorkloadInspectorService instance
func NewWorkloadInspectorService(
	clusterRepo repository.ClusterRepository,
	encryptKey string,
	log *logger.Logger,
) *WorkloadInspectorService {
	return &WorkloadInspectorService{
		clusterRepo:   clusterRepo,
		encryptKey:    encryptKey,
		clientManager: utils.NewK8sClientManager(),
		log:           log,
	}
}

// GetWorkloadPods retrieves all pods for a given workload (Deployment/StatefulSet/DaemonSet)
//
// Parameters:
//   - ctx: Context for cancellation and deadline
//   - cluster: The Cluster containing kubeconfig (must be decrypted before calling)
//   - workloadTarget: WorkloadTarget containing workload info
//
// Returns:
//   - []*PodInfo: List of pods with their status information
//   - error: Non-nil if retrieval fails
func (s *WorkloadInspectorService) GetWorkloadPods(ctx context.Context, cluster *models.Cluster, workloadTarget *models.WorkloadTarget) ([]*PodInfo, error) {
	if cluster == nil || workloadTarget == nil {
		return nil, fmt.Errorf("invalid cluster or workload target")
	}

	namespace := workloadTarget.K8sNamespace
	workloadName := workloadTarget.K8sWorkload
	workloadType := workloadTarget.WorkloadType

	s.log.Info("Getting workload pods",
		"cluster", cluster.Name,
		"namespace", namespace,
		"workload", workloadName,
		"type", workloadType)

	// Get or create K8s client
	clientset, err := s.getOrCreateClient(cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s client: %w", err)
	}

	// Get label selector from workload
	labelSelector, err := s.getWorkloadLabelSelector(ctx, clientset, namespace, workloadName, workloadType)
	if err != nil {
		s.log.Warn("Failed to get workload label selector",
			"namespace", namespace,
			"workload", workloadName,
			"error", err)
		// Fallback: use workload name as label
		labelSelector = fmt.Sprintf("app=%s", workloadName)
	}

	s.log.Info("Using label selector for pod query", "selector", labelSelector)

	// Query pods using label selector
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Convert to PodInfo
	var pods []*PodInfo
	for _, pod := range podList.Items {
		podInfo := s.convertPodToPodInfo(&pod)
		pods = append(pods, podInfo)
	}

	s.log.Info("Retrieved workload pods",
		"cluster", cluster.Name,
		"workload", workloadName,
		"podCount", len(pods))

	return pods, nil
}

// GetPodInfo retrieves detailed information about a specific pod
func (s *WorkloadInspectorService) GetPodInfo(ctx context.Context, cluster *models.Cluster, namespace, podName string) (*PodInfo, error) {
	if cluster == nil || namespace == "" || podName == "" {
		return nil, fmt.Errorf("missing required parameters")
	}

	clientset, err := s.getOrCreateClient(cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s client: %w", err)
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return s.convertPodToPodInfo(pod), nil
}

// ==================== Private Helper Methods ====================

// getOrCreateClient retrieves or creates a K8s client for the cluster
func (s *WorkloadInspectorService) getOrCreateClient(cluster *models.Cluster) (kubernetes.Interface, error) {
	if cluster.Kubeconfig == nil || *cluster.Kubeconfig == "" {
		return nil, fmt.Errorf("kubeconfig is empty for cluster %s", cluster.Name)
	}

	// Decrypt kubeconfig
	decryptedKubeconfig, err := utils.DecryptAES(*cluster.Kubeconfig, s.encryptKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt kubeconfig: %w", err)
	}

	// Get or create client from manager
	cacheKey := fmt.Sprintf("%d_%s", cluster.ID, cluster.Name)
	clientset, err := s.clientManager.GetOrCreateK8sClient(cacheKey, decryptedKubeconfig, cluster.KubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s client: %w", err)
	}

	return clientset, nil
}

// getWorkloadLabelSelector extracts label selector from the workload spec
func (s *WorkloadInspectorService) getWorkloadLabelSelector(ctx context.Context, clientset kubernetes.Interface, namespace, workloadName, workloadType string) (string, error) {
	switch workloadType {
	case "Deployment":
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, workloadName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		return metav1.FormatLabelSelector(deployment.Spec.Selector), nil

	case "StatefulSet":
		sts, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, workloadName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		return metav1.FormatLabelSelector(sts.Spec.Selector), nil

	case "DaemonSet":
		ds, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, workloadName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		return metav1.FormatLabelSelector(ds.Spec.Selector), nil

	default:
		return "", fmt.Errorf("unsupported workload type: %s", workloadType)
	}
}

// convertPodToPodInfo converts a K8s Pod to PodInfo
func (s *WorkloadInspectorService) convertPodToPodInfo(pod *corev1.Pod) *PodInfo {
	podInfo := &PodInfo{
		Name:           pod.Name,
		Namespace:      pod.Namespace,
		Status:         string(pod.Status.Phase),
		CreatedAt:      pod.CreationTimestamp.Time,
		ContainerCount: len(pod.Spec.Containers),
		NodeName:       pod.Spec.NodeName,
	}

	// Get restart count and ready status
	if len(pod.Status.ContainerStatuses) > 0 {
		podInfo.RestartCount = pod.Status.ContainerStatuses[0].RestartCount
		for _, status := range pod.Status.ContainerStatuses {
			if status.Ready {
				podInfo.ReadyContainers++
			}
		}
	}

	// Get image from first container
	if len(pod.Spec.Containers) > 0 {
		podInfo.Image = pod.Spec.Containers[0].Image
	}

	// Get Ready condition status
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			readyStatus := string(condition.Status)
			podInfo.ReadyCondition = &readyStatus
			break
		}
	}

	return podInfo
}
