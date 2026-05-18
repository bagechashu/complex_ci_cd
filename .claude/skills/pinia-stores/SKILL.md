---
name: pinia-stores
description: 发布控制系统 - Pinia状态管理架构
keywords: Pinia, 状态管理, Store, Composition API, 类型安全, 异步操作
---

# Pinia 状态管理指南

## 概览

使用Pinia (v2.1+) 管理应用全局状态，采用Composition API风格，提供类型安全和出色的开发体验。

## 四个核心Store

### 0. shellStore - Shell 命令执行（临时执行 + 历史查询）

**职责**: 管理 Shell 命令的发布、执行、历史查询

```typescript
// src/stores/shellStore.ts

export const useShellStore = defineStore('shell', () => {
  // 状态
  const shellServers = ref<ShellServer[]>([])
  const shellCommands = ref<ShellCommand[]>([])  // 已发布的命令
  const shellTaskExecutions = ref<ShellTaskExecution[]>([])
  
  // ============ 命令执行相关 ============
  
  /**
   * executeShellCommand - 直接执行已发布命令
   * 用途：ShellCommandExecution.vue 中用户选择命令后执行
   * 工作流：选择命令 → 点击执行 → 获得 execution_id → 显示加载状态
   * 返回：新的 ShellTaskExecution 记录（status=pending）
   */
  const executeShellCommand = async (commandId: number, serverId: number) => {
    const result = await api.post('/shell-commands/execute', {
      command_id: commandId,
      server_id: serverId
    })
    return result.data  // { id, status: 'pending', started_at }
  }
  
  /**
   * getCommandExecutions - 获取某条命令的执行历史
   * 用途：ShellCommandExecution.vue 中右侧面板显示该命令的最近执行
   * 限制：只返回最近 5 条，用于快速查看
   */
  const getCommandExecutions = async (commandId: number, limit = 5) => {
    const result = await api.get('/shell-tasks/executions', {
      command_id: commandId,
      limit
    })
    return result.data
  }
  
  // ============ 全局执行历史相关 ============
  
  /**
   * listAllExecutions - 查询全局执行历史（分页）
   * 用途：ExecutionHistory.vue 中显示所有执行记录表格
   * 支持过滤：按命令、状态、服务器过滤
   * 返回：分页列表 { data, pagination }
   */
  const listAllExecutions = async (filters: {
    limit?: number
    offset?: number
    command_id?: number
    status?: string
    server_id?: number
  }) => {
    const result = await api.get('/shell-tasks/executions', filters)
    return result.data
  }
  
  /**
   * getExecutionDetail - 获取单条执行的详细信息
   * 用途：ExecutionHistory.vue 中模态框显示执行详情（输出、错误、耗时等）
   */
  const getExecutionDetail = async (executionId: number) => {
    const result = await api.get(`/shell-tasks/${executionId}`)
    return result.data
  }
  
  // ============ 发布命令相关 ============
  
  /**
   * fetchPublishedCommands - 获取已发布的 Shell 命令列表
   * 用途：ShellCommandExecution.vue 初始化时加载
   * 返回：按服务器分组的已发布命令
   */
  const fetchPublishedCommands = async () => {
    const result = await api.get('/shell-commands/published')
    shellCommands.value = result.data
  }
  
  /**
   * fetchShellServers - 获取 Shell 服务器列表
   * 用途：ShellCommandExecution.vue 显示命令执行的目标服务器
   */
  const fetchShellServers = async () => {
    const result = await api.get('/shell-servers')
    shellServers.value = result.data
  }
  
  return {
    // 状态
    shellServers,
    shellCommands,
    shellTaskExecutions,
    
    // 命令执行方法
    executeShellCommand,
    getCommandExecutions,
    
    // 全局历史方法
    listAllExecutions,
    getExecutionDetail,
    
    // 数据加载方法
    fetchPublishedCommands,
    fetchShellServers
  }
})
```

**使用场景**:

