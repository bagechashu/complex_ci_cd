---
name: api-design
description: 发布控制系统 - REST API设计与实现规范
keywords: REST API, go-chi, HTTP状态码, 错误处理, 请求追踪, API文档
---

# REST API 设计与实现指南

## API概览

### 基础信息
- **基础URL**: `http://localhost:8080/api/v1`
- **协议**: HTTP/HTTPS
- **内容类型**: application/json
- **版本控制**: URI路径版本 (/api/v1)
- **链路追踪**: X-Request-ID头

## 核心API端点

### 1. 应用管理 (Applications)

#### 列表应用
```
GET /api/v1/applications

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "api-service",
      "git_repo": "https://github.com/company/api-service",
      "build_type": "docker",
      "description": "Core API service",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 创建应用
```
POST /api/v1/applications

请求体:
{
  "name": "api-service",
  "git_repo": "https://github.com/company/api-service",
  "build_type": "docker",
  "description": "Core API service"
}

响应: 201 Created
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

### 2. 环境管理 (Environments)

#### 列表环境
```
GET /api/v1/environments

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "development",
      "priority": 1,
      "description": "Development environment"
    },
    {
      "id": 2,
      "name": "staging",
      "priority": 2,
      "description": "Staging environment"
    },
    {
      "id": 3,
      "name": "production",
      "priority": 3,
      "description": "Production environment"
    }
  ]
}
```

### 3. 集群管理 (Clusters)

#### 列表集群
```
GET /api/v1/clusters

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "cluster-prod-1",
      "type": "kubernetes",
      "k8s_connection_status": "connected",  // ★ kubeconfig已隐藏
      "description": "Production K8s cluster",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 获取集群详情
```
GET /api/v1/clusters/{id}

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "cluster-prod-1",
    "type": "kubernetes",
    "k8s_connection_status": "connected",
    "description": "Production K8s cluster"
  }
}
```

#### 创建集群
```
POST /api/v1/clusters

请求体:
{
  "name": "cluster-prod-1",
  "type": "kubernetes",
  "kubeconfig": "---\napiVersion: v1\nkind: Config\n...",
  "description": "Production K8s cluster"
}

响应: 201 Created
```

#### 更新集群
```
PUT /api/v1/clusters/{id}

请求体:
{
  "name": "cluster-prod-1",
  "type": "kubernetes",
  "kubeconfig": "...",  // 可选，不填保持原值
  "description": "Updated description"
}

响应: 200 OK
```

#### 删除集群
```
DELETE /api/v1/clusters/{id}

响应: 204 No Content
```

### 4. 工作负载配置 (Workload Targets - 核心)

#### 列表配置
```
GET /api/v1/app-cluster-configs

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "app_id": 1,
      "env_id": 1,
      "cluster_id": 1,
      "namespace": "default",
      "workload_name": "api-service",
      "workload_type": "Deployment",
      "container_name": "api-service",
      "registry_domain": "harbor.example.com",
      "image_repo": "company/api-service"
    }
  ]
}
```

#### 获取应用的所有配置
```
GET /api/v1/app-cluster-configs/by-app/{app_id}

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": [
    { ... 所有该应用的配置 ... }
  ]
}
```

#### 创建配置
```
POST /api/v1/app-cluster-configs

请求体:
{
  "app_id": 1,
  "env_id": 1,
  "cluster_id": 1,
  "namespace": "default",
  "workload_name": "api-service",
  "workload_type": "Deployment",
  "container_name": "api-service",
  "registry_domain": "harbor.example.com",
  "image_repo": "company/api-service"
}

响应: 201 Created
```

#### 更新配置
```
PUT /api/v1/app-cluster-configs/{id}

请求体: { ... }

响应: 200 OK
```

### 5. 发布管理 (Releases - 核心业务流程)

#### 创建发布（启动发布流程）
```
POST /api/v1/releases

请求体:
{
  "app_id": 1,
  "env_id": 3,
  "cluster_id": 1,
  "image": "harbor.example.com/company/api-service:v1.2.3"
}

