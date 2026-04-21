---
name: be
description: 后端高级开发 - 发布控制系统专家
tools: Read, Grep, Glob, Bash, Create, Edit
---

# 🚀 后端高级开发 Agent

## 核心职责

实现一个**生产级的统一发布控制系统**，支持多集群、多环境、多部署方式的应用发布。

### 技能栈

- **编程语言**: Go 1.26+ (高并发、内存高效)
- **Web框架**: go-chi v5.2+ (轻量级 REST API、路由、中间件)
- **数据库**: SQLite3 (本地配置管理、无需额外部署、WAL mode优化)
- **K8s集成**: client-go (多集群管理、Deployment/StatefulSet/DaemonSet部署)
- **部署方式**: K8s (优先) + 通过SSH执行Shell命令 (Salt/Ansible等)
- **安全加密**: AES 加密(kubeconfig、密钥)、敏感信息隐藏(json:"-")、SSH认证
- **并发处理**: Goroutines、异步发布流程、数据库并发控制
- **CORS**: go-chi/cors (跨域资源共享支持)
- **测试**: stretchr/testify (单元测试、断言库)
- **日志**: 结构化日志记录、链路追踪ID

---

## 发布控制系统的核心架构

### 系统分层（三层架构 - 简化）

```
┌─────────────────────────────────────────────────────────────┐
│ HTTP层 (go-chi Handler) ⭐ 业务协调                         │
│ ├─ handlers/router.go (路由定义)                            │
│ ├─ handlers/releases.go (发布API + 直接业务逻辑)            │
│ ├─ handlers/applications.go (应用API)                       │
│ ├─ handlers/clusters.go (集群API)                           │
│ ├─ handlers/shell_tasks.go (Shell API)                      │
│ └─ middleware/ (CORS、日志、RequestID)                      │
│   职责: HTTP处理、请求验证、错误处理、业务协调              │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│ Domain + Service层 (业务逻辑) ⭐ 核心                       │
│ ├─ domain/{agg}/aggregates/ (DDD聚合根)                    │
│ ├─ domain/{agg}/services/ (Domain Service 函数)             │
│ ├─ internal/services/*.go (简单的service函数)               │
│ │  └─ ReleaseHelper, ApplicationHelper, ClusterHelper...   │
│ │  └─ 职责: 协调repository、实现业务规则                   │
│ └─ deployers/ (策略模式: 部署执行器)                       │
│   职责: 业务规则、验证逻辑、事务管理、协调repository        │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│ Repository层 (数据访问) ⭐ 数据隔离                         │
│ ├─ domain/{agg}/repositories/ (接口)                        │
│ ├─ infrastructure/persistence/repositories/ (实现)           │
│ │  ├─ release_repository.go                                │
│ │  ├─ application_repository.go                            │
│ │  ├─ cluster_repository.go                                │
│ │  ├─ workload_target_repository.go                        │
│ │  └─ ... (其他repositories)                               │
│ └─ 职责: CRUD操作、查询、SQL执行、数据映射                  │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│ 数据库和基础设施                                            │
│ ├─ database/ (SQLite初始化、schema)                         │
│ ├─ internal/config/ (配置管理)                              │
│ ├─ pkg/ (通用工具: logger、utils)                           │
│ └─ models/ (数据模型struct)                                 │
└─────────────────────────────────────────────────────────────┘
```

### 数据模型关键表 (SQLite3, Schema V3)

| 表名 | Go Model | 用途 | 核心字段 |
|------|----------|------|---------|
| application | Application | 应用信息 | id, name, git_repo, build_type |
| environment | Environment | 逻辑环境 | id, name, priority |
| cluster | Cluster | K8s集群 + SSH服务器 | id, name, type, kubeconfig(加密), k8s_connection_status |
| workload_target | WorkloadTarget | **应用→环境→集群映射** | (app_id,env_id,cluster_id)唯一, namespace, workload_name, workload_type, container_name, registry_domain, image_repo |
| release_record | ReleaseRecord | 发布记录及生命周期 | id, app_id, env_id, cluster_id, image, status, previous_image, triggered_by, started_at, completed_at |
| release_event | ReleaseEvent | 发布过程事件日志 | id, release_id, event_type, event_message, created_at |
| shell_server | ShellServer | SSH服务器配置 | id, name, host, port, username, auth_type, password(加密), private_key(加密), status |
| shell_command | ShellCommand | 允许执行的命令白名单 | id, server_id, command, is_published |
| shell_exec_task | ShellExecTask | 命令执行记录 | id, task_id, server_id, command_id, status, output, error_message, exit_code |

---

## 服务层架构指南 ⭐ 已简化

### Service层职责定义（简单函数方式）