```typescript
// ===== ShellCommandExecution.vue 中的使用 =====
const shellStore = useShellStore()

// 1. 初始化：加载已发布命令列表
onMounted(async () => {
  await shellStore.fetchPublishedCommands()
  await shellStore.fetchShellServers()
})

// 2. 用户选择命令后：加载该命令的执行历史
const selectCommand = async (cmd: ShellCommand) => {
  const executions = await shellStore.getCommandExecutions(cmd.id, 5)
  // 显示最近 5 次执行
}

// 3. 用户点击执行按钮
const executeCommand = async () => {
  const execution = await shellStore.executeShellCommand(cmd.id, server.id)
  // execution.id 用于轮询获取执行结果
}

// ===== ExecutionHistory.vue 中的使用 =====
const shellStore = useShellStore()

// 1. 加载执行历史（支持分页和过滤）
const loadExecutions = async () => {
  const result = await shellStore.listAllExecutions({
    limit: 20,
    offset: (currentPage.value - 1) * 20,
    command_id: selectedCommandFilter.value,
    status: selectedStatusFilter.value,
    server_id: selectedServerFilter.value
  })
  executions.value = result.data
  totalCount.value = result.pagination.total
}

// 2. 用户点击查看详情
const showDetail = async (executionId: number) => {
  const detail = await shellStore.getExecutionDetail(executionId)
  // 在模态框显示完整的输出、错误、耗时等信息
}
```

**职责分工总结**:

| 方法 | 页面 | 用途 | 频率 |
|------|------|------|------|
| `executeShellCommand()` | ShellCommandExecution | 直接执行命令 | 用户点击时 |
| `getCommandExecutions()` | ShellCommandExecution | 查看该命令的历史 | 选择命令时 |
| `listAllExecutions()` | ExecutionHistory | 查看全局执行记录 | 分页查询 |
| `getExecutionDetail()` | ExecutionHistory | 查看执行详情 | 点击详情时 |

---

## 三个核心Store

### 1. appStore - 应用元数据（只读、缓存）

**职责**: 管理应用、环境、集群的元数据，一次性加载缓存

**职责**: 管理应用、环境、集群的元数据，一次性加载缓存

```typescript
// src/stores/appStore.ts

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiService } from '@/api'

export const useAppStore = defineStore('app', () => {
  // 状态
  const applications = ref<Application[]>([])
  const environments = ref<Environment[]>([])
  const clusters = ref<Cluster[]>([])
  const workloadTargets = ref<WorkloadTarget[]>([])
  const allMappingsByApp = ref<Map<number, WorkloadTarget[]>>(new Map())
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 计算属性
  const applicationMap = computed(() => {
    const map = new Map<number, Application>()
    applications.value.forEach(app => {
      map.set(app.id, app)
    })
    return map
  })

  const environmentMap = computed(() => {
    const map = new Map<number, Environment>()
    environments.value.forEach(env => {
      map.set(env.id, env)
    })
    return map
  })

  const clusterMap = computed(() => {
    const map = new Map<number, Cluster>()
    clusters.value.forEach(cluster => {
      map.set(cluster.id, cluster)
    })
    return map
  })

  // 方法 - 初始化
  const initializeMetadata = async () => {
    if (applications.value.length > 0) {
      return // 已加载，直接返回
    }

    loading.value = true
    error.value = null

    try {
      // 并行加载所有元数据
      const [apps, envs, clusters, mappings] = await Promise.all([
        apiService.applications.list(),
        apiService.metadata.environments(),
        apiService.clusters.list(),
        apiService.metadata.workloadTargets(),
      ])

      applications.value = apps
      environments.value = envs.sort((a, b) => a.priority - b.priority)
      clusters.value = clusters
      workloadTargets.value = mappings

      // 构建按应用索引的映射缓存
      const mapByApp = new Map<number, WorkloadTarget[]>()
      mappings.forEach(mapping => {
        if (!mapByApp.has(mapping.app_id)) {
          mapByApp.set(mapping.app_id, [])
        }
        mapByApp.get(mapping.app_id)!.push(mapping)
      })
      allMappingsByApp.value = mapByApp
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load metadata'
    } finally {
      loading.value = false
    }
  }

  // 方法 - 查询
  const getApplication = (id: number) => applicationMap.value.get(id)
  
  const getEnvironment = (id: number) => environmentMap.value.get(id)
  
  const getCluster = (id: number) => clusterMap.value.get(id)

  const getAvailableClustersForApp = (appId: number): Cluster[] => {
    const mappings = allMappingsByApp.value.get(appId) || []
    const clusterIds = new Set(mappings.map(m => m.cluster_id))
    return clusters.value.filter(c => clusterIds.has(c.id))
  }

  const getWorkloadTargets = (appId: number, envId: number, clusterId: number) => {
    return workloadTargets.value.find(
      t => t.app_id === appId && t.env_id === envId && t.cluster_id === clusterId
    )
  }

  return {
    // 状态
    applications,
    environments,
    clusters,
    workloadTargets,
    loading,
    error,

    // 计算属性
    applicationMap,
    environmentMap,
    clusterMap,

    // 方法
    initializeMetadata,
    getApplication,
    getEnvironment,
    getCluster,
    getAvailableClustersForApp,
    getWorkloadTargets,
  }
})
```

