package deployers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// EventBus 事件总线接口（导入自services包的定义）
type EventBus interface {
	Publish(topic string, event *models.ReleaseEvent)
}

// ReleaseEventRepository 事件仓储接口
type ReleaseEventRepository interface {
	Create(ctx context.Context, event *models.ReleaseEvent) error
}

// K8s 部署超时设置
const (
	// DeploymentRolloutTimeout 是 Deployment 发布完成的超时时间
	DeploymentRolloutTimeout = 2 * time.Minute
	// StatefulSetRolloutTimeout 是 StatefulSet 发布完成的超时时间
	StatefulSetRolloutTimeout = 2 * time.Minute
	// DaemonSetRolloutTimeout 是 DaemonSet 发布完成的超时时间
	DaemonSetRolloutTimeout = 2 * time.Minute
)

// K8sDeployer implements workload strategy for Kubernetes
type K8sDeployer struct {
	BaseDeployer
	log           *logger.Logger
	encryptionKey string
	clientManager *utils.K8sClientManager
	
	// Phase 2: 事件记录（可选）
	releaseID  int
	eventBus   EventBus
	eventRepo  ReleaseEventRepository
}

// NewK8sDeployer creates a new Kubernetes deployer with encryption key for kubeconfig decryption
func NewK8sDeployer(log *logger.Logger, encryptionKey string) *K8sDeployer {
	return &K8sDeployer{
		BaseDeployer:  BaseDeployer{name: "kubernetes"},
		log:           log,
		encryptionKey: encryptionKey,
		clientManager: utils.NewK8sClientManager(),
	}
}

// Phase 2: SetEventRecorder 设置事件记录器（可选）
func (d *K8sDeployer) SetEventRecorder(releaseID int, eventBus EventBus) {
	d.releaseID = releaseID
	d.eventBus = eventBus
}

// Phase 2: SetEventRecorder with repo
func (d *K8sDeployer) SetEventRecorderWithRepo(releaseID int, eventBus EventBus, eventRepo ReleaseEventRepository) {
	d.releaseID = releaseID
	d.eventBus = eventBus
	d.eventRepo = eventRepo
}

// Phase 2: recordDeploymentEvent 记录部署事件到EventBus和数据库
func (d *K8sDeployer) recordDeploymentEvent(eventType, message string) {
	if d.eventBus == nil || d.releaseID == 0 {
		d.log.Warn("Cannot record deployment event: eventBus is nil or releaseID is 0", 
			"eventType", eventType, "message", message)
		return // 不记录事件
	}
	event := &models.ReleaseEvent{
		ReleaseID: d.releaseID,
		Type:      eventType,
		Message:   message,
		CreatedAt: time.Now(),
	}
	
	// 1. 发送到EventBus供实时SSE订阅者使用
	d.log.Debug("Publishing deployment event via EventBus", 
		"releaseID", d.releaseID, "eventType", eventType)
	d.eventBus.Publish(fmt.Sprintf("release:%d", d.releaseID), event)
	
	// 2. 如果有eventRepo，也保存到数据库以确保持久化
	if d.eventRepo != nil {
		// 使用context.Background因为Deploy已经有超时控制
		if err := d.eventRepo.Create(context.Background(), event); err != nil {
			d.log.Warn("Failed to persist event to database", "releaseID", d.releaseID, 
				"eventType", eventType, "error", err)
		}
	}
}