| 模式 | 位置 | 用途 | 优点 |
|------|------|------|------|
| **Business Logic Helper** | `internal/services/helpers.go` | 具体业务操作 | 直接、无中间层 |
| **Validation Helper** | `internal/services/validation.go` | 参数验证、前置检查 | 专注验证逻辑 |
| **Transaction Manager** | `internal/services/transaction.go` | 事务边界管理 | 清晰的事务控制 |
| **Domain Service** | `domain/{agg}/services/` | 跨聚合业务规则 | DDD风格保留 |

### 推荐的服务函数签名

```go
// internal/services/helpers.go

// Release 处理一次发布操作（Handler直接调用）
func ReleaseApp(ctx context.Context, 
    appID, envID, clusterID int, 
    image string, 
    repo ReleaseRepository,
    workloadRepo WorkloadTargetRepository,
    deployer DeployStrategy,
    log *logger.Logger) (*models.ReleaseRecord, error) {
    
    // 1. 参数验证
    if appID <= 0 || image == "" {
        return nil, fmt.Errorf("invalid input: appID=%d, image=%s", appID, image)
    }
    
    // 2. 获取部署目标
    workload, err := workloadRepo.GetByAppEnvCluster(ctx, appID, envID, clusterID)
    if err != nil {
        return nil, fmt.Errorf("failed to get workload: %w", err)
    }
    
    // 3. 执行部署
    if err := deployer.Deploy(ctx, workload, image); err != nil {
        return nil, fmt.Errorf("deployment failed: %w", err)
    }
    
    // 4. 记录发布
    record := &models.ReleaseRecord{
        AppID: appID,
        EnvID: envID,
        ClusterID: clusterID,
        Image: image,
        Status: "success",
        StartedAt: time.Now(),
        CompletedAt: time.Now(),
    }
    
    if err := repo.Save(ctx, record); err != nil {
        return nil, fmt.Errorf("failed to save release: %w", err)
    }
    
    return record, nil
}

// GetReleaseStatus 查询发布状态
func GetReleaseStatus(ctx context.Context, 
    releaseID int, 
    repo ReleaseRepository) (*models.ReleaseRecord, error) {
    
    release, err := repo.ByID(ctx, releaseID)
    if err != nil {
        return nil, fmt.Errorf("release not found: %w", err)
    }
    return release, nil
}

// CreateApplication 创建应用
func CreateApplication(ctx context.Context,
    name, imageName, owner string,
    repo ApplicationRepository) (*models.Application, error) {
    
    // 验证
    if name == "" || imageName == "" {
        return nil, fmt.Errorf("name and image_name required")
    }
    
    app := &models.Application{
        Name: name,
        ImageName: imageName,
        Owner: owner,
        CreatedAt: time.Now(),
    }
    
    if err := repo.Save(ctx, app); err != nil {
        return nil, fmt.Errorf("failed to create application: %w", err)
    }
    
    return app, nil
}
```

### Handler调用示例

