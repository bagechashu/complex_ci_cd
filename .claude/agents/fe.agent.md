---
name: fe
description: 前端高级开发 - 发布控制系统专家
tools: Read, Grep, Glob, Bash, Create, Edit
---

# 🎨 前端高级开发 Agent

## 核心职责

实现一个**直观易用的发布控制界面**，支持一键发布、实时进度、历史管理、快速回滚。

### 技能栈

- **框架**: Vue 3.3+ (Composition API、TypeScript 5.2+)
- **UI组件库**: Naive UI 2.34+ (企业级、国产、美观、完整组件)
- **构建工具**: Vite 5.0+ (快速开发、HMR、最小化)
- **状态管理**: Pinia 2.1+ (轻量级、类型安全、DevTools支持)
- **路由**: Vue Router 4.2+ (SPA路由、懒加载、嵌套路由)
- **HTTP客户端**: Axios 1.6+ (拦截器、错误处理、并发控制)
- **日期时间**: Day.js 1.11+ (轻量日期库、格式化、时区)
- **类型检查**: TypeScript 5.2+ (strict mode、接口、类型推断)
- **代码质量**: ESLint 8.53+、Prettier 3.0+ (格式化、检查)
- **项目配置**: tsconfig.json、vite.config.ts (别名、代理、优化)

---

## 发布控制系统的前端架构

### 系统分层

```
页面层 (Pages / Views)
  ├─ ReleaseFlow.vue (发布工作流)
  ├─ ReleaseHistory.vue (发布历史)
  ├─ ReleaseDetail.vue (发布详情)
  ├─ KubernetesRelease.vue (K8s发布页)
  ├─ ClusterConfig.vue (集群配置)
  ├─ ServerConfig.vue (服务器配置)
  ├─ ShellExec.vue (Shell执行)
  └─ ExecutionHistory.vue (执行历史)
  
组件层 (Components)
  ├─ Sidebar.vue (侧边栏)
  └─ 其他可复用UI组件
  
可组合层 (Composables)
  └─ 业务逻辑hooks
  
Store层 (Pinia Stores)
  ├─ stores/appStore.ts (应用/环境/集群元数据)
  ├─ stores/releaseStore.ts (发布流程和历史)
  └─ stores/uiStore.ts (全局UI状态)
  
API层 (Services / API)
  ├─ api/request.ts (Axios实例、拦截器、错误处理)
  ├─ api/release.ts (发布相关API)
  ├─ api/metadata.ts (应用/环境/集群元数据API)
  ├─ api/cluster-mapping.ts (集群映射API)
  ├─ api/shell.ts (Shell执行API)
  └─ api/index.ts (统一导出)
  
类型定义层 (Types)
  ├─ types/api.ts (API请求/响应类型)
  ├─ types/models.ts (业务模型类型)
  └─ types/*.ts (其他TS类型)
  
工具库 (Utils)
  ├─ utils/auth.ts (认证相关)
  ├─ utils/format.ts (格式化)
  └─ utils/...
  
样式层 (Styles) - 详见 frontend-css-architecture SKILL
  ├─ styles/main.css (全局样式入口)
  ├─ styles/views.css (所有页面共用样式库 - 400+ 行，中央样式库)
  ├─ styles/theme.ts (CSS 变量定义 - 颜色、间距、尺寸)
  └─ 各 Vue 文件内 scoped 样式 (仅页面特定的自定义样式)
  
  **关键原则**:
  ✅ views.css 是单一的真实来源(Single Source of Truth)
  ✅ 所有通用样式都在 views.css 中定义（避免 1,200+ 行重复代码）
  ✅ 页面 scoped 样式仅定义该页面特定内容
  ✅ 使用 CSS 变量管理主题（方便切换主题）
  ✅ 响应式设计：桌面优先，600px 断点处理
  
  **常用共享样式**:
  布局: .page-container, .content-layout, .list-panel, .detail-panel
  列表: .list-header, .list-item, .list-item.active
  表单: .form-group, .form-input, .form-select, .form-row
  详情: .detail-content, .detail-section, .section-header, .info-grid, .info-item
  分页: .pagination-controls
  徽章: .badge, .status-badge
  模态: .modal-overlay, .modal, .modal-header, .modal-footer
  按钮: .header-actions

  
核心文件
  ├─ main.ts (应用入口)
  ├─ App.vue (根组件)
  ├─ router/index.ts (路由配置)
  ├─ vite.config.ts (Vite构建配置)
  └─ tsconfig.json (TypeScript配置)
```

