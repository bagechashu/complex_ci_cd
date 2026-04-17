# Release Control System - Backend

生产级的统一发布控制系统，支持多集群、多环境、多部署方式的应用发布。

## 项目概览

这是一个 Go + go-chi 框架的 REST API 服务，用于管理应用发布流程：

- **多环境支持**: dev/staging/production 等逻辑环境
- **多集群管理**: 支持多个 Kubernetes 集群或其他基础设施
- **发布追踪**: 完整的发布历史和事件日志
- **灵活的部署策略**: 支持插件式部署实现（K8s/Salt/Ansible）

## 项目结构

```
backend/
├── cmd/server/             # 应用入口
│   └── main.go
├── internal/
│   ├── config/            # 配置管理
│   ├── database/          # 数据库初始化
│   ├── models/            # 数据模型
│   ├── repository/        # 数据访问层
│   ├── services/          # 业务逻辑层
│   ├── handlers/          # API处理层
│   └── deployers/         # 部署策略实现
├── pkg/
│   ├── logger/            # 日志工具
│   ├── utils/             # 工具函数 (加密、错误处理)
│   └── middleware/        # HTTP中间件
├── db/
│   ├── schema.sql         # 数据库表结构
│   └── init.sql           # 初始化数据
├── go.mod / go.sum        # 依赖管理
├── .env.example           # 环境变量模板
└── README.md              # 本文件
```

## 快速开始

### 前置要求

- Go 1.21+
- SQLite 3
- （可选）Kubernetes 集群 (用于完整功能测试)

### 安装依赖

```bash
cd backend
go mod download
go mod tidy
```

### 配置环境

```bash
# 复制环境配置模板
cp .env.example .env

# 编辑 .env 文件（可选）
# SERVER_PORT=8080
# DATABASE_PATH=./release_control.db
```

### 运行服务

```bash
# 开发模式
go run cmd/server/main.go

# 或者编译后运行
go build -o release-control cmd/server/main.go
./release-control
```

服务将在 `http://localhost:8080` 启动。

### 验证服务

```bash
# 检查健康指标
curl http://localhost:8080/health

# 查看发布历史
curl http://localhost:8080/api/v1/releases

# 发起发布
curl -X POST http://localhost:8080/api/v1/releases \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": 1,
    "env_id": 2,
    "image": "myapp:v1.0.0",
    "user": "admin"
  }'
```

## API 文档

### 核心端点

#### 1. 发起发布
```
POST /api/v1/releases
```

**请求体**:
```json
{
  "app_id": 1,
  "env_id": 2,
  "image": "registry.example.com/myapp:v1.0.0",
  "user": "admin"
}
```

**响应** (202 Accepted):
```json
{
  "id": 1,
  "app_id": 1,
  "env_id": 2,
  "cluster_id": 1,
  "image": "registry.example.com/myapp:v1.0.0",
  "status": "pending",
  "triggered_by": "admin",
  "started_at": null,
  "completed_at": null,
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

#### 2. 查询发布状态
```
GET /api/v1/releases/{id}
```

#### 3. 获取发布事件
```
GET /api/v1/releases/{id}/events
```

#### 4. 发布历史列表
```
GET /api/v1/releases?limit=20&offset=0
```

#### 5. 回滚发布
```
POST /api/v1/releases/{id}/rollback
```

#### 6. 健康检查
```
GET /health
```

## 数据模型

### application - 应用
```sql
CREATE TABLE application (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    repo TEXT,
    build_type TEXT,
    created_at DATETIME,
    updated_at DATETIME
);
```

### environment - 环境
```sql
CREATE TABLE environment (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    rank INTEGER,
    created_at DATETIME,
    updated_at DATETIME
);
```

### cluster - 集群
```sql
CREATE TABLE cluster (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,  -- kubernetes|salt|ansible
    kubeconfig_encrypted TEXT,
    created_at DATETIME,
    updated_at DATETIME
);
```

### workload_target - 部署目标 (核心映射表)
```sql
CREATE TABLE workload_target (
    id INTEGER PRIMARY KEY,
    app_id INTEGER,
    env_id INTEGER,
    cluster_id INTEGER,
    k8s_namespace TEXT,
    k8s_workload TEXT,
    container_name TEXT,
    registry_domain TEXT,
    image_repo TEXT,
    UNIQUE(app_id, env_id, cluster_id)
);
```

### release_record - 发布记录
```sql
CREATE TABLE release_record (
    id INTEGER PRIMARY KEY,
    app_id INTEGER,
    env_id INTEGER,
    cluster_id INTEGER,
    image TEXT,
    status TEXT,  -- pending|validating|deploying|success|failed|rolled_back
    previous_image TEXT,
    error_msg TEXT,
    triggered_by TEXT,
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);
```

## 状态流转图

```
┌─────────┐
│ pending │
└────┬────┘
     │
     ▼