响应: 202 Accepted  ★ 异步操作返回202
{
  "code": 0,
  "message": "release accepted",
  "data": {
    "id": 42,
    "app_id": 1,
    "env_id": 3,
    "cluster_id": 1,
    "image": "harbor.example.com/company/api-service:v1.2.3",
    "status": "pending",
    "triggered_by": "john.doe",
    "started_at": "2024-01-15T10:30:00Z",
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### 查询发布状态
```
GET /api/v1/releases/{id}

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 42,
    "app_id": 1,
    "env_id": 3,
    "cluster_id": 1,
    "image": "harbor.example.com/company/api-service:v1.2.3",
    "status": "deploying",  // 'pending','validating','deploying','success','failed','rolled_back'
    "previous_image": "harbor.example.com/company/api-service:v1.2.2",
    "triggered_by": "john.doe",
    "started_at": "2024-01-15T10:30:00Z",
    "completed_at": null,
    "error_message": null,
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

#### 获取发布事件（实时进度）
```
GET /api/v1/releases/{id}/events

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "release_id": 42,
      "type": "started",
      "message": "Release started",
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "release_id": 42,
      "type": "validating",
      "message": "Validating image: harbor.example.com/company/api-service:v1.2.3",
      "created_at": "2024-01-15T10:30:05Z"
    },
    {
      "id": 3,
      "release_id": 42,
      "type": "deploying",
      "message": "Updating pod: api-service-5d4c8f7bc9-x7k8m",
      "created_at": "2024-01-15T10:30:10Z"
    }
  ]
}
```

#### 列表发布历史
```
GET /api/v1/releases?limit=20&offset=0

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": [
    { ... 发布记录 ... }
  ],
  "pagination": {
    "total": 150,
    "limit": 20,
    "offset": 0
  }
}
```

#### 回滚发布
```
POST /api/v1/releases/{id}/rollback

请求体:
{
  "reason": "Production issue"
}

响应: 202 Accepted
{
  "code": 0,
  "message": "rollback accepted",
  "data": {
    "id": 43,  // 新的发布记录ID
    "status": "pending",
    "previous_status": "success",
    "rollback_from_release_id": 42,
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### 6. Shell执行 (Shell Commands)

#### 列表Shell服务器
```
GET /api/v1/shell-servers

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "prod-server-1",
      "host": "192.168.1.10",
      "port": 22,
      "username": "deploy",
      "auth_type": "key",
      "status": "active",
      "last_connected": "2024-01-15T10:00:00Z"
      // password 和 private_key 已隐藏
    }
  ]
}
```

#### 创建Shell任务
```
POST /api/v1/shell-commands

请求体:
{
  "name": "Deploy backup",
  "command": "cd /opt/backup && ./backup.sh",
  "server_ids": [1, 2],
  "execution_method": "parallel",
  "requires_approval": true
}

响应: 201 Created
```

#### 执行Shell命令
```
POST /api/v1/shell-commands/{id}/execute

响应: 202 Accepted
{
  "code": 0,
  "message": "execution accepted",
  "data": {
    "execution_id": 101,
    "command_id": 5,
    "status": "pending",
    "started_at": "2024-01-15T10:30:00Z"
  }
}
```

#### 获取已发布命令列表
```
GET /api/v1/shell-commands/published

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "server_id": 1,
      "server_name": "prod-server-1",
      "command": "systemctl status api-service",
      "description": "Check API service status",
      "is_published": true,
      "created_at": "2024-01-15T10:00:00Z"
    },
    {
      "id": 2,
      "server_id": 1,
      "server_name": "prod-server-1",
      "command": "docker logs -f api-service",
      "description": "View API service logs",
      "is_published": true,
      "created_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

#### 执行已发布命令
```
POST /api/v1/shell-commands/execute

请求体:
{
  "command_id": 1,
  "server_id": 1
}

响应: 202 Accepted
{
  "code": 0,
  "message": "command execution accepted",
  "data": {
    "execution_id": 101,
    "command_id": 1,
    "server_id": 1,
    "status": "pending",
    "started_at": "2024-01-15T10:30:00Z"
  }
}
```

#### 查询Shell命令执行状态
```
GET /api/v1/shell-commands/{execution_id}

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 101,
    "command_id": 1,
    "server_id": 1,
    "server_name": "prod-server-1",
    "command": "systemctl status api-service",
    "status": "success",
    "output": "● api-service.service - API Service\n   Loaded: loaded (/etc/systemd/system/api-service.service; enabled)\n   Active: active (running) since Mon 2024-01-15 10:30:00 UTC; 5 days ago",
    "error": null,
    "started_at": "2024-01-15T10:30:00Z",
    "completed_at": "2024-01-15T10:30:05Z",
    "duration_seconds": 5
  }
}
```

#### 查询Shell命令执行历史
```
GET /api/v1/shell-commands/executions?limit=20&offset=0&command_id=1&server_id=1

