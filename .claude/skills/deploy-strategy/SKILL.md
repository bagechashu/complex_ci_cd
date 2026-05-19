---
name: deploy-strategy
description: 发布控制系统 - 部署策略模式与Go并发实现
keywords: 策略模式, 多部署方式, Kubernetes, 异步发布, Goroutine, 事件日志
---

# 部署策略与并发实现指南

## 部署架构概览

### 策略模式设计

项目使用**策略模式**支持K8s部署。Salt/Ansible等其他部署方式通过SSH执行Shell命令处理：

```
┌─────────────────────────────────────────────┐
│         DeployStrategy Interface             │
├─────────────────────────────────────────────┤
│ + Deploy(ctx, info, image) error            │
│ + Validate(ctx, info) error                 │
│ + Rollback(ctx, info, prevImage) error      │
│ + GetStatus(ctx, info) string               │
│ + HealthCheck(ctx, info) bool               │
│ + Type() string                             │
└─────────────────────────────────────────────┘
         ▲
         │
    ┌────┴──┐
    │ K8s  │
    │Deploy│
    └──────┘

其他部署方式 (Salt/Ansible) 通过:
    ShellService (SSH连接) → 执行远程Shell命令
```

## 关键接口与实现

### 1. DeployStrategy 接口

```go
// internal/deployers/deployer.go

package deployers

import (
    "built-and-deploy/internal/models"
    "context"
)

// DeployStrategy 定义部署策略接口
type DeployStrategy interface {
    // Deploy 执行应用部署
    // 输入: 工作负载信息、目标镜像
    // 输出: 错误(如果部署失败)
    Deploy(ctx context.Context, info *models.WorkloadInfo, image string) error

    // Validate 验证工作负载配置是否有效
    Validate(ctx context.Context, info *models.WorkloadInfo) error

    // Rollback 回滚到上一个镜像版本
    Rollback(ctx context.Context, info *models.WorkloadInfo, previousImage string) error

    // GetStatus 获取当前工作负载状态
    // 返回状态字符串: 'running', 'pending', 'error'
    GetStatus(ctx context.Context, info *models.WorkloadInfo) (string, error)

    // HealthCheck 健康检查 - 检查Pod是否就绪
    HealthCheck(ctx context.Context, info *models.WorkloadInfo) (bool, error)

    // Type 返回部署器类型标识
    Type() string
}

// BaseDeployer 提供通用功能
type BaseDeployer struct {
    name string
}

func (b *BaseDeployer) Type() string {
    return b.name
}
```

### 2. K8s 部署器实现

