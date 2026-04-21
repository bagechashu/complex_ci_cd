---
name: service-layer
description: 发布控制系统 - 简化的业务服务层（Service Helper函数方式）
keywords: Service函数, 业务逻辑, 事务管理, Handler协调, Go习惯
---

# 简化的Service层指南（Service Helper函数方式）

## 核心理念 ⭐ 已简化

**目标**: 摒弃复杂的Application Service类，采用 Go 风格的 **Service Helper 函数**。

### 为什么简化？

```
❌ 之前（Java风格，过度抽象）      ✅ 现在（Go风格，简单直白）
────────────────────────────────────────────────────
Handler → ApplicationService         Handler → ServiceHelper函数
  ↓         ├─ coordinate             ↓    
  └─────────├─ validate               └──── 直接调用Repository
            ├─ transaction                   
            └─ Repository            清晰、直白、没有多余中间层
            
问题（之前）：                     改进（现在）：
- 多一层无必要的抽象              - 直接的函数调用
- 新人需要理解"是什么"             - 新人一眼看清逻辑流
- 代码跳转多，理解成本高           - Go的"简即是美"
- 测试需要mock多个层              - 测试只需mock repository
```

## 新架构对比

```
层级        之前（ApplicationService）          现在（ServiceHelper）
───────────────────────────────────────────────────────────────────
Handler      POST /releases                    POST /releases
             └─ handlers.go                    └─ handlers.go
                 
 中间层       services.ReleaseApplicationService ❌ 删除此层
             ├─ application/services/          
             ├─ application/dto/               
             ├─ domain/release/services/       
             └─ infrastructure/.../repositories/
             
Service      ❌ 复杂的类 + DI容器              ✅ 简单的helper函数
             ├─ ReleaseApplicationService     ├─ ReleaseHelper()
             ├─ ApplicationApplicationService ├─ ApplicationHelper()
             ├─ ClusterApplicationService     ├─ ClusterHelper()
             └─ WorkloadApplicationService    └─ WorkloadHelper()
             
Domain       ✅ DDD聚合根保留                  ✅ DDD聚合根保留
             ├─ domain/release/aggregates/   ├─ domain/release/aggregates/
             ├─ domain/release/services/      ├─ domain/release/services/
             ├─ domain/release/repositories/  ├─ domain/release/repositories/
             └─ value_objects/                └─ value_objects/
             
Repository   ✅ 保留所有repository接口          ✅ 保留所有repository接口
             └─ infrastructure/.../repositories/
             
DB          SQLite                           SQLite
```

## 三类Service Helper函数

### 1. 业务操作函数（Business Logic）

**位置**: `internal/services/release_helper.go`

**职责**: 完整的业务操作，涉及多个repository协调

**示例**:

```go
package services

// ReleaseApp 执行一次完整的应用发布
// 返回发布记录ID，交由前端轮询查询详情
func ReleaseApp(ctx context.Context, 
    appID, envID, clusterID int, 
    image, operator string,
    deps *ServiceDeps) (int, error) {
    
    // 1. 参数验证
    if appID <= 0 || image == "" {
        return 0, fmt.Errorf("invalid input: appID=%d, image=%s", appID, image)
    }
    
    // 2. 获取部署目标
    workload, err := deps.workloadRepo.GetByAppEnvCluster(ctx, appID, envID, clusterID)
    if err != nil {
        return 0, fmt.Errorf("workload not found: %w", err)
    }
    
    // 3. 获取集群信息
    cluster, err := deps.clusterRepo.ByID(ctx, clusterID)
    if err != nil {
        return 0, fmt.Errorf("cluster not found: %w", err)
    }
    
    // 4. 创建发布记录
    release := &models.ReleaseRecord{
        AppID:       appID,
        EnvID:       envID,
        ClusterID:   clusterID,
        Image:       image,
        Status:      "pending",
        TriggeredBy: operator,
        StartedAt:   time.Now(),
    }
    
    if err := deps.releaseRepo.Save(ctx, release); err != nil {
        return 0, fmt.Errorf("failed to save release: %w", err)
    }
    
    deps.log.Info("Release created", 
        "releaseID", release.ID, 
        "appID", appID, 
        "image", image)
    
    // 5. 返回发布ID，异步执行部署
    go deployAsync(context.Background(), release.ID, cluster.Type, workload, image, deps)
    
    return release.ID, nil
}

// GetReleaseStatus 查询发布状态
func GetReleaseStatus(ctx context.Context, 
    releaseID int, 
    deps *ServiceDeps) (*models.ReleaseRecord, error) {
    
    release, err := deps.releaseRepo.ByID(ctx, releaseID)
    if err != nil {
        return nil, fmt.Errorf("release not found: %w", err)
    }
    return release, nil
}
```