### 2. releaseStore - 发布流程和历史

**职责**: 管理当前发布、历史记录、轮询状态

```typescript
// src/stores/releaseStore.ts

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiService } from '@/api'

export const useReleaseStore = defineStore('release', () => {
  // 状态
  const releases = ref<ReleaseRecord[]>([])
  const currentRelease = ref<ReleaseRecord | null>(null)
  const releaseEvents = ref<ReleaseEvent[]>([])
  const isPolling = ref(false)
  const pollInterval = ref<number | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 计算属性
  const currentReleaseProgress = computed(() => {
    if (!currentRelease.value) return 0

    const status = currentRelease.value.status
    const statusMap: Record<string, number> = {
      pending: 10,
      validating: 30,
      deploying: 60,
      success: 100,
      failed: 100,
      rolled_back: 100,
    }
    return statusMap[status] || 0
  })

  const isCurrentReleaseComplete = computed(() => {
    if (!currentRelease.value) return false
    return ['success', 'failed', 'rolled_back'].includes(currentRelease.value.status)
  })

  // 方法 - 创建发布
  const createRelease = async (appId: number, envId: number, clusterId: number, image: string) => {
    loading.value = true
    error.value = null

    try {
      const response = await apiService.releases.create({
        app_id: appId,
        env_id: envId,
        cluster_id: clusterId,
        image,
      })

      currentRelease.value = response
      releaseEvents.value = []
      loading.value = false

      // 自动启动轮询
      startPolling(response.id)

      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create release'
      loading.value = false
      throw err
    }
  }

  // 方法 - 轮询获取发布状态
  const startPolling = (releaseId: number) => {
    if (isPolling.value) {
      return // 已在轮询中
    }

    isPolling.value = true
    const pollFunc = async () => {
      try {
        const release = await apiService.releases.getStatus(releaseId)
        currentRelease.value = release

        // 更新事件列表
        const events = await apiService.releases.getEvents(releaseId)
        releaseEvents.value = events

        // 发布完成，停止轮询
        if (isCurrentReleaseComplete.value) {
          stopPolling()
        }
      } catch (err) {
        console.error('Poll error:', err)
        // 轮询错误继续，不停止
      }
    }

    // 第一次立即执行
    pollFunc()

    // 之后每2秒轮询一次
    pollInterval.value = window.setInterval(pollFunc, 2000)
  }

  // 方法 - 停止轮询
  const stopPolling = () => {
    if (pollInterval.value !== null) {
      clearInterval(pollInterval.value)
      pollInterval.value = null
    }
    isPolling.value = false
  }

  // 方法 - 回滚
  const rollback = async (releaseId: number, reason: string) => {
    loading.value = true
    error.value = null

    try {
      const response = await apiService.releases.rollback(releaseId, { reason })
      
      // 更新当前发布为回滚发布
      currentRelease.value = response
      releaseEvents.value = []

      // 重新启动轮询
      startPolling(response.id)

      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to rollback'
      throw err
    } finally {
      loading.value = false
    }
  }

  // 方法 - 获取发布历史
  const fetchReleaseHistory = async (limit = 20, offset = 0) => {
    loading.value = true
    error.value = null

    try {
      const response = await apiService.releases.list({ limit, offset })
      releases.value = response.data
      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch history'
      throw err
    } finally {
      loading.value = false
    }
  }

  // 方法 - 清空当前发布
  const clearCurrentRelease = () => {
    stopPolling()
    currentRelease.value = null
    releaseEvents.value = []
  }

  return {
    // 状态
    releases,
    currentRelease,
    releaseEvents,
    isPolling,
    loading,
    error,

    // 计算属性
    currentReleaseProgress,
    isCurrentReleaseComplete,

    // 方法
    createRelease,
    startPolling,
    stopPolling,
    rollback,
    fetchReleaseHistory,
    clearCurrentRelease,
  }
})
```

