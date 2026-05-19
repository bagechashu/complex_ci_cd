# 🎯 .claude 配置快速参考卡

> 快速查看 Skills 和 Agents 的业务对应关系、关键文件位置、核心概念

---

## 📊 核心概念速查表

| 业务概念 | BE 实现 | FE 实现 | 相关 Skill |
|---------|--------|--------|-----------|
| **发布流程** | ReleaseService | releaseStore | service-layer, deploy-strategy, api-design |
| **异步部署** | Goroutine + context | 轮询 event API | deploy-strategy, pinia-stores |
| **命令执行** | ShellService | shellStore | shell-service, api-design |
| **元数据缓存** | Repository | appStore | pinia-stores |
| **安全加密** | crypto/aes | N/A | database-design, shell-service |
| **样式管理** | N/A | CSS 变量 + views.css | frontend-css-architecture |
| **UI 组件** | N/A | Naive UI | naive-ui |

---

## 🗂️ 关键文件位置速查

### 后端关键文件

| 功能 | 文件路径 | 行数 | 说明 |
|------|---------|------|------|
| **Service 层** | `backend/internal/services/container.go` | ~100 | DI 容器实现 |
| | `backend/internal/services/release_service.go` | ~200 | 发布管理核心 |
| | `backend/internal/services/shell_service.go` | ~150 | SSH 执行服务 |
| **数据模型** | `backend/internal/models/*.go` | 多文件 | 业务模型定义 |
| **数据访问** | `backend/internal/repository/*.go` | 多文件 | 数据库操作 |
| **API 处理** | `backend/internal/handlers/*/handlers.go` | 多文件 | HTTP 请求处理 |
| **部署策略** | `backend/internal/deployers/k8s_deployer.go` | ~200 | K8s 部署实现 |
| **数据库** | `backend/internal/database/sqlite.go` | ~300 | Schema 定义 |

### 前端关键文件

| 功能 | 文件路径 | 行数 | 说明 |
|------|---------|------|------|
| **Stores** | `frontend/src/stores/appStore.ts` | ~80 | 元数据缓存 |
| | `frontend/src/stores/releaseStore.ts` | ~100 | 发布流程 |
| | `frontend/src/stores/shellStore.ts` | ~100 | 命令执行 |
| **API 集成** | `frontend/src/api/release.ts` | ~60 | 发布 API |
| | `frontend/src/api/shell.ts` | ~50 | Shell API |
| | `frontend/src/api/request.ts` | ~100 | Axios 实例 |
| **页面** | `frontend/src/views/ReleaseFlow.vue` | ~200 | 发布工作流 |
| | `frontend/src/views/ShellCommandExecution.vue` | ~200 | 命令执行 |
| **样式** | `frontend/src/styles/views.css` | ~400 | 中央样式库 |
| | `frontend/src/theme.ts` | ~50 | CSS 变量 |

---

## 🔀 业务流程快速对应

### 场景 1: 用户发起一个发布

```
用户交互                后端处理                       数据存储
─────────────────────────────────────────────────────────────

1. 打开 ReleaseFlow.vue
   ↓ 加载应用、环境、集群
   ↓ (GET /api/v1/applications 等)
   ↓
   ├─→ handlers/metadata_handlers.go
       └─→ ApplicationService.GetAll()
           └─→ application_repo.GetAll()
               └─→ SQLite: SELECT * FROM application

2. 用户填表单 (appId=1, envId=2, clusterId=3, image=v2.0.0)
   └─ releaseStore.createRelease()

3. 点击 "发布" 按钮
   ↓ POST /api/v1/release
   ├─→ handlers/release_handlers.go:CreateRelease()
       ├─→ ReleaseService.Release()
       │   ├─ 验证 workload_target 配置
       │   ├─ 创建 release_record (status=pending)
       │   └─ 异步执行部署 (go s.deployAsync)
       └─→ release_record_repo.Create()
           └─ SQLite: INSERT INTO release_record

4. 后台异步执行
   ├─ deployers/k8s_deployer.go:Deploy()
   │  └─ 更新 K8s Deployment 镜像
   │
   └─ release_event_repo.Create()
      └─ SQLite: INSERT INTO release_event

5. 前端轮询获取进度
   ↓ GET /api/v1/release/{id}/events
   ├─→ handlers/release_handlers.go:GetReleaseEvents()
       └─→ ReleaseService.ListReleaseEvents()
           └─→ release_event_repo.GetByReleaseID()
               └─ SQLite: SELECT * FROM release_event

6. ReleaseDetail.vue 显示实时日志
   └─ 当 event_type='deployment_complete' 时停止轮询
```

### 场景 2: 用户执行一个 Shell 命令

