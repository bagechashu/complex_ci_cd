---
name: service-layer
description: 发布控制系统 - Service 类 + DI 容器架构
keywords: Service类, DI容器, ServiceContainer, 依赖注入, Go习惯
---

# Service 层指南（Service 类 + DI 容器）

## 核心理念 ⭐ 与实际实现一致

**目标**: 采用 **Service 类 + ServiceContainer** 的依赖注入模式。

### 架构对比

```
当前项目：Service 类 + DI 容器（推荐）
───────────────────────────────────
Handler
  └─ 从容器获取 Service 实例
     └─ Service 方法包含业务逻辑
        └─ Repository (接口)
           └─ SQLite 实现

优势：
✅ 依赖集中在结构体，参数不会过多
✅ Service 间可相互调用
✅ DI 容器统一管理生命周期
✅ 接口注入，便于单元测试
✅ 适合中大型复杂业务
```

---

## Service 类的实现

### 1. ReleaseService（核心业务）

```go
// internal/services/release_service.go

type ReleaseService struct {
    releaseRepo   repository.ReleaseRecordRepository    // 发布记录
    workloadRepo  repository.WorkloadTargetRepository   // 部署目标
    clusterRepo   repository.ClusterRepository          // 集群信息
    appRepo       repository.ApplicationRepository      // 应用信息
    eventRepo     repository.ReleaseEventRepository     // 事件日志
    deployerFact  *deployers.DeployerFactory           // 部署器工厂
    log           *logger.Logger                        // 日志
    db            interface{}                           // 数据库（事务用）
}

// 构造函数
func NewReleaseService(
    releaseRepo repository.ReleaseRecordRepository,
    workloadRepo repository.WorkloadTargetRepository,
    clusterRepo repository.ClusterRepository,
    appRepo repository.ApplicationRepository,
    eventRepo repository.ReleaseEventRepository,
    deployerFact *deployers.DeployerFactory,
    log *logger.Logger,
    db interface{},
) *ReleaseService {
    return &ReleaseService{
        releaseRepo:  releaseRepo,
        workloadRepo: workloadRepo,
        clusterRepo:  clusterRepo,
        appRepo:      appRepo,
        eventRepo:    eventRepo,
        deployerFact: deployerFact,
        log:          log,
        db:           db,
    }
}

// Release 发布应用 - 主流程
func (s *ReleaseService) Release(ctx context.Context, appID, clusterID int, image string) (*models.ReleaseRecord, error) {
    // 1. 参数验证
    if appID <= 0 || image == "" {
        return nil, fmt.Errorf("invalid input: appID=%d, image=%s", appID, image)
    }
    
    // 2. 获取部署目标
    workload, err := s.workloadRepo.GetByAppCluster(ctx, appID, clusterID)
    if err != nil {
        return nil, fmt.Errorf("workload not found: %w", err)
    }
    
    // 3. 创建发布记录
    release := &models.ReleaseRecord{
        AppID:       appID,
        ClusterID:   clusterID,
        Image:       image,
        Status:      "pending",
        TriggeredBy: "system",
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    if err := s.releaseRepo.Create(ctx, release); err != nil {
        return nil, fmt.Errorf("failed to create release: %w", err)
    }
    
    s.log.Info("Release created", "releaseID", release.ID, "appID", appID, "image", image)
    
    // 4. 异步执行部署
    go s.deployAsync(context.Background(), release.ID, workload, image)
    
    return release, nil
}

// deployAsync 异步执行部署
func (s *ReleaseService) deployAsync(ctx context.Context, releaseID int, workload *models.WorkloadTarget, image string) {
    cluster, err := s.clusterRepo.GetByID(ctx, workload.ClusterID)
    if err != nil {
        s.recordEvent(ctx, releaseID, "deploy_failed", fmt.Sprintf("Cluster not found: %v", err))
        return
    }
    
    deployer := s.deployerFact.Create(cluster.Type)
    
    if err := deployer.Deploy(ctx, workload, image); err != nil {
        s.recordEvent(ctx, releaseID, "deploy_failed", fmt.Sprintf("Error: %v", err))
        return
    }
    
    s.recordEvent(ctx, releaseID, "deploy_success", "Deployment completed")
}

// recordEvent 记录发布事件
func (s *ReleaseService) recordEvent(ctx context.Context, releaseID int, eventType, message string) {
    event := &models.ReleaseEvent{
        ReleaseID: releaseID,
        EventType: eventType,
        EventMsg:  message,
        CreatedAt: time.Now(),
    }
    
    if err := s.eventRepo.Create(ctx, event); err != nil {
        s.log.Error("Failed to record event", "error", err)
    }
}

// Rollback 回滚发布
func (s *ReleaseService) Rollback(ctx context.Context, releaseID int) (*models.ReleaseRecord, error) {
    release, err := s.releaseRepo.GetByID(ctx, releaseID)
    if err != nil {
        return nil, fmt.Errorf("release not found: %w", err)
    }
    
    release.Status = "rolled_back"
    release.UpdatedAt = time.Now()
    
    if err := s.releaseRepo.Update(ctx, release); err != nil {
        return nil, fmt.Errorf("failed to rollback: %w", err)
    }
    
    return release, nil
}

// ListReleaseEvents 查询发布事件
func (s *ReleaseService) ListReleaseEvents(ctx context.Context, releaseID string) ([]interface{}, error) {
    id, err := strconv.Atoi(releaseID)
    if err != nil {
        return nil, fmt.Errorf("invalid releaseID: %w", err)
    }
    
    events, err := s.eventRepo.GetByReleaseID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get events: %w", err)
    }
    
    result := make([]interface{}, len(events))
    for i, e := range events {
        result[i] = e
    }
    
    return result, nil
}
```

