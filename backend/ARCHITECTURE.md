# 发布控制系统 - 架构设计文档

## 系统概述

发布控制系统是一个生产级的多集群、多环境应用发布管理平台，支持灵活的部署策略和完整的发布生命周期追踪。

## 核心概念

### 1. 应用 (Application)
- **定义**: 要部署的服务单元
- **属性**: 名称、Git 仓库、构建类型
- **例子**: api-service, web-ui, data-processor

### 2. 环境 (Environment)
- **定义**: 逻辑部署环境
- **属性**: 环境名称、优先级排序
- **例子**: development → staging → production
- **用途**: 区分不同的发布目标

### 3. 集群 (Cluster)
- **定义**: 物理基础设施资源
- **属性**: 集群名称、类型(k8s/salt/ansible)、kubeconfig(加密)
- **例子**: cluster-prod-1, cluster-staging, cluster-dev
- **类型**: 
  - Kubernetes: 容器编排平台
  - Salt: 配置管理
  - Ansible: 自动化工具

### 4. 部署目标 (DeploymentTarget) **[核心]**
- **定义**: 应用到环境到集群的三维映射关系
- **唯一性**: (app_id, env_id, cluster_id) 构成唯一键
- **属性**:
  - K8s 命名空间 (namespace)
  - K8s 部署 (deployment)
  - 容器名称 (container_name) → 防止误更新 sidecar
  - 镜像仓库域名 (registry_domain)
  - 镜像名称 (image_repo)

**例子**:
```
api-service + production + cluster-prod-1
→ namespace: production
→ deployment: api-service
→ container_name: api-service
→ registry_domain: harbor.example.com
→ image_repo: company/api-service
```

### 5. 发布记录 (ReleaseRecord)
- **定义**: 单次发布操作的完整记录
- **关键字段**:
  - image: 目标镜像版本
  - status: 发布生命周期状态
  - previous_image: 上一版本 (用于回滚)
  - triggered_by: 发起人
  - started_at/completed_at: 时间戳

### 6. 发布事件 (ReleaseEvent)
- **定义**: 发布过程中的详细事件日志
- **事件类型**: started, validating, deploying, success, failed, rolled_back
- **用途**: 前端实时展示发布进度，错误追踪

## 分层架构

```
┌────────────────────────────────────────────────────────┐
│               REST API 层 (go-chi)                      │
│  ┌─────────────────────────────────────────────────┐   │
│  │ /api/v1/releases          (POST/GET/DELETE)    │   │
│  │ /api/v1/releases/{id}     (GET)                │   │
│  │ /api/v1/releases/{id}/events (GET)             │   │
│  │ /api/v1/releases/{id}/rollback (POST)          │   │
│  └─────────────────────────────────────────────────┘   │
│              handlers/ 数据序列化/HTTP处理              │
└────────────────────┬─────────────────────────────────┘
                     │
┌────────────────────▼─────────────────────────────────┐
│           业务逻辑层 (services/)                       │
│  ┌─────────────────────────────────────────────────┐  │
│  │ ReleaseService                                  │  │
│  │  - Release()      // 发布主流程                │  │
│  │  - Rollback()     // 回滚逻辑                  │  │
│  │  - GetStatus()    // 查询状态                  │  │
│  │  - deployAsync()  // 异步部署执行              │  │
│  └─────────────────────────────────────────────────┘  │
│         核心业务规则、事务管理、权限检查               │
└────────────────────┬─────────────────────────────────┘
                     │
┌────────────────────▼─────────────────────────────────┐
│           数据访问层 (repository/)                     │
│  ┌─────────────────────────────────────────────────┐  │
│  │ ApplicationRepository                          │  │
│  │ EnvironmentRepository                          │  │
│  │ ClusterRepository                              │  │
│  │ DeploymentTargetRepository (★ 核心)            │  │
│  │ ReleaseRecordRepository                        │  │
│  └─────────────────────────────────────────────────┘  │
│  数据库抽象、CRUD 操作、查询封装                       │
└────────────────────┬─────────────────────────────────┘
                     │
┌────────────────────▼─────────────────────────────────┐
│         部署策略层 (deployers/) ← 策略模式            │
│  ┌──────────────┬──────────────┬──────────────┐      │
│  │ DeployStrategy              │ (接口)       │      │
│  │ ├─ Deploy()                 │              │      │
│  │ ├─ Validate()               │              │      │
│  │ ├─ Rollback()               │              │      │
│  │ ├─ GetStatus()              │              │      │
│  │ └─ HealthCheck()            │              │      │
│  └──────────────┼──────────────┼──────────────┘      │
│                 │              │                      │
│          ┌──────▼────┐  ┌──────▼────┐               │
│          │K8sDeployer│  │SaltDeploy.│ (可扩展)      │
│          └───────────┘  └───────────┘               │
│  具体的部署实现：client-go, API 调用等               │
└────────────────────┬─────────────────────────────────┘
                     │
┌────────────────────▼─────────────────────────────────┐
│            数据库层 (SQLite + WAL)                    │
│  ┌─────────────────────────────────────────────────┐ │
│  │ application | environment | cluster             │ │
│  │ deployment_target | release_record              │ │
│  │ release_event     | audit_log                   │ │
│  └─────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

## 关键流程

### 1. 发布流程 (Release Process)

```
用户请求 POST /api/v1/releases
    │
    ▼
