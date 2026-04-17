# 发布控制系统 - 前端架构设计文档

## 整体架构

```
┌─────────────────────────────────────────────────────┐
│            Pages (Vue Components)                   │
│  ├─ ReleaseFlow (4步向导)                          │
│  ├─ ReleaseHistory (历史列表)                       │
│  └─ ReleaseDetail (详情展示)                        │
└───────────────────┬─────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────┐
│        Components & Composables                     │
│  ├─ 可复用 UI 组件                                 │
│  ├─ 业务逻辑 Hook                                  │
│  └─ 共享功能                                        │
└───────────────────┬─────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────┐
│       Store Layer (Pinia)                           │
│  ├─ appStore       (应用/环境/集群元数据)          │
│  ├─ releaseStore   (发布流程和历史)                │
│  └─ uiStore        (全局 UI 状态)                  │
└───────────────────┬─────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────┐
│      API Service Layer                              │
│  ├─ request.ts     (Axios 实例和拦截器)           │
│  ├─ release.ts     (发布相关 API)                  │
│  └─ metadata.ts    (元数据 API)                    │
└───────────────────┬─────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────┐
│      Backend API (go-chi)                           │
│  /api/v1/releases
│  /api/v1/applications
│  /api/v1/environments
│  /api/v1/clusters
│  /api/v1/workload-targets
└─────────────────────────────────────────────────────┘
```

## 数据流向

### 初始化流程

```
应用启动
  │
  ├─> main.ts
  │    ├─> createApp()
  │    ├─> use(Pinia)
  │    ├─> use(Router)
  │    └─> mount()
  │
  └─> App.vue
       ├─> onMounted()
       │    └─> appStore.initializeMetadata()
       │         ├─> fetchApplications()
       │         ├─> fetchEnvironments()
       │         ├─> fetchClusters()
       │         └─> fetchWorkloadTargets()
       │
       └─> router-view (当前页面)
```

### 发布流程

```
用户在 ReleaseFlow 页面
  │
  1. 选择应用 (Step 1)
     └─> form.app_id = selected
  
  2. 选择环境 (Step 2)
     └─> form.env_id = selected
     └─> appStore.getAvailableClusters() 加载可用集群
  
  3. 选择集群 (Step 3)
     └─> form.cluster_id = selected
  
  4. 输入镜像 (Step 4)
     └─> form.image = input
  
  5. 提交发布 (Step 5)
     │
     └─> releaseStore.createRelease()
         │
         ├─> API: POST /api/v1/releases
         │    └─> backend async goroutine
         │
         └─> startPolling(releaseId)
             │
             └─> 每 2 秒调用
                 ├─> releaseStore.fetchReleaseStatus()
                 │    └─> API: GET /api/v1/releases/{id}
                 │
                 ├─> releaseStore.fetchReleaseEvents()
                 │    └─> API: GET /api/v1/releases/{id}/events
                 │
                 └─> 当状态为 success/failed 时停止轮询
                     └─> stopPolling()
                     └─> frontend 显示完成状态
```

### 历史查看流程

```
用户在 ReleaseHistory 页面
  │
  ├─> 点击"详情"
  │    └─> showDetail(release)
  │         └─> 弹窗显示发布信息
  │         └─> 加载并显示事件日志
  │
  └─> 点击"回滚"
       └─> releaseStore.rollback(releaseId)
            │
            ├─> API: POST /api/v1/releases/{id}/rollback
            │    └─> backend async goroutine
            │
            └─> startPolling() 追踪回滚进度
                 └─> 同发布流程
```

## 核心 Store 设计

### appStore (应用元数据)

负责管理全局的应用元数据，减少重复请求：

```typescript
defineStore('app', () => {
  // 只读数据
  const applications = ref([])
  const environments = ref([])
  const clusters = ref([])
  const workloadTargets = ref([])
  
  // Computed getters (缓存计算)
  const getApplicationById = computed(...)  // O(1) 查询
  const getEnvironmentById = computed(...)
  const getAvailableClusters = computed(...)  // 动态计算集群列表
  
  // Actions
  const fetchApplications = async () => {}  // 从 API 加载
  const initializeMetadata = async () => {}  // 初始化所有
})
```