响应: 200 OK
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 105,
      "command_id": 1,
      "server_id": 1,
      "server_name": "prod-server-1",
      "command": "systemctl status api-service",
      "status": "success",
      "output": "● api-service.service - API Service\n   Active: active (running)",
      "error": null,
      "started_at": "2024-01-15T11:00:00Z",
      "completed_at": "2024-01-15T11:00:02Z",
      "duration_seconds": 2
    },
    {
      "id": 104,
      "command_id": 1,
      "server_id": 1,
      "server_name": "prod-server-1",
      "command": "systemctl status api-service",
      "status": "failed",
      "output": null,
      "error": "SSH connection timeout",
      "started_at": "2024-01-15T10:45:00Z",
      "completed_at": "2024-01-15T10:45:30Z",
      "duration_seconds": 30
    }
  ],
  "pagination": {
    "total": 256,
    "limit": 20,
    "offset": 0
  }
}
```

## 响应格式规范 (方案 B: 业务码模式)

**设计原则**：
- ✅ **总是返回 HTTP 200** - 简化客户端处理
- ✅ **用 `code` 字段表示业务状态** - 便于细粒度错误处理
- ✅ **避免冗余** - 不混合 HTTP 状态码和业务码

### 标准响应结构

```json
{
  "code": 0,              // 业务状态码: 0=成功, 其他=各种错误
  "message": "success",  // 可读的错误/成功消息
  "data": {...}          // 响应数据(成功时) 或 null(失败时)
}
```

### 业务状态码表

| 状态码 | 含义 | HTTP原映射 | 用途 | 示例 |
|--------|------|-----------|------|------|
| **0** | **成功** | 200/201/202/204 | 所有成功操作 | 列表、创建、更新、删除 |
| **1001** | 应用不存在 | 404 | 应用查询/操作 | GET /applications/999 |
| **1002** | 集群不存在 | 404 | 集群查询/操作 | GET /clusters/999 |
| **1003** | 环境不存在 | 404 | 环境查询/操作 | GET /environments/999 |
| **1004** | 发布不存在 | 404 | 发布查询 | GET /releases/999 |
| **2001** | 应用名重复 | 409 | 应用创建 | POST /applications (重名) |
| **2002** | 集群名重复 | 409 | 集群创建 | POST /clusters (重名) |
| **2003** | 环境名重复 | 409 | 环境创建 | POST /environments (重名) |
| **3001** | 参数验证失败 | 400 | 所有请求 | 缺少必要字段 |
| **3002** | 无效的参数值 | 400 | 所有请求 | 格式错误的 JSON |
| **3003** | 非法的参数组合 | 400 | 业务逻辑检查 | 环境优先级冲突 |
| **4001** | 权限不足 | 403 | 敏感操作 | 删除生产环境 |
| **4002** | 认证失败 | 401 | 所有请求 | 缺少认证凭证 |
| **5001** | 发布进行中 | 409 | 发布状态检查 | 无法回滚进行中的发布 |
| **5002** | 发布已完成 | 409 | 发布状态检查 | 无法修改已完成的发布 |
| **5003** | 集群连接失败 | 503 | 部署执行 | Kubernetes 连接超时 |
| **5004** | 部署执行失败 | 500 | 部署执行 | Pod 创建失败 |
| **5005** | Shell 执行失败 | 500 | Shell 命令执行 | SSH 连接错误 |
| **9999** | 服务器内部错误 | 500 | 数据库错误或异常 | 数据库连接失败 |

### 响应示例

**成功响应 (code=0)**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "production",
    "priority": 3
  }
}
```

**参数错误 (code=3001)**
```json
{
  "code": 3001,
  "message": "Validation failed: cluster_name is required",
  "data": null
}
```

**资源不存在 (code=1002)**
```json
{
  "code": 1002,
  "message": "Cluster not found",
  "data": null
}
```

**业务冲突 (code=5001)**
```json
{
  "code": 5001,
  "message": "Release in progress, cannot perform rollback",
  "data": null
}
```

### 错误处理指南

**所有响应都是 HTTP 200**，客户端需要检查 `code` 字段来判断成功/失败：