### 2. ApplicationService

```go
type ApplicationService struct {
    appRepo     repository.ApplicationRepository
    releaseRepo repository.ReleaseRecordRepository
    log         *logger.Logger
}

func NewApplicationService(
    appRepo repository.ApplicationRepository,
    releaseRepo repository.ReleaseRecordRepository,
    log *logger.Logger,
) *ApplicationService {
    return &ApplicationService{
        appRepo:     appRepo,
        releaseRepo: releaseRepo,
        log:         log,
    }
}

// Create 创建应用
func (s *ApplicationService) Create(ctx context.Context, name, gitRepo, buildType string) (*models.Application, error) {
    if name == "" {
        return nil, fmt.Errorf("application name required")
    }
    
    app := &models.Application{
        Name:      name,
        GitRepo:   gitRepo,
        BuildType: buildType,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    if err := s.appRepo.Create(ctx, app); err != nil {
        return nil, fmt.Errorf("failed to create application: %w", err)
    }
    
    s.log.Info("Application created", "id", app.ID, "name", app.Name)
    return app, nil
}
```

### 3. ServiceContainer（DI 容器）

```go
// internal/services/container.go

type ServiceContainer struct {
    // Service instances
    applicationService *ApplicationService
    clusterService     *ClusterService
    releaseService     *ReleaseService
    shellService       *ShellService
    workloadService    *WorkloadService
    
    // Repositories
    releaseRepo       repository.ReleaseRecordRepository
    appRepo           repository.ApplicationRepository
    clusterRepo       repository.ClusterRepository
    workloadRepo      repository.WorkloadTargetRepository
    eventRepo         repository.ReleaseEventRepository
    // ... 其他 repos
    
    log *logger.Logger
    db  *sql.DB
}

// 初始化选项
type Option func(*ServiceContainer) error

func WithApplicationRepository(repo repository.ApplicationRepository) Option {
    return func(c *ServiceContainer) error {
        c.appRepo = repo
        return nil
    }
}

// ... 其他 WithXxx 选项

// NewServiceContainer 初始化 DI 容器
func NewServiceContainer(db *sql.DB, log *logger.Logger, opts ...Option) (*ServiceContainer, error) {
    c := &ServiceContainer{
        log: log,
        db:  db,
    }
    
    // 应用所有选项
    for _, opt := range opts {
        if err := opt(c); err != nil {
            return nil, err
        }
    }
    
    // 验证必要的 repository
    if c.appRepo == nil || c.clusterRepo == nil || c.releaseRepo == nil {
        return nil, fmt.Errorf("required repositories not initialized")
    }
    
    // 初始化 Service
    c.applicationService = NewApplicationService(c.appRepo, c.releaseRepo, log)
    c.clusterService = NewClusterService(c.clusterRepo, log)
    c.releaseService = NewReleaseService(
        c.releaseRepo, c.workloadRepo, c.clusterRepo,
        c.appRepo, c.eventRepo,
        deployers.NewDeployerFactory(log),
        log, db,
    )
    c.shellService = NewShellService(c.shellServerRepo, c.shellCommandRepo, c.shellCommandExecRepo, log)
    c.workloadService = NewWorkloadService(c.workloadRepo, log)
    
    return c, nil
}

// Getter 方法
func (c *ServiceContainer) Application() *ApplicationService { return c.applicationService }
func (c *ServiceContainer) Cluster() *ClusterService { return c.clusterService }
func (c *ServiceContainer) Release() *ReleaseService { return c.releaseService }
func (c *ServiceContainer) Shell() *ShellService { return c.shellService }
func (c *ServiceContainer) Workload() *WorkloadService { return c.workloadService }
```