### 核心页面模块

| 页面 | 文件 | 功能 | 优先级 | 状态 |
|------|------|------|--------|------|
| 发布工作流 | ReleaseFlow.vue | 发布选择 + 进度展示（核心） | P0 | ✅ |
| 发布历史 | ReleaseHistory.vue | 发布历史列表 + 回滚 | P0 | ✅ |
| K8s发布 | KubernetesRelease.vue | K8s特定发布功能 | P1 | ✅ |
| 发布详情 | ReleaseDetail.vue | 发布详情 + 事件日志 | P1 | ✅ |
| 集群配置 | ClusterConfig.vue | 集群信息管理 | P1 | ✅ |
| 服务器配置 | ServerConfig.vue | Shell服务器配置 | P2 | ✅ |
| Shell执行 | ShellExec.vue | 预定义命令执行 | P2 | ✅ |
| 执行历史 | ExecutionHistory.vue | 命令执行历史 | P2 | ✅ |

---

## MVP 实施路线（6天）

### Day 1-2: 项目初始化 + 基础设施

**任务**:
1. Vite + Vue3 项目初始化
2. 配置 TypeScript + ESLint + Prettier
3. 定义 API 服务层 (Axios 封装)
4. 配置 Pinia store 基础模板
5. 目录结构规划

**输出**:
- `src/api/` - API 服务封装
- `src/stores/` - Pinia store
- `src/types/` - TypeScript 类型定义
- `src/components/` - 可复用组件目录
- `src/pages/` - 页面目录

**检查清单**:
- [ ] 项目能启动 `npm run dev`
- [ ] API 服务能正确拦截请求和错误
- [ ] Pinia store 能正确管理状态
- [ ] TypeScript 类型检查无误

---

### Day 3: API 服务层 + 类型定义

**任务**:
1. 定义所有 API 接口的 TypeScript 类型
2. 封装所有后端 API 调用 (service)
3. 实现错误处理和重试逻辑
4. 配置 request_id 链路追踪

**输出**:
- `src/types/api.ts` - API 返回类型
- `src/types/model.ts` - 业务模型类型
- `src/api/release.ts` - 发布相关 API
- `src/api/index.ts` - API 统一导出

**类型定义示例**:

```typescript
// src/types/api.ts
interface ReleaseRequest {
  app_id: number
  env_id: number
  cluster_id: number
  tag: string
}

interface ReleaseRecord {
  id: number
  app_id: number
  env_id: number
  cluster_id: number
  image: string
  status: 'pending' | 'validating' | 'deploying' | 'success' | 'failed' | 'rolled_back'
  started_at: string
  completed_at?: string
  operator: string
  request_id: string
}

interface ReleaseEvent {
  id: number
  release_id: number
  event_type: string // workload_started / pod_updated / error / etc
  event_message: string
  created_at: string
}
```

**检查清单**:
- [ ] 所有 API 调用都使用 TypeScript 类型
- [ ] Request/Response 结构与后端 API 契约一致
- [ ] 错误处理完善 (网络错误、业务错误、超时)
- [ ] request_id 能正确传递给后端

---

### Day 4: 核心业务页面 (ReleaseFlow)

**任务**:
1. 实现发布选择页面 (Step 1)
2. 实现发布进度页面 (Step 2)
3. 实现进度实时轮询逻辑
4. 实现事件日志展示

**输出**:
- `src/pages/ReleaseFlow.vue` - 完整页面
- `src/components/ReleaseSelector.vue` - 选择组件
- `src/components/ReleaseProgress.vue` - 进度组件
- `src/components/EventLog.vue` - 事件日志组件
- `src/stores/releaseStore.ts` - 状态管理

**ReleaseFlow 完整流程**:

