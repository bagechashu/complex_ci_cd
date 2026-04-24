# 🤝 前后端 Agent 协作约定

> 确保 BE Agent 和 FE Agent 高效协作，快速交付发布控制系统 MVP

---

## 核心原则

### 分工清晰

| 领域 | BE Agent | FE Agent |
|------|----------|----------|
| **数据模型** | ✅ 设计 + 实现 | ❌ |
| **业务逻辑** | ✅ 实现 | ❌ |
| **API 设计** | ✅ 实现 | ✅ 消费 |
| **UI 组件** | ❌ | ✅ 实现 |
| **状态管理** | ❌ | ✅ 实现 (Pinia) |
| **HTTP 通讯** | ✅ 返回 JSON | ✅ 发送 JSON |
| **实时更新** | ✅ 提供事件日志 API | ✅ 轮询/WebSocket |

### 接口契约优先

在代码实现前，必须有**明确、文档化的后端 API 契约**（OpenAPI/Swagger），FE 基于契约进行开发。

---

## API 契约规范

> 📌 **API 响应格式的完整定义详见**: `skills/api-design/SKILL.md`

### 1. 统一的 Request 格式

```typescript
// 所有 POST 请求都遵循此结构
{
  "field1": value1,
  "field2": value2,
  // ...
  "request_id"?: string // 可选，由前端生成，后端记录用于链路追踪
}
```

### 2. 统一的 Response 格式

> 响应格式详见 `skills/api-design/SKILL.md`

#### 成功 (200/201/202)

```typescript
interface SuccessResponse<T> {
  code: 0                // HTTP 状态码或固定 0 表示成功
  message: string        // 用户可读消息
  data: T               // 实际数据
}
```

例:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "app_id": 1,
    "status": "pending"
  }
}
```

#### 错误 (400/404/500)

```typescript
interface ErrorResponse {
  code: number          // HTTP 状态码 (400, 404, 500 等)
  message: string       // 用户可读错误信息
  error?: any           // 错误详情（可选）
}
```

例:
```json
{
  "code": 400,
  "message": "invalid app_id"
}
```

### 3. HTTP 状态码规范

| 场景 | 状态码 | 说明 |
|------|--------|------|
| 发起发布 | **202 Accepted** | 异步操作，发布进行中 |
| 查询资源 | 200 OK | 成功 |
| 参数错误 | 400 Bad Request | 格式/必填项错误 |
| 资源不存在 | 404 Not Found | 应用/环境/集群不存在 |
| 无权限 | 403 Forbidden | 用户无权发布此环境 |
| 服务错误 | 500 Internal Server Error | 后端异常 |

### 4. 所有日期字段

**格式**: ISO 8601 (带时区)
```
"2026-04-10T10:30:45.123Z"  // ✅ 标准格式
"2026-04-10 10:30:45"       // ❌ 不规范，会导致前后端时差问题
```

前端必须使用服务器返回的时间戳，不要用本地时间。

### 5. 枚举字段

**格式**: 小写，下划线分隔
```typescript
status: "pending" | "validating" | "deploying" | "success" | "failed" | "rolled_back"
event_type: "workload_started" | "pod_updated" | "rollout_complete" | "error"
```

### 6. ID 和主键

**格式**: 整数 (NOT 字符串)
```json
{
  "app_id": 1,
  "env_id": 2,
  "cluster_id": 1,
  "release_id": 123
}
```

---

## 关键 API 约定

### POST /api/v1/release (发起发布)

**请求**:
```json
{
  "app_id": 1,
  "env_id": 2,
  "cluster_id": 1,
  "tag": "v1.2.3"
}
```

**响应** (202 Accepted):
```json
{
  "code": 0,
  "message": "accepted",
  "data": {
    "id": 123,
    "app_id": 1,
    "env_id": 2,
    "cluster_id": 1,
    "status": "pending"
  }
}
```

**FE 期望**: 
- 立即返回 (不要阻塞)
- 获得 `release_id` 
- 开始轮询 GET /api/v1/release/{release_id}

---

### GET /api/v1/release/{release_id} (查询进度)

**响应** (200):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 123,
    "app_id": 1,
    "env_id": 2,
    "cluster_id": 1,
    "image": "harbor.com/project/service:v1.2.3",
    "status": "deploying",
    "started_at": "2026-04-10T10:00:00Z",
    "completed_at": null,
    "operator": "user@example.com",
    "events": [
      {
        "id": 1,
        "release_id": 123,
        "event_type": "workload_started",
        "event_message": "K8s workload patch started",
        "created_at": "2026-04-10T10:00:05Z"
      },
      {
        "id": 2,
        "release_id": 123,
        "event_type": "pod_updated",
        "event_message": "1/3 pods updated",
        "created_at": "2026-04-10T10:00:10Z"
      }
    ],
    "error_msg": null
  }
}
```