┌─────────────────────────────────┐
│ Handler: validateRequest()      │  ← 参数验证
│ - app_id, env_id, image 必填    │
└────────────┬────────────────────┘
             │
             ▼
┌─────────────────────────────────┐
│ Service: Release()              │  ← 事务开始
│ - 查询 Application/Environment  │
│ - 创建 ReleaseRecord (pending)  │
└────────────┬────────────────────┘
             │
             ▼ (立即返回 202 Accepted + releaseID)
┌─────────────────────────────────┐
│ 异步 goroutine: deployAsync()   │  ← 后台执行
└────────────┬────────────────────┘
             │
             ├─→ 1. 记录事件: "started"
             │
             ├─→ 2. 更新状态: "validating"
             │    - 调用 Deployer.Validate()
             │    - 检查 kubeconfig、namespace、镜像等
             │
             ├─→ 3. 更新状态: "deploying"
             │    - 获取 DeploymentTarget
             │    - 获取 Deployer (factory)
             │    - 调用 Deployer.Deploy()
             │      - 更新 K8s Deployment 镜像
             │      - 监控 Pod 状态
             │
             ├─→ 4. HealthCheck
             │    - 检查所有 Pod ready
             │    - 验证应用响应正常
             │
             ├─→ 5. 更新状态: "success"
             │    - 记录事件: "success"
             │    - 保存 completed_at
             │
             └─→ 6. 错误处理
                  - 任何步骤失败
                  - 更新状态: "failed"
                  - 记录错误日志和堆栈
                  - 触发告警 (可选)

前端轮询 GET /api/v1/releases/{id}
    ↓
获取实时状态 + 进度事件
```

### 2. 回滚流程 (Rollback Process)

```
用户请求 POST /api/v1/releases/{id}/rollback
    │
    ▼
┌──────────────────────────────────┐
│ Service: Rollback()              │
│ - 获取当前 ReleaseRecord         │
│ - 读取 previous_image 字段       │
└────────────┬─────────────────────┘
             │
             ▼
┌──────────────────────────────────┐
│ 异步执行回滚                       │
│ 1. 更新状态: "validating"        │
│ 2. 获取 Deployer                 │
│ 3. 调用 Deployer.Rollback()      │
│    - 撤销到 previous_image       │
│    - 监控部署进度                │
│ 4. HealthCheck                  │
│ 5. 更新状态: "rolled_back"      │
└──────────────────────────────────┘
```

## 状态机

```
         ┌─── (validation failed) → ┐
         │                          │
    pending → validating          failed ← (any step failed)
         │        │                │
         │        │                │
         │   deploying ← (validation success)
         │        │
         │        ├─→ (deployment failed) → failed
         │        │
         │        └─→ (deployment success) → success
         │
         └────────────────────────┘
                                  ↑
                    (rollback request)
                                  │
                          rolled_back