---

### 2. 验证函数（Validation）

**位置**: `internal/services/validation.go`

**职责**: 参数验证、业务规则检查

**示例**:

```go
// ValidateImageFormat 验证镜像格式
func ValidateImageFormat(image string) error {
    // 格式: [registry/]repository[:tag]
    // 例: harbor.example.com/company/api-service:v1.2.3
    
    parts := strings.Split(image, "/")
    if len(parts) < 2 {
        return fmt.Errorf("invalid image: missing repository")
    }
    
    repo := parts[len(parts)-1]
    if !strings.Contains(repo, ":") {
        return fmt.Errorf("invalid image: tag required (e.g., image:v1.0)")
    }
    
    return nil
}

// CanRollback 检查是否可以回滚
func CanRollback(ctx context.Context, 
    releaseID int, 
    deps *ServiceDeps) (bool, error) {
    
    release, err := deps.releaseRepo.ByID(ctx, releaseID)
    if err != nil {
        return false, fmt.Errorf("release not found: %w", err)
    }
    
    // 规则: 只能回滚已成功的发布
    if release.Status != "success" {
        return false, fmt.Errorf("release status is %s, not success", release.Status)
    }
    
    if release.PreviousImage == "" {
        return false, fmt.Errorf("no previous image to rollback")
    }
    
    return true, nil
}
```

---

### 3. 事务管理函数（Transaction Helpers）

**位置**: `internal/services/transaction.go`

**职责**: 事务边界管理

**示例**:

```go
// WithTx 事务包装器
func WithTx(ctx context.Context, db *sql.DB, fn func(context.Context) error) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    if err := fn(ctx); err != nil {
        return err
    }
    
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    
    return nil
}
```

---

## Handler层如何调用Service Helper

### 设计原则

1. **Handler**: HTTP处理、参数解析、错误包装  
2. **ServiceHelper**: 业务逻辑、repository协调  
3. **Repository**: 数据访问  

**不再有ApplicationService中间层**

### Handler示例

```go
// internal/handlers/releases/handler.go

type ReleaseHandler struct {
    deps *services.ServiceDeps
    log  *logger.Logger
}

// POST /api/v1/releases - 创建发布
func (h *ReleaseHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req struct {
        AppID     int    `json:"app_id"`
        EnvID     int    `json:"env_id"`
        ClusterID int    `json:"cluster_id"`
        Image     string `json:"image"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    
    // 直接调用Service Helper（不经过ApplicationService）
    releaseID, err := services.ReleaseApp(
        r.Context(),
        req.AppID, req.EnvID, req.ClusterID, req.Image,
        getUser(r),
        h.deps,
    )
    if err != nil {
        h.log.Error("release failed", "error", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "code": 0,
        "data": map[string]int{"id": releaseID},
    })
}

// GET /api/v1/releases/{id} - 查询发布状态
func (h *ReleaseHandler) Get(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.Atoi(chi.URLParam(r, "id"))
    
    // 简单查询：直接调用repository，无需service helper
    release, err := h.deps.releaseRepo.ByID(r.Context(), id)
    if err != nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "code": 0,
        "data": release,
    })
}
```

---

## ServiceDeps 依赖结构

```go
// internal/services/deps.go