### 3. uiStore - 全局UI状态

**职责**: 管理侧边栏展开/收起、主题、通知等UI状态

```typescript
// src/stores/uiStore.ts

import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  message: string
  duration?: number
}

export const useUIStore = defineStore('ui', () => {
  // 状态
  const sidebarCollapsed = ref(false)
  const currentTheme = ref<'light' | 'dark'>('light')
  const notifications = ref<Notification[]>([])

  // 方法 - 切换侧边栏
  const toggleSidebar = () => {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  // 方法 - 切换主题
  const toggleTheme = () => {
    currentTheme.value = currentTheme.value === 'light' ? 'dark' : 'light'
    localStorage.setItem('theme', currentTheme.value)
  }

  // 方法 - 初始化主题
  const initializeTheme = () => {
    const saved = localStorage.getItem('theme') as 'light' | 'dark' | null
    if (saved) {
      currentTheme.value = saved
    } else {
      // 检测系统主题
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      currentTheme.value = prefersDark ? 'dark' : 'light'
    }
  }

  // 方法 - 添加通知
  const addNotification = (notification: Notification) => {
    const id = Math.random().toString(36).substr(2, 9)
    const notif = { ...notification, id }
    notifications.value.push(notif)

    // 自动移除
    if (notification.duration !== 0) {
      setTimeout(() => {
        removeNotification(id)
      }, notification.duration || 3000)
    }

    return id
  }

  // 方法 - 移除通知
  const removeNotification = (id: string) => {
    notifications.value = notifications.value.filter(n => n.id !== id)
  }

  // 方法 - 便利方法
  const success = (message: string) => {
    addNotification({
      type: 'success',
      message,
      duration: 3000,
    } as Notification)
  }

  const error = (message: string) => {
    addNotification({
      type: 'error',
      message,
      duration: 5000,
    } as Notification)
  }

  const warning = (message: string) => {
    addNotification({
      type: 'warning',
      message,
      duration: 4000,
    } as Notification)
  }

  const info = (message: string) => {
    addNotification({
      type: 'info',
      message,
      duration: 3000,
    } as Notification)
  }

  return {
    // 状态
    sidebarCollapsed,
    currentTheme,
    notifications,

    // 方法
    toggleSidebar,
    toggleTheme,
    initializeTheme,
    addNotification,
    removeNotification,
    success,
    error,
    warning,
    info,
  }
})
```

## 在组件中使用

### 基础使用

```vue
<template>
  <div>
    <h1>{{ app.name }}</h1>
    <p v-if="appStore.loading">加载中...</p>
    <n-select 
      v-model:value="selectedApp"
      :options="applicationOptions"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/appStore'
import { useUIStore } from '@/stores/uiStore'

const appStore = useAppStore()
const uiStore = useUIStore()

const selectedApp = computed({
  get: () => appStore.applications[0]?.id || null,
  set: (id) => {
    // 更新逻辑
  },
})

const applicationOptions = computed(() => {
  return appStore.applications.map(app => ({
    label: app.name,
    value: app.id,
  }))
})

onMounted(() => {
  appStore.initializeMetadata()
})
</script>
```

### 发布流程

```vue
<script setup lang="ts">
const releaseStore = useReleaseStore()
const appStore = useAppStore()
const uiStore = useUIStore()

const handleSubmitRelease = async () => {
  try {
    await releaseStore.createRelease(
      form.appId,
      form.envId,
      form.clusterId,
      form.image
    )
    uiStore.success('发布已启动，请稍候...')
  } catch (error) {
    uiStore.error('发布失败: ' + error.message)
  }
}

const handleRollback = async () => {
  if (!releaseStore.currentRelease) return
  
  try {
    await releaseStore.rollback(
      releaseStore.currentRelease.id,
      '用户回滚'
    )
    uiStore.success('回滚已启动')
  } catch (error) {
    uiStore.error('回滚失败: ' + error.message)
  }
}
</script>
```

## 最佳实践