```go
// internal/deployers/k8s_deployer.go

package deployers

import (
    "context"
    "fmt"
    "sync"
    
    "built-and-deploy/internal/models"
    "built-and-deploy/pkg/logger"
    
    corev1 "k8s.io/api/core/v1"
    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

type K8sDeployer struct {
    BaseDeployer
    // 客户端缓存: host → clientset
    clientCache map[string]kubernetes.Interface
    cacheLock   sync.RWMutex
    log         *logger.Logger
}

func NewK8sDeployer(log *logger.Logger) *K8sDeployer {
    return &K8sDeployer{
        BaseDeployer: BaseDeployer{name: "kubernetes"},
        clientCache:  make(map[string]kubernetes.Interface),
        log:          log,
    }
}

// getOrCreateClient 获取或创建K8s客户端(缓存优化)
func (d *K8sDeployer) getOrCreateClient(kubeconfig string) (kubernetes.Interface, error) {
    host := "default" // 简化版本，实际应提取kubeconfig的server地址
    
    d.cacheLock.RLock()
    if client, exists := d.clientCache[host]; exists {
        d.cacheLock.RUnlock()
        return client, nil
    }
    d.cacheLock.RUnlock()

    // 构建新客户端
    config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
    if err != nil {
        d.log.Error("Failed to build kubeconfig", "error", err)
        return nil, err
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        d.log.Error("Failed to create clientset", "error", err)
        return nil, err
    }

    // 缓存
    d.cacheLock.Lock()
    d.clientCache[host] = clientset
    d.cacheLock.Unlock()

    return clientset, nil
}

// Validate 验证K8s配置
func (d *K8sDeployer) Validate(ctx context.Context, info *models.WorkloadInfo) error {
    d.log.Info("Validating K8s workload", "namespace", info.Namespace, "workload", info.WorkloadName)
    
    client, err := d.getOrCreateClient(info.Kubeconfig)
    if err != nil {
        return fmt.Errorf("failed to get K8s client: %w", err)
    }

    // 检查命名空间是否存在
    _, err = client.CoreV1().Namespaces().Get(ctx, info.Namespace, metav1.GetOptions{})
    if err != nil {
        return fmt.Errorf("namespace not found: %s", info.Namespace)
    }

    // 检查Deployment/StatefulSet是否存在
    if info.WorkloadType == "Deployment" {
        _, err = client.AppsV1().Deployments(info.Namespace).Get(ctx, info.WorkloadName, metav1.GetOptions{})
    } else if info.WorkloadType == "StatefulSet" {
        _, err = client.AppsV1().StatefulSets(info.Namespace).Get(ctx, info.WorkloadName, metav1.GetOptions{})
    } else if info.WorkloadType == "DaemonSet" {
        _, err = client.AppsV1().DaemonSets(info.Namespace).Get(ctx, info.WorkloadName, metav1.GetOptions{})
    }

    if err != nil {
        return fmt.Errorf("workload not found: %s/%s", info.Namespace, info.WorkloadName)
    }

    d.log.Info("Validation successful")
    return nil
}

// Deploy 执行K8s部署
func (d *K8sDeployer) Deploy(ctx context.Context, info *models.WorkloadInfo, image string) error {
    d.log.Info("Starting K8s deployment",
        "namespace", info.Namespace,
        "workload", info.WorkloadName,
        "image", image,
    )

    client, err := d.getOrCreateClient(info.Kubeconfig)
    if err != nil {
        return err
    }

    // 1. 获取Deployment
    deployment, err := client.AppsV1().Deployments(info.Namespace).Get(ctx, info.WorkloadName, metav1.GetOptions{})
    if err != nil {
        return fmt.Errorf("failed to get deployment: %w", err)
    }

    // 2. 更新指定容器的镜像
    updated := false
    for i, container := range deployment.Spec.Template.Spec.Containers {
        if container.Name == info.ContainerName {
            deployment.Spec.Template.Spec.Containers[i].Image = image
            updated = true
            break
        }
    }

    if !updated {
        return fmt.Errorf("container '%s' not found in deployment", info.ContainerName)
    }

    // 3. 更新Deployment
    _, err = client.AppsV1().Deployments(info.Namespace).Update(ctx, deployment, metav1.UpdateOptions{})
    if err != nil {
        d.log.Error("Failed to update deployment", "error", err)
        return fmt.Errorf("failed to update deployment: %w", err)
    }

    d.log.Info("Deployment updated successfully")
    
    // 4. 等待Rollout完成(可选，最多等待5分钟)
    return d.waitForRollout(ctx, client, info, 300)
}

// waitForRollout 等待Rollout完成
func (d *K8sDeployer) waitForRollout(ctx context.Context, client kubernetes.Interface, 
    info *models.WorkloadInfo, timeoutSeconds int) error {
    
    d.log.Info("Waiting for rollout", "timeout", timeoutSeconds)
    
    // 实现等待逻辑...
    // 检查Pod ready状态，定期查询deployment.status.conditions
    
    return nil
}

// Rollback 回滚到上一个镜像
func (d *K8sDeployer) Rollback(ctx context.Context, info *models.WorkloadInfo, previousImage string) error {
    d.log.Info("Rolling back deployment", "previous_image", previousImage)
    return d.Deploy(ctx, info, previousImage)
}

// GetStatus 获取部署状态
func (d *K8sDeployer) GetStatus(ctx context.Context, info *models.WorkloadInfo) (string, error) {
    client, err := d.getOrCreateClient(info.Kubeconfig)
    if err != nil {
        return "error", err
    }

    deployment, err := client.AppsV1().Deployments(info.Namespace).Get(ctx, info.WorkloadName, metav1.GetOptions{})
    if err != nil {
        return "error", err
    }

    // 检查ready replicas
    if deployment.Status.ReadyReplicas == deployment.Status.Replicas {
        return "running", nil
    }

    return "pending", nil
}

// HealthCheck 健康检查
func (d *K8sDeployer) HealthCheck(ctx context.Context, info *models.WorkloadInfo) (bool, error) {
    status, err := d.GetStatus(ctx, info)
    if err != nil {
        return false, err
    }
    return status == "running", nil
}
```

### 3. 部署器工厂

```go
// internal/deployers/factory.go

package deployers

import (
    "fmt"
    "built-and-deploy/pkg/logger"
)

type DeployerFactory struct {
    log *logger.Logger
}

func NewDeployerFactory(log *logger.Logger) *DeployerFactory {
    return &DeployerFactory{log: log}
}

// CreateDeployer 根据集群类型获取对应的部署器
// 当前仅支持Kubernetes deployer
// Salt/Ansible等其他部署方式通过SSH执行Shell命令处理
func (f *DeployerFactory) CreateDeployer(clusterType string) (DeployStrategy, error) {
    switch clusterType {
    case "kubernetes":
        return NewK8sDeployer(f.log), nil
    default:
        return nil, fmt.Errorf("unsupported cluster type: %s (only 'kubernetes' is supported; use shell execution for other methods)", clusterType)
    }
}
```