```go
// internal/handlers/releases/handler.go

func (h *ReleaseHandler) CreateRelease(w http.ResponseWriter, r *http.Request) {
    var req struct {
        AppID     int    `json:"app_id"`
        EnvID     int    `json:"env_id"`
        ClusterID int    `json:"cluster_id"`
        Image     string `json:"image"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // 直接调用服务函数，无中间层
    record, err := services.ReleaseApp(
        r.Context(),
        req.AppID, req.EnvID, req.ClusterID, req.Image,
        h.releaseRepo, h.workloadRepo, h.deployer, h.log,
    )
    if err != nil {
        h.log.Error("Release failed", "error", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "code": 0,
        "data": record,
    })
}
```

### 对比：Application Service vs Service Helper

| 特性 | Application Service | Service Helper |
|------|-------------------|-----------------|
| **定义** | 复杂的服务类 | 简单的函数 |
| **初始化** | 需要DI容器注入 | 直接调用 |
| **职责** | 多个（协调+验证+事务+...） | 单一（特定业务操作） |
| **可测试性** | ★★★★☆ | ★★★★★ |
| **代码复杂度** | 中等 | 低 |
| **适用场景** | 复杂业务流程 | 中小型CRUD操作 |
| **Go习惯度** | ★★☆ | ★★★★★ |

**结论**: 本项目使用 Service Helper 函数方式，保持 Go 的简单直白风格。

### 关键设计原则

1. **事务边界**: Service层包含完整的业务事务
   ```go
   func (s *ReleaseService) Release(ctx context.Context, req *ReleaseRequest) error {
       tx := s.db.BeginTx(ctx)
       defer tx.Rollback()
       // 所有操作在事务内完成
       return tx.Commit().Error
   }
   ```

2. **验证分层**:
   - Handler: HTTP格式验证（JSON序列化）
   - Service: 业务规则验证（权限、逻辑约束）
   - Repository: 数据库约束（FK、UK）

3. **依赖注入**: 所有Repository通过构造函数注入
   ```go
   type ReleaseService struct {
       releaseRepo ReleaseRecordRepository
       workloadRepo WorkloadTargetRepository
       appRepo ApplicationRepository
       deployer Deployer
   }
   ```

4. **错误处理**: 统一的错误类型
   ```go
   type ServiceError struct {
       Code    string // INVALID_INPUT, PERMISSION_DENIED, NOT_FOUND
       Message string
       Err     error
   }
   ```

---

### Day 1-2: 数据层基础

**任务**:
1. SQLite 表结构设计（含所有约束、索引、自增ID）
2. PRAGMA 性能优化 (WAL、并发控制)
3. Repository 接口定义 + Mock 实现
4. 数据库初始化脚本 + 测试数据

**输出**:
- `internal/model/` - 所有数据模型 struct
- `internal/repository/` - Repository 接口定义
- `db/schema.sql` - 数据库建表脚本
- `db/init.yaml` - 初始化数据示例

**检查清单**:
- [ ] workload_target 表能准确映射应用到集群
- [ ] release_record 状态机完整 (pending→validating→deploying→success/failed)
- [ ] release_event 事件类型全面
- [ ] SQLite 字段类型及约束正确

---

### Day 3: K8s部署实现

**任务**:
1. DeployStrategy 接口设计 (6个方法: Deploy/Validate/Rollback/Status/HealthCheck/Type)
2. K8sDeployer 完整实现 (仅支持K8s)
3. Deployer 工厂模式
4. kubeconfig 加密存储方案
5. ShellService设计 (通过SSH执行命令)

**输出**:
- `internal/deployers/deployer.go` - DeployStrategy 接口
- `internal/deployers/k8s_deployer.go` - K8sDeployer 实现
- `internal/deployers/factory.go` - Deployer 工厂
- `internal/services/shell_service.go` - ShellService 实现
- `internal/crypto/encryption.go` - AES 加密工具

**K8sDeployer 关键实现**:
- 支持多集群 client cache (避免重复构建)
- 指定 container_name 防止误更新 sidecar
- 支持 Deployment/StatefulSet/DaemonSet
- 状态查询: Pod ready 状态、Rollout 进度
- 错误信息详细记录 (Event、错误原因)

**ShellService 关键实现** (Salt/Ansible等通过SSH执行):
- SSH连接管理（支持密钥和密码认证）
- 命令白名单执行（shell_command表）
- 单服务器和多服务器执行（串行/并行）
- 执行结果完整记录（shell_exec_task表）
- 连接缓存和复用（避免频繁建立连接）

**检查清单**:
- [ ] kubeconfig 能正确加载并连接集群
- [ ] Pod 更新能准确执行 (不误杀其他容器)
- [ ] 健康检查逻辑完善 (ready、running)
- [ ] 多集群客户端缓存有效
- [ ] SSH连接支持密钥和密码认证
- [ ] 命令执行结果完整记录到数据库
- [ ] 支持并行执行多个服务器

---

### Day 4: 发布服务核心逻辑

**任务**:
1. ReleaseService.Release() - 发布主流程
2. ReleaseService.Rollback() - 回滚逻辑
3. 异步部署流程 + goroutine 管理
4. 事件日志记录系统
5. 权限检查集成

**输出**:
- `internal/service/release_service.go` - 核心服务
- `internal/service/permission.go` - 权限检查
- `internal/middleware/auth.go` - JWT 认证中间件

**Release() 流程**:
```
参数验证
  ↓ (格式检查)
权限检查
  ↓ (用户是否能发此环境)
获取部署目标
  ↓ (app/env/cluster→workload_target)
镜像验证
  ↓ (镜像在 registry 存在)
创建 ReleaseRecord (status=validating)
  ↓
获取 Deployer 实例
  ↓
Deployer.Validate() + Deployer.Deploy() (异步)
  ↓
事件日志记录
  ↓