### 1. 初始化时机
- 在App.vue的onMounted中调用appStore.initializeMetadata()
- 只加载一次，后续读缓存

### 2. 数据持久化
- UI状态(主题)保存到localStorage
- 避免将服务器数据缓存到localStorage

### 3. 类型安全
- 定义完整的TypeScript接口
- 使用computed类型推断

### 4. 性能优化
- 使用Map索引加速查询
- 避免过度响应化大数据

### 5. 错误处理
- 每个async方法设置error状态
- 在组件中监听error并显示通知

### 6. 内存管理
- stopPolling()及时清理定时器
- clearCurrentRelease()释放大数据

---

## Store交互流程

### 架构概览

```
┌─────────────────────────────────────────────────────┐
│              Vue组件层                              │
│  (ReleaseFlow.vue, ReleaseHistory.vue, etc)        │
└──────────────┬──────────────────────────────────────┘
               │ 读写状态
               ▼
┌─────────────────────────────────────────────────────┐
│           Pinia Store层                             │
│  ┌──────────────┬──────────────┬──────────────┐    │
│  │  appStore    │ releaseStore │   uiStore    │    │
│  │  (元数据)    │  (发布流程)  │  (UI状态)    │    │
│  └──────────────┴──────────────┴──────────────┘    │
└──────────────┬──────────────────────────────────────┘
               │ 调用方法
               ▼
┌─────────────────────────────────────────────────────┐
│           API层                                      │
│  (apiService/releases, apiService/metadata)        │
└──────────────┬──────────────────────────────────────┘
               │ HTTP请求
               ▼
        ┌──────────────┐
        │  后端服务    │
        │  (Go-chi)    │
        └──────────────┘
```

### 数据流向

#### 初始化流程
```
App.vue (onMounted)
  ↓
appStore.initializeMetadata() [并行获取]
  ├─ apiService.applications.list()
  ├─ apiService.metadata.environments()
  ├─ apiService.clusters.list()
  └─ apiService.metadata.workloadTargets()
  ↓
应用、环境、集群、映射缓存到 appStore
  ↓
组件读appStore中的缓存数据
```

#### 发布流程
```
用户操作 (ReleaseFlow.vue)
  ↓ 用户输入 (appId, envId, clusterId, image)
  ↓
handleSubmitRelease()
  ├─ releaseStore.createRelease(...) [发送HTTP]
  │  ↓
  │  apiService.releases.create()  [后端创建ReleaseRecord]
  │  ↓
  │  返回 release_record 对象
  │  ↓
  │  releaseStore.currentRelease = response
  │  releaseStore.startPolling(release_id)
  │
  └─ uiStore.success('发布已启动')  [显示通知]
  ↓
轮询循环 (每2秒)
  ├─ apiService.releases.getStatus(releaseId)  [获取状态]
  ├─ apiService.releases.getEvents(releaseId)  [获取事件日志]
  ├─ releaseStore.currentRelease = updated
  ├─ releaseStore.releaseEvents = events
  └─ 如果status=success/failed → stopPolling()
  ↓
组件响应式更新 (currentRelease, releaseEvents变化)
  ├─ 进度条更新
  ├─ 事件列表滚动加载
  └─ 最后一个事件动画显示
  ↓
发布完成 → 显示成功/失败通知
```

#### 回滚流程
```
用户点击"回滚"按钮 (ReleaseDetail.vue)
  ↓
handleRollback()
  ├─ releaseStore.rollback(releaseId, reason)
  │  ↓
  │  apiService.releases.rollback(releaseId)  [后端执行回滚]
  │  ↓
  │  返回新的 rollback_release_record
  │  ↓
  │  releaseStore.currentRelease = new_release
  │  releaseStore.releaseEvents = []  [重置事件]
  │  releaseStore.startPolling(new_release.id)  [重新轮询]
  │
  └─ uiStore.success('回滚已启动')
  ↓
页面重新进入轮询循环 (同发布流程)
```

### 关键交互场景