┌──────────────┐
│ validating   │
└────┬────────┬┘
     │        │
     │        └──→ ┌────────┐
     │             │ failed │
     │             └────────┘
     ▼
┌──────────────┐
│ deploying    │
└────┬─────┬──┘
     │     │
     │     │
     │     └──→ ┌────────┐
     │          │ failed │
     │          └────────┘
     ▼
┌─────────┐
│ success │
└────┬────┘
     │
     ▼ (rollback triggered)
┌──────────────┐
│ rolled_back  │
└──────────────┘
```

## 架构设计

### 分层架构

```
┌─────────────────────────────────┐
│     API 处理层 (handlers)        │
│  ├── releaseHandler             │
│  ├── applicationHandler         │
│  └── workload_targetHandler   │
└──────────────┬──────────────────┘
               │
┌──────────────▼───────────────────┐
│     业务逻辑层 (services)         │
│  ├── ReleaseService             │
│  ├── ApplicationService         │
│  └── DeployerFactory            │
└──────────────┬──────────────────┘
               │
┌──────────────▼────────────────────┐
│     数据访问层 (repository)        │
│  ├── ApplicationRepository       │
│  ├── EnvironmentRepository       │
│  ├── WorkloadTargetRepository  │
│  └── ReleaseRecordRepository     │
└──────────────┬───────────────────┘
               │
┌──────────────▼────────────────────┐
│     数据库层 (sqlite)              │
│  ├── Application table           │
│  ├── Environment table           │
│  ├── Cluster table               │
│  ├── WorkloadTarget table      │
│  ├── ReleaseRecord table         │
│  └── ReleaseEvent table          │
└────────────────────────────────────┘
```

### 部署策略 (策略模式)

```
┌─────────────────────────┐
│   DeployStrategy        │
│   (Interface)           │
└────────────┬────────────┘
             │
   ┌─────────┼─────────┬─────────┐
   │         │         │         │
   ▼         ▼         ▼         ▼
K8sDeployer SaltDeploy...ansible...
```

## 开发指南

### 添加新的部署方式

1. 在 `internal/deployers/` 目录下创建新的 deployer

```go
type MySaltDeployer struct {
    BaseDeployer
    log *logger.Logger
}

func (d *MySaltDeployer) Deploy(ctx context.Context, info *models.WorkloadInfo, image string) error {
    // 实现部署逻辑
    return nil
}
```

2. 在工厂方法中注册:

```go
func (f *DeployerFactory) CreateDeployer(clusterType string) (DeployStrategy, error) {
    case "salt":
        return NewSaltDeployer(f.log), nil
    //...
}
```

### 添加新的 API 端点

1. 在 `internal/handlers/` 下创建新的 handler 文件
2. 在 `handlers/router.go` 中注册路由

```go
router.Route("/api/v1", func(r chi.Router) {
    r.Post("/my-endpoint", myHandler(service, log))
})
```

## 测试

### 单元测试 (TODO)

```bash
go test ./...
```

### 集成测试 (TODO)

```bash
go test -tags=integration ./...
```

### 本地手工测试

```bash
# 1. 启动服务
go run cmd/server/main.go &

# 2. 查看已有的应用
curl http://localhost:8080/api/v1/releases

# 3. 发起发布
curl -X POST http://localhost:8080/api/v1/releases \
  -H "Content-Type: application/json" \
  -d '{"app_id":1,"env_id":1,"image":"test:v1.0"}'

# 4. 查看发布状态
curl http://localhost:8080/api/v1/releases/1
```

## 下一步

### MVP 功能实现路线

- [ ] **Day 1-2**: 完成数据库基础设施
  - [ ] SQLite 初始化完成
  - [ ] 所有 Repository 实现完成
  - [ ] 测试数据导入完成

- [ ] **Day 3**: 部署抽象实现
  - [ ] DeployStrategy 接口设计完成
  - [ ] K8s Deployer 骨架代码完成
  - [ ] Kubeconfig 加密存储完成

- [ ] **Day 4**: 发布服务核心
  - [ ] ReleaseService.Release() 完成
  - [ ] 异步部署流程完成
  - [ ] 事件日志系统完成

- [ ] **Day 5**: API 层完善
  - [ ] 所有 API 端点实现
  - [ ] 错误处理标准化
  - [ ] 请求日志和链路追踪

- [ ] **Day 6**: 集成测试
  - [ ] 端到端测试完成
  - [ ] 真实环境验证
  - [ ] 性能测试

## 常见问题

### Q: 如何连接真实的 Kubernetes 集群?
A: 在 workload_target 表中配置集群信息，并将 kubeconfig 文件加密后存储到 cluster 表。

### Q: 数据库支持并发吗?
A: SQLite 通过 WAL (Write-Ahead Logging) 模式支持基本并发，适合中小规模应用。

### Q: 如何扩展支持其他部署方式?
A: 实现 DeployStrategy 接口，在 factory.go 中注册即可。

## 许可证

MIT License

## 联系方式

For issues and questions, please visit the project repository.