// Deploy deploys an application to Kubernetes by updating the container image
func (d *K8sDeployer) Deploy(ctx context.Context, info *models.WorkloadInfo, image string) error {
	d.log.Info("Starting K8s deployment", "app", info.App.Name, "namespace", info.Target.K8sNamespace, 
		"workload", info.Target.K8sWorkload, "cluster", info.Cluster.Name, "image", image)

	// Validate inputs
	if info.App == nil || info.Target == nil || info.Cluster == nil {
		errMsg := "invalid workload info: missing required fields"
		d.log.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	if image == "" {
		errMsg := "image cannot be empty"
		d.log.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	if info.Cluster.Kubeconfig == nil || *info.Cluster.Kubeconfig == "" {
		errMsg := fmt.Sprintf("kubeconfig not configured for cluster: %s", info.Cluster.Name)
		d.log.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	// Get or create K8s client
	clientset, err := d.getOrCreateClient(info.Cluster)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create K8s client for cluster %s: %v", info.Cluster.Name, err)
		d.log.Error(errMsg, "error", err)
		return fmt.Errorf(errMsg)
	}

	namespace := info.Target.K8sNamespace
	workloadName := info.Target.K8sWorkload
	containerName := info.Target.ContainerName
	workloadType := info.Target.WorkloadType

	// Determine which resource type to update
	switch workloadType {
	case "Deployment":
		return d.deployDeployment(ctx, clientset, namespace, workloadName, containerName, image)
	case "StatefulSet":
		return d.deployStatefulSet(ctx, clientset, namespace, workloadName, containerName, image)
	case "DaemonSet":
		return d.deployDaemonSet(ctx, clientset, namespace, workloadName, containerName, image)
	default:
		return fmt.Errorf("unsupported workload type: %s", workloadType)
	}
}

// Validate validates the workload configuration before deployment
func (d *K8sDeployer) Validate(ctx context.Context, info *models.WorkloadInfo) error {
	d.log.Info("Validating K8s workload configuration", "app", info.App.Name, "cluster", info.Cluster.Name)

	// Basic validation
	if info == nil || info.App == nil || info.Target == nil || info.Cluster == nil {
		return fmt.Errorf("invalid workload info: missing required fields")
	}

	if info.Target.K8sNamespace == "" {
		return fmt.Errorf("kubernetes namespace not configured")
	}

	if info.Target.K8sWorkload == "" {
		return fmt.Errorf("kubernetes workload name not configured")
	}

	if info.Cluster.Kubeconfig == nil || *info.Cluster.Kubeconfig == "" {
		return fmt.Errorf("kubeconfig not configured for cluster: %s", info.Cluster.Name)
	}

	// Create K8s client
	clientset, err := d.getOrCreateClient(info.Cluster)
	if err != nil {
		return fmt.Errorf("failed to create K8s client: %w", err)
	}

	namespace := info.Target.K8sNamespace
	workloadName := info.Target.K8sWorkload

	// Check if namespace exists
	_, err = clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("namespace %q not found or not accessible: %w", namespace, err)
	}

	// Check if workload exists
	switch info.Target.WorkloadType {
	case "Deployment":
		_, err = clientset.AppsV1().Deployments(namespace).Get(ctx, workloadName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("deployment %q not found in namespace %q: %w", workloadName, namespace, err)
		}
	case "StatefulSet":
		_, err = clientset.AppsV1().StatefulSets(namespace).Get(ctx, workloadName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("statefulset %q not found in namespace %q: %w", workloadName, namespace, err)
		}
	case "DaemonSet":
		_, err = clientset.AppsV1().DaemonSets(namespace).Get(ctx, workloadName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("daemonset %q not found in namespace %q: %w", workloadName, namespace, err)
		}
	default:
		return fmt.Errorf("unsupported workload type: %s", info.Target.WorkloadType)
	}

	d.log.Info("Validation successful", "workload", workloadName, "namespace", namespace)
	return nil
}

// Rollback rollbacks the workload to previous image version
func (d *K8sDeployer) Rollback(ctx context.Context, info *models.WorkloadInfo, previousImage string) error {
	d.log.Info("Rolling back K8s deployment", "app", info.App.Name, "cluster", info.Cluster.Name, 
		"previousImage", previousImage)

	// Validate inputs
	if info == nil || info.App == nil || info.Target == nil {
		return fmt.Errorf("invalid workload info")
	}

	if previousImage == "" {
		return fmt.Errorf("previous image version not specified")
	}

	// Rollback is essentially a deploy with the previous image
	return d.Deploy(ctx, info, previousImage)
}

// GetStatus returns the current workload status
func (d *K8sDeployer) GetStatus(ctx context.Context, info *models.WorkloadInfo) (string, error) {
	d.log.Info("Getting K8s workload status", "app", info.App.Name, "workload", info.Target.K8sWorkload)

	if info == nil || info.Target == nil || info.Cluster == nil {
		return "", fmt.Errorf("invalid workload info")
	}

	clientset, err := d.getOrCreateClient(info.Cluster)
	if err != nil {
		return "", fmt.Errorf("failed to create K8s client: %w", err)
	}

	namespace := info.Target.K8sNamespace
	workloadName := info.Target.K8sWorkload

	switch info.Target.WorkloadType {
	case "Deployment":
		return d.getDeploymentStatus(ctx, clientset, namespace, workloadName)
	case "StatefulSet":
		return d.getStatefulSetStatus(ctx, clientset, namespace, workloadName)
	case "DaemonSet":
		return d.getDaemonSetStatus(ctx, clientset, namespace, workloadName)
	default:
		return "", fmt.Errorf("unsupported workload type: %s", info.Target.WorkloadType)
	}
}

// HealthCheck checks the health of deployed pods in the workload
func (d *K8sDeployer) HealthCheck(ctx context.Context, info *models.WorkloadInfo) (bool, error) {
	d.log.Info("Checking health of K8s workload", "app", info.App.Name, "workload", info.Target.K8sWorkload)

	if info == nil || info.Target == nil || info.Cluster == nil {
		return false, fmt.Errorf("invalid workload info")
	}

	clientset, err := d.getOrCreateClient(info.Cluster)
	if err != nil {
		return false, fmt.Errorf("failed to create K8s client: %w", err)
	}

	namespace := info.Target.K8sNamespace
	workloadName := info.Target.K8sWorkload

	// Get pods for this workload
	labelSelector := fmt.Sprintf("app=%s", workloadName) // Common label pattern
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		// If label selector fails, try to get pods with different pattern
		pods, err = clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, fmt.Errorf("failed to list pods: %w", err)
		}
	}

	if len(pods.Items) == 0 {
		d.log.Warn("No pods found for workload", "workload", workloadName, "namespace", namespace)
		return false, nil
	}

	// Check if all pods are ready and running
	healthy := true
	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			d.log.Warn("Pod not running", "pod", pod.Name, "phase", pod.Status.Phase)
			healthy = false
			break
		}

		// Check container readiness
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status != corev1.ConditionTrue {
				d.log.Warn("Pod not ready", "pod", pod.Name, "reason", condition.Reason)
				healthy = false
				break
			}
		}

		if !healthy {
			break
		}
	}

	d.log.Info("Health check result", "workload", workloadName, "healthy", healthy, "podCount", len(pods.Items))
	return healthy, nil
}

// ==================== Private Helper Methods ====================