---

## Handler 调用 Service

### 设计原则

1. **Handler**: HTTP 处理、参数解析
2. **Service**: 业务逻辑、repository 协调
3. **Repository**: 数据访问
4. **Container**: 生命周期管理

### Handler 示例

```go
// internal/handlers/applications/handlers.go

func Create(service *services.ApplicationService, log *logger.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            Name      string `json:"name"`
            GitRepo   string `json:"git_repo"`
            BuildType string `json:"build_type"`
        }
        
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid request", http.StatusBadRequest)
            return
        }
        
        app, err := service.Create(r.Context(), req.Name, req.GitRepo, req.BuildType)
        if err != nil {
            log.Error("Failed to create application", "error", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{
            "code": 0,
            "data": app,
        })
    }
}

// internal/handlers/releases/handlers.go

func Create(releaseService *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req struct {
            AppID     int    `json:"app_id"`
            ClusterID int    `json:"cluster_id"`
            Image     string `json:"image"`
        }
        
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid request", http.StatusBadRequest)
            return
        }
        
        release, err := releaseService.Release(r.Context(), req.AppID, req.ClusterID, req.Image)
        if err != nil {
            log.Error("Release failed", "error", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusAccepted)  // 202: 异步操作
        json.NewEncoder(w).Encode(map[string]any{
            "code": 0,
            "data": release,
        })
    }
}
```

### main.go 初始化

```go
func main() {
    db, err := database.Init(cfg.DatabasePath)
    if err != nil {
        log.Fatal("Database initialization failed", "error", err)
    }
    
    // 初始化 repository
    appRepo := repository.NewApplicationRepository(db)
    clusterRepo := repository.NewClusterRepository(db)
    releaseRepo := repository.NewReleaseRecordRepository(db)
    // ... 其他 repositories
    
    // 初始化 DI 容器
    container, err := services.NewServiceContainer(
        db, log,
        services.WithApplicationRepository(appRepo),
        services.WithClusterRepository(clusterRepo),
        services.WithReleaseRepository(releaseRepo),
        // ... 其他 options
    )
    if err != nil {
        log.Fatal("ServiceContainer initialization failed", "error", err)
    }
    
    // 设置路由
    router := chi.NewRouter()
    router.Route("/api/v1", func(r chi.Router) {
        r.Post("/applications", applications.Create(container.Application(), log))
        r.Post("/releases", releases.Create(container.Release(), log))
        r.Post("/clusters", clusters.Create(container.Cluster(), log))
        // ... 其他路由
    })
    
    http.ListenAndServe(":8080", router)
}
```

---

## 关键设计原则

### 1. 单一职责
每个 Service 类负责一个领域的业务逻辑。

### 2. 依赖通过接口注入
```go
type ReleaseService struct {
    releaseRepo repository.ReleaseRecordRepository  // interface
}
```

### 3. 错误处理 - 使用 %w 包装
```go
if err := s.releaseRepo.Create(ctx, release); err != nil {
    return nil, fmt.Errorf("failed to create release: %w", err)
}
```

### 4. 日志记录
在关键操作点记录日志，便于问题排查。

---

## 总结

**优势**:
- ✅ 依赖管理清晰
- ✅ Service 间可协作
- ✅ 生命周期统一管理
- ✅ 单元测试容易
- ✅ 适合复杂业务

**注意**:
- ⚠️ 避免 Service 循环依赖
- ⚠️ Repository 使用 interface
- ⚠️ Handler 中处理 HTTP 细节
- ⚠️ Service 中处理业务逻辑

这就是当前项目采用的架构模式！