```typescript
// 正确的错误处理
const response = await api.get('/applications')
if (response.data.code === 0) {
  // 成功 - 使用 response.data.data
  console.log(response.data.data)
} else if (response.data.code === 1001) {
  // 应用不存在
  toast.error('应用已删除')
} else if (response.data.code === 3001) {
  // 参数验证失败
  toast.error(response.data.message)
} else {
  // 其他业务错误
  toast.error(`错误(${response.data.code}): ${response.data.message}`)
}
```

## 请求追踪

### X-Request-ID 头
所有API响应都包含唯一的request ID，用于日志追踪：

```
Request:
GET /api/v1/applications
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000

Response:
HTTP/1.1 200 OK
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

### 链路追踪流程
1. 生成或提取 X-Request-ID
2. 注入到上下文
3. 记录到所有日志
4. 返回给客户端

## API实现细节 (Go)

### Handler 模板（直接调用 Service Helper）

```go
// internal/handlers/releases/handler.go

type ReleaseHandler struct {
    deps *services.ServiceDeps  // 一次性注入所有依赖
    log  *logger.Logger
}

// 简单查询：直接调用Repository
func (h *ReleaseHandler) List(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 直接调用repository
    releases, err := h.deps.releaseRepo.List(ctx)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{
        "code": 0,
        "data": releases,
    })
}