```

## 数据持久化

### SQLite 性能优化

1. **WAL 模式 (Write-Ahead Logging)**
   - 支持并发读写
   - 提高吞吐量
   - 配置: `PRAGMA journal_mode = WAL`

2. **同步机制**
   - 使用 NORMAL 同步 (而非 FULL)
   - 权衡数据安全和性能
   - 配置: `PRAGMA synchronous = NORMAL`

3. **缓存大小**
   - 增加页缓存
   - 减少磁盘 I/O
   - 配置: `PRAGMA cache_size = 10000`

4. **外键约束**
   - 启用外键检查
   - 保证数据完整性
   - 配置: `PRAGMA foreign_keys = ON`

### 索引策略

```sql
-- 发布查询优化
CREATE INDEX idx_release_app_env ON release_record(app_id, env_id);
CREATE INDEX idx_release_status ON release_record(status);
CREATE INDEX idx_release_created ON release_record(created_at DESC);

-- 事件查询优化
CREATE INDEX idx_event_release ON release_event(release_id);

-- 部署目标查询优化
CREATE INDEX idx_deployment_target ON deployment_target(app_id, env_id, cluster_id);

-- 审计日志优化
CREATE INDEX idx_audit_log_created ON audit_log(created_at DESC);
```

## 扩展性设计

### 1. 部署策略扩展

通过 **策略模式** 实现灵活的部署支持：

```go
// 定义接口
type DeployStrategy interface {
    Deploy(ctx context.Context, info *models.DeploymentInfo, image string) error
    Validate(ctx context.Context, info *models.DeploymentInfo) error
    Rollback(ctx context.Context, info *models.DeploymentInfo, previousImage string) error
    GetStatus(ctx context.Context, info *models.DeploymentInfo) (string, error)
    HealthCheck(ctx context.Context, info *models.DeploymentInfo) (bool, error)
    Type() string
}

// 实现具体策略
type K8sDeployer struct { /* ... */ }
type SaltDeployer struct { /* ... */ }
type AnsibleDeployer struct { /* ... */ }

// 工厂方法
func (f *DeployerFactory) CreateDeployer(clusterType string) DeployStrategy {
    switch clusterType {
        case "kubernetes": return NewK8sDeployer()
        case "salt": return NewSaltDeployer()
        case "ansible": return NewAnsibleDeployer()
    }
}
```

### 2. 权限管理扩展

- **预留接口**: Permission.CanRelease(user, app, env)
- **可集成**: LDAP/AD/IAM 系统
- **RBAC 模型**: 角色-权限映射

### 3. 通知系统扩展

- **事件驱动**: 发布事件发布到消息队列
- **集成**: Slack / 钉钉 / Email 通知
- **监控**: Prometheus 指标采集

## 安全考虑

### 1. Kubeconfig 管理

```
明文 kubeconfig
         ↓
    AES-GCM 加密
         ↓
   数据库存储
         ↓
   需要时解密
         ↓
   构建 k8s client
```

**关键点**:
- 密钥从环境变量读取 (不在代码中)
- 定期密钥轮换
- 审计所有 kubeconfig 访问

### 2. API 认证

- JWT token 验证 (待实现)
- 操作审计日志
- 请求链路追踪 (X-Request-ID)

### 3. 镜像验证

- 验证镜像在 registry 存在
- 检查镜像签名 (可选)
- 防止恶意镜像部署

## 性能指标

### 目标 KPI

| 指标 | 目标值 |
|------|--------|
| 发布响应时间 | < 500ms (返回 202) |
| K8s 部署耗时 | 2-5 分钟 (取决于应用) |
| 单应用并发发布 | >= 10 |
| 数据库查询延迟 | < 100ms |
| 系统可用性 | >= 99.9% |

### 监控点

- 发布成功率 (success/total)
- 平均部署耗时
- K8s API 调用延迟
- 数据库连接池占用率
- Go runtime: 内存、GC 频率

## 下一步优化方向

### 短期 (MVP + 1)
1. 实现 JWT 认证和 RBAC
2. 完整的 K8s Deployer 实现
3. 性能测试和优化
4. 错误恢复机制

### 中期
1. Salt/Ansible Deployer 实现
2. 灰度发布策略
3. 流量切换管理
4. 自动回滚决策

### 长期
1. 多区域部署支持
2. 成本优化和资源调度
3. AI 驱动的部署优化
4. 完整的 SaaS 平台化