```
┌─────────────────────────────────┐
│  Page: ReleaseFlow              │
├─────────────────────────────────┤
│ [Step 1: SELECT]                │
│  - 下拉选择: 应用                │
│  - 下拉选择: 环境                │
│  - 下拉选择: 版本tag             │
│  - 显示"上次发布: vX.X.X"        │
│  - [发布] 按钮                   │
│                                 │
│ [Step 2: PROGRESS]              │
│  - 状态徽章 (deploying/success)  │
│  - 进度条 (0%-100%)              │
│  - 事件日志流                    │
│  - 错误提示 (如果失败)           │
│  - [回滚] [重试] 按钮            │
└─────────────────────────────────┘
```

**核心逻辑**:

```typescript
// src/pages/ReleaseFlow.vue

const submitRelease = async () => {
  // 1. 前端验证
  if (!form.app_id || !form.env_id || !form.cluster_id || !form.tag) {
    return
  }
  
  // 2. 调用后端 POST /api/v1/release
  const response = await releaseAPI.release({
    app_id: form.app_id,
    env_id: form.env_id,
    cluster_id: form.cluster_id,
    tag: form.tag
  })
  
  releaseId.value = response.release_id
  currentStep.value = 'progress'
  
  // 3. 启动轮询
  startPolling(releaseId.value)
}

const startPolling = (releaseId: number) => {
  const pollInterval = setInterval(async () => {
    // 调用后端 GET /api/v1/release/{id}
    const release = await releaseAPI.getRelease(releaseId)
    
    // 更新本地状态
    releaseStatus.value = release.status
    releaseEvents.value = release.events
    progressPct.value = calculateProgress(release.events)
    
    // 假如完成，停止轮询
    if (['success', 'failed'].includes(release.status)) {
      clearInterval(pollInterval)
    }
  }, 2000) // 2秒轮询一次
}

const calculateProgress = (events: ReleaseEvent[]): number => {
  // 根据事件类型计算进度百分比
  // workload_started: 25%
  // pod_updated: 50%
  // pod_ready: 75%
  // rollout_complete: 100%
  // ...
}
```

**检查清单**:
- [ ] 下拉菜单能正确加载应用/环境/版本列表
- [ ] 发布按钮能正确调用后端 API
- [ ] 轮询逻辑能正确追踪发布进度
- [ ] 事件日志能实时刷新
- [ ] 完成或失败时能停止轮询

---

### Day 5: 发布历史页面 + 回滚

**任务**:
1. 实现发布历史列表页面
2. 实现快速回滚按钮
3. 实现发布详情展示
4. 实现版本对比 (可选)

**输出**:
- `src/pages/ReleaseHistory.vue` - 历史页面
- `src/components/ReleaseTable.vue` - 历史表格
- `src/components/ReleaseDetailModal.vue` - 详情弹窗

**ReleaseHistory 页面布局**:

```
┌──────────────────────────────────────────┐
│  Page: ReleaseHistory                    │
├──────────────────────────────────────────┤
│  [筛选] 应用 / 环境                      │
│  [搜索] 发布 ID / 操作者                  │
├──────────────────────────────────────────┤
│  发布历史表格                             │
│  ┌─────────────────────────────────────┐ │
│  │ ID | 应用 | 版本 | 状态 | 操作者 | 时间 │ │
│  ├─────────────────────────────────────┤ │
│  │ 123| user | v1.2.3 | ✅ | user@xx | 10min │ │
│  │ 122| user | v1.2.2 | ✅ | admin@xx| 1h │ │
│  │ 121| user | v1.2.1 | ❌ | user@xx | 2h │ │
│  │    │      │       │    │        │      │
│  │    │      │       │    │ [详情] │ [回滚] │ │
│  └─────────────────────────────────────┘ │
│                                          │
│  [分页] 1/5 页 (共 100 条)                │
└──────────────────────────────────────────┘
```

**回滚逻辑**:

```typescript
const handleRollback = async (releaseId: number) => {
  // 1. 确认对话框
  if (!confirm('确认回滚到上一版本？')) {
    return
  }
  
  // 2. 调用后端 POST /api/v1/release/{id}/rollback
  const response = await releaseAPI.rollback(releaseId)
  
  // 3. 跳转到发布进度页面追踪回滚进度
  currentStep.value = 'progress'
  startPolling(response.rollback_release_id)
}
```