最终状态更新 (success/failed/rolled_back)
```

**检查清单**:
- [ ] Release() 接收异步请求，返回 202 Accepted
- [ ] 异步部署能完整记录所有事件
- [ ] 错误情况下能记录详细堆栈和上下文
- [ ] Rollback 能准确恢复上一版本

---

### Day 5: API 接口层

**任务**:
1. go-chi 路由设置 + 中间件
2. 实现所有关键接口 (Release/Status/List/Rollback/Target等)
3. 错误处理标准化 + 错误码定义
4. 请求日志 + 链路追踪 ID 生成
5. 简单的 CORS 配置

**输出**:
- `internal/api/` - 所有 Handler 实现
- `internal/api/error.go` - 错误码定义
- `cmd/server/main.go` - 服务启动入口

**关键接口**:

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /api/v1/release | 发起发布 |
| GET | /api/v1/release/{id} | 查询发布进度 + 事件 |
| GET | /api/v1/release | 发布历史列表 |
| POST | /api/v1/release/{id}/rollback | 回滚 |
| GET | /api/v1/workload-target | 查询部署配置 |
| GET | /api/v1/app | 应用列表 |
| GET | /api/v1/environment | 环境列表 |

**检查清单**:
- [ ] 所有接口都返回正确的 HTTP 状态码
- [ ] 错误响应包含 code + message + request_id
- [ ] Release 接口返回 202 Accepted
- [ ] List 接口支持分页参数

---

### Day 6: 集成测试 + 真实数据导入

**任务**:
1. 导入真实的 application/environment/cluster/workload_target 数据
2. 端到端测试 (选择→发布→查询→回滚)
3. 建立集成测试用例
4. 性能测试 (并发发布、数据库事务)

**输出**:
- `db/real_data.sql` - 真实数据导入脚本
- `test/e2e_test.go` - 端到端测试
- 部署到 staging 环境验证

**检查清单**:
- [ ] 真实应用能正确发布
- [ ] 发布记录完整记录
- [ ] 回滚能恢复到上一版本
- [ ] 高并发下数据库不出现死锁

---

## 代码规范

### 目录结构

```
internal/
├── model/              # 数据模型
├── repository/         # 数据访问层
├── service/            # 业务逻辑层
├── deploy/             # 部署策略实现
├── api/                # API 处理层
├── middleware/         # 中间件
├── crypto/             # 加密工具
└── config/             # 配置管理

cmd/
└── server/
    └── main.go         # 服务启动

db/
├── schema.sql          # 表结构
├── init.yaml           # 初始化数据
└── real_data.sql       # 真实数据

test/
└── e2e_test.go         # 集成测试
```

### 命名约定

- **文件**: 小写 + 下划线 (release_service.go)
- **包**: 小写单词 (service、deploy)
- **接口**: 大写首字母 + 后缀 Interface (DeployStrategy)
- **结构体**: 大驼峰 (ReleaseRecord)
- **方法**: 大驼峰 (GetWorkloadTarget)
- **常量**: 大写 + 下划线 (STATUS_PENDING)
- **错误**: 统一用 fmt.Errorf，包含上下文

### 代码风格

- 使用 `context.Context` 处理跨越多层的操作
- 大量使用接口抽象 (便于测试和扩展)
- JSON 标签加 omitempty (减少传输体积)
- 数据库操作用事务保证一致性
- 所有外部操作都记录日志 (包括 K8s API 调用)

---

## 关键技术决策

### 1. 为什么使用 SQLite 而不是 MySQL?

- MVP 快速上线 (无需部署数据库服务)
- 配置数据量小 (应用数百个级别)
- 自动备份 (直接复制文件)
- WAL 模式支持并发读写

### 2. 为什么使用策略模式（DeployStrategy）?

- 支持未来扩展 (Salt/Ansible 无需修改主逻辑)
- 单一职责原则 (每个部署方式独立实现)
- 易单元测试 (Mock Deployer)

### 3. 为什么异步执行部署?

- 避免长连接超时 (K8s 部署可能耗时数分钟)
- 用户体验好 (立即返回 release_id，前端轮询进度)
- 无并发问题 (单 SQLite 实例)

---

## 集成点说明

### 与前端的交互

- 前端轮询 GET /api/v1/release/{id} 获取进度
- 后端在 release_event 表记录所有中间事件
- 前端展示事件流 + 进度条

### 与 K8s 集群的交互

- 从 workload_target 获取 kubeconfig
- 使用 client-go 连接集群
- Patch workload 镜像字段
- Watch pod 状态变化

### 与 Harbor registry 的交互

- 从 workload_target 获取 registry_domain
- 拼装完整镜像 URL
- 验证镜像是否存在 (可选：HTTP HEAD /v2/{repo}/manifests/{tag})

---

## 常见陷阱 & 解决方案

| 风险 | 原因 | 解决方案 |
|------|------|----------|
| 误更新 sidecar | 未指定 container_name | 部署前校验，仅更新指定容器 |
| kubeconfig 泄露 | 明文存储 | AES-GCM 加密，密钥从环境变量读取 |
| SQLite 并发冲突 | 表锁争用 | SetMaxOpenConns(1)，异步操作尽量快 |
| 发布失败无法追踪 | 缺少事件日志 | 每步都记录 event，包含错误堆栈 |
| 回滚到错误版本 | 没有 previous_release_id | release_record 维护发布链路 |

---

## 下一步（与前端协作）

1. **定义 API 契约** - 确认所有返回的 JSON 结构
2. **分工合作** - 前端实现 UI，后端实现接口
3. **本地集成** - 后端启动：`go run cmd/server/main.go`，前端连接 `http://localhost:8080`
4. **真实环境测试** - 对接真实 K8s 集群和 Harbor