```
用户交互                后端处理                       数据存储
─────────────────────────────────────────────────────────────

1. 打开 ShellCommandExecution.vue
   ↓ 加载已发布命令
   ↓ (GET /api/v1/shell-commands/published)
   ↓
   ├─→ handlers/shell_handlers.go
       └─→ ShellService.GetPublishedCommands()
           └─→ shell_command_repo.GetPublished(is_published=true)
               └─ SQLite: SELECT * FROM shell_command WHERE is_published=1

2. 显示命令列表（按服务器分组）
   └─ shellStore.shellCommands = [...]

3. 用户选择命令和服务器
   └─ shellStore 中记录选择

4. 点击 "执行" 按钮
   ↓ POST /api/v1/shell-commands/execute
   ├─→ handlers/shell_handlers.go:ExecuteCommand()
       ├─→ ShellService.ExecuteCommand(commandId=5, serverId=2)
       │   ├─ 验证命令是否已发布
       │   ├─ 获取服务器配置 (SSH 连接信息)
       │   ├─ 建立 SSH 连接 (缓存)
       │   ├─ 执行远程命令
       │   └─ 捕获输出和退出码
       └─→ shell_command_execution_repo.Create()
           └─ SQLite: INSERT INTO shell_command_execution

5. 前端轮询获取执行结果
   ↓ GET /api/v1/shell-commands/executions/{executionId}
   ├─→ handlers/shell_handlers.go:GetExecution()
       └─→ ShellService.GetExecutionDetail()
           └─→ shell_command_execution_repo.GetByID()
               └─ SQLite: SELECT * FROM shell_command_execution

6. ExecutionHistory.vue 显示结果
   └─ 当 status='success' 或 'failed' 时停止轮询
```

---

## 🎯 Skill 对应的三个关键问题

| Skill | 解决的核心问题 | 输出物 | 代码位置 |
|-------|-------------|--------|---------|
| **service-layer** | 如何组织业务逻辑？ | Service 类设计 + DI 容器 | `backend/internal/services/` |
| **api-design** | 如何定义 API 契约？ | 端点设计 + 响应格式 + 错误码 | `backend/internal/handlers/routes.go` |
| **database-design** | 如何设计数据库？ | Schema + 表关系 + 加密方案 | `backend/internal/database/sqlite.go` |
| **deploy-strategy** | 如何支持多部署方式？ | 策略模式 + K8s 实现 + 异步执行 | `backend/internal/deployers/` |
| **shell-service** | 如何安全执行 Shell？ | 白名单 + SSH 连接管理 + 日志记录 | `backend/internal/services/shell_service.go` |
| **pinia-stores** | 如何管理前端状态？ | 4 个核心 Store 的设计 | `frontend/src/stores/` |
| **naive-ui** | 如何高效使用组件库？ | 常用组件范例 + 主题定制 | `frontend/src/views/` |
| **frontend-css-architecture** | 如何统一样式管理？ | CSS 变量系统 + views.css | `frontend/src/styles/` |
| **tech-stack** | 项目用了什么技术？ | 完整的技术选型清单 | 各文件 |

---

## 🔗 Skill 间的依赖关系（简化版）

```
┌──────────────────┐
│   tech-stack     │ ← 底层参考（所有 Skills）
└────────┬─────────┘

┌────────▼────────────────┐
│  database-design         │ ← 数据层基础
└────────┬────────────────┘

     ┌───┴────┬──────────┐
     │        │          │
┌────▼──┐  ┌──▼──┐  ┌──▼──────────┐
│service │  │shell│  │deploy       │
│-layer  │  │-service  │-strategy    │
└────┬──┘  └──┬──┘  └──┬──────────┘
     │        │        │
     └────┬───┴────┬───┘
          │        │
      ┌───▼────────▼──┐
      │   api-design   │ ← API 层定义（FE 依赖）
      └────────┬───────┘
               │
         ┌─────▼──────────────┐
         │  pinia-stores      │ ← FE 状态管理
         └─────┬──────────────┘
              │
       ┌──────┴──────────┐
       │                 │
    ┌──▼───┐         ┌──▼─────────┐
    │naive │         │frontend-css │
    │-ui   │         │-architecture│
    └──────┘         └─────────────┘
```

---

## ⚙️ 核心实现模式速记

### 后端模式

**Service 构造函数**:
```go
func New{Service}(
    // 依赖注入的 Repository 接口
    repo repository.{Repo}Repository,
    log *logger.Logger,
    // 可选的其他 Service
    otherService *{OtherService},
) *{Service} {
    return &{Service}{
        repo: repo,
        log: log,
        otherService: otherService,
    }
}
```

**Service 方法**:
```go
func (s *{Service}) {Method}(ctx context.Context, input {InputType}) ({OutputType}, error) {
    // 1. 验证输入
    // 2. 调用 Repository
    // 3. 业务逻辑处理
    // 4. 返回结果或错误
    return result, nil
}
```