// monitorPodEvents watches Pod events for a deployment and detects critical errors
// like ImagePullBackOff, CrashLoopBackOff, etc.
// 关键改进：只监听新创建的 Pod（来自当前 ReplicaSet），不监听旧 Pod
func (d *K8sDeployer) monitorPodEvents(ctx context.Context, clientset kubernetes.Interface, namespace, deploymentName string) error {
	podsClient := clientset.CoreV1().Pods(namespace)
	eventsClient := clientset.CoreV1().Events(namespace)
	replicasetsClient := clientset.AppsV1().ReplicaSets(namespace)
	
	// Get the deployment to extract selector and get the latest ReplicaSet
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		d.log.Warn("Failed to get deployment for Pod monitoring", "deployment", deploymentName, "error", err)
		return nil
	}
	
	// 获取最新的 ReplicaSet（通过 ownerReferences 关联）
	// 这样可以准确区分新旧 Pod
	replicasetList, err := replicasetsClient.List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		d.log.Warn("Failed to list replicasets for deployment", "deployment", deploymentName, "error", err)
		return nil
	}
	
	// 找到最新的 ReplicaSet（按创建时间排序）
	var latestRS *appsv1.ReplicaSet
	var latestRSTime metav1.Time
	for i := range replicasetList.Items {
		rs := &replicasetList.Items[i]
		if latestRS == nil || rs.CreationTimestamp.After(latestRSTime.Time) {
			latestRS = rs
			latestRSTime = rs.CreationTimestamp
		}
	}
	
	if latestRS == nil {
		d.log.Warn("No replicaset found for deployment", "deployment", deploymentName)
		return nil
	}
	
	d.log.Info("Pod event monitoring started", 
		"deployment", deploymentName, 
		"latestReplicaSet", latestRS.Name,
		"latestRSReplicas", latestRS.Status.Replicas,
		"namespace", namespace)
	
	// Track pods we've already reported errors for to avoid duplicates
	reportedErrors := make(map[string]bool)
	
	// 1. First, check for existing Pods from the LATEST ReplicaSet only
	// 仅检查最新 ReplicaSet 的 Pod（不检查旧 Pod）
	podList, err := podsClient.List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(latestRS.Spec.Selector),
	})
	if err == nil && podList != nil {
		d.log.Info("Checking existing Pods from latest ReplicaSet for errors", 
			"deployment", deploymentName, 
			"replicaset", latestRS.Name,
			"podCount", len(podList.Items))
		
		for _, pod := range podList.Items {
			// 确保这个 Pod 确实属于最新的 ReplicaSet
			if !isPodOwnedByReplicaSet(&pod, latestRS) {
				d.log.Debug("Skipping pod from old replicaset", "pod", pod.Name, "replicaset", latestRS.Name)
				continue
			}
			
			d.log.Debug("Checking pod status", "pod", pod.Name, "phase", pod.Status.Phase)
			
			// Check Pod container statuses for error conditions
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil {
					reason := containerStatus.State.Waiting.Reason
					message := containerStatus.State.Waiting.Message
					
					d.log.Warn("Pod container waiting", "pod", pod.Name, "reason", reason, "message", message)
					
					switch reason {
					case "ImagePullBackOff":
						errorMsg := fmt.Sprintf("[ImagePullBackOff] Failed to pull image: %s (Pod: %s)", message, pod.Name)
						d.log.Error(errorMsg, "deployment", deploymentName)
						// Phase 2: 记录Pod错误检测事件
						d.recordDeploymentEvent("pod_error_detected", 
							fmt.Sprintf("Pod %s: ImagePullBackOff - %s", pod.Name, message))
						return fmt.Errorf(errorMsg)
					case "CrashLoopBackOff":
						errorMsg := fmt.Sprintf("[CrashLoopBackOff] Container crashed repeatedly: %s (Pod: %s)", message, pod.Name)
						d.log.Error(errorMsg, "deployment", deploymentName)
						// Phase 2: 记录Pod错误检测事件
						d.recordDeploymentEvent("pod_error_detected", 
							fmt.Sprintf("Pod %s: CrashLoopBackOff - Container crashed repeatedly", pod.Name))
						return fmt.Errorf(errorMsg)
					case "ErrImagePull":
						errorMsg := fmt.Sprintf("[ErrImagePull] Error pulling image: %s (Pod: %s)", message, pod.Name)
						d.log.Warn(errorMsg, "deployment", deploymentName)
						// Phase 2: 记录Pod错误检测事件
						d.recordDeploymentEvent("pod_error_detected", 
							fmt.Sprintf("Pod %s: ErrImagePull - %s", pod.Name, message))
						return fmt.Errorf(errorMsg)
					}
				}
			}
		}
	}
	
	// 2. Then watch for new events ONLY from the latest ReplicaSet Pods
	// 只监听最新 ReplicaSet 的 Pod 事件
	watcher, err := eventsClient.Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.kind=Pod,involvedObject.namespace=%s", namespace),
		Watch:         true,
	})
	if err != nil {
		d.log.Warn("Failed to watch Pod events", "deployment", deploymentName, "error", err)
		return nil
	}
	defer watcher.Stop()
	
	d.log.Info("Pod event watcher started", "deployment", deploymentName, "replicaset", latestRS.Name)
	
	// Track last check time to avoid processing old events
	watchStartTime := time.Now()
	
	for {
		select {
		case <-ctx.Done():
			d.log.Debug("Pod event monitoring context done", "deployment", deploymentName)
			return nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				d.log.Debug("Pod event watcher closed", "deployment", deploymentName)
				return nil
			}
			
			if event.Type == watch.Error {
				d.log.Warn("Pod event watch error", "deployment", deploymentName)
				continue
			}
			
			k8sEvent, ok := event.Object.(*corev1.Event)
			if !ok {
				continue
			}
			
			// Only process events for Pods from LATEST ReplicaSet
			// Pod names typically follow pattern: <deployment>-<replicaset-hash>-<pod-hash>
			if !strings.Contains(k8sEvent.InvolvedObject.Name, deploymentName) {
				continue
			}
			
			// Phase 1: 改进Pod-ReplicaSet关联验证
			// 对于关键事件（如错误事件），获取Pod对象来验证ownerReference
			podName := k8sEvent.InvolvedObject.Name
			
			// 对于错误类事件，进行更严格的验证
			isCriticalEvent := strings.Contains(k8sEvent.Reason, "Failed") || 
				strings.Contains(k8sEvent.Reason, "Error") ||
				strings.Contains(k8sEvent.Reason, "BackOff") ||
				k8sEvent.Reason == "FailedCreate"
			
			if isCriticalEvent {
				// 获取Pod对象以验证ownership
				pod, err := podsClient.Get(ctx, podName, metav1.GetOptions{})
				if err == nil {
					if !isPodOwnedByReplicaSet(pod, latestRS) {
						d.log.Debug("Skipping event: pod not owned by latest replicaset",
							"pod", podName, "latestRS", latestRS.Name)
						continue
					}
				} else {
					// 如果无法获取Pod，用字符串匹配作为备选
					if !strings.Contains(podName, latestRS.Name) {
						d.log.Debug("Skipping event from old replicaset pod",
							"pod", podName, "latestRS", latestRS.Name)
						continue
					}
				}
			} else {
				// 非关键事件，使用快速的字符串匹配
				if !strings.Contains(podName, latestRS.Name) {
					d.log.Debug("Skipping event from old replicaset pod",
						"pod", podName, "latestRS", latestRS.Name)
					continue
				}
			}
			
			// Skip events that occurred before monitoring started
			if k8sEvent.FirstTimestamp.Time.Before(watchStartTime) && k8sEvent.LastTimestamp.Time.Before(watchStartTime.Add(5*time.Second)) {
				continue
			}
			
			eventKey := fmt.Sprintf("%s-%s", podName, k8sEvent.Reason)
			
			d.log.Debug("Pod event received from latest replicaset", 
				"pod", podName, 
				"replicaset", latestRS.Name,
				"reason", k8sEvent.Reason, 
				"message", k8sEvent.Message)
			
			// Detect critical error conditions
			switch k8sEvent.Reason {
			case "ImagePullBackOff":
				if !reportedErrors[eventKey] {
					errorMsg := fmt.Sprintf("[ImagePullBackOff] Failed to pull image: %s (Pod: %s)", k8sEvent.Message, podName)
					d.log.Error(errorMsg, "deployment", deploymentName, "replicaset", latestRS.Name)
					reportedErrors[eventKey] = true
					// Phase 2: 发送Pod错误检测事件
					d.recordDeploymentEvent("pod_error_detected", 
						fmt.Sprintf("Pod %s: ImagePullBackOff - %s", podName, k8sEvent.Message))
					return fmt.Errorf(errorMsg)
				}
			
			case "CrashLoopBackOff":
				if !reportedErrors[eventKey] {
					errorMsg := fmt.Sprintf("[CrashLoopBackOff] Container crashed repeatedly: %s (Pod: %s)", k8sEvent.Message, podName)
					d.log.Error(errorMsg, "deployment", deploymentName, "replicaset", latestRS.Name)
					reportedErrors[eventKey] = true
					// Phase 2: 发送Pod错误检测事件
					d.recordDeploymentEvent("pod_error_detected", 
						fmt.Sprintf("Pod %s: CrashLoopBackOff - Container crashed repeatedly", podName))
					return fmt.Errorf(errorMsg)
				}
			
			case "ErrImagePull":
				if !reportedErrors[eventKey] {
					errorMsg := fmt.Sprintf("[ErrImagePull] Error pulling image: %s (Pod: %s)", k8sEvent.Message, podName)
					d.log.Warn(errorMsg, "deployment", deploymentName, "replicaset", latestRS.Name)
					reportedErrors[eventKey] = true
					// Phase 2: 发送Pod错误检测事件
					d.recordDeploymentEvent("pod_error_detected", 
						fmt.Sprintf("Pod %s: ErrImagePull - %s", podName, k8sEvent.Message))
					return fmt.Errorf(errorMsg)
				}
			
			case "Failed":
				if !reportedErrors[eventKey] && strings.Contains(k8sEvent.Message, "pull") {
					errorMsg := fmt.Sprintf("[Failed] Failed to pull image: %s (Pod: %s)", k8sEvent.Message, podName)
					d.log.Error(errorMsg, "deployment", deploymentName, "replicaset", latestRS.Name)
					reportedErrors[eventKey] = true
					// Phase 2: 发送Pod错误检测事件
					d.recordDeploymentEvent("pod_error_detected", 
						fmt.Sprintf("Pod %s: Failed to pull image - %s", podName, k8sEvent.Message))
					return fmt.Errorf(errorMsg)
				}
			
			case "BackOff":
				if !reportedErrors[eventKey] && strings.Contains(k8sEvent.Message, "pull") {
					errorMsg := fmt.Sprintf("[BackOff] Image pull backoff: %s (Pod: %s)", k8sEvent.Message, podName)
					d.log.Warn(errorMsg, "deployment", deploymentName, "replicaset", latestRS.Name)
					reportedErrors[eventKey] = true
					d.log.Info("Image pull backoff detected, waiting for more info", "pod", podName)
				}
			
			default:
				// Log other events for debugging
				if strings.Contains(k8sEvent.Message, "pull") || strings.Contains(k8sEvent.Reason, "Pull") {
					d.log.Debug("Pod pull-related event", 
						"pod", podName, 
						"reason", k8sEvent.Reason,
						"message", k8sEvent.Message,
						"count", k8sEvent.Count)
				}
			}
		}
	}
}

