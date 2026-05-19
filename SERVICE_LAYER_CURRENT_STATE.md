# Service 层架构 - 当前实现分析

> 基于实际后端代码的深度分析

---

## 📊 当前实现方式

### 使用的模式：**Service 类 + DI 容器**（混合式）

```
当前实现                           
─────────────────────────────────────
✅ Service 是结构体类（不是函数）
✅ 有 DI 容器（ServiceContainer）
✅ 每个 Service 有 New() 构造函数
❌ 没有 domain/{agg}/services/ 目录
❌ 没有 DDD 聚合根
❌ 不是 Helper 函数方式
```

---

## 📁 实际文件结构

```
backend/internal/
├── services/
│   ├── container.go                   ← DI 容器（ServiceContainer）
│   ├── application_service.go         ← Service 类
│   ├── application_service_test.go
│   ├── cluster_service.go             ← Service 类
│   ├── cluster_service_test.go
│   ├── release_service.go             ← Service 类（存在但未集成！）
│   ├── release_service_test.go
│   ├── shell_service.go               ← Service 类
│   ├── shell_service_test.go
│   ├── workload_service.go            ← Service 类
│   └── ...
├── repository/                        ← 数据访问层
│   ├── application_repo.go
│   ├── cluster_repo.go
│   ├── release_record_repo.go
│   ├── workload_target_repo.go
│   └── ...
├── handlers/                          ← HTTP 处理层
│   ├── applications/
│   ├── clusters/
│   ├── workloads/
│   └── ...
└── models/                            ← 数据模型
    ├── application.go
    ├── cluster.go
    ├── release_record.go
    └── ...

❌ 缺失：domain/ 目录（没有 DDD 实现）
```

---

## 🔍 Service 类的具体实现

### 1. ApplicationService

```go
// internal/services/application_service.go

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

// 包含业务方法如：
// func (s *ApplicationService) Create(ctx, req) error
// func (s *ApplicationService) Update(ctx, req) error
// func (s *ApplicationService) Delete(ctx, id) error
```

### 2. ClusterService

```go
type ClusterService struct {
    clusterRepo  repository.ClusterRepository
    deployerFact *deployers.DeployerFactory
    log          *logger.Logger
}
```

### 3. ReleaseService

```go
// internal/services/release_service.go

type ReleaseService struct {
    releaseRepo   repository.ReleaseRecordRepository
    workloadRepo  repository.WorkloadTargetRepository
    clusterRepo   repository.ClusterRepository
    appRepo       repository.ApplicationRepository
    eventRepo     repository.ReleaseEventRepository
    deployerFact  *deployers.DeployerFactory
    log           *logger.Logger
    db            interface{}
}

func NewReleaseService(
    releaseRepo repository.ReleaseRecordRepository,
    workloadRepo repository.WorkloadTargetRepository,
    clusterRepo repository.ClusterRepository,
    appRepo repository.ApplicationRepository,
    eventRepo repository.ReleaseEventRepository,
    deployerFact *deployers.DeployerFactory,
    log *logger.Logger,
    db interface{},
) *ReleaseService { ... }

// 方法包括：
func (s *ReleaseService) Release(ctx, appID, clusterID, image) (*ReleaseRecord, error)
func (s *ReleaseService) ListReleaseEvents(ctx, releaseID) ([]interface{}, error)
func (s *ReleaseService) Rollback(ctx, releaseID) (*ReleaseRecord, error)
```

### 4. ShellService

```go
type ShellService struct {
    serverRepo     repository.ShellServerRepository
    commandRepo    repository.ShellCommandRepository
    executionRepo  repository.ShellCommandExecutionRepository
    encryptionKey  string
    log            *logger.Logger
}
```

### 5. WorkloadService

```go
type WorkloadService struct {
    workloadRepo repository.WorkloadTargetRepository
    appRepo      repository.ApplicationRepository
    envRepo      *repository.EnvironmentRepository
    clusterRepo  repository.ClusterRepository
    log          *logger.Logger
}

// ⚠️ 注意：WorkloadService 在 Container 中是动态创建的
func (c *ServiceContainer) Workload() *WorkloadService {
    return NewWorkloadService(c.workloadRepo, c.appRepo, c.envRepo, c.clusterRepo, c.log)
}
```