type ServiceDeps struct {
    // Repositories
    releaseRepo   ReleaseRepository
    appRepo       ApplicationRepository
    clusterRepo   ClusterRepository
    workloadRepo  WorkloadTargetRepository
    eventRepo     ReleaseEventRepository
    
    // Infrastructure
    deployer      DeployerFactory
    log           *logger.Logger
    db            *sql.DB
}

// 在main.go中一次性创建
func main() {
    db, _ := database.Init(cfg.DatabasePath)
    repos := createRepositories(db)
    
    deps := &services.ServiceDeps{
        releaseRepo:  repos.Release,
        appRepo:      repos.App,
        workloadRepo: repos.Workload,
        deployer:     deployers.NewDeployerFactory(log),
        log:          log,
        db:           db,
    }
    
    // 传递给handler
    mux.HandleFunc("/api/v1/releases", 
        handlers.NewReleaseHandler(deps).Create)
}
```

---

## 错误处理（简单直白）

**原则**: 不要包装成ServiceError，直接返回Go error

```go
// ✅ 推荐（Go习惯）
func GetApplication(ctx context.Context, id int, repo ApplicationRepository) (*Application, error) {
    app, err := repo.ByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get application: %w", err)  // 直接wrap
    }
    if app == nil {
        return nil, fmt.Errorf("application %d not found", id)
    }
    return app, nil
}

// 然后在Handler里正常处理
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

---

## 何时使用Service Helper？

| 场景 | 使用Helper? | 理由 |
|------|-----------|------|
| 简单GET | ❌ | 一条SQL，直接调用repository |
| 创建+关系检查 | ✅ | 多个repository协调 |
| 发布操作 | ✅ | 复杂业务逻辑 |
| 删除 | ❌ | 一条SQL |
| 更新+验证 | ✅ | 需要前置检查 |

---

## 改进对比

| 维度 | 之前（ApplicationService） | 现在（ServiceHelper） |
|------|---------------------------|----------------------|
| 代码层数 | 5层 Handler→AppService→DomainService→Repo→DB | 3层 Handler→Helper→Repo→DB |
| 新手理解成本 | ⭐⭐⭐⭐ 中等 | ⭐⭐ 低 |
| 跳转复杂度 | 多次跳转 | 直接调用 |
| 测试复杂度 | Mock多个层 | Mock repository + deps |
| Go习惯度 | ⭐⭐☆ 偏离 | ⭐⭐⭐⭐⭐ 完全符合 |
| 代码量 | 较多 | 较少 |

**结论**: Service Helper 函数方式是Go的最佳实践。


---

## Service 架构

### 1. ReleaseService（核心业务）

#### 职责
- 发布流程控制（验证 → 部署 → 记录事件）
- 发布状态管理
- 回滚逻辑
- 权限检查

#### 方法设计