// getOrCreateClient creates or retrieves a cached K8s client for the cluster
// getOrCreateClient 使用 utils.K8sClientManager 获取或创建 K8s 客户端
func (d *K8sDeployer) getOrCreateClient(cluster *models.Cluster) (kubernetes.Interface, error) {
	clusterKey := fmt.Sprintf("%d_%s", cluster.ID, cluster.Name)

	// Validate kubeconfig exists
	if cluster.Kubeconfig == nil || *cluster.Kubeconfig == "" {
		return nil, fmt.Errorf("kubeconfig is empty for cluster %s (id=%d)", cluster.Name, cluster.ID)
	}

	// Decrypt kubeconfig
	decryptedKubeconfig, err := utils.DecryptAES(*cluster.Kubeconfig, d.encryptionKey)
	if err != nil {
		errMsg := fmt.Sprintf("failed to decrypt kubeconfig for cluster %s: %v", cluster.Name, err)
		d.log.Error(errMsg, "clusterID", cluster.ID)
		return nil, fmt.Errorf(errMsg)
	}

	// Use K8sClientManager to get or create client (handles caching)
	clientset, err := d.clientManager.GetOrCreateK8sClient(clusterKey, decryptedKubeconfig, cluster.KubernetesVersion)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create K8s client for cluster %s (id=%d): %v", cluster.Name, cluster.ID, err)
		d.log.Error(errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	d.log.Info("K8s client created and cached", "cluster", cluster.Name)
	return clientset, nil
}

// deployDeployment updates a Kubernetes Deployment with new image
func (d *K8sDeployer) deployDeployment(ctx context.Context, clientset kubernetes.Interface, namespace, deploymentName string, containerName *string, image string) error {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	// Get current deployment
	deployment, err := deploymentsClient.Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("failed to get deployment %q in namespace %q: %v\nDeployment may not exist or namespace may be inaccessible", deploymentName, namespace, err)
		d.log.Error(errMsg, "error", err, "namespace", namespace, "deployment", deploymentName)
		return fmt.Errorf(errMsg)
	}

	d.log.Info("Deployment found", "deployment", deploymentName, "namespace", namespace, "replicas", deployment.Spec.Replicas)

	// Log container information
	d.log.Info("Current deployment containers", "containerCount", len(deployment.Spec.Template.Spec.Containers))
	for i, container := range deployment.Spec.Template.Spec.Containers {
		d.log.Info(fmt.Sprintf("Container %d", i), "name", container.Name, "image", container.Image)
	}

	d.log.Info("Updating deployment", "deployment", deploymentName, "targetImage", image, "containerName", containerName)

	// Update container image
	updated := false
	for i := range deployment.Spec.Template.Spec.Containers {
		// If containerName is specified, only update that container
		if containerName != nil && *containerName != "" {
			if deployment.Spec.Template.Spec.Containers[i].Name == *containerName {
				deployment.Spec.Template.Spec.Containers[i].Image = image
				updated = true
				d.log.Info("Container image updated", "container", *containerName, "image", image)
				break
			}
		} else {
			// If no container name specified, update the first container (for backward compatibility)
			if i == 0 {
				deployment.Spec.Template.Spec.Containers[i].Image = image
				updated = true
				d.log.Info("Container image updated (first container)", "containerName", deployment.Spec.Template.Spec.Containers[i].Name, "image", image)
				break
			}
		}
	}

	if !updated && containerName != nil && *containerName != "" {
		errMsg := fmt.Sprintf("container %q not found in deployment %q. Available containers: %v", 
			*containerName, deploymentName, getContainerNames(deployment))
		d.log.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	// Apply the update
	_, err = deploymentsClient.Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("failed to update deployment %q: %v", deploymentName, err)
		d.log.Error(errMsg, "error", err)
		return fmt.Errorf(errMsg)
	}

	d.log.Info("Deployment update applied", "deployment", deploymentName)

	// Wait for rollout to complete
	return d.waitForDeploymentRollout(ctx, clientset, namespace, deploymentName, DeploymentRolloutTimeout)
}

