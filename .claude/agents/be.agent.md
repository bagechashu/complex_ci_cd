---
name: be
description: 后端高级开发 - 发布控制系统专家
tools: Read, Grep, Glob, Bash, Create, Edit
---

# 🚀 后端高级开发 Agent

## 📚 相关文档

**系统级规划**（必读）：
- [架构设计](../../prompt/架构设计.md) - 系统 4 阶段建设路线（MVP/增强/优化/进阶）
- [核心问题和目标](../../prompt/核心问题和目标.md) - 为什么要做这个系统

**协作指南**：
- [前后端协作约定](../AGENT_COLLABORATION.md) - 分工、沟通、质量标准
- [FE Agent](./fe.agent.md) - 前端同步实现计划

**本文档关注**：⚡ **Day 1-6 MVP 代码实现路线**

### 📌 设计新业务域时

若要从零设计新的 Service、Repository 或数据模型（如权限系统、通知系统等），
使用 [System Architecture Design Prompt](../prompts/design.prompt.md) 获得完整的架构提案。

---

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

### 系统分层（三层架构 + 依赖注入）

```
┌─────────────────────────────────────────────────────────────┐
│ HTTP层 (go-chi Handler) ⭐ 请求处理                        │
│ ├─ handlers/{domain}/handlers.go (各领域的处理函数)         │
│ ├─ router.go (路由定义)                                     │
│ └─ middleware/ (CORS、日志、RequestID)                      │
│   职责: HTTP处理、参数解析、请求验证、错误转化            │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│ Service层 (业务逻辑) ⭐ 核心                               │
│ ├─ services/container.go (DI容器 - ServiceContainer)       │
│ ├─ services/application_service.go (ApplicationService)     │
│ ├─ services/cluster_service.go (ClusterService)             │
│ ├─ services/release_service.go (ReleaseService)             │
│ ├─ services/shell_service.go (ShellService)                 │
│ ├─ services/workload_service.go (WorkloadService)           │
│ └─ deployers/ (策略模式: 多部署方式执行器)                 │
│   职责: 业务规则、验证逻辑、事务管理、repository协调        │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│ Repository层 (数据访问) ⭐ 数据隔离                         │
│ ├─ repository/application_repo.go (应用数据访问)             │
│ ├─ repository/cluster_repo.go (集群数据访问)                │
│ ├─ repository/release_record_repo.go (发布记录访问)         │
│ ├─ repository/workload_target_repo.go (工作负载映射访问)    │
│ ├─ repository/shell_*_repo.go (Shell相关数据访问)           │
│ └─ 职责: CRUD操作、查询、SQL执行、数据映射                  │
└────────────────┬────────────────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────────────────┐
│ 数据库和基础设施                                            │
│ ├─ database/ (SQLite初始化、schema)                         │
│ ├─ internal/config/ (配置管理)                              │
│ ├─ models/ (数据模型struct)                                 │
│ └─ pkg/ (通用工具: logger、middleware、utils)               │
└─────────────────────────────────────────────────────────────┘
```

### 数据模型关键表 (SQLite3, Schema V3)

| 表名 | Go Model | 用途 | 核心字段 |
| application | Application | 应用信息 | id, name, git_repo, build_type |
| environment | Environment | 逻辑环境 | id, name, priority |
| cluster | Cluster | K8s集群 + SSH服务器 | id, name, type, kubeconfig(加密), k8s_connection_status |
| workload_target | WorkloadTarget | **应用→环境→集群映射** | (app_id,env_id,cluster_id)唯一, namespace, workload_name, workload_type, container_name, registry_domain, image_repo |
| release_record | ReleaseRecord | 发布记录及生命周期 | id, app_id, env_id, cluster_id, image, status, previous_image, triggered_by, started_at, completed_at |
| release_event | ReleaseEvent | 发布过程事件日志 | id, release_id, event_type, event_message, created_at |
| shell_server | ShellServer | SSH服务器配置 | id, name, host, port, username, auth_type, password(加密), private_key(加密), status |
| shell_command | ShellCommand | 允许执行的命令白名单 | id, server_id, command, is_published |
| shell_task | ShellTask | 命令执行任务 | id, name, description, command_id, server_id, requires_approval |
| shell_task_execution | ShellTaskExecution | 命令执行记录 | id, task_id, command_id, server_id, status, output, error_message, command_params, exit_code |

---

## API 端点设计

### Shell 命令执行端点
```
POST /v1/shell-commands/execute
- 请求: { task_id, command_id, server_id, command_params }
- 响应: ShellTaskExecution (id, status=pending, created_at, updated_at)
- 说明: 创建并返回执行记录，实际执行在后台异步进行
```

---

## 服务层架构指南 ⭐ Service 类 + DI 容器

### Service层职责定义（Service 类 + 依赖注入）

当前项目采用 **Service 类 + ServiceContainer** 的设计模式：