**API 处理函数**:
```go
func {Handler}(w http.ResponseWriter, r *http.Request) {
    // 1. 解析请求参数
    // 2. 验证请求
    // 3. 从容器获取 Service
    service := container.Get(ServiceKey)
    // 4. 调用 Service 方法
    result, err := service.{Method}(r.Context(), params)
    // 5. 返回响应 (200 + code 字段)
    response.Success(w, result)
}
```

### 前端模式

**Store 定义**:
```typescript
export const use{Store}Store = defineStore('{store}', () => {
  // 状态
  const state1 = ref<Type1>(defaultValue)
  const state2 = ref<Type2>(defaultValue)
  
  // 计算属性
  const computed1 = computed(() => ...)
  
  // 方法
  const action1 = async (param: T) => {
    const result = await api.get('/endpoint')
    state1.value = result.data
  }
  
  return {
    // 导出状态
    state1, state2,
    // 导出方法
    action1,
  }
})
```

**页面使用**:
```vue
<script setup lang="ts">
const store = use{Store}Store()

onMounted(async () => {
  await store.fetchData()
})
</script>

<template>
  <!-- 使用状态 -->
  <div>{{ store.state1 }}</div>
  <!-- 调用方法 -->
  <button @click="store.action1()">Click</button>
</template>

<style scoped>
/* 页面特定样式 */
</style>
```

**轮询模式**:
```typescript
const pollData = async () => {
  try {
    const result = await api.get(`/api/{resource}/{id}/detail`)
    if (TERMINAL_STATUSES.includes(result.status)) {
      // 终态：停止轮询
      clearInterval(pollTimer)
      return
    }
    // 继续轮询
    pollTimer = setTimeout(pollData, POLL_INTERVAL)
  } catch (error) {
    console.error('轮询失败:', error)
    clearInterval(pollTimer)
  }
}
```

---

## 📋 每天一个 Agent 需要实现的任务

### BE Agent 的 6 天计划

| Day | 主要任务 | 涉及 Skill | 输出物 |
|-----|---------|-----------|--------|
| 1-2 | 数据库设计 + Repository 实现 | database-design, tech-stack | Schema + Repos |
| 3-4 | Service 层实现 | service-layer, shell-service | Services + DI 容器 |
| 5 | API 端点实现 | api-design | Go-chi 路由 + Handlers |
| 6 | Deployer 实现 + 测试 | deploy-strategy, tech-stack | K8sDeployer + Tests |

### FE Agent 的 6 天计划

| Day | 主要任务 | 涉及 Skill | 输出物 |
|-----|---------|-----------|--------|
| 1-2 | Stores 设计 + 初始化 | pinia-stores, tech-stack | 4 个 Stores |
| 3 | API 集成 | api-design, tech-stack | Axios + API 模块 |
| 4-5 | 页面实现 | naive-ui, pinia-stores | Release + Shell 页面 |
| 5-6 | 样式调整 + 测试 | frontend-css-architecture, naive-ui | CSS + E2E |

---

## 🚀 快速启动检查清单

### 新开发者快速理解

- [ ] 读 `架构设计.md` (系统为什么这样设计)
- [ ] 读 `AGENT_COLLABORATION.md` (前后端如何协作)
- [ ] 读本文档了解业务逻辑
- [ ] 浏览代码目录结构对应
- [ ] 运行 `make test` 验证代码运行

### 新增功能快速参考

| 如果要增加... | 需要修改的文件 | 涉及 Skills |
|-------------|-------------|-----------|
| 新的数据模型 | `models/`, `db/sqlite.go`, `repository/` | database-design |
| 新的业务逻辑 | `services/` | service-layer |
| 新的 API 端点 | `handlers/`, `routes.go` | api-design |
| 新的发布方式 | `deployers/` | deploy-strategy |
| 新的前端页面 | `views/`, `stores/` | pinia-stores, naive-ui |
| 新的样式 | `styles/views.css`, `theme.ts` | frontend-css-architecture |

---

## 📞 问题排查快速索引

| 问题 | 可能原因 | 查看位置 |
|------|---------|---------|
| 后端启动失败 | 数据库 Schema 错误、端口被占用 | `database/sqlite.go`, `config/config.go` |
| API 返回 null | Repository 查询返回空 | `repository/` 中的查询逻辑 |
| 前端无法加载数据 | API 端点不存在、CORS 未配置 | `api-design` SKILL, `middleware/` |
| 发布卡住 | K8s 连接超时、Goroutine 泄漏 | `deploy-strategy` SKILL, `deployers/` |
| 样式混乱 | CSS 冲突、Naive UI 样式未覆盖 | `styles/views.css`, `theme.ts` |
| 轮询不停止 | 终态判断错误 | `pinia-stores` SKILL 中的轮询逻辑 |

---

**最后更新**: 2026-05-19  
**用途**: 快速查阅 .claude 配置与前后端代码的对应关系