```go
// services/release_service.go

type ReleaseService struct {
    releaseRepo ReleaseRecordRepository
    workloadRepo WorkloadTargetRepository
    clusterRepo ClusterRepository
    appRepo ApplicationRepository
    deployer DeployerFactory
    eventRepo ReleaseEventRepository
    log *logger.Logger
    db *sql.DB  // 用于事务
}

// 新建发布
func (s *ReleaseService) Release(ctx context.Context, req *ReleaseRequest) (*ReleaseResponse, error) {
    // 1. 权限检查
    if !hasPermission(ctx, req.EnvID) {
        return nil, NewServiceError("PERMISSION_DENIED", "No permission to deploy to this environment")
    }
    
    // 2. 获取部署目标
    workload, err := s.workloadRepo.GetByAppEnvCluster(ctx, req.AppID, req.EnvID, req.ClusterID)
    if err != nil {
        return nil, NewServiceError("NOT_FOUND", "Deployment target not found")
    }
    
    // 3. 镜像验证
    if !isValidImage(req.Image) {
        return nil, NewServiceError("INVALID_INPUT", "Invalid image format")
    }
    
    // 4. 启动事务
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, NewServiceError("DATABASE", "Failed to start transaction", err)
    }
    defer tx.Rollback()
    
    // 5. 创建发布记录（status=pending）
    release := &models.ReleaseRecord{
        AppID: req.AppID,
        EnvID: req.EnvID,
        ClusterID: req.ClusterID,
        Image: req.Image,
        Status: "pending",
        TriggeredBy: getUser(ctx),
        StartedAt: time.Now(),
    }
    releaseID, err := s.releaseRepo.Create(ctx, release)
    if err != nil {
        return nil, NewServiceError("DATABASE", "Failed to create release record", err)
    }
    
    // 6. 记录事件：started
    s.recordEvent(ctx, releaseID, "started", "Release process started")
    
    // 7. 提交事务
    if err := tx.Commit(); err != nil {
        return nil, NewServiceError("DATABASE", "Failed to commit transaction", err)
    }
    
    // 8. 异步部署（goroutine）
    go s.deployAsync(context.Background(), releaseID, workload)
    
    // 9. 返回响应
    return &ReleaseResponse{
        ID: releaseID,
        Status: "pending",
        StartedAt: release.StartedAt,
    }, nil
}

// 异步部署（在goroutine中运行）
func (s *ReleaseService) deployAsync(ctx context.Context, releaseID string, workload *models.WorkloadTarget) {
    // 更新状态为validating
    s.releaseRepo.UpdateStatus(ctx, releaseID, "validating")
    s.recordEvent(ctx, releaseID, "validating", "Validating deployment...")
    
    // 获取Deployer
    release, _ := s.releaseRepo.GetByID(ctx, releaseID)
    cluster, _ := s.clusterRepo.GetByID(ctx, release.ClusterID)
    
    deployer := s.deployer.CreateDeployer(cluster.Type)
    
    // 验证
    if err := deployer.Validate(ctx, workload); err != nil {
        s.releaseRepo.UpdateStatus(ctx, releaseID, "failed")
        s.recordEvent(ctx, releaseID, "failed", fmt.Sprintf("Validation failed: %v", err))
        return
    }
    
    // 部署
    s.releaseRepo.UpdateStatus(ctx, releaseID, "deploying")
    s.recordEvent(ctx, releaseID, "deploying", "Deploying...")
    
    if err := deployer.Deploy(ctx, workload); err != nil {
        s.releaseRepo.UpdateStatus(ctx, releaseID, "failed")
        s.recordEvent(ctx, releaseID, "failed", fmt.Sprintf("Deployment failed: %v", err))
        return
    }
    
    // 成功
    s.releaseRepo.UpdateStatus(ctx, releaseID, "success")
    s.recordEvent(ctx, releaseID, "success", "Deployment completed successfully")
}

// 回滚
func (s *ReleaseService) Rollback(ctx context.Context, releaseID string) error {
    // 权限检查
    if !hasPermission(ctx, "") {  // 回滚需要特殊权限
        return NewServiceError("PERMISSION_DENIED", "No permission to rollback")
    }
    
    // 获取发布记录和上一版本
    release, err := s.releaseRepo.GetByID(ctx, releaseID)
    if err != nil {
        return NewServiceError("NOT_FOUND", "Release not found")
    }
    
    if release.PreviousImage == "" {
        return NewServiceError("INVALID_STATE", "No previous version to rollback")
    }
    
    // 创建新的回滚发布
    rollback := &models.ReleaseRecord{
        AppID: release.AppID,
        EnvID: release.EnvID,
        ClusterID: release.ClusterID,
        Image: release.PreviousImage,
        Status: "deploying",
        TriggeredBy: getUser(ctx),
        StartedAt: time.Now(),
    }
    
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    rollbackID, _ := s.releaseRepo.Create(ctx, rollback)
    s.recordEvent(ctx, rollbackID, "rollback_started", "Rolling back to previous version")
    tx.Commit()
    
    // 异步执行回滚
    go s.deployAsync(context.Background(), rollbackID, nil)
    
    return nil
}

// 记录事件
func (s *ReleaseService) recordEvent(ctx context.Context, releaseID, eventType, message string) {
    event := &models.ReleaseEvent{
        ReleaseID: releaseID,
        EventType: eventType,
        EventMessage: message,
        CreatedAt: time.Now(),
    }
    s.eventRepo.Create(ctx, event)
    s.log.Info("Release event recorded", "releaseID", releaseID, "event", eventType)
}
```