// 业务操作：调用 Service Helper（不再经过ApplicationService）
func (h *ReleaseHandler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // 1. 解析请求
    var req struct {
        AppID     int    `json:"app_id"`
        EnvID     int    `json:"env_id"`
        ClusterID int    `json:"cluster_id"`
        Image     string `json:"image"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    
    // 2. 直接调用 Service Helper（核心改进）
    releaseID, err := services.ReleaseApp(
        ctx,
        req.AppID, req.EnvID, req.ClusterID, req.Image,
        getUser(r),
        h.deps,  // ServiceDeps包含所有repository + deployer
    )
    if err != nil {
        h.log.Error("release failed", "error", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 3. 异步操作返回202 Accepted
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]any{
        "code": 0,
        "message": "release accepted",
        "data": map[string]int{
            "id": releaseID,
        },
    })
}

// 回滚：调用 Service Helper
func (h *ReleaseHandler) Rollback(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id, _ := strconv.Atoi(chi.URLParam(r, "id"))
    
    // 直接调用Service Helper
    newReleaseID, err := services.RollbackApp(ctx, id, getUser(r), h.deps)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]any{
        "code": 0,
        "data": map[string]int{"id": newReleaseID},
    })
}
```

### ServiceDeps 结构

```go
// internal/services/deps.go

type ServiceDeps struct {
    releaseRepo   domain.ReleaseRepository
    appRepo       domain.ApplicationRepository
    clusterRepo   domain.ClusterRepository
    workloadRepo  domain.WorkloadTargetRepository
    eventRepo     domain.ReleaseEventRepository
    deployer      *deployers.DeployerFactory
    log           *logger.Logger
    db            *sql.DB
}
```

### 错误处理（Go习惯）

```go
// Service Helper 返回 error，Handler负责HTTP响应

// ✅ Service Helper层
func ReleaseApp(ctx context.Context, appID, envID, clusterID int, image, operator string, deps *ServiceDeps) (int, error) {
    if appID <= 0 {
        return 0, fmt.Errorf("invalid app_id: %d", appID)
    }
    
    workload, err := deps.workloadRepo.GetByAppEnvCluster(ctx, appID, envID, clusterID)
    if err != nil {
        return 0, fmt.Errorf("workload not found: %w", err)
    }
    
    // ... 业务逻辑 ...
    return releaseID, nil
}

// ✅ Handler层（HTTP处理）
func (h *ReleaseHandler) Create(w http.ResponseWriter, r *http.Request) {
    releaseID, err := services.ReleaseApp(r.Context(), ...)
    if err != nil {
        // 直接返回error信息，让HTTP层处理
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // ... 成功响应 ...
}
```

### 中间件

```go
// pkg/middleware/middleware.go

// RequestID 中间件
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-ID")
        if id == "" {
            id = uuid.New().String()
        }
        ctx := context.WithValue(r.Context(), "request_id", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Logging 中间件
func Logging(log *logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            requestID := GetRequestID(r.Context())
            
            next.ServeHTTP(w, r)
            
            duration := time.Since(start)
            log.Info("Request processed",
                "method", r.Method,
                "path", r.URL.Path,
                "duration_ms", duration.Milliseconds(),
                "request_id", requestID,
            )
        })
    }
}
```

## 客户端集成 (Frontend - Axios)

```typescript
// src/api/request.ts

import axios from 'axios'
import { v4 as uuidv4 } from 'uuid'

const request = axios.create({
    baseURL: '/api/v1',
    timeout: 30000,
})

// 请求拦截器：添加Request ID
request.interceptors.request.use((config) => {
    config.headers['X-Request-ID'] = uuidv4()
    return config
})

// 响应拦截器：处理错误
request.interceptors.response.use(
    (response) => {
        if (response.data.code !== 0) {
            throw new Error(response.data.message)
        }
        return response.data.data
    },
    (error) => {
        if (error.response?.status === 202) {
            // 异步操作，返回202，继续轮询
            return error.response.data.data
        }
        throw error
    }
)

export default request
```

---

## Service层集成指南 ⭐ 新增

### Handler + Service 集成模式

从**直接调用Repository**升级到**通过Service调用Repository**：

#### ❌ 旧模式（Repo直接）
```go
func ListApplicationsHandler(repo repository.ApplicationRepository) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        apps, _ := repo.List(r.Context())
        respondJSON(w, 200, apps)
    }
}
```

#### ✅ 新模式（Service中介）
```go
func ListApplicationsHandler(svc *ApplicationService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Service层处理：权限、业务验证
        apps, err := svc.ListApplications(r.Context(), &ListRequest{
            Page: getParam(r, "page"),
            PageSize: getParam(r, "pageSize"),
        })
        if err != nil {
            handleServiceError(w, err, requestID)
            return
        }
        respondJSON(w, 200, apps)
    }
}
```

### Service错误到HTTP映射

```go
// handlers/errors.go

type ServiceError struct {
    Code    string // 业务错误码
    Message string
    Err     error
}

func handleServiceError(w http.ResponseWriter, err error, requestID string) {
    if svcErr, ok := err.(*ServiceError); ok {
        switch svcErr.Code {
        case "NOT_FOUND":
            respondError(w, 404, svcErr.Message, svcErr.Err, requestID)
        case "PERMISSION_DENIED":
            respondError(w, 403, svcErr.Message, svcErr.Err, requestID)
        case "INVALID_INPUT":
            respondError(w, 400, svcErr.Message, svcErr.Err, requestID)
        case "CONFLICT":
            respondError(w, 409, svcErr.Message, svcErr.Err, requestID)
        default:
            respondError(w, 500, "Internal server error", svcErr.Err, requestID)
        }
        return
    }
    respondError(w, 500, "Internal server error", err, requestID)
}
```

### Handler DTO定义

```go
// handlers/dtos.go

// Request DTOs
type CreateApplicationRequest struct {
    Name        string `json:"name" validate:"required,min=1,max=100"`
    Repository  string `json:"repository" validate:"required,url"`
    BuildType   string `json:"build_type" validate:"required,oneof=docker helm"`
    Description string `json:"description" validate:"max=500"`
}

type CreateReleaseRequest struct {
    WorkloadTargetID string `json:"workload_target_id" validate:"required"`
    ImageTag         string `json:"image_tag" validate:"required,len=40"` // SHA长度
}

// Response DTOs (通常与Model相同)
type ApplicationResponse struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Repository  string    `json:"repository"`
    BuildType   string    `json:"build_type"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// 统一响应格式
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
    Type      string `json:"type"`
    Details   string `json:"details"`
    Timestamp string `json:"timestamp"`
    RequestID string `json:"request_id"`
}
```

### 请求验证中间件

```go
// pkg/middleware/validator.go

func ValidateRequest(r *http.Request, req interface{}) error {
    if err := json.NewDecoder(r.Body).Decode(req); err != nil {
        return NewServiceError("INVALID_JSON", "Invalid request body", err)
    }
    
    if err := validate.Struct(req); err != nil {
        // 获取所有验证错误
        var details []string
        for _, err := range err.(validator.ValidationErrors) {
            details = append(details, fmt.Sprintf(
                "Field '%s' failed '%s' validation",
                err.Field(), err.Tag(),
            ))
        }
        return NewServiceError("INVALID_INPUT", 
            "Request validation failed", 
            fmt.Errorf(strings.Join(details, "; ")))
    }
    
    return nil
}
```