**特点**:
- 数据只加载一次（应用启动时）
- 提供快速的 getter 方法
- computed 缓存，避免重复计算
- 支持动态过滤（如根据 app+env 获取集群）

### releaseStore (发布状态)

管理发布流程和历史：

```typescript
defineStore('release', () => {
  // 当前发布
  const currentRelease = ref(null)
  const releaseEvents = ref([])
  
  // 发布历史
  const releaseHistory = ref([])
  
  // 加载和轮询状态
  const isCreatingRelease = ref(false)
  const isPolling = ref(false)
  const pollingInterval = ref(null)
  
  // Computed getters
  const progressPercentage = computed(...)  // 根据事件计算进度
  const currentReleaseFormatted = computed(...)  // 关联名称
  
  // Actions
  const createRelease = async () => {}  // 发起发布
  const startPolling = async () => {}  // 启动轮询
  const stopPolling = () => {}  // 停止轮询
  const rollback = async () => {}  // 回滚
})
```

**特点**:
- 支持异步操作（async/await）
- 轮询管理（防止内存泄漏）
- 自动清理（cleanup on unmount）
- 带错误处理

### uiStore (UI 状态)

全局 UI 通知和状态：

```typescript
defineStore('ui', () => {
  const messages = ref([])
  
  const showMessage = (content, type, duration) => {}
  const success = (content) => {}  // 快捷方法
  const error = (content) => {}
  const warning = (content) => {}
})
```

**特点**:
- 统一的消息通知系统
- 自动过期（timeout）
- 快捷方法减少重复代码

## 组件设计模式

### 页面组件结构

```vue
<template>
  <!-- 页面内容 -->
</template>

<script setup lang="ts">
// 1. 导入 stores
const store1 = useStore1()
const store2 = useStore2()

// 2. 定义本地 state
const localState = ref()

// 3. 定义 computed
const derivedValue = computed(...)

// 4. 定义 methods
const handleAction = async () => {
  try {
    // store 操作
  } catch (error) {
    uiStore.error()
  }
}

// 5. 生命周期
onMounted(async () => {
  // 初始化加载
})

onUnmounted(() => {
  // 清理（如停止轮询）
})
</script>
```

### 响应式流程

```
用户交互 (click, input)
  │
  └─> 触发 method
       │
       ├─> 验证输入
       ├─> 更新 local state 或 store
       ├─> 发送 API 请求
       └─> 显示通知
            │
            └─> 组件自动重新渲染 (reactive)
```

## API 拦截器

### 请求拦截

```typescript
request.interceptors.request.use((config) => {
  // 1. 生成 request ID
  config.headers['X-Request-ID'] = generateUUID()
  
  // 2. 添加认证信息 (TODO)
  // config.headers['Authorization'] = `Bearer ${token}`
  
  // 3. 调试日志
  console.log(`[request-id] method url`)
  
  return config
})
```

### 响应拦截

```typescript
request.interceptors.response.use(
  (response) => {
    // 成功响应 - 返回 data
    return response.data
  },
  (error) => {
    // 错误处理
    if (error.response?.data?.error) {
      // 业务错误
      return reject(new Error(message))
    } else if (error.code === 'ECONNABORTED') {
      // 超时错误
      return reject(new Error('请求超时'))
    } else {
      // 网络错误
      return reject(error)
    }
  }
)
```

## 轮询机制

### 启动轮询

