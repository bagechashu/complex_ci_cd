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
      "event_type": "started",
      "event_message": "Release started",
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "release_id": 42,
      "event_type": "validating",
      "event_message": "Validating image: harbor.example.com/company/api-service:v1.2.3",
      "created_at": "2024-01-15T10:30:05Z"
    },
    {
      "id": 3,
      "release_id": 42,
      "event_type": "deploying",
      "event_message": "Updating pod: api-service-5d4c8f7bc9-x7k8m",
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
POST /api/v1/shell-tasks

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

#### 执行Shell任务
```
POST /api/v1/shell-tasks/{id}/execute

响应: 202 Accepted
{
  "code": 0,
  "message": "execution accepted",
  "data": {
    "execution_id": 101,
    "task_id": 5,
    "status": "pending",
    "started_at": "2024-01-15T10:30:00Z"
  }
}
```

## HTTP 状态码规范

| 状态码 | 含义 | 用途 |
|--------|------|------|
| 200 | OK | 同步操作成功 |
| 201 | Created | 资源创建成功 |
| 202 | Accepted | 异步操作接受(发布、回滚、Shell执行) |
| 204 | No Content | 删除成功 |
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | 缺少认证 |
| 403 | Forbidden | 无权限 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 冲突(如重复的唯一键) |
| 500 | Internal Server Error | 服务器错误 |

## 错误响应格式

```json
{
  "code": 400,
  "message": "Invalid cluster name",
  "data": null,
  "error": {
    "type": "ValidationError",
    "details": "Cluster name must not be empty",
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
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