// deployStatefulSet updates a Kubernetes StatefulSet with new image
func (d *K8sDeployer) deployStatefulSet(ctx context.Context, clientset kubernetes.Interface, namespace, statefulSetName string, containerName *string, image string) error {
	stsClient := clientset.AppsV1().StatefulSets(namespace)

	// Get current StatefulSet
	sts, err := stsClient.Get(ctx, statefulSetName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get statefulset %q: %w", statefulSetName, err)
	}

	d.log.Info("Updating statefulset", "statefulset", statefulSetName, "image", image)

	// Update container image
	updated := false
	for i := range sts.Spec.Template.Spec.Containers {
		if containerName != nil && *containerName != "" {
			if sts.Spec.Template.Spec.Containers[i].Name == *containerName {
				sts.Spec.Template.Spec.Containers[i].Image = image
				updated = true
				d.log.Info("Container image updated", "container", *containerName, "image", image)
				break
			}
		} else {
			if i == 0 {
				sts.Spec.Template.Spec.Containers[i].Image = image
				updated = true
				break
			}
		}
	}

	if !updated && containerName != nil && *containerName != "" {
		return fmt.Errorf("container %q not found in statefulset spec", *containerName)
	}

	// Apply the update
	_, err = stsClient.Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update statefulset: %w", err)
	}

	d.log.Info("StatefulSet update applied", "statefulset", statefulSetName)

	// Wait for rollout to complete
	return d.waitForStatefulSetRollout(ctx, clientset, namespace, statefulSetName, StatefulSetRolloutTimeout)
}

// deployDaemonSet updates a Kubernetes DaemonSet with new image
func (d *K8sDeployer) deployDaemonSet(ctx context.Context, clientset kubernetes.Interface, namespace, daemonSetName string, containerName *string, image string) error {
	dsClient := clientset.AppsV1().DaemonSets(namespace)

	// Get current DaemonSet
	ds, err := dsClient.Get(ctx, daemonSetName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get daemonset %q: %w", daemonSetName, err)
	}

	d.log.Info("Updating daemonset", "daemonset", daemonSetName, "image", image)

	// Update container image
	updated := false
	for i := range ds.Spec.Template.Spec.Containers {
		if containerName != nil && *containerName != "" {
			if ds.Spec.Template.Spec.Containers[i].Name == *containerName {
				ds.Spec.Template.Spec.Containers[i].Image = image
				updated = true
				d.log.Info("Container image updated", "container", *containerName, "image", image)
				break
			}
		} else {
			if i == 0 {
				ds.Spec.Template.Spec.Containers[i].Image = image
				updated = true
				break
			}
		}
	}

	if !updated && containerName != nil && *containerName != "" {
		return fmt.Errorf("container %q not found in daemonset spec", *containerName)
	}

	// Apply the update
	_, err = dsClient.Update(ctx, ds, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update daemonset: %w", err)
	}

	d.log.Info("DaemonSet update applied", "daemonset", daemonSetName)

	// DaemonSet doesn't have replicas in the traditional sense, so we just wait for pods
	return d.waitForDaemonSetRollout(ctx, clientset, namespace, daemonSetName, DaemonSetRolloutTimeout)
}