---

## 🏗️ DI 容器 (ServiceContainer)

### 容器中的 Service 字段

```go
type ServiceContainer struct {
    // Service instances
    applicationService *ApplicationService    ✅ 存储
    clusterService     *ClusterService        ✅ 存储
    shellService       *ShellService          ✅ 存储
    // ❌ 缺失：releaseService   

    // Repository instances
    releaseRepo  repository.ReleaseRecordRepository
    appRepo      repository.ApplicationRepository
    clusterRepo  repository.ClusterRepository
    workloadRepo repository.WorkloadTargetRepository
    // ... 其他 repo

    deployerFact *deployers.DeployerFactory
    log          *logger.Logger
    db           *sql.DB
}
```

### 容器的 Getter 方法

```go
// internal/services/container.go

// ✅ 存在的 getter
func (c *ServiceContainer) Application() *ApplicationService { return c.applicationService }
func (c *ServiceContainer) Cluster() *ClusterService { return c.clusterService }
func (c *ServiceContainer) Shell() *ShellService { return c.shellService }

// 动态创建
func (c *ServiceContainer) Workload() *WorkloadService {
    return NewWorkloadService(c.workloadRepo, c.appRepo, c.envRepo, c.clusterRepo, c.log)
}

// ✅ Repository 直接暴露
func (c *ServiceContainer) ReleaseRepo() repository.ReleaseRecordRepository { return c.releaseRepo }
func (c *ServiceContainer) ApplicationRepo() repository.ApplicationRepository { return c.appRepo }
// ... 其他 repo getter

// ❌ 缺失：没有 Release() 或 ReleaseService() getter！
```

---

## 🔴 关键问题

### 问题 1️⃣ : ReleaseService 存在但未集成

**代码事实**:
```
✅ ReleaseService 类存在      → internal/services/release_service.go
✅ 有 NewReleaseService()      → 可以创建实例
✅ 有完整的业务方法           → Release(), Rollback(), ListReleaseEvents()
✅ 有单元测试                  → release_service_test.go

❌ 但 ServiceContainer 中：
   - 没有 releaseService 字段
   - 没有初始化逻辑
   - 没有 Release() getter 方法
   - 没有在 NewServiceContainer 时创建

❌ 在 routes.go 中：
   - 没有使用 ReleaseService 的路由
   - 没有 /api/v1/releases 相关端点
```

**现象**: ReleaseService 是"孤立"的，完全没有集成到系统中

---

### 问题 2️⃣ : 架构不一致

**代码实现**: Service 类模式（OOP风格）
```go
type ApplicationService struct { ... }
func NewApplicationService(...) *ApplicationService { ... }
```

**SKILL 文档说**: Helper 函数模式（Go风格）
```go
func ReleaseApp(ctx, appID, envID, ...) error { ... }
func ValidateImageFormat(image) bool { ... }
```

**be.agent.md 说**: 混合（DDD + Service类）
```
domain/{agg}/services/
internal/services/helpers.go
```

**三者不一致！**

---

### 问题 3️⃣ : 没有 Handler 调用 ReleaseService

**检查点**:
- ❌ routes.go 中没有 release 相关路由
- ❌ handlers/ 目录中没有 release handler 目录
- ❌ 没有 `POST /api/v1/releases` 端点
- ❌ container.Release() 不存在

**影响**: 前端无法调用发布 API

---

## 📋 ServiceContainer 初始化流程

```go
// cmd/server/main.go

// 1. 创建所有 Repository
appRepo := repository.NewSQLiteApplicationRepository(db)
clusterRepo := repository.NewSQLiteClusterRepository(db, encryptionKey)
releaseRepo := repository.NewSQLiteReleaseRecordRepository(db)
workloadRepo := repository.NewWorkloadTargetRepository(db)
// ... 其他 repo

// 2. 创建 ServiceContainer with functional options
container, err := services.NewServiceContainer(
    db, log,
    services.WithApplicationRepository(appRepo),
    services.WithClusterRepository(clusterRepo),
    services.WithReleaseRepository(releaseRepo),
    // ... 其他 options
)

// 3. 初始化只发生在 NewServiceContainer 中
//    但只初始化了 3 个 service：
//    - applicationService
//    - clusterService  
//    - shellService
//
//    ❌ releaseService 从未初始化！
```