**FE 期望**:
- events 按时间排序
- 可以计算进度百分比 (基于 event_type)
- 完成时 status = "success" 或 "failed"

---

### GET /api/v1/release (列表查询)

**查询参数**:
```
GET /api/v1/release?app_id=1&env_id=2&limit=20&offset=0
```

**响应** (200):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "limit": 20,
    "offset": 0,
    "items": [
      {
        "id": 123,
        "app_id": 1,
        "env_id": 2,
        "image": "...:v1.2.3",
        "status": "success",
        "started_at": "2026-04-10T10:00:00Z",
        "completed_at": "2026-04-10T10:05:00Z",
        "operator": "user@example.com"
      }
    ]
  }
}
```

**FE 期望**:
- 支持分页 (limit + offset)
- 支持排序 (order_by + sort)
- 支持筛选 (app_id, env_id, status)

---

### POST /api/v1/release/{release_id}/rollback (回滚)

**请求**:
```json
{
  "operator": "user@example.com"
}
```

**响应** (202 Accepted):
```json
{
  "code": 0,
  "message": "accepted",
  "data": {
    "id": 124,
    "status": "pending"
  }
}
```

**FE 期望**:
- 返回新的 release_id
- 跳转到发布进度页，轮询新 ID

---

### GET /api/v1/workload-target

**查询参数**:
```
GET /api/v1/workload-target?app_id=1&env_id=2
```

**响应** (200):
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "app_id": 1,
      "env_id": 2,
      "cluster_id": 1,
      "namespace": "prod",
      "workload_name": "user-service",
      "workload_type": "Deployment",
      "container_name": "user-service",
      "registry_domain": "harbor.com",
      "image_repo": "project/user-service"
    }
  ]
}
```

**FE 期望**:
- 用于展示发布目标的信息
- 帮助用户确认将要发布到哪个环境/集群

---

## 数据类型定义

### 前端必须定义的 TypeScript 类型

```typescript
// src/types/api.ts

// ============ Request Types ============

interface ReleaseRequest {
  app_id: number
  env_id: number
  cluster_id: number
  tag: string
}

interface RollbackRequest {
  operator: string
}

// ============ Response Types ============

enum ReleaseStatus {
  PENDING = 'pending',
  VALIDATING = 'validating',
  DEPLOYING = 'deploying',
  SUCCESS = 'success',
  FAILED = 'failed',
  ROLLED_BACK = 'rolled_back'
}

enum EventType {
  workload_STARTED = 'workload_started',
  POD_UPDATED = 'pod_updated',
  ROLLOUT_COMPLETE = 'rollout_complete',
  ERROR = 'error'
}

interface ReleaseEvent {
  id: number
  release_id: number
  event_type: EventType
  event_message: string
  created_at: string   // ISO 8601
}

interface ReleaseRecord {
  id: number
  app_id: number
  env_id: number
  cluster_id: number
  image: string
  status: ReleaseStatus
  started_at: string
  completed_at?: string
  operator: string
  events: ReleaseEvent[]
  error_msg?: string
}

interface WorkloadTarget {
  id: number
  app_id: number
  env_id: number
  cluster_id: number
  cluster_name: string
  k8s_namespace: string
  k8s_workload: string
  registry_domain: string
  image_repo: string
}

interface Application {
  id: number
  name: string
  repo: string
  build_type: string
}

interface Environment {
  id: number
  name: string
  rank: number  // 1=prod, 2=staging, 3=dev
}

interface Cluster {
  id: number
  name: string
  type: string  // 仅支持 kubernetes，其他部署方式通过 shell_server 实现
}

// ============ API Response Wrappers ============

interface ApiResponse<T> {
  code: 0 | number       // 0 表示成功，HTTP 状态码表示错误
  message: string
  data?: T
  error?: {
    type?: string
    details?: string
  }
}

interface ListResponse<T> {
  total: number
  limit: number
  offset: number
  items: T[]
}

interface ListResponse<T> {
  total: number
  limit: number
  offset: number
  items: T[]
}
```