// waitForDeploymentRollout waits for deployment rollout to complete
func (d *K8sDeployer) waitForDeploymentRollout(ctx context.Context, clientset kubernetes.Interface, namespace, deploymentName string, timeout time.Duration) error {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	d.log.Info("Waiting for deployment rollout", "deployment", deploymentName, "timeout", timeout)

	// Channel to signal Pod errors (e.g., ImagePullBackOff)
	podErrorCh := make(chan error, 1)
	
	// Start monitoring Pod events in a separate goroutine
	go func() {
		if err := d.monitorPodEvents(ctx, clientset, namespace, deploymentName); err != nil {
			select {
			case podErrorCh <- err:
			case <-ctx.Done():
			}
		}
	}()

	// Use watch to monitor deployment updates
	watcher, err := deploymentsClient.Watch(ctx, metav1.ListOptions{
		FieldSelector: "metadata.name=" + deploymentName,
	})
	if err != nil {
		return fmt.Errorf("failed to watch deployment: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			// Phase 1: 增强超时诊断信息
			// 收集诊断信息而不是简单地返回超时错误
			diagnostics := d.collectDeploymentDiagnostics(ctx, clientset, namespace, deploymentName)
			if diagnostics != "" {
				d.log.Error("Deployment rollout timeout with diagnostics", 
					"deployment", deploymentName, "diagnostics", diagnostics)
				return fmt.Errorf("deployment rollout timeout after %v: %s", timeout, diagnostics)
			}
			return fmt.Errorf("deployment rollout timeout after %v", timeout)
		
		// Check for Pod errors (ImagePullBackOff, CrashLoopBackOff, etc.)
		case podErr := <-podErrorCh:
			return podErr
		
		case event := <-watcher.ResultChan():
			if event.Type == watch.Error {
				return fmt.Errorf("watch error: %v", event.Object)
			}

			deployment, ok := event.Object.(*appsv1.Deployment)
			if !ok {
				continue
			}

			// Check if rollout is complete
			if deployment.Status.ObservedGeneration >= deployment.Generation &&
				deployment.Status.Replicas == deployment.Status.UpdatedReplicas &&
				deployment.Status.Replicas == deployment.Status.ReadyReplicas &&
				deployment.Status.AvailableReplicas == deployment.Status.ReadyReplicas {
				d.log.Info("Deployment rollout completed", "deployment", deploymentName)
				// Phase 2: 发送部署成功事件
				d.recordDeploymentEvent("deployment_success", 
					fmt.Sprintf("Deployment %s rolled out successfully with %d replicas", deploymentName, deployment.Status.Replicas))
				return nil
			}

			d.log.Debug("Deployment status", "deployment", deploymentName,
				"replicas", deployment.Status.Replicas,
				"updated", deployment.Status.UpdatedReplicas,
				"ready", deployment.Status.ReadyReplicas,
				"available", deployment.Status.AvailableReplicas)
		}
	}
}

// waitForStatefulSetRollout waits for statefulset rollout to complete
func (d *K8sDeployer) waitForStatefulSetRollout(ctx context.Context, clientset kubernetes.Interface, namespace, statefulSetName string, timeout time.Duration) error {
	stsClient := clientset.AppsV1().StatefulSets(namespace)
	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	d.log.Info("Waiting for statefulset rollout", "statefulset", statefulSetName, "timeout", timeout)

	// Channel to signal Pod errors (e.g., ImagePullBackOff)
	podErrorCh := make(chan error, 1)
	
	// Start monitoring Pod events in a separate goroutine
	go func() {
		if err := d.monitorPodEvents(ctx, clientset, namespace, statefulSetName); err != nil {
			select {
			case podErrorCh <- err:
			case <-ctx.Done():
			}
		}
	}()

	watcher, err := stsClient.Watch(ctx, metav1.ListOptions{
		FieldSelector: "metadata.name=" + statefulSetName,
	})
	if err != nil {
		return fmt.Errorf("failed to watch statefulset: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			// Phase 1: 增强超时诊断信息（StatefulSet版本）
			diagnostics := d.collectDeploymentDiagnostics(ctx, clientset, namespace, statefulSetName)
			if diagnostics != "" {
				d.log.Error("StatefulSet rollout timeout with diagnostics", 
					"statefulset", statefulSetName, "diagnostics", diagnostics)
				return fmt.Errorf("statefulset rollout timeout after %v: %s", timeout, diagnostics)
			}
			return fmt.Errorf("statefulset rollout timeout after %v", timeout)
		
		// Check for Pod errors (ImagePullBackOff, CrashLoopBackOff, etc.)
		case podErr := <-podErrorCh:
			return podErr
		
		case event := <-watcher.ResultChan():
			if event.Type == watch.Error {
				return fmt.Errorf("watch error: %v", event.Object)
			}

			sts, ok := event.Object.(*appsv1.StatefulSet)
			if !ok {
				continue
			}

			// Check if rollout is complete
			if sts.Status.ObservedGeneration >= sts.Generation &&
				sts.Status.Replicas == sts.Status.UpdatedReplicas &&
				sts.Status.Replicas == sts.Status.ReadyReplicas {
				d.log.Info("StatefulSet rollout completed", "statefulset", statefulSetName)
				// Phase 2: 发送部署成功事件
				d.recordDeploymentEvent("deployment_success", 
					fmt.Sprintf("StatefulSet %s rolled out successfully with %d replicas", statefulSetName, sts.Status.Replicas))
				return nil
			}

			d.log.Debug("StatefulSet status", "statefulset", statefulSetName,
				"replicas", sts.Status.Replicas,
				"updated", sts.Status.UpdatedReplicas,
				"ready", sts.Status.ReadyReplicas)
		}
	}
}

// waitForDaemonSetRollout waits for daemonset rollout to complete
func (d *K8sDeployer) waitForDaemonSetRollout(ctx context.Context, clientset kubernetes.Interface, namespace, daemonSetName string, timeout time.Duration) error {
	dsClient := clientset.AppsV1().DaemonSets(namespace)
	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	d.log.Info("Waiting for daemonset rollout", "daemonset", daemonSetName, "timeout", timeout)

	// Channel to signal Pod errors (e.g., ImagePullBackOff)
	podErrorCh := make(chan error, 1)
	
	// Start monitoring Pod events in a separate goroutine
	go func() {
		if err := d.monitorPodEvents(ctx, clientset, namespace, daemonSetName); err != nil {
			select {
			case podErrorCh <- err:
			case <-ctx.Done():
			}
		}
	}()

	watcher, err := dsClient.Watch(ctx, metav1.ListOptions{
		FieldSelector: "metadata.name=" + daemonSetName,
	})
	if err != nil {
		return fmt.Errorf("failed to watch daemonset: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			// Phase 1: 增强超时诊断信息（DaemonSet版本）
			diagnostics := d.collectDeploymentDiagnostics(ctx, clientset, namespace, daemonSetName)
			if diagnostics != "" {
				d.log.Error("DaemonSet rollout timeout with diagnostics", 
					"daemonset", daemonSetName, "diagnostics", diagnostics)
				return fmt.Errorf("daemonset rollout timeout after %v: %s", timeout, diagnostics)
			}
			return fmt.Errorf("daemonset rollout timeout after %v", timeout)
		
		// Check for Pod errors (ImagePullBackOff, CrashLoopBackOff, etc.)
		case podErr := <-podErrorCh:
			return podErr
		
		case event := <-watcher.ResultChan():
			if event.Type == watch.Error {
				return fmt.Errorf("watch error: %v", event.Object)
			}

			ds, ok := event.Object.(*appsv1.DaemonSet)
			if !ok {
				continue
			}

			// Check if rollout is complete
			if ds.Status.ObservedGeneration >= ds.Generation &&
				ds.Status.DesiredNumberScheduled == ds.Status.UpdatedNumberScheduled &&
				ds.Status.DesiredNumberScheduled == ds.Status.NumberReady {
				d.log.Info("DaemonSet rollout completed", "daemonset", daemonSetName)
				// Phase 2: 发送部署成功事件
				d.recordDeploymentEvent("deployment_success", 
					fmt.Sprintf("DaemonSet %s rolled out successfully on %d nodes", daemonSetName, ds.Status.NumberReady))
				return nil
			}

			d.log.Debug("DaemonSet status", "daemonset", daemonSetName,
				"scheduled", ds.Status.DesiredNumberScheduled,
				"updated", ds.Status.UpdatedNumberScheduled,
				"ready", ds.Status.NumberReady)
		}
	}
}