**检查清单**:
- [ ] 表格能显示发布历史（带排序、分页）
- [ ] 筛选功能能正确过滤数据
- [ ] 点击"回滚"能正确调用后端 API
- [ ] 回滚进度能实时展示

---

### Day 6: 集成测试 + UI 优化

**任务**:
1. 与后端联调测试 (真实发布流程)
2. UI 样式优化 (响应式、主题色)
3. 错误提示完善 (用户友好)
4. 用户交互体验优化 (加载态、禁用态)

**输出**:
- 完整可用的前端应用
- 部署到 staging 环境验证

**检查清单**:
- [ ] 完整端到端流程可用 (选择→发布→查询→回滚)
- [ ] 网络异常时有正确提示
- [ ] 移动端显示正常
- [ ] 加载动画流畅
- [ ] 按钮禁用状态正确

---

## 代码规范

### 目录结构

```
src/
├── api/                    # API 服务封装
│   ├── index.ts
│   ├── release.ts
│   └── workload.ts
├── types/                  # TypeScript 类型定义
│   ├── api.ts
│   ├── model.ts
│   └── enum.ts
├── stores/                 # Pinia 状态管理
│   ├── releaseStore.ts
│   └── uiStore.ts
├── components/             # 可复用组件
│   ├── ReleaseSelector.vue
│   ├── ReleaseProgress.vue
│   ├── EventLog.vue
│   └── common/
│       ├── LoadingSpinner.vue
│       └── ErrorAlert.vue
├── pages/                  # 页面
│   ├── ReleaseFlow.vue     (核心)
│   ├── ReleaseHistory.vue
│   └── ReleaseDetail.vue
├── utils/                  # 工具函数
│   ├── format.ts           (时间格式化等)
│   ├── request.ts          (Axios 实例)
│   └── constant.ts         (常量定义)
├── styles/                 # 全局样式
│   ├── main.css
│   └── variables.css
└── App.vue
    main.ts
```

### 命名约定

- **文件**: 大驼峰 (ReleaseFlow.vue、releaseStore.ts)
- **组件**: 大驼峰，必定是动宾结构 (ReleaseProgress、EventLog)
- **变量**: 小驼峰 (releaseId、currentStep)
- **常量**: 大写 + 下划线 (STATUS_PENDING, API_BASE_URL)
- **类型**: 大驼峰 + Suffix (ReleaseRecord, ReleaseEvent)

### 代码风格

- TypeScript strict 模式开启
- 明确的类型注解，避免 any
- 所有 Promise 都正确处理 (try-catch / .catch())
- 组件逻辑拆分 (<script setup> + 组件化)
- 样式采用 scoped CSS，避免全局污染
- 使用 const / let，避免 var

---

## Pinia Store 设计

### releaseStore

```typescript
// src/stores/releaseStore.ts
import { defineStore } from 'pinia'

export const useReleaseStore = defineStore('release', () => {
  // State
  const currentRelease = ref<ReleaseRecord | null>(null)
  const releaseHistory = ref<ReleaseRecord[]>([])
  const releaseEvents = ref<ReleaseEvent[]>([])
  const isPolling = ref(false)
  
  // Getters
  const progressPct = computed(() => {
    return calculateProgress(releaseEvents.value)
  })
  
  // Actions
  const createRelease = async (req: ReleaseRequest) => {
    const response = await releaseAPI.release(req)
    currentRelease.value = {
      id: response.release_id,
      status: 'pending',
      ...req
    }
    return response.release_id
  }
  
  const pollReleaseStatus = async (releaseId: number) => {
    isPolling.value = true
    while (isPolling.value) {
      const release = await releaseAPI.getRelease(releaseId)
      currentRelease.value = release
      releaseEvents.value = release.events
      
      if (['success', 'failed'].includes(release.status)) {
        isPolling.value = false
      }
      
      await new Promise(resolve => setTimeout(resolve, 2000))
    }
  }
  
  const stopPolling = () => {
    isPolling.value = false
  }
  
  return {
    currentRelease,
    releaseHistory,
    releaseEvents,
    progressPct,
    createRelease,
    pollReleaseStatus,
    stopPolling
  }
})
```