---

### 2. ApplicationService

#### 职责
- 应用的CRUD操作
- 应用名称唯一性检查
- 应用删除前的依赖检查（有发布记录不能删除）

#### 关键方法

```go
// services/application_service.go

type ApplicationService struct {
    appRepo ApplicationRepository
    releaseRepo ReleaseRecordRepository
    log *logger.Logger
}

// 创建应用
func (s *ApplicationService) CreateApplication(ctx context.Context, req *CreateAppRequest) (*models.Application, error) {
    // 验证输入
    if req.Name == "" || len(req.Name) > 100 {
        return nil, NewServiceError("INVALID_INPUT", "Application name must be 1-100 characters")
    }
    
    if !isValidURL(req.Repository) {
        return nil, NewServiceError("INVALID_INPUT", "Invalid repository URL")
    }
    
    // 检查名称唯一性
    existing, _ := s.appRepo.GetByName(ctx, req.Name)
    if existing != nil {
        return nil, NewServiceError("CONFLICT", fmt.Sprintf("Application '%s' already exists", req.Name))
    }
    
    // 创建
    app := &models.Application{
        Name: req.Name,
        Repository: req.Repository,
        BuildType: req.BuildType,
        Description: req.Description,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    id, err := s.appRepo.Create(ctx, app)
    if err != nil {
        return nil, NewServiceError("DATABASE", "Failed to create application", err)
    }
    
    app.ID = id
    s.log.Info("Application created", "id", id, "name", req.Name)
    return app, nil
}

// 删除应用（需要检查依赖）
func (s *ApplicationService) DeleteApplication(ctx context.Context, appID string) error {
    // 权限检查（仅admin）
    if !isAdmin(ctx) {
        return NewServiceError("PERMISSION_DENIED", "Only admins can delete applications")
    }
    
    // 检查是否有发布记录
    releases, _ := s.releaseRepo.ListByApp(ctx, appID)
    if len(releases) > 0 {
        return NewServiceError("CONFLICT", 
            "Cannot delete application with existing release records. Archive the releases first.")
    }
    
    // 删除
    err := s.appRepo.Delete(ctx, appID)
    if err != nil {
        return NewServiceError("DATABASE", "Failed to delete application", err)
    }
    
    s.log.Info("Application deleted", "id", appID)
    return nil
}
```

---

### 3. ClusterService

#### 职责
- 集群的CRUD操作
- 连接验证（kubeconfig是否有效）
- 健康检查

#### 关键方法

```go
// services/cluster_service.go

type ClusterService struct {
    clusterRepo ClusterRepository
    deployer DeployerFactory
    log *logger.Logger
}

// 创建集群
func (s *ClusterService) CreateCluster(ctx context.Context, req *CreateClusterRequest) (*models.Cluster, error) {
    // 验证输入
    if req.Type != "kubernetes" {
        return nil, NewServiceError("INVALID_INPUT", 
            "Only 'kubernetes' cluster type is supported. Use Shell execution for other methods.")
    }
    
    // 测试连接（同步）
    if err := s.testConnection(ctx, req.Kubeconfig); err != nil {
        return nil, NewServiceError("CONNECTION_FAILED", 
            fmt.Sprintf("Failed to connect to cluster: %v", err))
    }
    
    // 加密kubeconfig
    encryptedConfig, err := encryptAES(req.Kubeconfig, os.Getenv("ENCRYPTION_KEY"))
    if err != nil {
        return nil, NewServiceError("ENCRYPTION", "Failed to encrypt kubeconfig")
    }
    
    // 创建
    cluster := &models.Cluster{
        Name: req.Name,
        Type: req.Type,
        Kubeconfig: encryptedConfig,  // 加密存储
        Status: "connected",
        CreatedAt: time.Now(),
    }
    
    id, _ := s.clusterRepo.Create(ctx, cluster)
    cluster.ID = id
    s.log.Info("Cluster created", "id", id, "name", req.Name)
    return cluster, nil
}

// 测试连接
func (s *ClusterService) TestConnection(ctx context.Context, clusterID string) error {
    cluster, err := s.clusterRepo.GetByID(ctx, clusterID)
    if err != nil {
        return NewServiceError("NOT_FOUND", "Cluster not found")
    }
    
    // 解密kubeconfig
    kubeconfig, err := decryptAES(cluster.Kubeconfig, os.Getenv("ENCRYPTION_KEY"))
    if err != nil {
        return NewServiceError("DECRYPTION", "Failed to decrypt kubeconfig")
    }
    
    // 测试连接
    if err := s.testConnection(ctx, kubeconfig); err != nil {
        return NewServiceError("CONNECTION_FAILED", fmt.Sprintf("Connection test failed: %v", err))
    }
    
    // 更新状态
    s.clusterRepo.UpdateStatus(ctx, clusterID, "connected")
    return nil
}

func (s *ClusterService) testConnection(ctx context.Context, kubeconfig string) error {
    deployer := s.deployer.CreateDeployer("kubernetes")
    return deployer.HealthCheck(ctx)
}
```