```typescript
const startPolling = async (releaseId, interval = 2000) => {
  if (isPolling.value) return  // 防止重复
  
  isPolling.value = true
  
  // 立即执行一次
  await poll()
  
  // 设置间隔轮询
  pollingInterval.value = setInterval(poll, interval)
}

const poll = async () => {
  try {
    const release = await fetchReleaseStatus(releaseId)
    
    // 状态完成时停止轮询
    if (['success', 'failed'].includes(release.status)) {
      stopPolling()
    }
  } catch (err) {
    // 轮询出错继续尝试
    console.error(err)
  }
}
```

### 停止轮询

```typescript
const stopPolling = () => {
  if (pollingInterval.value) {
    clearInterval(pollingInterval.value)
    pollingInterval.value = null
  }
  isPolling.value = false
}

// 组件卸载时必须停止
onUnmounted(() => {
  stopPolling()
})
```

**关键点**:
- 防止内存泄漏（清理 interval）
- 防止重复轮询（check flag）
- 错误恢复（catch 不中断）
- 优雅停止（检查状态自动停止）

## 进度计算

```typescript
const progressPercentage = computed(() => {
  if (!releaseEvents.value || events.length === 0) return 0
  
  const eventTypes = releaseEvents.value.map(e => e.type)
  
  // 根据事件类型计算进度
  if (eventTypes.includes('failed')) return 0
  if (eventTypes.includes('success')) return 100
  if (eventTypes.includes('deploying')) return 75
  if (eventTypes.includes('validating')) return 50
  if (eventTypes.includes('started')) return 25
  
  return 10  // 未开始
})
```

## 类型安全

### API 类型

```typescript
// api.ts - 从后端 API 契约定义
interface ReleaseResponse {
  id: number
  app_id: number
  status: 'pending' | 'success' | 'failed'
  // ...
}
```

### 业务模型类型

```typescript
// models.ts - 前端使用的扩展类型
interface ReleaseRecord extends ReleaseResponse {
  app_name: string  // 关联的应用名称
  env_name: string
  cluster_name: string
}
```

**优势**:
- 编译时类型检查
- IDE 自动补全
- 错误早发现

## 路由设计

### 路由表

```typescript
const routes = [
  { path: '/', redirect: '/release' },
  { path: '/release', component: ReleaseFlow },
  { path: '/history', component: ReleaseHistory },
  { path: '/detail/:id', component: ReleaseDetail }
]
```

### 路由守卫

```typescript
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = `${to.meta.title} - 发布控制系统`
  next()
})
```

## 性能优化

### 1. 代码分割

```typescript
// Vite 自动按路由分割代码
const ReleaseFlow = () => import('@/views/ReleaseFlow.vue')
const ReleaseHistory = () => import('@/views/ReleaseHistory.vue')
```

### 2. 计算缓存

```typescript
// computed 自动缓存计算结果
const derivedData = computed(() => {
  return expensiveComputation()
})
```

### 3. 轮询优化

- 2秒轮询间隔（不过频）
- 完成时立即停止
- 错误不中断（继续尝试）

### 4. 列表虚拟化

对于大列表，可使用虚拟滚动：

```typescript
// 只渲染可见项
<virtual-list :items="largeList" :item-height="50" />
```

## 扩展性设计

### 添加新页面

1. 创建 `src/views/NewPage.vue`
2. 在 `router/index.ts` 注册
3. 在 `App.vue` 添加导航

### 添加新的 API 模块

1. 在 `src/api/` 创建模块文件
2. 定义类型
3. 在 store 中召用

### 添加新的 Store

1. 在 `src/stores/` 创建 store
2. 在需要的地方 `useXxxStore()`

## 常见问题

**Q: 为什么轮询设置为 2 秒?**
A: 平衡实时性和性能，过短的间隔会增加服务器压力和网络开销。

**Q: 如何处理 API 错误?**
A: 在 request 拦截器中统一处理，页面层通过 try-catch 捕获。

**Q: 发布历史如何保持同步?**
A: releaseStore 中的 releaseHistory 是内存数据，刷新页面时重新加载。

**Q: 如何调试轮询?**
A: 在浏览器开发者工具的 Network 标签中看 API 请求，或在 console 打印日志。