---

## 工作流程

### Day 1: API 契约评审 ✅

1. **BE 负责**: 编写完整 OpenAPI/Swagger 文档，定义所有接口
2. **FE 负责**: 审核接口设计，提出不合理之处
3. **输出**: 双方签署的 API 文档 (final.openapi.yaml)

### Day 2-3: 并行开发

**后端**:
- 数据库设计、表结构创建
- Repository 实现
- API 框架搭建
- Mock 数据准备

**前端**:
- 类型定义 (基于 API 文档)
- API 服务封装 (基于 Mock)
- Store 实现
- UI 组件开发 (使用 Mock 数据)

### Day 4-5: 集成测试

1. 后端更新 Mock → 真实逻辑
2. 前端指向真实后端地址
3. 端到端流程测试
4. 错误场景测试

### Day 6: 优化 & 上线

1. 性能测试 (并发发布)
2. 用户体验优化
3. 部署到 staging 环境

---

## 沟通规范

### Slack/钉钉 快速反馈

| 情况 | 频道 | 内容 |
|------|------|------|
| 接口设计问题 | #release-system | 具体的 API 改进建议 |
| 测试发现的 Bug | #release-system | 完整的复现步骤 |
| 阻塞问题 | @mention | 紧急问题直接 @ |
| 日版本同步 | #release-system | 每日更新进度 (EOD) |

### 每日进度同步

**格式**:
```
✅ 完成: 
  - API 框架搭建
  - workload_target 表设计

🚀 今日进行中:
  - K8sDeployer 实现

🔴 阻塞:
  - 等待 Harbor registry 访问权限

📅 明日计划:
  - Deploy Service 完整实现
```

---

## Bug 报告规范

### 后端发现的前端 Bug

```
标题: [FE BUG] 发布列表分页不工作

详情:
- 环境: localhost:5173
- 步骤:
  1. 打开发布历史页面
  2. 设置 limit=5
  3. 点击下一页
- 预期: 显示第 2 页
- 实际: 还是显示第 1 页
- 网络: GET /api/v1/release?limit=5&offset=5 返回 200, 正确数据

代码位置: src/pages/ReleaseHistory.vue:123
```

### 前端发现的后端 Bug

```
标题: [BE BUG] Rollback API 返回错误响应格式

详情:
- 请求: POST /api/v1/release/123/rollback
- 期望返回: { code, data: { rollback_release_id, ... } }
- 实际返回: 直接返回 { release_id } (格式不一致)

影响: 前端 response.data.rollback_release_id 获取 undefined
```

---

## 部署前检查清单

### BE 部署前

- [ ] 所有接口都返回约定的 JSON 格式
- [ ] 错误时含有 code 和 request_id
- [ ] 202 Accepted 用于异步操作
- [ ] request_id 来自请求头 X-Request-ID
- [ ] 数据库事务保证一致性
- [ ] 日期字段都是 ISO 8601
- [ ] SQL 性能通过了压力测试

### FE 部署前

- [ ] 所有 API 调用都使用 TypeScript 类型
- [ ] 错误处理完善 (网络错误、业务错误)
- [ ] 轮询逻辑在页面卸载时清理
- [ ] 所有输入都经过前端验证
- [ ] 响应式设计适配移动端
- [ ] 加载态和禁用态显示正确
- [ ] 浏览器控制台无 TypeScript 错误

---

## 总结

✅ **API 契约是唯一的真理**  
✅ **并行开发，减少依赖**  
✅ **每日同步进度和问题**  
✅ **统一的数据格式和错误处理**  
✅ **充分的类型检查 (TypeScript)**  

🎯 **目标**: 6 天高质量交付发布控制系统 MVP