---

### 4. WorkloadService

#### 职责
- 部署目标的CRUD
- 关系验证（app/env/cluster都必须存在）
- 唯一性检查（app + env + cluster）

#### 关键方法

```go
// services/workload_service.go

type WorkloadService struct {
    workloadRepo WorkloadTargetRepository
    appRepo ApplicationRepository
    envRepo EnvironmentRepository
    clusterRepo ClusterRepository
    log *logger.Logger
}

// 创建部署目标
func (s *WorkloadService) CreateWorkloadTarget(ctx context.Context, req *CreateWorkloadRequest) (*models.WorkloadTarget, error) {
    // 验证应用存在
    app, err := s.appRepo.GetByID(ctx, req.AppID)
    if err != nil || app == nil {
        return nil, NewServiceError("NOT_FOUND", "Application not found")
    }
    
    // 验证环境存在
    env, err := s.envRepo.GetByID(ctx, req.EnvID)
    if err != nil || env == nil {
        return nil, NewServiceError("NOT_FOUND", "Environment not found")
    }
    
    // 验证集群存在
    cluster, err := s.clusterRepo.GetByID(ctx, req.ClusterID)
    if err != nil || cluster == nil {
        return nil, NewServiceError("NOT_FOUND", "Cluster not found")
    }
    
    // 检查唯一性（app + env + cluster）
    existing, _ := s.workloadRepo.GetByAppEnvCluster(ctx, req.AppID, req.EnvID, req.ClusterID)
    if existing != nil {
        return nil, NewServiceError("CONFLICT", 
            "Workload target for this app/environment/cluster combination already exists")
    }
    
    // 创建
    workload := &models.WorkloadTarget{
        AppID: req.AppID,
        EnvID: req.EnvID,
        ClusterID: req.ClusterID,
        Namespace: req.Namespace,
        WorkloadName: req.WorkloadName,
        WorkloadType: req.WorkloadType,
        ContainerName: req.ContainerName,
        RegistryDomain: req.RegistryDomain,
        ImageRepo: req.ImageRepo,
        CreatedAt: time.Now(),
    }
    
    id, _ := s.workloadRepo.Create(ctx, workload)
    workload.ID = id
    s.log.Info("Workload target created", "id", id, "app", req.AppID, "env", req.EnvID)
    return workload, nil
}
```

---

## 错误处理模型

### 统一的错误类型

