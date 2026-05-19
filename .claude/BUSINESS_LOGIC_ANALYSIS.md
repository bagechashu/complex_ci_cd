# 🎯 .claude 配置业务逻辑分析报告

> 整理日期: 2026-05-19  
> 分析范围: agents/, skills/, 及其与前后端代码的对应关系

---

## 📋 目录

1. [整体架构图](#整体架构图)
2. [核心业务域分解](#核心业务域分解)
3. [Agents 业务职责](#agents-业务职责)
4. [Skills 业务对应](#skills-业务对应)
5. [业务流程图](#业务流程图)
6. [Agent-Skill 映射矩阵](#agent-skill-映射矩阵)
7. [问题与改进建议](#问题与改进建议)

---

## 整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                    发布控制系统 - 完整业务流                    │
└─────────────────────────────────────────────────────────────────┘

前端 (FE Agent)                    后端 (BE Agent)
════════════════                   ════════════════

┌─────────────────┐                ┌─────────────────────────────┐
│  Vue 3 + Pinia  │◄──────JSON─────►│   Go + SQLite + K8s         │
│                 │   (HTTP)        │                             │
│  ┌───────────┐  │                 │  ┌──────────────────────┐   │
│  │ Stores    │  │                 │  │ Service Layer ⭐     │   │
│  │ (appStore,│  │                 │  │ ┌──────────────────┐ │   │
│  │ release   │  │                 │  │ │ReleaseService    │ │   │
│  │ shellStore)  │                 │  │ │ApplicationService │ │   │
│  └───────────┘  │                 │  │ │ClusterService    │ │   │
│                 │                 │  │ │WorkloadService   │ │   │
│  ┌───────────┐  │                 │  │ │ShellService      │ │   │
│  │ Views     │  │                 │  │ └──────────────────┘ │   │
│  │ (Release  │  │                 │  └──────────┬───────────┘   │
│  │ Shell     │  │                 │             │               │
│  │ History)  │  │                 │  ┌──────────▼──────────┐   │
│  └───────────┘  │                 │  │ Repository Layer    │   │
│                 │                 │  │ (数据访问)          │   │
│  ┌───────────┐  │                 │  └──────────┬──────────┘   │
│  │ API Client│  │                 │             │               │
│  │ (Axios)   │  │                 │  ┌──────────▼──────────┐   │
│  └───────────┘  │                 │  │ SQLite Database     │   │
│                 │                 │  │ (Schema V3)         │   │
│  ┌───────────┐  │                 │  └─────────────────────┘   │
│  │ CSS Arch  │  │                 │                             │
│  │ (views.css)  │                 │  ┌──────────────────────┐   │
│  └───────────┘  │                 │  │ Deployer Strategy    │   │
│                 │                 │  │ ┌──────────────────┐ │   │
└─────────────────┘                 │  │ │K8sDeployer      │ │   │
                                    │  │ │(via client-go)  │ │   │
                                    │  │ │ShellService     │ │   │
                                    │  │ │(via SSH)        │ │   │
                                    │  │ └──────────────────┘ │   │
                                    │  └─────────────────────┘   │
                                    │                             │
                                    └─────────────────────────────┘
```

---

## 核心业务域分解

### 业务域 1️⃣: 应用发布管理 (Release Management)

**核心问题**: 如何在多集群、多环境中安全、可追踪地发布应用？

**职责链路**:
```
FE Views (ReleaseFlow)
  ↓ (1. 用户选择应用、环境、集群、镜像)
Store (releaseStore)
  ↓ (2. 保存发布参数)
API (release.ts)
  ↓ (3. POST /api/v1/release)
Handler (handlers/release_handlers.go)
  ↓ (4. 解析请求、验证)
Service (ReleaseService)
  ↓ (5. 业务逻辑：创建记录、验证配置)
Repository (release_record_repo.go, workload_target_repo.go)
  ↓ (6. 持久化数据)
Database (SQLite)
  ↓ (7. 存储发布记录)
Deployer (k8s_deployer.go / ShellService)
  ↓ (8. 异步执行部署)
K8s API / SSH Server
  ↓ (9. 实际更新工作负载)
ReleaseEventRepository
  ↓ (10. 记录事件日志)
Database (release_event 表)
  ↓ (FE 轮询查询事件)
Frontend (ReleaseDetail.vue)
```

**涉及 Skills**:
- 🔵 `service-layer` - ReleaseService 的设计
- 🔵 `api-design` - POST /api/v1/release 端点设计
- 🔵 `deploy-strategy` - K8s 部署器
- 🔵 `database-design` - release_record/release_event 表

**涉及 Stores**:
- 🔵 `releaseStore` - 管理发布流程和历史

---

### 业务域 2️⃣: Shell 命令执行 (Shell Command Execution)

**核心问题**: 如何安全地在远程服务器上执行预定义的命令？

**职责链路**:
```
FE Views (ShellCommandExecution)
  ↓ (1. 用户选择已发布命令、服务器)
Store (shellStore)
  ↓ (2. 加载已发布命令列表)
API (shell.ts)
  ↓ (3. GET /api/v1/shell-commands/published)
Handler (handlers/shell_handlers.go)
  ↓ (4. 查询已发布的命令)
Service (ShellService)
  ↓ (5. 业务逻辑：验证权限、检查白名单)
Repository (shell_command_repo.go)
  ↓ (6. 获取命令配置)
Database (SQLite)
  ↓ (7. 取回已发布的命令)
  ↓
  ↓ (用户点击执行)
  ↓
FE 再次调用
  ↓ (8. POST /api/v1/shell-commands/execute)
Handler
  ↓ (9. 解析 command_id、server_id)
Service (ShellService.ExecuteCommand)
  ↓ (10. 验证命令、获取服务器配置)
SSH 连接 (ssh.Client)
  ↓ (11. 执行远程命令)
Remote Shell
  ↓ (12. 返回结果)
ShellCommandExecutionRepository
  ↓ (13. 保存执行结果和输出)
Database (shell_command_execution 表)
  ↓ (FE 轮询查询执行状态)
Frontend (ExecutionHistory.vue)
```

**涉及 Skills**:
- 🔵 `shell-service` - ShellService 的设计
- 🔵 `api-design` - Shell 相关端点设计
- 🔵 `database-design` - shell_server/shell_command/shell_command_execution 表

**涉及 Stores**:
- 🔵 `shellStore` - 管理命令执行和历史

---

### 业务域 3️⃣: 元数据管理 (Metadata Management)

**核心问题**: 如何管理和缓存系统的基础配置（应用、环境、集群）？

**职责链路**:
```
FE (App 初始化)
  ↓ (1. 加载应用、环境、集群列表)
Store (appStore)
  ↓ (2. 缓存到本地状态)
API (metadata.ts)
  ↓ (3. GET /api/v1/applications/environments/clusters)
Handler (handlers/metadata_handlers.go)
  ↓ (4. 查询请求)
Service (ApplicationService / ClusterService)
  ↓ (5. 业务逻辑：过滤、排序)
Repository (application_repo.go / cluster_repo.go)
  ↓ (6. 数据查询)
Database (SQLite)
  ↓ (7. 返回静态数据)
```

**涉及 Skills**:
- 🔵 `pinia-stores` - appStore 的设计
- 🔵 `api-design` - 元数据端点设计
- 🔵 `database-design` - application/environment/cluster 表

**涉及 Stores**:
- 🔵 `appStore` - 应用、环境、集群的元数据缓存

---

### 业务域 4️⃣: 部署目标配置 (Workload Target Mapping)

**核心问题**: 如何定义"应用在某环境的某集群上的部署配置"？

**职责链路**:
```
FE Views (ClusterConfig)
  ↓ (1. 管理员配置应用→环境→集群映射)
API (cluster-mapping.ts)
  ↓ (2. POST /api/v1/app-cluster-configs)
Handler (handlers/cluster_mapping_handlers.go)
  ↓ (3. 解析配置)
Service (WorkloadService)
  ↓ (4. 业务逻辑：验证唯一性、集群连接状态)
Repository (workload_target_repo.go)
  ↓ (5. 持久化)
Database (SQLite workload_target 表)
  ↓ (6. 保存 (app_id, env_id, cluster_id) 映射关系)
```

**涉及 Skills**:
- 🔵 `database-design` - workload_target 表设计（★核心表）
- 🔵 `service-layer` - WorkloadService 的设计
- 🔵 `api-design` - /api/v1/app-cluster-configs 端点

---

## Agents 业务职责

### BE Agent (后端高级开发专家)

**核心职责**: 实现生产级的统一发布控制系统后端

#### 1. **数据层** (Day 1-2)
- 设计 SQLite Schema V3（application/environment/cluster/workload_target/release_record/release_event/shell_*）
- 实现 Repository 层（CRUD 操作、查询、事务）
- 关键表: `workload_target` (app_id, env_id, cluster_id 唯一约束)

#### 2. **业务层** (Day 3-4)
- 实现 Service 层 + DI 容器
- ReleaseService: 发布流程管理、异步发布、回滚
- ShellService: SSH 连接、命令白名单执行
- ApplicationService/ClusterService: 元数据管理
- WorkloadService: 部署目标配置

#### 3. **API 层** (Day 5)
- 实现 REST API 端点（go-chi）
- 应用管理: GET/POST /applications
- 发布流程: POST /release, GET /release/{id}/events
- Shell 执行: POST /shell-commands/execute, GET /shell-commands/published
- 集群配置: GET/POST /app-cluster-configs

#### 4. **部署层** (Day 6)
- 策略模式: K8sDeployer (使用 client-go)
- 异步执行: Goroutines + context 管理
- 事件日志: 记录发布全生命周期

**关键设计决策**:
- ✅ Service 类 + DI 容器架构
- ✅ 异步发布 + 事件日志（前端轮询）
- ✅ 敏感信息加密存储（kubeconfig/SSH密钥）
- ✅ SQLite WAL mode 优化并发

---

### FE Agent (前端高级开发专家)

**核心职责**: 实现直观易用的发布控制界面

#### 1. **状态管理** (Day 1-2)
- appStore: 应用/环境/集群元数据缓存
- releaseStore: 发布流程和历史
- shellStore: 命令执行和历史
- uiStore: 全局 UI 状态

#### 2. **API 集成** (Day 2-3)
- request.ts: Axios 实例、拦截器、错误处理
- release.ts: 发布 API
- metadata.ts: 元数据 API
- shell.ts: Shell 执行 API
- cluster-mapping.ts: 集群映射 API

#### 3. **页面实现** (Day 3-5)
- ReleaseFlow.vue: 一键发布（选择应用、环境、镜像）
- ReleaseHistory.vue: 发布历史（分页、过滤）
- ReleaseDetail.vue: 发布详情（实时事件日志）
- ShellCommandExecution.vue: Shell 命令执行
- ExecutionHistory.vue: 执行历史
- ClusterConfig.vue: 集群配置管理

#### 4. **样式与 UI** (Day 5-6)
- 采用 Naive UI 组件库
- CSS 架构: theme.ts (CSS 变量) + views.css (统一样式库) + scoped 样式
- 响应式设计、深色主题支持

**关键设计决策**:
- ✅ Pinia stores 集中状态管理
- ✅ views.css 单一真理来源（避免样式重复）
- ✅ 轮询方案获取实时进度（简单、易调试）
- ✅ TypeScript strict mode 保证类型安全

---

## Skills 业务对应

### 1️⃣ `service-layer` SKILL

**业务场景**: 如何组织后端业务逻辑？

**解决的问题**:
- ❓ Service 应该怎么写？
- ❓ 如何管理 Service 间的依赖？
- ❓ 如何实现 DI 容器？

**核心内容**:
```
Service 类 + DI 容器模式
├─ ReleaseService
│  ├─ Release() - 创建发布、验证配置、记录事件
│  ├─ Rollback() - 回滚到上一个镜像
│  └─ ListReleaseEvents() - 查询发布事件
├─ ApplicationService
├─ ClusterService
├─ WorkloadService
└─ ShellService

ServiceContainer (DI容器)
├─ 集中管理所有 Service 实例
├─ 处理 Service 间的依赖注入
└─ 提供统一的生命周期管理
```

**输出**:
- 🎯 明确的 Service 类实现范式
- 🎯 DI 容器的具体构造方式
- 🎯 ReleaseService 的核心方法及流程

**映射代码**:
- `backend/internal/services/container.go` - ServiceContainer 实现
- `backend/internal/services/release_service.go` - ReleaseService 完整实现
- `backend/internal/services/application_service.go` - ApplicationService 实现

---

### 2️⃣ `api-design` SKILL

**业务场景**: 如何设计 RESTful API？

**解决的问题**:
- ❓ 有哪些端点？
- ❓ 请求和响应格式是什么？
- ❓ 如何处理错误？

**核心内容**:
```
API 端点设计（分类）
├─ 应用管理
│  ├─ GET /api/v1/applications
│  └─ POST /api/v1/applications
├─ 环境管理
│  └─ GET /api/v1/environments
├─ 集群管理
│  ├─ GET /api/v1/clusters
│  ├─ POST /api/v1/clusters
│  └─ PUT /api/v1/clusters/{id}
├─ 部署目标配置 ⭐
│  ├─ GET /api/v1/app-cluster-configs
│  └─ POST /api/v1/app-cluster-configs
├─ 发布流程 ⭐⭐
│  ├─ POST /api/v1/release (创建发布)
│  ├─ GET /api/v1/release/{id} (获取发布详情)
│  ├─ GET /api/v1/release/{id}/events (获取事件日志)
│  ├─ POST /api/v1/release/{id}/rollback (回滚)
│  └─ GET /api/v1/releases (列表)
└─ Shell 命令执行 ⭐
   ├─ GET /api/v1/shell-servers (服务器列表)
   ├─ GET /api/v1/shell-commands/published (已发布命令)
   ├─ POST /api/v1/shell-commands/execute (执行命令)
   └─ GET /api/v1/shell-commands/executions (执行历史)

业务状态码 (code 字段)
├─ 0 = 成功
├─ 1000-1999 = 资源不存在
├─ 2000-2999 = 业务冲突
├─ 3000-3999 = 参数验证错误
├─ 4000-4999 = 权限/认证错误
├─ 5000-5999 = 业务状态错误
└─ 9999 = 服务器错误
```

**输出**:
- 🎯 完整的 API 端点清单
- 🎯 请求/响应格式范例
- 🎯 业务状态码定义
- 🎯 错误处理规范

**映射代码**:
- `backend/internal/handlers/routes.go` - 路由定义
- `backend/internal/handlers/{domain}/handlers.go` - 各域处理函数

---

### 3️⃣ `database-design` SKILL

**业务场景**: 如何设计 SQLite 数据库 Schema？

**解决的问题**:
- ❓ 需要哪些表？
- ❓ 表之间什么关系？
- ❓ 如何处理加密字段？

**核心内容**:
```
Schema V3 核心表
├─ 应用管理
│  ├─ application (应用信息)
│  ├─ environment (逻辑环境)
│  └─ cluster (物理集群)
├─ 发布管理 ⭐
│  ├─ workload_target (app→env→cluster 映射)
│  ├─ release_record (发布记录)
│  └─ release_event (发布事件日志)
└─ Shell 执行
   ├─ shell_server (SSH 服务器配置)
   ├─ shell_command (命令白名单)
   └─ shell_command_execution (执行记录)

关键约束
├─ workload_target: UNIQUE(app_id, env_id, cluster_id)
├─ 敏感字段加密: kubeconfig, password, private_key
└─ 引用完整性: FOREIGN KEY 约束
```

**输出**:
- 🎯 完整的 SQL Schema 定义
- 🎯 表之间的关系图
- 🎯 加密字段处理方案
- 🎯 性能优化建议（索引、查询）

**映射代码**:
- `backend/internal/database/sqlite.go` - Schema 定义
- `backend/db/create-sample-data.sql` - 初始数据

---

### 4️⃣ `deploy-strategy` SKILL

**业务场景**: 如何支持多种部署方式？

**解决的问题**:
- ❓ K8s 部署如何实现？
- ❓ 如何扩展新的部署方式？
- ❓ 如何异步执行部署？

**核心内容**:
```
策略模式 DeployStrategy 接口
├─ Deploy() - 执行部署
├─ Validate() - 验证工作负载配置
├─ Rollback() - 回滚到上一个镜像
├─ GetStatus() - 获取当前状态
├─ HealthCheck() - 健康检查
└─ Type() - 返回类型标识

具体实现
├─ K8sDeployer (使用 client-go)
│  ├─ 读取 kubeconfig
│  ├─ 更新 Deployment/StatefulSet 的镜像
│  ├─ 监听 Pod 的 ready 状态
│  └─ 缓存 K8s Client 优化连接
└─ ShellService 方式 (Salt/Ansible等)
   ├─ 通过 SSH 执行远程命令
   ├─ 白名单机制安全控制
   └─ 完整的执行日志记录

并发处理
├─ Goroutines 执行异步部署
├─ context.Context 管理超时
├─ 发布状态生命周期: pending → validating → deploying → success/failed
└─ 事件日志记录每个阶段
```

**输出**:
- 🎯 DeployStrategy 接口定义
- 🎯 K8sDeployer 的完整实现
- 🎯 异步发布的模式
- 🎯 Goroutine 安全实践

**映射代码**:
- `backend/internal/deployers/deployer.go` - 接口定义
- `backend/internal/deployers/k8s_deployer.go` - K8s 实现
- `backend/internal/deployers/factory.go` - 工厂模式

---

### 5️⃣ `shell-service` SKILL

**业务场景**: 如何安全地执行远程 Shell 命令？

**解决的问题**:
- ❓ 如何建立 SSH 连接？
- ❓ 如何确保命令安全（白名单）？
- ❓ 如何记录执行结果？

**核心内容**:
```
ShellService 核心操作
├─ ExecuteCommand(commandID, serverID)
│  ├─ 1. 验证命令是否已发布
│  ├─ 2. 验证服务器是否活跃
│  ├─ 3. 获取 SSH 连接（缓存复用）
│  ├─ 4. 执行命令并捕获输出
│  ├─ 5. 记录执行结果
│  └─ 6. 返回退出码和输出
└─ SSH 连接管理
   ├─ 密码认证 (加密存储)
   ├─ 密钥认证 (加密存储)
   ├─ 连接缓存 (避免重复建立)
   └─ 自动清理 (连接失败时)

数据模型关系
shell_server (SSH 配置)
  ↓ (1:N)
shell_command (命令白名单，按服务器归组)
  ↓ (1:N)
shell_command_execution (执行记录)

安全机制
├─ 白名单: 只执行 is_published=true 的命令
├─ 加密: 密码/密钥 AES 加密存储
├─ 审计: 完整记录每次执行和结果
└─ 隔离: 单服务器单命令的执行模型
```

**输出**:
- 🎯 ShellService 的接口和实现
- 🎯 SSH 连接管理方案
- 🎯 执行日志记录机制
- 🎯 安全最佳实践

**映射代码**:
- `backend/internal/services/shell_service.go` - ShellService 实现
- `backend/internal/repository/shell_*_repo.go` - 数据访问层

---

### 6️⃣ `pinia-stores` SKILL

**业务场景**: 如何管理前端全局状态？

**解决的问题**:
- ❓ 应该有哪些 Store？
- ❓ 每个 Store 管理什么状态？
- ❓ Store 之间如何通信？

**核心内容**:
```
4 个核心 Store

1. appStore (应用元数据)
   状态:
   ├─ applications: 应用列表 (缓存)
   ├─ environments: 环境列表 (缓存)
   ├─ clusters: 集群列表 (缓存)
   └─ workloadTargets: 部署目标映射
   
   操作:
   ├─ fetchApplications() - 一次性加载
   ├─ fetchEnvironments()
   └─ fetchClusters()

2. releaseStore (发布流程)
   状态:
   ├─ currentRelease: 当前发布记录
   ├─ releases: 发布历史列表
   ├─ events: 发布事件日志
   └─ loading: 加载状态
   
   操作:
   ├─ createRelease(appId, envId, clusterId, image) - 创建发布
   ├─ getReleaseDetail(releaseId)
   ├─ listReleaseEvents(releaseId) - 轮询获取事件
   ├─ rollback(releaseId)
   └─ listReleases() - 分页查询历史

3. shellStore (Shell 命令执行)
   状态:
   ├─ shellServers: 服务器列表
   ├─ shellCommands: 已发布命令列表
   ├─ shellCommandExecutions: 执行记录
   └─ currentExecution: 当前执行状态
   
   操作:
   ├─ fetchPublishedCommands() - 加载已发布命令
   ├─ fetchShellServers() - 加载服务器列表
   ├─ executeShellCommand(commandId, serverId) - 执行命令
   ├─ getCommandExecutions(commandId) - 获取该命令的历史
   ├─ listAllExecutions(filters) - 全局执行历史
   └─ getExecutionDetail(executionId) - 获取详情

4. uiStore (全局 UI 状态)
   状态:
   ├─ sidebar_collapsed: 侧边栏收起状态
   ├─ theme: 当前主题
   └─ notifications: 通知队列
   
   操作:
   ├─ toggleSidebar()
   ├─ setTheme(theme)
   └─ addNotification()
```

**输出**:
- 🎯 4 个 Store 的完整设计
- 🎯 每个 Store 的状态和操作
- 🎯 Store 间通信的模式
- 🎯 异步操作和轮询实现

**映射代码**:
- `frontend/src/stores/appStore.ts` - 元数据缓存
- `frontend/src/stores/releaseStore.ts` - 发布流程
- `frontend/src/stores/shellStore.ts` - Shell 执行
- `frontend/src/stores/uiStore.ts` - UI 状态

---

### 7️⃣ `frontend-css-architecture` SKILL

**业务场景**: 如何统一管理前端样式？

**解决的问题**:
- ❓ CSS 如何组织？
- ❓ 如何避免样式重复？
- ❓ 如何支持主题切换？

**核心内容**:
```
CSS 分层架构

1. theme.ts (CSS 变量定义)
   ├─ 颜色变量
   │  ├─ --color-primary: #2d8659 (主操作色)
   │  ├─ --color-text-primary/secondary/muted
   │  ├─ --color-bg-card/dark/light
   │  ├─ --color-success/error/warning/info
   │  └─ --color-border/border-light
   ├─ 间距和尺寸
   │  ├─ --spacing-xs/sm/md/lg/xl/xxl/3xl
   │  ├─ --font-size-xs/sm/base/lg/xl
   │  ├─ --border-radius (直角风格: 0px)
   │  └─ --shadow-sm/md
   └─ 其他
      └─ 字体家族、行高等

2. main.css (全局入口)
   @import views.css  ← 中央样式库
   @import 全局重置
   @import 通用工具类

3. views.css (中央样式库 400+ 行)
   ├─ .page-container - 页面根容器
   ├─ .content-layout - 列表+详情布局
   ├─ .list-panel / .list-item - 列表样式
   ├─ .form-group / .form-input - 表单样式
   ├─ .detail-panel / .detail-section - 详情页样式
   ├─ .pagination-controls - 分页样式
   ├─ .badge / .status-badge - 徽章样式
   ├─ .modal-overlay / .modal - 模态框样式
   └─ .header-actions - 按钮组样式

4. Vue 文件中的 <style scoped> (页面特定)
   └─ 仅包含该页面的自定义样式
```

**核心原则**:
- ✅ views.css 是单一真理来源 (DRY 原则)
- ✅ 避免在各 Vue 文件中重复定义相同的样式
- ✅ 使用 CSS 变量进行主题化
- ✅ scoped 样式仅用于页面特定的定制

**输出**:
- 🎯 CSS 变量系统完整定义
- 🎯 views.css 样式库的结构
- 🎯 主题切换的实现方式
- 🎯 响应式设计规范

**映射代码**:
- `frontend/src/styles/main.css` - 全局入口
- `frontend/src/styles/views.css` - 中央样式库
- `frontend/src/theme.ts` - 主题定义

---

### 8️⃣ `naive-ui` SKILL

**业务场景**: 如何高效使用 Naive UI 组件库？

**解决的问题**:
- ❓ 常用组件有哪些？
- ❓ 如何组合组件完成界面？
- ❓ 如何定制组件主题？

**核心内容**:
```
常用组件分类

1. 布局组件
   ├─ <n-layout> - 页面布局
   ├─ <n-layout-sider> - 侧边栏
   ├─ <n-layout-content> - 内容区
   └─ <n-layout-header> - 页眉

2. 表单组件
   ├─ <n-input> - 文本输入
   ├─ <n-select> - 下拉选择
   ├─ <n-form> - 表单容器
   ├─ <n-form-item> - 表单项
   ├─ <n-input-number> - 数字输入
   ├─ <n-checkbox> - 复选框
   ├─ <n-radio> - 单选框
   └─ <n-switch> - 开关

3. 数据展示
   ├─ <n-data-table> - 数据表格 ⭐
   ├─ <n-list> - 列表
   ├─ <n-card> - 卡片
   ├─ <n-tag> - 标签
   ├─ <n-badge> - 徽章
   ├─ <n-progress> - 进度条
   ├─ <n-timeline> - 时间线
   └─ <n-statistic> - 统计数据

4. 反馈组件
   ├─ <n-modal> - 模态框
   ├─ <n-popconfirm> - 确认框
   ├─ <n-drawer> - 抽屉
   ├─ <n-message> - 消息提示
   ├─ <n-notification> - 通知
   ├─ <n-alert> - 警告框
   └─ <n-result> - 结果页

5. 导航组件
   ├─ <n-menu> - 菜单
   ├─ <n-breadcrumb> - 面包屑
   └─ <n-tabs> - 标签页

关键特性
├─ 水平滚动表格: :scroll-x="1200"
├─ 虚拟滚动: 大数据列表优化
├─ 主题定制: naive-ui 内置主题系统
├─ 国际化: 支持多语言
└─ TypeScript: 完整类型定义
```

**输出**:
- 🎯 常用组件使用范例
- 🎯 组件参数和插槽说明
- 🎯 表格、表单等常见模式
- 🎯 主题定制和配置

**映射代码**:
- `frontend/src/views/*.vue` - 各页面使用了大量 Naive UI 组件
- `frontend/src/theme.ts` - 主题配置

---

### 9️⃣ `tech-stack` SKILL

**业务场景**: 全面了解系统的技术选型和架构

**解决的问题**:
- ❓ 后端用什么框架？
- ❓ 前端用什么框架？
- ❓ 数据库用什么？
- ❓ 整个系统如何集成？

**核心内容**:
```
后端技术栈
├─ Go 1.26+ (高并发、内存高效)
├─ go-chi v5.2+ (轻量级 Web 框架)
├─ SQLite3 + mattn/go-sqlite3 (本地数据库)
├─ client-go (K8s 集群操作)
├─ crypto/aes (敏感信息加密)
├─ stretchr/testify (单元测试)
└─ pkg/logger (结构化日志)

前端技术栈
├─ Vue 3.3+ (Composition API)
├─ TypeScript 5.2+ (类型安全)
├─ Vite 5.0+ (快速构建)
├─ Pinia 2.1+ (状态管理)
├─ Vue Router 4.2+ (SPA 路由)
├─ Axios 1.6+ (HTTP 客户端)
├─ Naive UI 2.34+ (UI 组件库)
├─ Day.js 1.11+ (日期处理)
├─ ESLint 8.53+ (代码检查)
└─ Prettier 3.0+ (代码格式化)

架构集成
├─ 前端 → HTTP(s) → 后端 API
├─ 后端 → SQLite → 本地持久化
├─ 后端 → K8s API → 更新工作负载
├─ 后端 → SSH → 远程执行命令
└─ 发布事件 → 前端轮询 → 实时进度展示
```

**输出**:
- 🎯 完整的技术栈清单
- 🎯 各技术的选型理由
- 🎯 项目结构和代码组织
- 🎯 开发环境和部署指南

---

## 业务流程图

### 流程 1: 完整的发布流程

```
┌─────────────┐
│  用户界面   │
│ ReleaseFlow │
└──────┬──────┘
       │ 1. 选择 应用/环境/集群/镜像
       ↓
┌──────────────────┐
│  releaseStore    │
│  保存参数        │
└──────┬───────────┘
       │ 2. POST /api/v1/release
       ↓
┌──────────────────┐
│  handlers        │
│  release_handler │
└──────┬───────────┘
       │ 3. 验证请求
       ↓
┌──────────────────┐
│  ReleaseService  │
│  Release()       │
└──────┬───────────┘
       │ 4. 创建 release_record (status=pending)
       ↓
┌──────────────────┐
│  Repository      │
│  保存到 SQLite   │
└──────┬───────────┘
       │ 5. 异步执行发布 (Goroutine)
       ↓
┌──────────────────┐
│  Deployer        │
│  K8sDeployer或   │
│  ShellService    │
└──────┬───────────┘
       │ 6. 更新工作负载
       ↓
┌──────────────────┐
│  K8s API/SSH     │
│  实际部署        │
└──────┬───────────┘
       │ 7. 记录事件 (release_event)
       ↓
┌──────────────────┐
│  前端轮询        │
│  GET /release/   │
│  {id}/events     │
└──────┬───────────┘
       │ 8. 获取事件列表
       ↓
┌──────────────────┐
│  ReleaseDetail   │
│  实时显示进度    │
└──────────────────┘
```

### 流程 2: Shell 命令执行流程

```
┌──────────────────────┐
│ ShellCommandExecution│
└────────┬─────────────┘
         │ 1. 加载已发布命令列表
         ↓
    ┌────────────┐
    │ 前端初始化  │
    │ 显示命令列表│
    └────┬───────┘
         │ 2. 用户选择命令和服务器
         ↓
    ┌────────────────┐
    │ 点击执行按钮   │
    └────┬───────────┘
         │ 3. POST /api/v1/shell-commands/execute
         ↓
    ┌────────────────┐
    │ handlers       │
    │ shell_handler  │
    └────┬───────────┘
         │ 4. 验证权限、白名单
         ↓
    ┌──────────────────┐
    │ ShellService     │
    │ ExecuteCommand()  │
    └────┬─────────────┘
         │ 5. 建立 SSH 连接
         ↓
    ┌──────────────────┐
    │ SSH 连接         │
    │ 执行远程命令     │
    └────┬─────────────┘
         │ 6. 捕获输出和退出码
         ↓
    ┌──────────────────┐
    │ Repository       │
    │ 保存执行结果     │
    └────┬─────────────┘
         │ 7. 前端轮询
         │ GET /shell-commands/executions/{id}
         ↓
    ┌──────────────────┐
    │ ExecutionHistory │
    │ 显示执行结果     │
    └──────────────────┘
```

---

## Agent-Skill 映射矩阵

### 后端 Agent (BE) 依赖的 Skills

| 优先级 | Skill | 用途 | 阶段 |
|-------|-------|------|------|
| ⭐⭐⭐ | `database-design` | 设计 SQLite Schema | Day 1-2 |
| ⭐⭐⭐ | `service-layer` | 实现 Service 层 | Day 3-4 |
| ⭐⭐⭐ | `api-design` | 定义 REST API 端点 | Day 5 |
| ⭐⭐ | `deploy-strategy` | K8s 部署和异步实现 | Day 6 |
| ⭐⭐ | `shell-service` | Shell 命令执行 | Day 3-4 |
| ⭐ | `tech-stack` | 整体技术选型理解 | Day 0 |

### 前端 Agent (FE) 依赖的 Skills

| 优先级 | Skill | 用途 | 阶段 |
|-------|-------|------|------|
| ⭐⭐⭐ | `pinia-stores` | 设计 4 个核心 Store | Day 1-2 |
| ⭐⭐⭐ | `naive-ui` | 使用 UI 组件库 | Day 3-5 |
| ⭐⭐⭐ | `frontend-css-architecture` | 统一样式管理 | Day 5-6 |
| ⭐⭐ | `api-design` | 理解后端 API 契约 | Day 2-3 |
| ⭐ | `tech-stack` | 整体技术选型理解 | Day 0 |

### Skill 之间的依赖关系

```
.claude/skills/
    │
    ├─ tech-stack (底层参考)
    │   ├─ 为 be.agent 提供技术背景
    │   ├─ 为 fe.agent 提供技术背景
    │   └─ 为其他 skills 提供上下文
    │
    ├─ database-design (数据层)
    │   ├─ 被 service-layer 依赖 (Repository 设计)
    │   ├─ 被 shell-service 依赖 (表设计)
    │   └─ 被 api-design 参考 (数据模型)
    │
    ├─ service-layer (业务层)
    │   ├─ 依赖 database-design
    │   ├─ 被 api-design 依赖 (端点实现)
    │   ├─ 依赖 deploy-strategy (部署执行)
    │   └─ 依赖 shell-service (Shell 执行)
    │
    ├─ deploy-strategy (部署策略)
    │   ├─ 被 service-layer 依赖
    │   └─ 实现 ReleaseService 的发布执行
    │
    ├─ shell-service (Shell 执行)
    │   ├─ 依赖 database-design
    │   ├─ 被 service-layer 依赖
    │   └─ 替代的部署方式 (Salt/Ansible)
    │
    ├─ api-design (API 层)
    │   ├─ 依赖 service-layer
    │   ├─ 依赖 database-design
    │   ├─ 被 pinia-stores 参考
    │   └─ 定义前后端契约
    │
    ├─ pinia-stores (前端状态)
    │   ├─ 依赖 api-design (API 调用)
    │   ├─ 被 fe.agent 实现
    │   └─ 为前端 views 提供数据
    │
    ├─ naive-ui (前端组件)
    │   ├─ 被 fe.agent 实现 (页面构建)
    │   ├─ 与 frontend-css-architecture 配合
    │   └─ 无依赖关系
    │
    └─ frontend-css-architecture (前端样式)
        ├─ 与 naive-ui 配合
        ├─ 被 fe.agent 实现 (样式管理)
        └─ 定义 CSS 变量系统
```

---

## 问题与改进建议

### 问题 1: 发布状态生命周期文档不完整

**当前状态**:
- ❌ release_record.status 字段的所有可能值没有完整定义
- ❌ 状态转移的规则不清晰（如是否允许从 failed 回到 pending）

**建议改进**:
- ✅ 在 database-design SKILL 中补充状态机图
- ✅ 在 release_record 表的注释中明确列出所有状态和转移规则

```sql
-- 建议的状态定义
-- pending: 发布初始状态，等待 Deployer 处理
-- validating: 验证工作负载配置中
-- deploying: 正在更新镜像
-- success: 发布成功
-- failed: 发布失败（可修复）
-- rolled_back: 已回滚
```

---

### 问题 2: Shell 命令的参数化执行没有定义

**当前状态**:
- ❌ ShellService 的 ExecuteCommand 方法无法接收参数
- ❌ 如何处理动态命令（如 `salt "prod-*" state.apply webserver:version=1.2.3`）

**建议改进**:
- ✅ 在 shell-service SKILL 中补充参数模板机制
- ✅ 定义参数白名单和验证规则

```go
// 建议的改进
type ShellCommandExecution struct {
    CommandID int
    ServerID int
    Params map[string]string  // 参数
}

// 执行时进行参数替换
// command: "salt 'prod-*' state.apply {{ environment }}"
// params: { "environment": "staging" }
// 结果: "salt 'prod-*' state.apply staging"
```

---

### 问题 3: 错误处理在各 Skill 中分散定义

**当前状态**:
- ❌ service-layer SKILL 未定义 Service 方法的错误返回规范
- ❌ shell-service SKILL 未定义 SSH 连接失败时的重试策略

**建议改进**:
- ✅ 在 api-design SKILL 中统一定义所有业务错误码
- ✅ 在各 service skill 中明确列出可能的错误及其含义
- ✅ 在 backend/pkg/errors/errors.go 中实现统一的错误类型

---

### 问题 4: 发布事件粒度不清晰

**当前状态**:
- ❌ release_event 表的 event_type 有哪些值？
- ❌ 事件如何从 Deployer 回调到 Repository？

**建议改进**:
- ✅ 在 database-design SKILL 中定义完整的事件类型
- ✅ 在 deploy-strategy SKILL 中定义 Deployer 的事件回调接口

```
建议的事件类型:
- release_started: 发布开始
- validating: 正在验证工作负载配置
- deploying: 正在更新镜像
- pod_updating: Pod 正在更新
- pod_ready: Pod 已就绪
- deployment_complete: 发布完成
- deployment_failed: 发布失败
- rollback_started: 回滚开始
- rollback_complete: 回滚完成
```

---

### 问题 5: 前端轮询策略没有规范

**当前状态**:
- ❌ 轮询间隔是多少？
- ❌ 最多轮询多少次？
- ❌ 如何检测轮询结束？

**建议改进**:
- ✅ 在 pinia-stores SKILL 中定义轮询常数
- ✅ 在 fe.agent.md 中补充轮询实现范例

```typescript
// 建议的轮询参数
const POLL_INTERVAL = 1000 // 1 秒轮询一次
const POLL_MAX_ATTEMPTS = 600 // 最多轮询 10 分钟
const POLL_TERMINAL_STATUSES = ['success', 'failed', 'rolled_back'] // 终态

// 当 status 进入终态时停止轮询
```

---

### 问题 6: Naive UI 主题定制不完整

**当前状态**:
- ❌ 如何在运行时切换深色/浅色主题？
- ❌ 如何定制 Naive UI 内置组件的样式？

**建议改进**:
- ✅ 在 naive-ui SKILL 中补充主题切换示例
- ✅ 展示如何使用 Naive UI 的 ConfigProvider 定制主题

```vue
<n-config-provider :theme="theme" :theme-overrides="themeOverrides">
  <!-- 应用内容 -->
</n-config-provider>
```

---

### 问题 7: 测试覆盖和单元测试范例不足

**当前状态**:
- ❌ tech-stack 中提到 stretchr/testify，但没有示例
- ❌ 没有定义后端 Service 层的单元测试规范
- ❌ 前端组件和 Store 的测试范例缺失

**建议改进**:
- ✅ 在 service-layer SKILL 中补充单元测试范例
- ✅ 在 pinia-stores SKILL 中补充 Store 的测试范例
- ✅ 补充前端组件的测试最佳实践

---

### 问题 8: 性能优化建议缺失

**当前状态**:
- ❌ K8s 集群连接是否应该缓存？（已在 k8s_deployer.go 中实现）
- ❌ SSH 连接是否应该复用？（已在 shell_service 中实现）
- ❌ 发布历史列表如何分页优化？

**建议改进**:
- ✅ 在 database-design SKILL 中补充索引优化建议
- ✅ 在 deploy-strategy SKILL 中明确提及连接缓存策略
- ✅ 在 api-design SKILL 中明确列表 API 的分页参数

---

### 问题 9: 安全性最佳实践不完整

**当前状态**:
- ✅ kubeconfig 加密存储（已实现）
- ✅ SSH 密码/密钥加密存储（已实现）
- ❌ API 认证/授权没有定义（假设已内置）
- ❌ Shell 命令白名单的版本化没有定义

**建议改进**:
- ✅ 补充 API 认证方式（如 JWT、API Key）
- ✅ 定义基于角色的访问控制（RBAC）
- ✅ 在 shell-service SKILL 中补充白名单审计日志

---

### 问题 10: 部署文档和故障排查指南缺失

**当前状态**:
- ❌ 如何部署后端？（go build / Docker）
- ❌ 如何部署前端？（npm build / Docker）
- ❌ 常见问题和排查方法？

**建议改进**:
- ✅ 补充部署指南（支持开发/生产环境）
- ✅ 补充故障排查指南（如 K8s 连接失败、SSH 超时）
- ✅ 补充性能监控和告警配置

---

## 总结

### ✅ 当前配置的优点

1. **清晰的分层架构**: 前后端分工明确、协作约定详细
2. **完整的 Skill 体系**: 覆盖从数据库到 API 到 UI 的全栈
3. **明确的业务逻辑**: 每个 Skill 对应具体的业务问题
4. **实现基础完善**: Agent 和 Skill 指导的代码已基本实现

### 🔄 需要改进的方向

1. **文档细节补充**: 状态机、错误码、参数化、轮询策略
2. **测试规范补充**: 单元测试、集成测试、测试覆盖
3. **性能优化建议**: 查询优化、连接缓存、分页
4. **安全加固**: 认证授权、审计日志、敏感信息处理
5. **运维支持**: 部署指南、故障排查、监控告警

### 📌 建议行动清单

- [ ] 补充发布状态机定义（database-design SKILL）
- [ ] 补充 Shell 参数化执行机制（shell-service SKILL）
- [ ] 补充错误码统一映射表（api-design SKILL）
- [ ] 补充发布事件类型完整定义（database-design SKILL）
- [ ] 补充前端轮询最佳实践（pinia-stores SKILL）
- [ ] 补充 Naive UI 主题切换示例（naive-ui SKILL）
- [ ] 补充单元测试范例（service-layer, pinia-stores SKILLs）
- [ ] 补充部署和故障排查指南（新建 SKILL 或补充到 tech-stack）
- [ ] 补充 API 认证授权方案（api-design SKILL）
- [ ] 补充性能优化建议（database-design, deploy-strategy SKILLs）

---

**生成时间**: 2026-05-19  
**分析范围**: .claude 目录下的 9 个 Skills + 2 个 Agents + 协作文档  
**覆盖代码**: backend/, frontend/ 完整实现  
**下一步**: 基于本分析逐步补充和完善各 Skill 文档