### 4. Shell/SSH部署方式（Salt/Ansible等）

> **设计决策**: 不实现Salt/Ansible的DeployStrategy，而是通过ShellService执行SSH命令

**ShellService 架构**:

```go
// internal/services/shell_service.go

type ShellService struct {
    serverRepo    repository.ShellServerRepository    // SSH服务器配置
    commandRepo   repository.ShellCommandRepository   // 命令白名单
    execRepo      repository.ShellCommandExecutionRepository  // 执行记录
    clientCache   map[string]*ssh.Client              // SSH连接缓存
}

// 单服务器执行
func (s *ShellService) ExecuteCommand(
    ctx context.Context,
    commandID int,
    serverID int,
) (exitCode int, output string, err error)


**关键特性**:

1. **命令白名单**: 只能执行shell_command表中已发布的命令
   ```sql
   -- 命令定义示例
   INSERT INTO shell_command (server_id, command, description, is_published)
   VALUES (1, 'salt "prod-*" state.apply webserver', 'Apply webserver state', true);
   ```

2. **认证支持**: 密钥和密码均使用AES加密存储
   ```go
   // 密码认证
   server.AuthType = "password"
   server.Password = encryptedPassword  // AES加密
   
   // 密钥认证
   server.AuthType = "key"
   server.PrivateKey = encryptedPrivateKey  // AES加密
   ```

3. **连接缓存**: 避免频繁建立SSH连接
   ```go
   // 再次连接到同一服务器时复用缓存
   // 连接失败时自动清理并重试
   ```


**使用场景**:

```go
// 在发布流程中执行Salt或Ansible
exitCode, output, err := shellService.ExecuteCommand(
    ctx,
    commandID=5,   // 预定义的Salt命令
    serverID=2,    // Salt Master服务器
    nil,
)

// 在多个集群并行部署
results, err := shellService.ExecuteCommandParallel(
    ctx,
    commandID=10,  // Ansible部署命令
    serverIDs=[]int{clusterA, clusterB, clusterC},
    nil,
)
```

**详细指南**: 参考 `shell-service` skill 获取完整的API文档和集成示例

## 异步发布流程

### 发布生命周期

```
发布请求 (POST /api/v1/releases)
  │
  ├─> 1. 参数验证
  ├─> 2. 创建ReleaseRecord (status=pending)
  ├─> 3. 返回202 Accepted (立即)
  │
  └─> 4. 启动Goroutine执行以下步骤:
      │
      ├─ 记录事件: started
      ├─ 获取Deployer实例
      │
      ├─ 5. 执行Validate
      │   └─ 记录事件: validating
      │   └─ 成功/失败更新状态
      │
      ├─ 6. 执行Deploy
      │   └─ 记录事件: deploying
      │   └─ 记录多个步骤事件(pod_updated等)
      │
      ├─ 7. 执行HealthCheck
      │   └─ 记录事件: health_checking
      │   └─ 确认应用就绪
      │
      └─ 8. 更新ReleaseRecord (status=success/failed)
         └─ 记录最终事件
```

### 异步执行实现

```go
// internal/handlers/api_handlers.go