```go
// pkg/errors/errors.go

type ServiceError struct {
    Code    string        // 业务错误码
    Message string        // 用户友好的消息
    Err     error         // 原始错误（用于日志）
    Status  int           // HTTP状态码（可选）
}

func (e *ServiceError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

// 工厂方法
func NewServiceError(code, message string) *ServiceError {
    return &ServiceError{
        Code: code,
        Message: message,
        Status: mapCodeToStatus(code),
    }
}

func NewServiceErrorWithCause(code, message string, err error) *ServiceError {
    return &ServiceError{
        Code: code,
        Message: message,
        Err: err,
        Status: mapCodeToStatus(code),
    }
}

// 映射错误码到HTTP状态码
func mapCodeToStatus(code string) int {
    mapping := map[string]int{
        "NOT_FOUND": 404,
        "PERMISSION_DENIED": 403,
        "INVALID_INPUT": 400,
        "CONFLICT": 409,
        "CONNECTION_FAILED": 503,
        "DATABASE": 500,
        "ENCRYPTION": 500,
    }
    
    if status, ok := mapping[code]; ok {
        return status
    }
    return 500
}
```

---

## 依赖注入模式

### 服务容器

```go
// internal/container/container.go

type ServiceContainer struct {
    // Repositories
    applicationRepo ApplicationRepository
    clusterRepo ClusterRepository
    releaseRepo ReleaseRecordRepository
    workloadRepo WorkloadTargetRepository
    eventRepo ReleaseEventRepository
    
    // Services
    releaseService *ReleaseService
    appService *ApplicationService
    clusterService *ClusterService
    workloadService *WorkloadService
    
    // Dependencies
    deployer DeployerFactory
    db *sql.DB
    log *logger.Logger
}

// 初始化容器
func NewServiceContainer(db *sql.DB, log *logger.Logger) *ServiceContainer {
    // 初始化Repositories
    appRepo := repository.NewApplicationRepository(db)
    clusterRepo := repository.NewClusterRepository(db)
    releaseRepo := repository.NewReleaseRecordRepository(db)
    workloadRepo := repository.NewWorkloadTargetRepository(db)
    eventRepo := repository.NewReleaseEventRepository(db)
    
    // 初始化Services
    return &ServiceContainer{
        applicationRepo: appRepo,
        clusterRepo: clusterRepo,
        releaseRepo: releaseRepo,
        workloadRepo: workloadRepo,
        eventRepo: eventRepo,
        
        releaseService: &ReleaseService{
            releaseRepo: releaseRepo,
            workloadRepo: workloadRepo,
            clusterRepo: clusterRepo,
            appRepo: appRepo,
            deployer: deployers.NewDeployerFactory(log),
            eventRepo: eventRepo,
            log: log,
            db: db,
        },
        appService: &ApplicationService{
            appRepo: appRepo,
            releaseRepo: releaseRepo,
            log: log,
        },
        clusterService: &ClusterService{
            clusterRepo: clusterRepo,
            deployer: deployers.NewDeployerFactory(log),
            log: log,
        },
        workloadService: &WorkloadService{
            workloadRepo: workloadRepo,
            appRepo: appRepo,
            clusterRepo: clusterRepo,
            log: log,
        },
    }
}

// 获取Service
func (c *ServiceContainer) ReleaseService() *ReleaseService {
    return c.releaseService
}

func (c *ServiceContainer) ApplicationService() *ApplicationService {
    return c.appService
}

// ... 其他service getter
```

---

## Handler集成

### 使用Service

```go
// handlers/api_handlers.go

func ListApplicationsHandler(appService *ApplicationService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        requestID := middleware.GetRequestID(ctx)
        
        // 调用Service
        apps, err := appService.ListApplications(ctx)
        if err != nil {
            if svcErr, ok := err.(*ServiceError); ok {
                respondError(w, svcErr.Status, svcErr.Message, svcErr.Err, requestID)
            } else {
                respondError(w, 500, "Internal server error", err, requestID)
            }
            return
        }
        
        respondJSON(w, 200, apps)
    }
}

func CreateApplicationHandler(appService *ApplicationService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        requestID := middleware.GetRequestID(ctx)
        
        // 解析请求
        var req CreateApplicationRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            respondError(w, 400, "Invalid request body", err, requestID)
            return
        }
        
        // 调用Service
        app, err := appService.CreateApplication(ctx, &req)
        if err != nil {
            if svcErr, ok := err.(*ServiceError); ok {
                respondError(w, svcErr.Status, svcErr.Message, svcErr.Err, requestID)
            } else {
                respondError(w, 500, "Internal server error", err, requestID)
            }
            return
        }
        
        respondJSON(w, 201, app)
    }
}
```