---

## 📍 与文档的对比

### 当前代码 vs .claude 文档

| 方面 | 当前代码 | be.agent.md | service-layer/SKILL.md | 
|------|--------|-----------|----------------------|
| Service 位置 | `internal/services/` | `internal/services/` 和 `domain/{agg}/services/` | `internal/services/helpers.go` |
| Service 形式 | 类 (struct) | 类 (struct) | 函数 (Helper) |
| DI 方式 | ServiceContainer | ServiceContainer | 不涉及 (直接调用) |
| Release 端点 | ❌ 没有 | ✅ 已定义 | ✅ 已定义 |
| Release Service | 存在但孤立 | ✅ 完整 | ✅ 完整 |
| domain/ 目录 | ❌ 没有 | ✅ 应该有 | ❌ 不用 |

---

## 🎯 当前模式总结

### 实际采用的架构

```
HTTP 请求
    ↓
Handler 层 (internal/handlers/)
    ├─ 调用 container.Application()
    ├─ 调用 container.Cluster()
    ├─ 调用 container.Shell()
    └─ 调用 container.WorkloadRepo() 直接操作
    ↓
Service 层 (internal/services/)
    ├─ ApplicationService
    ├─ ClusterService
    ├─ ShellService
    └─ ❌ ReleaseService (孤立，未使用)
    ↓
Repository 层 (internal/repository/)
    ├─ ApplicationRepository
    ├─ ClusterRepository
    ├─ ReleaseRecordRepository
    └─ ...
    ↓
数据库 (SQLite)
```

### 问题

1. **不完整**: ReleaseService 没有被使用
2. **不一致**: Handler 有时调用 Service，有时直接调用 Repository
3. **文档偏差**: 当前代码与 SKILL.md 和 be.agent.md 描述的都不一样

---

## 💡 根本原因分析

### 为什么会这样?

1. **开发阶段性**: 项目可能是逐步实现的
   - ApplicationService ✅ 完整
   - ClusterService ✅ 完整
   - ReleaseService ✅ 写了，但没集成

2. **设计决策变更**: 
   - 初期可能计划所有 domain 都有 Service
   - 后来可能改变策略，但没有更新 ReleaseService

3. **文档与代码脱离**:
   - SKILL.md 描述了 ReleaseService 的完整实现
   - 但代码中的实现方式与 SKILL.md 不完全一致

---

## 🔧 下一步建议

### 选项 A: 完成当前模式（推荐）

1. 在 ServiceContainer 中添加 ReleaseService 字段和初始化
2. 添加 Release() getter 方法
3. 创建 Release Handler（internal/handlers/releases/）
4. 添加 Release 路由到 routes.go
5. 更新 .claude 文档以匹配实现

```go
// 需要修改的地方
type ServiceContainer struct {
    // ... 现有字段
    releaseService *ReleaseService  ← 添加此字段
}

// 在 NewServiceContainer 中初始化
if c.releaseRepo != nil && c.workloadRepo != nil {
    c.releaseService = NewReleaseService(...)
}

// 添加 getter
func (c *ServiceContainer) Release() *ReleaseService {
    return c.releaseService
}
```

### 选项 B: 转换为 Helper 函数方式

1. 将所有 Service 类转换为函数
2. 删除 ServiceContainer
3. Handler 直接调用 helper 函数

### 选项 C: 采用纯 DDD 方式

1. 创建 domain/ 目录
2. 实现聚合根和 Value Objects
3. Service 变成 Domain Service
4. 增加复杂度但提高可维护性

---

## 📌 结论

**当前代码采用的是 "Service 类 + DI 容器" 的模式**，这是一个合理的混合方案。

**主要问题**:
- ❌ ReleaseService 虽然实现了，但完全没有集成到系统中
- ❌ 缺少对应的 Handler 和路由
- ❌ 文档与代码不一致

**快速修复**: 完成 ReleaseService 的集成（选项 A），这是最小改动，最快见效的方案。