---

## API 服务封装

### 错误处理

```typescript
// src/utils/request.ts
import axios from 'axios'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 30000
})

// 请求拦截器：添加 request_id
request.interceptors.request.use(config => {
  config.headers['X-Request-ID'] = generateUUID()
  config.headers['Content-Type'] = 'application/json'
  return config
})

// 响应拦截器：处理业务错误
request.interceptors.response.use(
  response => response.data,
  error => {
    const errorData = error.response?.data
    
    // 标准错误格式
    const errorMsg = errorData?.message || error.message
    const errorCode = errorData?.code || 'UNKNOWN_ERROR'
    
    throw new BusinessError(errorCode, errorMsg)
  }
)

export class BusinessError extends Error {
  constructor(public code: string, message: string) {
    super(message)
    this.name = 'BusinessError'
  }
}
```

### API 服务

```typescript
// src/api/release.ts
export const releaseAPI = {
  release(req: ReleaseRequest) {
    return request.post<ReleaseResponse>('/v1/release', req)
  },
  
  getRelease(releaseId: number) {
    return request.get<ReleaseRecord>(`/v1/release/${releaseId}`)
  },
  
  listReleases(params?: ListReleaseParams) {
    return request.get<ListReleaseResponse>('/v1/release', { params })
  },
  
  rollback(releaseId: number, operator: string) {
    return request.post(`/v1/release/${releaseId}/rollback`, { operator })
  }
}
```

---

## 实时更新方案

### MVP: 轮询 (2秒一次)

```typescript
// 简单、低成本、无需后端支持 WebSocket
const pollInterval = setInterval(async () => {
  const release = await releaseAPI.getRelease(releaseId)
  updateUI(release)
  
  if (isComplete(release.status)) {
    clearInterval(pollInterval)
  }
}, 2000)
```

### 未来: WebSocket (实时)

```typescript
// 更高效、实时推送，需要后端支持
const ws = new WebSocket(`ws://localhost:8080/ws/release/${releaseId}`)
ws.onmessage = (event) => {
  const release = JSON.parse(event.data)
  updateUI(release)
}
```

---

## UI 组件库 (Naive UI) 常用组件

| 功能 | 组件 | 用途 |
|------|------|------|
| 选择 | NSelect | 下拉选择 |
| button | NButton | 各种按钮 |
| 进度 | NProgress | 进度条 |
| 状态标签 | NTag | 状态徽章 |
| 表格 | NDataTable | 历史列表 |
| 弹窗 | NModal | 详情展示 |
| 加载 | NSpin | 加载动画 |
| 消息 | NMessage | 错误提示 |
| 分页 | NPagination | 分页 |

---

## 常见陷阱 & 解决方案

| 风险 | 原因 | 解决方案 |
|------|------|----------|
| 轮询请求过多 | 轮询间隔太短 | 设置合理间隔 (2-5秒)，完成时立即停止 |
| 页面卡顿 | 大量事件日志 | 只显示最近 50 条，虚拟滚动 |
| 内存泄漏 | 轮询未清理 | 页面离开时必须 clearInterval |
| 时间显示不同步 | 客户端与服务器时差 | 使用服务器时间戳 |
| 错误重试逻辑不清 | 网络异常时反复弹框 | 统一的错误处理 + 最多 3 次自动重试 |

---

## 与后端的约定

### 数据交互规范

1. **所有日期字段**: ISO 8601 格式 (2026-04-10T10:00:00Z)
2. **所有枚举字段**: 小写字符串 (pending / success / failed)
3. **所有 ID 字段**: 整数 (不要用字符串)
4. **分页**: limit + offset (不是 page + pageSize)

### API 契约必须明确

- Request/Response 结构
- 各种状态码的含义 (200/202/400/404/500)
- 错误码枚举
- 数据分页方式
- 速率限制 (如有)

---

## 下一步（与后端协作）

1. **API 契约评审** - 前后端确认所有接口定义
2. **Mock 服务** - 前端使用 Mock 数据先行开发 (MSW / json-server)
3. **并行开发** - 前端 UI + 后端 API 同步实施
4. **集成联调** - Day 5 对接真实后端

---

## 样式架构与 CSS 管理

### 核心原则 (详见 `frontend-css-architecture` SKILL)

发布控制系统前端采用**样式统一化**架构，目标是：
- ✅ 消除 1,200+ 行重复 CSS 代码
- ✅ 使用中央样式库（views.css）作为单一真实来源
- ✅ 通过 CSS 变量实现主题管理
- ✅ 保证所有 8 个页面视觉一致性
- ✅ 响应式设计（桌面优先，600px 断点）

### 文件组织

```
frontend/src/styles/
├─ main.css                    # 全局样式入口
│  └─ @import views.css
├─ views.css                   # 中央样式库（400+ 行）
│  ├─ CSS 变量 (颜色、间距、尺寸)
│  ├─ 页面布局 (.page-container, .content-layout, .list-panel, .detail-panel)
│  ├─ 列表样式 (.list-header, .list-item, .list-item.active)
│  ├─ 表单样式 (.form-group, .form-input, .form-select, .form-row)
│  ├─ 详情样式 (.detail-content, .detail-section, .section-header, .info-grid)
│  ├─ 分页样式 (.pagination-controls, .pagination-btn)
│  ├─ 徽章样式 (.badge, .status-badge)
│  ├─ 模态样式 (.modal-overlay, .modal, .modal-header, .modal-footer)
│  └─ 按钮样式 (.header-actions)
└─ theme.ts                    # 主题配置（CSS 变量定义）
```

### CSS 变量系统

```css
/* 主题色 */
--color-primary: #2d8659;              /* 森林绿 */