---

## 测试

### Service单元测试

```go
// services/release_service_test.go

func TestReleaseService_Release(t *testing.T) {
    // 准备Mock Repositories
    mockReleaseRepo := &MockReleaseRecordRepository{}
    mockWorkloadRepo := &MockWorkloadTargetRepository{}
    
    svc := &ReleaseService{
        releaseRepo: mockReleaseRepo,
        workloadRepo: mockWorkloadRepo,
        // ... 其他依赖
    }
    
    // 测试用例
    t.Run("successful release", func(t *testing.T) {
        req := &ReleaseRequest{
            AppID: "1",
            EnvID: "1",
            ClusterID: "1",
            Image: "harbor.com/app:v1.0.0",
        }
        
        resp, err := svc.Release(context.Background(), req)
        
        assert.NoError(t, err)
        assert.NotNil(t, resp)
        assert.Equal(t, "pending", resp.Status)
    })
    
    t.Run("permission denied", func(t *testing.T) {
        req := &ReleaseRequest{
            AppID: "1",
            EnvID: "99",  // 无权限环境
            ClusterID: "1",
        }
        
        err := svc.Release(context.Background(), req)
        
        assert.Error(t, err)
        svcErr := err.(*ServiceError)
        assert.Equal(t, "PERMISSION_DENIED", svcErr.Code)
    })
}
```

---

## 最佳实践

### ✅ 做这些事

1. **Service中集中业务规则**
   ```go
   // ✅ 好的做法
   func (s *ReleaseService) Release() {
       // 权限检查 - Service的职责
       // 验证关系 - Service的职责
       // 事务管理 - Service的职责
   }
   ```

2. **事务必须在Service包裹**
   ```go
   // ✅ 好的做法
   func (s *ReleaseService) Release() {
       tx := s.db.BeginTx(ctx)
       defer tx.Rollback()
       // ... 所有操作
       tx.Commit()
   }
   ```

3. **使用统一的错误类型**
   ```go
   // ✅ 好的做法
   return NewServiceError("CONFLICT", "Resource already exists")
   ```

4. **异步操作使用goroutine**
   ```go
   // ✅ 好的做法
   go s.deployAsync(context.Background(), releaseID, workload)
   ```

### ❌ 避免这些事

1. **不要在Service中处理HTTP细节**
   ```go
   // ❌ 不要
   func (s *ReleaseService) Release(w http.ResponseWriter) {
       w.Header().Set("Content-Type", "application/json")
   }
   ```

2. **不要直接返回Repository错误**
   ```go
   // ❌ 不要
   return releaseRepo.Create(ctx, release)
   
   // ✅ 应该
   if err := releaseRepo.Create(ctx, release); err != nil {
       return NewServiceErrorWithCause("DATABASE", "Failed to create", err)
   }
   ```

3. **不要跳过验证**
   ```go
   // ❌ 不要
   release := &ReleaseRecord{...}
   s.releaseRepo.Create(ctx, release)
   
   // ✅ 应该
   if err := s.validateRelease(release); err != nil {
       return NewServiceError("INVALID_INPUT", err.Error())
   }
   ```

4. **不要在Service中调用其他Service**
   ```go
   // ❌ 不要（造成循环依赖）
   func (s *ReleaseService) Release() {
       s.appService.ValidateApp()  // 不要这样
   }
   
   // ✅ 应该
   func (s *ReleaseService) Release() {
       s.appRepo.GetByID()  // 直接调用repo
   }
   ```

---

## 实施清单

- [ ] 创建 `services/` 目录
- [ ] 定义错误类型 `pkg/errors/service_error.go`
- [ ] 实现 `ReleaseService`（最核心）
- [ ] 实现 `ApplicationService`
- [ ] 实现 `ClusterService`
- [ ] 实现 `WorkloadService`
- [ ] 创建 `ServiceContainer`
- [ ] 更新 Handler 调用 Service 而非 Repository
- [ ] 添加单元测试
- [ ] 集成测试