| 组件 | 位置 | 用途 | 特点 |
|------|------|------|------|
| **Service 类** | `internal/services/{domain}_service.go` | 业务操作、协调repository | 结构体类，包含业务方法 |
| **DI 容器** | `internal/services/container.go` | 依赖管理、服务生命周期 | ServiceContainer 集中管理 |
| **Deployer** | `internal/deployers/` | 多部署方式的策略实现 | 策略模式，支持 K8s/SSH |
| **Repository** | `internal/repository/` | 数据访问 | 接口定义 + SQLite 实现 |

### Service 类实现示例

### ReleaseService 核心业务

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

// 发布应用 - 核心业务方法
func (s *ReleaseService) Release(ctx context.Context, appID, clusterID int, image string) (*models.ReleaseRecord, error) {
    // 参数验证
    if appID <= 0 || image == "" {
        return nil, fmt.Errorf("invalid input: appID=%d, image=%s", appID, image)
    }
    
    // 获取部署目标
    workload, err := s.workloadRepo.GetByAppCluster(ctx, appID, clusterID)
    if err != nil {
        return nil, fmt.Errorf("failed to get workload: %w", err)
    }
    
    // 创建发布记录
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
    
    // 异步执行部署
    go s.deployAsync(context.Background(), release.ID, workload, image)
    
    return release, nil
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

### Service 类 vs Helper 函数对比（当前采用Service类）

| 特性 | Service 类（当前采用） | Helper 函数 |
|------|-------|----------|
| **定义** | 结构体类 + 方法 | 独立函数 |
| **依赖管理** | 结构体字段 | 函数参数 |
| **参数数量** | 少（依赖已在结构体） | 多（所有 repo 作参数） |
| **可测试性** | ★★★★★ | ★★★★☆ |
| **代码复杂度** | 中等 | 低 |
| **适用场景** | 复杂业务流程 | 中小型CRUD操作 |
| **Go习惯度** | ★★★★☆ | ★★★★★ |
| **Service协作** | ✅ 可相互调用 | ❌ 困难 |
| **DI容器管理** | ✅ 统一管理 | ❌ 手工管理 |

**结论**: 本项目采用 **Service 类 + DI 容器** 模式，更适合中大型系统的复杂业务。

### 关键设计原则

1. **单一职责**: 每个 Service 类负责一个领域
   ```go
   type ReleaseService struct { }      // 只管发布
   type ApplicationService struct { }  // 只管应用
   type ClusterService struct { }      // 只管集群
   ```

2. **DI 容器统一管理**:
   ```go
   container := services.NewServiceContainer(db, log, opts...)
   releaseService := container.Release()    // 获取 Service
   appService := container.Application()
   ```

3. **验证分层**:
   - Handler: HTTP格式验证（JSON序列化）
   - Service: 业务规则验证（权限、逻辑约束）
   - Repository: 数据库约束（FK、UK）

4. **依赖通过接口注入**:
   ```go
   type ReleaseService struct {
       releaseRepo repository.ReleaseRecordRepository  // interface
       workloadRepo repository.WorkloadTargetRepository
       appRepo repository.ApplicationRepository
       deployer Deployer
   }
   ```

4. **错误处理**: 统一使用 Go error，用 %w 包装
   ```go
   // ✅ 推荐
   if err := s.releaseRepo.Create(ctx, release); err != nil {
       return nil, fmt.Errorf("failed to create release: %w", err)
   }
   
   // ❌ 避免
   type ServiceError struct { ... }  // 不需要自定义错误类型
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
- 命令白名单执行（shell_command表，must is_published=true）
- 单次命令执行（单服务器单命令）
- 执行结果完整记录（shell_task_execution表）
- 连接缓存和复用（避免频繁建立连接）
- 前端直接选择已发布命令执行（无需创建ShellTask）

**检查清单**:
- [ ] kubeconfig 能正确加载并连接集群
- [ ] Pod 更新能准确执行 (不误杀其他容器)
- [ ] 健康检查逻辑完善 (ready、running)
- [ ] 多集群客户端缓存有效
- [ ] SSH连接支持密钥和密码认证
- [ ] 命令执行结果完整记录到数据库
- [ ] 单次执行端点 POST /v1/shell-commands/execute 能正确创建执行记录

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
3. 错误处理标准化 + 业务码定义
4. 请求日志 + 链路追踪 ID 生成
5. 简单的 CORS 配置

**API 响应格式（业务码模式）**:

所有 API 使用 `code` 字段表示业务状态：

**HTTP 状态码规则**:
- **200 OK**: 同步操作成功或业务错误（通过 code 字段区分）
- **202 Accepted**: ⭐ 异步操作已接受（如发布、命令执行）
- **400 Bad Request**: 请求格式错误（JSON 解析失败）
- **404 Not Found**: 路由不存在
- **500 Internal Server Error**: 服务器异常

**响应体格式** - 所有响应统一使用 `{code, message, data}`：

```go
// 同步操作成功 (HTTP 200)
{
  "code": 0,
  "message": "success",
  "data": { /* 实际数据 */ }
}

// 同步操作失败 - 业务错误 (HTTP 200，code ≠ 0)
{
  "code": 1002,
  "message": "Cluster not found",
  "data": null
}

// 异步操作已接受 (HTTP 202 Accepted ⭐)
{
  "code": 0,
  "message": "release accepted",
  "data": {
    "id": 42,
    "app_id": 1,
    "status": "pending",
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}

// 参数验证失败 (HTTP 200)
{
  "code": 3001,
  "message": "app_id is required",
  "data": null
}

// 资源不存在 (HTTP 200)
{
  "code": 1002,
  "message": "Cluster not found",
  "data": null
}

// 权限不足 (HTTP 200)
{
  "code": 4001,
  "message": "Permission denied",
  "data": null
}
```

**业务码分类** (详见 skills/api-design/SKILL.md):
- **0** - 成功
- **1000-1999** - 资源不存在
- **2000-2999** - 业务冲突
- **3000-3999** - 参数/验证错误
- **4000-4999** - 权限/认证错误
- **5000-5999** - 业务状态错误
- **9999** - 服务器内部错误

**Handler 实现示例**:

```go
// ✅ 同步操作示例（HTTP 200）
func (h *ReleaseHandler) GetRelease(w http.ResponseWriter, r *http.Request) {
    releaseID := chi.URLParam(r, "id")
    
    release, err := h.releaseService.Get(r.Context(), releaseID)
    if err != nil {
        h.writeResponse(w, http.StatusOK, 1001, "Release not found", nil)
        return
    }
    
    h.writeResponse(w, http.StatusOK, 0, "success", release)
}

// ✅ 异步操作示例（HTTP 202 Accepted）
func (h *ReleaseHandler) CreateRelease(w http.ResponseWriter, r *http.Request) {
    var req CreateReleaseRequest
    
    // 参数解析
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeResponse(w, http.StatusOK, 3001, "Invalid request format", nil)
        return
    }
    
    // 业务逻辑 - 同步验证，异步执行
    release, err := h.releaseService.Create(r.Context(), req)
    if err != nil {
        // 根据错误类型返回对应的业务码
        if errors.Is(err, ErrAppNotFound) {
            h.writeResponse(w, http.StatusOK, 1001, "Application not found", nil)
        } else if errors.Is(err, ErrClusterNotAvailable) {
            h.writeResponse(w, http.StatusOK, 5003, "Cluster not available", nil)
        } else {
            h.writeResponse(w, http.StatusOK, 9999, err.Error(), nil)
        }
        return
    }
    
    // ⭐ 异步操作成功：返回 202 Accepted + code=0
    h.writeResponse(w, http.StatusAccepted, 0, "release accepted", release)
}

// 统一的响应写入方法 - 支持自定义 HTTP 状态码
func (h *ReleaseHandler) writeResponse(w http.ResponseWriter, statusCode int, code int, message string, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)  // ⭐ 关键：设置正确的 HTTP 状态码
    json.NewEncoder(w).Encode(map[string]interface{}{
        "code":    code,
        "message": message,
        "data":    data,
    })
}
```

**关键实现细节**:
- ✅ 所有同步响应：`writeResponse(w, http.StatusOK, ...)`
- ✅ 异步操作成功：`writeResponse(w, http.StatusAccepted, 0, "xxx accepted", data)`
- ✅ 业务错误总是 200：`writeResponse(w, http.StatusOK, codeXXX, ...)`
- ✅ code=0 表示成功（无论 HTTP 状态码是 200 还是 202）

**输出**:
- `internal/handlers/` - 所有 Handler 实现
- `internal/handlers/response.go` - 统一的响应工具函数
- `internal/errors.go` - 业务错误定义
- `cmd/server/main.go` - 服务启动入口

**关键接口**:

| 方法 | 路径 | 功能 | 状态码 |
|------|------|------|--------|
| POST | /api/v1/releases | 发起发布 | **202** ⭐ |
| GET | /api/v1/releases/{id} | 查询发布进度 + 事件 | 200 |
| GET | /api/v1/releases | 发布历史列表 | 200 |
| POST | /api/v1/releases/{id}/rollback | 回滚 | **202** ⭐ |
| GET | /api/v1/shell-commands/published | 已发布 Shell 命令列表 | 200 |
| POST | /api/v1/shell-commands/execute | 执行 Shell 命令 | **202** ⭐ |
| GET | /api/v1/shell-tasks/{id} | 查询命令执行状态 | 200 |
| GET | /api/v1/shell-tasks/executions | 查询执行历史（分页） | 200 |

**异步操作说明** (标记 ⭐ 的接口)：
- **立即返回 202 Accepted**，后端异步处理
- **code=0** 表示操作已接受（不是完成）
- **返回 request_id** 供客户端后续查询
- **前端轮询查询接口** 获取最终状态

**检查清单**:
- [ ] 所有接口返回统一的 {code, message, data} 格式
- [ ] 错误响应使用正确的业务码（不是 HTTP 状态码）
- [ ] ⭐ 异步操作返回 HTTP 202 Accepted + code=0
- [ ] ⭐ 异步操作返回 request_id，供客户端轮询
- [ ] List 接口支持 limit/offset 分页参数
- [ ] 所有错误情况都有详细的日志记录

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