// getDeploymentStatus returns the current deployment status with more details
func (d *K8sDeployer) getDeploymentStatus(ctx context.Context, clientset kubernetes.Interface, namespace, deploymentName string) (string, error) {
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get deployment: %w", err)
	}

	// 检查是否有更新正在进行
	if deployment.Status.ObservedGeneration < deployment.Generation {
		d.log.Debug("Deployment still updating", "deployment", deploymentName, 
			"observedGeneration", deployment.Status.ObservedGeneration,
			"generation", deployment.Generation)
		return "updating", nil
	}

	// 检查副本数是否为0
	if deployment.Status.Replicas == 0 {
		d.log.Debug("Deployment pending (no replicas)", "deployment", deploymentName)
		return "pending", nil
	}

	// 检查是否所有副本都已更新
	if deployment.Status.UpdatedReplicas < deployment.Status.Replicas {
		d.log.Debug("Deployment rolling update in progress", "deployment", deploymentName,
			"replicas", deployment.Status.Replicas,
			"updatedReplicas", deployment.Status.UpdatedReplicas)
		return "rolling", nil
	}

	// 检查是否所有副本都已就绪
	if deployment.Status.ReadyReplicas < deployment.Status.Replicas {
		d.log.Debug("Deployment waiting for pods to be ready", "deployment", deploymentName,
			"readyReplicas", deployment.Status.ReadyReplicas,
			"desiredReplicas", deployment.Status.Replicas)
		return "waiting", nil
	}

	// 检查是否所有副本都可用
	if deployment.Status.AvailableReplicas < deployment.Status.Replicas {
		d.log.Debug("Deployment waiting for pods to be available", "deployment", deploymentName,
			"availableReplicas", deployment.Status.AvailableReplicas,
			"desiredReplicas", deployment.Status.Replicas)
		return "waiting", nil
	}

	// 全部条件满足 - 发布完成
	d.log.Info("Deployment rollout successful", "deployment", deploymentName,
		"replicas", deployment.Status.Replicas,
		"updatedReplicas", deployment.Status.UpdatedReplicas,
		"readyReplicas", deployment.Status.ReadyReplicas,
		"availableReplicas", deployment.Status.AvailableReplicas)
	return "completed", nil
}

// getStatefulSetStatus returns the current statefulset status with more details
func (d *K8sDeployer) getStatefulSetStatus(ctx context.Context, clientset kubernetes.Interface, namespace, statefulSetName string) (string, error) {
	sts, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, statefulSetName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get statefulset: %w", err)
	}

	// 检查是否有更新正在进行
	if sts.Status.ObservedGeneration < sts.Generation {
		d.log.Debug("StatefulSet still updating", "statefulset", statefulSetName,
			"observedGeneration", sts.Status.ObservedGeneration,
			"generation", sts.Generation)
		return "updating", nil
	}

	// 检查副本数是否为0
	if sts.Status.Replicas == 0 {
		d.log.Debug("StatefulSet pending (no replicas)", "statefulset", statefulSetName)
		return "pending", nil
	}

	// 检查是否所有副本都已更新
	if sts.Status.UpdatedReplicas < sts.Status.Replicas {
		d.log.Debug("StatefulSet rolling update in progress", "statefulset", statefulSetName,
			"replicas", sts.Status.Replicas,
			"updatedReplicas", sts.Status.UpdatedReplicas)
		return "rolling", nil
	}

	// 检查是否所有副本都已就绪
	if sts.Status.ReadyReplicas < sts.Status.Replicas {
		d.log.Debug("StatefulSet waiting for pods to be ready", "statefulset", statefulSetName,
			"readyReplicas", sts.Status.ReadyReplicas,
			"desiredReplicas", sts.Status.Replicas)
		return "waiting", nil
	}

	// 全部条件满足 - 发布完成
	d.log.Info("StatefulSet rollout successful", "statefulset", statefulSetName,
		"replicas", sts.Status.Replicas,
		"updatedReplicas", sts.Status.UpdatedReplicas,
		"readyReplicas", sts.Status.ReadyReplicas)
	return "completed", nil
}