func CreateReleaseHandler(
    releaseRepo repository.ReleaseRecordRepository,
    workloadRepo *repository.WorkloadTargetRepository,
    deployer deployers.DeployStrategy,
    log *logger.Logger,
) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        requestID := middleware.GetRequestID(ctx)

        // 1. 解析请求
        var req models.ReleaseRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            respondError(w, 400, "Invalid request", err, requestID)
            return
        }

        // 2. 验证参数
        if req.AppID == 0 || req.EnvID == 0 || req.ClusterID == 0 || req.Image == "" {
            respondError(w, 400, "Missing required fields", nil, requestID)
            return
        }

        // 3. 获取工作负载配置
        workload, err := workloadRepo.GetByAppEnvCluster(ctx, req.AppID, req.EnvID, req.ClusterID)
        if err != nil {
            respondError(w, 404, "Workload target not found", err, requestID)
            return
        }

        // 4. 创建ReleaseRecord (status=pending)
        release := &models.ReleaseRecord{
            AppID:       req.AppID,
            EnvID:       req.EnvID,
            ClusterID:   req.ClusterID,
            Image:       req.Image,
            Status:      "pending",
            TriggeredBy: "system", // 实际应从auth获取
            StartedAt:   time.Now(),
        }

        releaseID, err := releaseRepo.Create(ctx, release)
        if err != nil {
            respondError(w, 500, "Failed to create release", err, requestID)
            return
        }

        release.ID = releaseID

        // 5. 立即返回202 Accepted
        w.WriteHeader(202)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "code":       0,
            "message":    "release accepted",
            "request_id": requestID,
            "data":       release,
        })

        // 6. 异步执行发布流程 (Goroutine)
        go func() {
            defer func() {
                if r := recover(); r != nil {
                    log.Error("Release panic", "panic", r, "release_id", releaseID)
                    releaseRepo.UpdateStatus(context.Background(), releaseID, "failed", "Internal error")
                }
            }()

            // 使用新context(不受请求超时影响)
            asyncCtx := context.Background()

            // 记录事件
            recordEvent := func(eventType, message string) {
                event := &models.ReleaseEvent{
                    ReleaseID:    releaseID,
                    EventType:    eventType,
                    EventMessage: message,
                    CreatedAt:    time.Now(),
                }
                // 保存事件到数据库...
                log.Info("Release event", "type", eventType, "message", message)
            }

            recordEvent("started", "Release started")

            // 7. 验证
            recordEvent("validating", fmt.Sprintf("Validating workload: %s/%s", workload.Namespace, workload.WorkloadName))
            if err := deployer.Validate(asyncCtx, &workload); err != nil {
                log.Error("Validation failed", "error", err, "release_id", releaseID)
                recordEvent("failed", fmt.Sprintf("Validation error: %v", err))
                releaseRepo.UpdateStatus(asyncCtx, releaseID, "failed", err.Error())
                return
            }

            // 8. 部署
            recordEvent("deploying", fmt.Sprintf("Deploying image: %s", req.Image))
            if err := deployer.Deploy(asyncCtx, &workload, req.Image); err != nil {
                log.Error("Deployment failed", "error", err, "release_id", releaseID)
                recordEvent("failed", fmt.Sprintf("Deployment error: %v", err))
                releaseRepo.UpdateStatus(asyncCtx, releaseID, "failed", err.Error())
                return
            }

            // 9. 健康检查
            recordEvent("health_checking", "Checking pod health")
            if healthy, err := deployer.HealthCheck(asyncCtx, &workload); err != nil || !healthy {
                log.Error("Health check failed", "error", err, "healthy", healthy)
                recordEvent("failed", "Health check failed")
                releaseRepo.UpdateStatus(asyncCtx, releaseID, "failed", "Health check failed")
                return
            }

            // 10. 成功
            recordEvent("success", "Release completed successfully")
            releaseRepo.UpdateStatus(asyncCtx, releaseID, "success", "")
            log.Info("Release succeeded", "release_id", releaseID)
        }()
    }
}
```

## 事件日志系统

### 事件类型

```go
// internal/models/release_event.go

const (
    // 发布生命周期事件
    EventStarted      = "started"
    EventValidating   = "validating"
    EventDeploying    = "deploying"
    EventPodUpdating  = "pod_updating"
    EventHealthCheck  = "health_checking"
    EventSuccess      = "success"
    EventFailed       = "failed"
    EventRolledBack   = "rolled_back"
)

type ReleaseEvent struct {
    ID           int       `json:"id" db:"id,primarykey"`
    ReleaseID    int       `json:"release_id" db:"release_id"`
    EventType    string    `json:"event_type" db:"event_type"`
    EventMessage string    `json:"event_message" db:"event_message"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
}
```

### 事件保存

```go
// 在发布流程中记录事件
recordEvent := func(eventType, message string) {
    event := &models.ReleaseEvent{
        ReleaseID:    releaseID,
        EventType:    eventType,
        EventMessage: message,
        CreatedAt:    time.Now(),
    }
    
    if err := releaseEventRepo.Create(asyncCtx, event); err != nil {
        log.Error("Failed to record event", "error", err)
        // 继续执行，不能因为事件记录失败而中断发布
    }
}
```

## 并发控制

### Goroutine 安全

1. **上下文隔离**: Goroutine中使用context.Background()，不受HTTP请求超时影响
2. **Panic恢复**: defer recover()防止崩溃
3. **客户端缓存**: 使用sync.RWMutex保护map访问
4. **定时器清理**: 明确cancel context，避免资源泄漏

### 最佳实践

```go
// 1. 使用context.Background()隔离Goroutine
go func() {
    asyncCtx := context.Background()
    // 执行长时间操作
}()

// 2. 使用defer恢复panic
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Error("Panic", "panic", r)
        }
    }()
    // 执行操作
}()

// 3. 使用sync.RWMutex保护共享资源
type Cache struct {
    mu    sync.RWMutex
    items map[string]interface{}
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.items[key]
    return val, ok
}
```