#### 场景1: 发布页面首次加载
```typescript
// ReleaseFlow.vue
const appStore = useAppStore()
const releaseStore = useReleaseStore()

onMounted(async () => {
  // 步骤1: 确保元数据已加载
  await appStore.initializeMetadata()
  
  // 步骤2: 读appStore中的缓存
  const apps = computed(() => appStore.applications)
  const envs = computed(() => appStore.environments)
  
  // 步骤3: 根据选中的app，得到可用的集群
  const availableClusters = computed(() => {
    if (!form.appId) return []
    return appStore.getAvailableClustersForApp(form.appId)
  })
  
  // 步骤4: 验证完整性
  const canSubmit = computed(() => {
    return form.appId && form.envId && form.clusterId && form.image
  })
})

const handleSubmit = async () => {
  // 步骤5: 提交到releaseStore
  try {
    await releaseStore.createRelease(
      form.appId,
      form.envId,
      form.clusterId,
      form.image
    )
    // 步骤6: 导航到详情页查看进度
    router.push(`/release/${releaseStore.currentRelease?.id}`)
  } catch (err) {
    uiStore.error(err.message)
  }
}
```

#### 场景2: 发布详情页实时更新
```typescript
// ReleaseDetail.vue
const route = useRoute()
const releaseStore = useReleaseStore()
const uiStore = useUIStore()
const releaseId = parseInt(route.params.id as string)

onMounted(() => {
  // 步骤1: 如果需要，手动启动轮询
  // (通常从发布页导航过来，轮询已启动)
  if (!releaseStore.isPolling) {
    releaseStore.startPolling(releaseId)
  }
})

// 步骤2: 响应式模板自动更新
const progressPercent = computed(() => releaseStore.currentReleaseProgress)
const events = computed(() => releaseStore.releaseEvents)
const isComplete = computed(() => releaseStore.isCurrentReleaseComplete)

onBeforeUnmount(() => {
  // 步骤3: 导航离开时，可选择停止轮询
  // (为避免浪费带宽，通常保留轮询，用户可主动停止)
})

// 步骤4: 用户操作
const handleRollback = async () => {
  try {
    await releaseStore.rollback(releaseId, '用户手动回滚')
    uiStore.success('回滚已启动')
  } catch (err) {
    uiStore.error(err.message)
  }
}
```

#### 场景3: 发布历史列表
```typescript
// ReleaseHistory.vue
const releaseStore = useReleaseStore()

onMounted(() => {
  // 步骤1: 加载历史记录
  releaseStore.fetchReleaseHistory(20, 0)
})

// 步骤2: 响应式列表
const releases = computed(() => releaseStore.releases)

// 步骤3: 分页加载
const handleLoadMore = async () => {
  const offset = releaseStore.releases.length
  await releaseStore.fetchReleaseHistory(20, offset)
}

// 步骤4: 点击进入详情
const handleViewDetail = (releaseId: number) => {
  router.push(`/release/${releaseId}`)
}
```

### Store间通信模式

#### 模式1: 组件协调 (推荐)
```typescript
// 组件作为协调器
const appStore = useAppStore()
const releaseStore = useReleaseStore()
const uiStore = useUIStore()

const handleAction = async () => {
  // 组件决定调用哪些store方法
  await appStore.initializeMetadata()
  const cluster = appStore.getCluster(clusterId)
  
  if (!cluster) {
    uiStore.error('集群不存在')
    return
  }
  
  await releaseStore.createRelease(...)
}
```

#### 模式2: Store内调用 (谨慎使用)
```typescript
// releaseStore中可调用appStore，但避免循环依赖
export const useReleaseStore = defineStore('release', () => {
  const appStore = useAppStore()  // 注入依赖
  
  const createRelease = async (...) => {
    // 可读appStore，但尽量少修改
    const app = appStore.getApplication(appId)
    // ...
  }
})
```

#### 模式3: 事件通知 (不推荐)
```typescript
// ❌ 避免在store中发射事件
// ✅ 改用：
// 1. 组件监听computed变化
// 2. 或显式调用方法
```

### 常见错误

| 错误 | 原因 | 解决 |
|------|------|------|
| 轮询泄漏 | 未调用stopPolling() | 在onBeforeUnmount清理 |
| 状态重复加载 | initializeMetadata()多次调用 | 检查是否已加载 |
| 循环依赖 | Store相互引用 | 用组件作协调层 |
| 内存溢出 | 长列表不分页 | 使用虚拟滚动+分页 |
| Race condition | 并发请求覆盖 | 使用abort token或版本号 |

