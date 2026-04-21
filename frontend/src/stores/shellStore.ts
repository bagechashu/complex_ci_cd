/**
 * shellStore.ts - Shell 任务状态管理 (DDD Aggregate)
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ShellTask, ShellServer, ShellCommand } from '@/types/api'
import {
  listShellTasks,
  getShellTask,
  createShellTask,
  updateShellTask,
  deleteShellTask,
  listShellServers,
  listShellCommands
} from '@/api/shell'

export const useShellStore = defineStore('shell', () => {
  // ============ State ============
  const shellTasks = ref<ShellTask[]>([])
  const shellServers = ref<ShellServer[]>([])
  const shellCommands = ref<ShellCommand[]>([])
  
  const currentShellTask = ref<ShellTask | null>(null)
  const selectedTaskIds = ref<number[]>([])

  // Pagination state
  const pagination = ref({
    page: 1,
    pageSize: 10,
    total: 0,
    totalPages: 0
  })

  // Loading states
  const tasksLoading = ref(false)
  const serversLoading = ref(false)
  const commandsLoading = ref(false)
  const createLoading = ref(false)
  const updateLoading = ref(false)
  const deleteLoading = ref(false)

  // Error state
  const error = ref<string | null>(null)

  // ============ Computed ============
  
  /**
   * 获取选定的任务
   */
  const selectedTasks = computed(() => {
    return shellTasks.value.filter(task => selectedTaskIds.value.includes(task.id))
  })

  /**
   * 按执行方式分类
   */
  const tasksByExecutionMethod = computed(() => {
    return {
      serial: shellTasks.value.filter(t => t.execution_method === 'serial'),
      parallel: shellTasks.value.filter(t => t.execution_method === 'parallel')
    }
  })

  /**
   * 需要审批的任务
   */
  const tasksRequiringApproval = computed(() => {
    return shellTasks.value.filter(t => t.requires_approval)
  })

  /**
   * 获取命令名称
   */
  const getCommandName = computed(() => (commandId: number) => {
    return shellCommands.value.find(c => c.id === commandId)?.command || `Command ${commandId}`
  })

  /**
   * 获取服务器名称
   */
  const getServerName = computed(() => (serverId: number) => {
    return shellServers.value.find(s => s.id === serverId)?.name || `Server ${serverId}`
  })

  // ============ Actions ============

  /**
   * 获取 Shell 任务列表
   */
  const fetchShellTasks = async (page: number = 1, pageSize: number = 10) => {
    tasksLoading.value = true
    error.value = null
    try {
      const response = await listShellTasks(page, pageSize)
      shellTasks.value = response.data
      pagination.value = {
        page: response.page,
        pageSize: response.pageSize,
        total: response.total,
        totalPages: Math.ceil(response.total / pageSize)
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取 Shell 任务列表失败'
      console.error('Error fetching shell tasks:', err)
    } finally {
      tasksLoading.value = false
    }
  }

  /**
   * 获取单个 Shell 任务
   */
  const fetchShellTask = async (id: number) => {
    try {
      const response = await getShellTask(id)
      currentShellTask.value = response
      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : `获取 Shell 任务 ${id} 失败`
      console.error(`Error fetching shell task ${id}:`, err)
      return null
    }
  }

  /**
   * 创建 Shell 任务
   */
  const createShellTaskAction = async (data: Omit<ShellTask, 'id' | 'created_at' | 'updated_at'>) => {
    createLoading.value = true
    error.value = null
    try {
      const response = await createShellTask(data)
      shellTasks.value.unshift(response)
      pagination.value.total += 1
      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : '创建 Shell 任务失败'
      console.error('Error creating shell task:', err)
      return null
    } finally {
      createLoading.value = false
    }
  }

  /**
   * 更新 Shell 任务
   */
  const updateShellTaskAction = async (id: number, data: Partial<ShellTask>) => {
    updateLoading.value = true
    error.value = null
    try {
      const response = await updateShellTask(id, data)
      const index = shellTasks.value.findIndex(t => t.id === id)
      if (index !== -1) {
        shellTasks.value[index] = response
      }
      if (currentShellTask.value?.id === id) {
        currentShellTask.value = response
      }
      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : `更新 Shell 任务 ${id} 失败`
      console.error(`Error updating shell task ${id}:`, err)
      return null
    } finally {
      updateLoading.value = false
    }
  }

  /**
   * 删除 Shell 任务
   */
  const deleteShellTaskAction = async (id: number) => {
    deleteLoading.value = true
    error.value = null
    try {
      await deleteShellTask(id)
      shellTasks.value = shellTasks.value.filter(t => t.id !== id)
      selectedTaskIds.value = selectedTaskIds.value.filter(tid => tid !== id)
      pagination.value.total -= 1
      if (currentShellTask.value?.id === id) {
        currentShellTask.value = null
      }
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : `删除 Shell 任务 ${id} 失败`
      console.error(`Error deleting shell task ${id}:`, err)
      return false
    } finally {
      deleteLoading.value = false
    }
  }

  /**
   * 批量删除 Shell 任务
   */
  const deleteMultipleShellTasks = async (ids: number[]) => {
    let successCount = 0
    for (const id of ids) {
      const success = await deleteShellTaskAction(id)
      if (success) successCount++
    }
    return successCount
  }

  /**
   * 获取 Shell 服务器列表
   */
  const fetchShellServers = async () => {
    serversLoading.value = true
    try {
      const response = await listShellServers(1, 100)
      shellServers.value = response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取 Shell 服务器列表失败'
      console.error('Error fetching shell servers:', err)
    } finally {
      serversLoading.value = false
    }
  }

  /**
   * 获取 Shell 命令列表
   */
  const fetchShellCommands = async () => {
    commandsLoading.value = true
    try {
      const response = await listShellCommands(1, 100)
      shellCommands.value = response.data
    } catch (err) {
      error.value = err instanceof Error ? err.message : '获取 Shell 命令列表失败'
      console.error('Error fetching shell commands:', err)
    } finally {
      commandsLoading.value = false
    }
  }

  /**
   * 初始化数据
   */
  const initializeData = async () => {
    await Promise.all([
      fetchShellTasks(1, 10),
      fetchShellServers(),
      fetchShellCommands()
    ])
  }

  /**
   * 清空选择
   */
  const clearSelection = () => {
    selectedTaskIds.value = []
  }

  /**
   * 切换选择
   */
  const toggleTaskSelection = (id: number) => {
    const index = selectedTaskIds.value.indexOf(id)
    if (index > -1) {
      selectedTaskIds.value.splice(index, 1)
    } else {
      selectedTaskIds.value.push(id)
    }
  }

  /**
   * 全选
   */
  const selectAllTasks = () => {
    selectedTaskIds.value = shellTasks.value.map(t => t.id)
  }

  /**
   * 清除错误信息
   */
  const clearError = () => {
    error.value = null
  }

  // ============ Return ============
  return {
    // State
    shellTasks,
    shellServers,
    shellCommands,
    currentShellTask,
    selectedTaskIds,
    pagination,
    error,

    // Loading states
    tasksLoading,
    serversLoading,
    commandsLoading,
    createLoading,
    updateLoading,
    deleteLoading,

    // Computed
    selectedTasks,
    tasksByExecutionMethod,
    tasksRequiringApproval,
    getCommandName,
    getServerName,

    // Actions
    fetchShellTasks,
    fetchShellTask,
    createShellTaskAction,
    updateShellTaskAction,
    deleteShellTaskAction,
    deleteMultipleShellTasks,
    fetchShellServers,
    fetchShellCommands,
    initializeData,
    clearSelection,
    toggleTaskSelection,
    selectAllTasks,
    clearError
  }
})