// getDaemonSetStatus returns the current daemonset status with more details
func (d *K8sDeployer) getDaemonSetStatus(ctx context.Context, clientset kubernetes.Interface, namespace, daemonSetName string) (string, error) {
	ds, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, daemonSetName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get daemonset: %w", err)
	}

	// 检查是否有更新正在进行
	if ds.Status.ObservedGeneration < ds.Generation {
		d.log.Debug("DaemonSet still updating", "daemonset", daemonSetName,
			"observedGeneration", ds.Status.ObservedGeneration,
			"generation", ds.Generation)
		return "updating", nil
	}

	// DaemonSet 的期望数量是节点数量
	// 检查是否所有Pod都已更新
	if ds.Status.UpdatedNumberScheduled < ds.Status.DesiredNumberScheduled {
		d.log.Debug("DaemonSet rolling update in progress", "daemonset", daemonSetName,
			"desiredNumberScheduled", ds.Status.DesiredNumberScheduled,
			"updatedNumberScheduled", ds.Status.UpdatedNumberScheduled)
		return "rolling", nil
	}

	// 检查是否所有Pod都已就绪
	if ds.Status.NumberReady < ds.Status.DesiredNumberScheduled {
		d.log.Debug("DaemonSet waiting for pods to be ready", "daemonset", daemonSetName,
			"desiredNumberScheduled", ds.Status.DesiredNumberScheduled,
			"numberReady", ds.Status.NumberReady)
		return "waiting", nil
	}

	// 检查是否所有Pod都可用
	if ds.Status.NumberAvailable < ds.Status.DesiredNumberScheduled {
		d.log.Debug("DaemonSet waiting for pods to be available", "daemonset", daemonSetName,
			"desiredNumberScheduled", ds.Status.DesiredNumberScheduled,
			"numberAvailable", ds.Status.NumberAvailable)
		return "waiting", nil
	}

	// 全部条件满足 - 发布完成
	d.log.Info("DaemonSet rollout successful", "daemonset", daemonSetName,
		"desiredNumberScheduled", ds.Status.DesiredNumberScheduled,
		"updatedNumberScheduled", ds.Status.UpdatedNumberScheduled,
		"numberReady", ds.Status.NumberReady,
		"numberAvailable", ds.Status.NumberAvailable)
	return "completed", nil
}

// getContainerNames extracts container names from a deployment spec
func getContainerNames(deployment *appsv1.Deployment) []string {
	var names []string
	if deployment == nil || deployment.Spec.Template.Spec.Containers == nil {
		return names
	}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		names = append(names, container.Name)
	}
	return names
}

// isPodOwnedByReplicaSet checks if a Pod is owned by a specific ReplicaSet
// by examining the ownerReferences
func isPodOwnedByReplicaSet(pod *corev1.Pod, rs *appsv1.ReplicaSet) bool {
	if pod == nil || rs == nil {
		return false
	}
	
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "ReplicaSet" && owner.Name == rs.Name {
			return true
		}
	}
	
	return false
}

// collectDeploymentDiagnostics 收集部署诊断信息
// Phase 1: 增强超时诊断
func (d *K8sDeployer) collectDeploymentDiagnostics(ctx context.Context, clientset kubernetes.Interface, namespace, deploymentName string) string {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	rsClient := clientset.AppsV1().ReplicaSets(namespace)
	podsClient := clientset.CoreV1().Pods(namespace)
	
	var diagMsg strings.Builder
	
	// 1. 获取Deployment状态
	deployment, err := deploymentsClient.Get(ctx, deploymentName, metav1.GetOptions{})
	if err == nil && deployment != nil {
		diagMsg.WriteString(fmt.Sprintf("Deployment status: [Desired=%d, Updated=%d, Ready=%d, Available=%d], ",
			deployment.Spec.Replicas, deployment.Status.UpdatedReplicas, deployment.Status.ReadyReplicas, deployment.Status.AvailableReplicas))
		
		// 检查是否有进度条件
		for _, cond := range deployment.Status.Conditions {
			if cond.Type == "Progressing" {
				diagMsg.WriteString(fmt.Sprintf("Progressing[%s]: %s, ", cond.Reason, cond.Message))
			}
		}
	}
	
	// 2. 获取ReplicaSet状态
	rsList, err := rsClient.List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err == nil && rsList != nil {
		diagMsg.WriteString(fmt.Sprintf("ReplicaSets: [Count=%d], ", len(rsList.Items)))
		for _, rs := range rsList.Items {
			diagMsg.WriteString(fmt.Sprintf("%s[Desired=%d, Current=%d, Ready=%d, Available=%d], ",
				rs.Name, rs.Spec.Replicas, rs.Status.Replicas, rs.Status.ReadyReplicas, rs.Status.AvailableReplicas))
		}
	}
	
	// 3. 获取Pod状态和错误信息
	podList, err := podsClient.List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err == nil && podList != nil {
		diagMsg.WriteString(fmt.Sprintf("Pods: [Total=%d], ", len(podList.Items)))
		
		failedPods := 0
		for _, pod := range podList.Items {
			// 检查Pod容器错误
			if errorReason := getPodErrorReason(&pod); errorReason != "" {
				failedPods++
				diagMsg.WriteString(fmt.Sprintf("Pod %s error: %s, ", pod.Name, errorReason))
			}
		}
		
		if failedPods == 0 && len(podList.Items) > 0 {
			diagMsg.WriteString("All pods are in progress (Pending or Running)")
		}
	}
	
	return diagMsg.String()
}

// getPodErrorReason 提取Pod的错误原因
// Phase 1: 容器状态检查
func getPodErrorReason(pod *corev1.Pod) string {
	if pod == nil {
		return ""
	}
	
	// 检查Pod阶段
	if pod.Status.Phase == corev1.PodFailed {
		return fmt.Sprintf("Pod phase Failed: %s", pod.Status.Reason)
	}
	
	// 检查容器状态
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			message := containerStatus.State.Waiting.Message
			
			// 记录所有的Waiting原因
			switch reason {
			case "ImagePullBackOff", "ErrImagePull", "RegistryUnavailable":
				return fmt.Sprintf("Container %s: %s - %s", containerStatus.Name, reason, message)
			case "CrashLoopBackOff":
				return fmt.Sprintf("Container %s: %s (restarted %d times)", containerStatus.Name, reason, containerStatus.RestartCount)
			case "CreateContainerConfigError":
				return fmt.Sprintf("Container %s: Config error - %s", containerStatus.Name, message)
			case "InvalidImageName":
				return fmt.Sprintf("Container %s: Invalid image name - %s", containerStatus.Name, message)
			default:
				// 其他Waiting原因
				if message != "" {
					return fmt.Sprintf("Container %s: %s - %s", containerStatus.Name, reason, message)
				}
			}
		}
		
		if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode != 0 {
			return fmt.Sprintf("Container %s terminated with exit code %d: %s", 
				containerStatus.Name, containerStatus.State.Terminated.ExitCode, containerStatus.State.Terminated.Reason)
		}
	}
	
	return ""
}