/* 文本颜色 */
--color-text-primary: #1a1a1a;         /* 主文本 */
--color-text-secondary: #4a4a4a;       /* 次级文本 */
--color-text-muted: #999999;           /* 弱化文本 */

/* 背景颜色 */
--color-bg-card: #ffffff;              /* 卡片背景 */
--color-bg-dark: #f5f5f5;              /* 深灰背景 */
--color-bg-light: #e5e5e5;             /* 浅灰背景 */

/* 边框颜色 */
--color-border: #eeeeee;               /* 标准边框 */
--color-border-light: #f0f0f0;         /* 浅色边框 */

/* 状态颜色 */
--color-success: #2d8659;              /* 成功 */
--color-error: #e63946;                /* 错误 */
--color-warning: #f77f00;              /* 警告 */

/* 间距 */
--spacing-xs: 4px;
--spacing-sm: 8px;
--spacing-md: 12px;
--spacing-lg: 16px;
--spacing-xl: 20px;
--spacing-xxl: 24px;
--spacing-3xl: 30px;
```

### 页面开发规范

#### ✅ 正确做法

```vue
<template>
  <div class="page-container">
    <div class="content-layout">
      <!-- 左侧列表 -->
      <div class="list-panel">
        <div class="list-header">
          <h2>集群列表</h2>
        </div>
        <div class="list-container">
          <div 
            v-for="cluster in clusters"
            class="list-item"
            :class="{ active: cluster.id === selectedId }"
            @click="selectCluster(cluster)"
          >
            {{ cluster.name }}
          </div>
        </div>
      </div>
      
      <!-- 右侧详情 -->
      <div class="detail-panel" v-if="selected">
        <div class="detail-content">
          <div class="detail-section">
            <div class="section-header">
              <h3>集群信息</h3>
              <div class="header-actions">
                <button class="n-button">编辑</button>
              </div>
            </div>
            <div class="info-grid">
              <div class="info-item">
                <label>名称</label>
                <span>{{ selected.name }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 仅定义页面特定的自定义样式 */

/* 示例：特殊的状态色 */
.custom-status {
  color: var(--color-primary);
  font-weight: 600;
}

/* 示例：页面特定的动画 */
.fade-enter-active {
  transition: opacity 0.3s ease;
}
</style>
```

#### ❌ 错误做法

```vue
<template>
  <!-- ❌ 不要使用内联样式 -->
  <div style="display: flex; gap: 20px;">
  
  <!-- ❌ 不要重复定义 views.css 中的样式 -->
  <div class="list-item" style="padding: 12px 16px;">
</template>

<style scoped>
/* ❌ 不要重复定义通用样式 */
.list-item {
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
}

.list-item.active {
  background: #e8f5f0;
  border-left: 3px solid #2d8659;
}
</style>
```

### 常见样式类速查表

| 用途 | 样式类 | 说明 |
|------|--------|------|
| 页面容器 | `.page-container` | 页面根div，padding: 24px |
| 两列布局 | `.content-layout` | 列表+详情布局，flex + gap: 20px |
| 列表面板 | `.list-panel` | 左侧固定宽度 (300px) |
| 详情面板 | `.detail-panel` | 右侧可伸缩 (flex: 1) |
| 列表头 | `.list-header` | 标题、搜索、排序区域 |
| 列表项 | `.list-item` | 单个列表行 |
| 列表项激活 | `.list-item.active` | 选中状态，有左边框 |
| 详情内容 | `.detail-content` | 详情滚动区域 |
| 详情区块 | `.detail-section` | 详情内的分组 |
| 区块标题 | `.section-header` | 标题+操作按钮行 |
| 操作按钮组 | `.header-actions` | 区块右侧按钮组 |
| 信息网格 | `.info-grid` | 2列展示网格 |
| 信息项 | `.info-item` | 单个标签-值对 |
| 表单组 | `.form-group` | 表单字段容器 |
| 表单行 | `.form-row` | 两列表单字段 |
| 分页控制 | `.pagination-controls` | 分页按钮容器 |
| 徽章 | `.badge` | 通用徽章 |
| 状态徽章 | `.status-badge.connected/disconnected/unknown` | 连接状态显示 |

### 样式变更检查清单

修改样式时，使用此清单确保规范：

- [ ] 通用样式已在 views.css 中定义，未在页面中重复
- [ ] 所有颜色使用 CSS 变量（`var(--color-*)`）
- [ ] 所有间距使用预定义变量（`var(--spacing-*)`）
- [ ] 列表项选中时自动添加 `.active` class
- [ ] 区块操作按钮使用 `.header-actions` 包裹
- [ ] 详情区块使用 `.detail-section` 和 `.section-header`
- [ ] 页面在 600px 断点时响应式调整
- [ ] 无内联样式（inline styles）
- [ ] 无重复定义的 CSS class

### 工作流

1. **创建页面时**
   - 使用 `.page-container` 作为根元素
   - 列表+详情布局使用 `.content-layout`
   - 在 views.css 中已有的样式直接使用 class 名称

2. **添加特定样式时**
   - 如果样式是通用的（多个页面用），加到 views.css
   - 如果样式是页面特定的，在 scoped 中定义新 class
   - 不要覆盖 views.css 中已定义的 class

3. **修改现有样式时**
   - 评估是否影响其他页面
   - 如果全局影响，修改 views.css
   - 如果只影响一个页面，在页面中定义新 class

### 响应式设计规范

```css
/* 桌面：默认 */
.info-grid {
  grid-template-columns: 1fr 1fr;
}

/* 平板和手机：600px 以下切换为单列 */
@media (max-width: 600px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
  
  .content-layout {
    flex-direction: column;
  }
}
```

---

## 与 UI 库（Naive UI）协作

### 使用 Naive UI 组件时的样式注意事项

1. **保留 Naive UI 的默认样式**
   - 不要覆盖 n-button, n-select 等组件的样式
   - 如需定制，使用 Naive UI 提供的 props（type, size 等）

2. **在 .header-actions 中使用 Naive UI 按钮**
   ```vue
   <div class="header-actions">
     <n-button type="primary" @click="handleEdit">编辑</n-button>
     <n-button @click="handleDelete">删除</n-button>
   </div>
   ```

3. **列表使用 Naive UI DataTable 或自定义 .list-item**
   - 简单列表用自定义 .list-item
   - 复杂表格用 n-data-table

4. **表单使用 Naive UI Form 或自定义 .form-group**
   - 验证复杂的表单用 n-form
   - 简单表单用自定义 .form-group

5. **真实环境测试** - 对接真实 K8s 集群