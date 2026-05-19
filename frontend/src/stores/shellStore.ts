/**
 * shellStore.ts - Shell 任务状态管理 (DDD Aggregate)
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ShellTask, ShellServer, ShellCommand, ShellTaskExecution } from '@/types/api'
import { getErrorMessage, isBusinessError } from '@/utils/error-handler'
import type { BusinessError } from '@/api/request'
import {
  listShellTasks,
  getShellTask,
  createShellTask,
  updateShellTask,
  deleteShellTask,
  listShellServers,
  listShellCommands,
  listShellTaskExecutions,
  executeShellCommand
} from '@/api/shell'

export const useShellStore = defineStore('shell', () => {
  // ============ State ============
  const shellTasks = ref<ShellTask[]>([])
  const shellServers = ref<ShellServer[]>([])
  const shellCommands = ref<ShellCommand[]>([])
  const shellTaskExecutions = ref<ShellTaskExecution[]>([])
  
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
  const errorCode = ref<number | null>(null)

  // ============ Computed ============
  
  /**
   * 获取选定的任务
   */
  const selectedTasks = computed(() => {
    return shellTasks.value.filter(task => selectedTaskIds.value.includes(task.id))
  })

  /**
   * 需要审批的任务
   */
  const tasksRequiringApproval = computed(() => {
    return shellTasks.value.filter(t => t.requires_approval)
  })

  /**
   * 按服务器分类的任务
   */
  const tasksByServer = computed(() => {
    const grouped: Record<number, ShellTask[]> = {}
    shellTasks.value.forEach(task => {
      if (!grouped[task.server_id]) {
        grouped[task.server_id] = []
      }
      grouped[task.server_id].push(task)
    })
    return grouped
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
    errorCode.value = null
    try {
      const response = await listShellTasks(page, pageSize)
      // 后端返回 {data: [...], page, pageSize, total, totalPages}
      shellTasks.value = Array.isArray(response.data) ? response.data : []
      pagination.value = {
        page: response.page || 1,
        pageSize: response.pageSize || 10,
        total: response.total || 0,
        totalPages: response.totalPages || 0
      }
    } catch (err) {
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
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
    errorCode.value = null
    try {
      const response = await createShellTask(data)
      shellTasks.value.unshift(response)
      pagination.value.total += 1
      return response
    } catch (err) {
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
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
    errorCode.value = null
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
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
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
    errorCode.value = null
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
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
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
  const fetchShellServers = async (page: number = 1, pageSize: number = 100) => {
    serversLoading.value = true
    error.value = null
    errorCode.value = null
    try {
      const response = await listShellServers(page, pageSize)
      // 后端返回 {data: [...], page, pageSize, total, totalPages}
      shellServers.value = Array.isArray(response.data) ? response.data : []
    } catch (err) {
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
      console.error('Error fetching shell servers:', err)
    } finally {
      serversLoading.value = false
    }
  }

  /**
   * 获取 Shell 命令列表
   */
  const fetchShellCommands = async (page: number = 1, pageSize: number = 100) => {
    commandsLoading.value = true
    error.value = null
    errorCode.value = null
    try {
      const response = await listShellCommands(page, pageSize)
      // 后端返回 {data: [...], page, pageSize, total, totalPages}
      shellCommands.value = Array.isArray(response.data) ? response.data : []
    } catch (err) {
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
      console.error('Error fetching shell commands:', err)
    } finally {
      commandsLoading.value = false
    }
  }

  /**
   * 获取 Shell 任务执行历史
   */
  const fetchShellTaskExecutions = async (
    page: number = 1,
    pageSize: number = 10,
    taskID?: number,
    commandID?: number
  ) => {
    try {
      const response = await listShellTaskExecutions(page, pageSize, taskID, commandID)
      // 后端返回 {data: [...], page, pageSize, total, totalPages}
      shellTaskExecutions.value = Array.isArray(response.data) ? response.data : []
      return shellTaskExecutions.value
    } catch (err) {
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
      console.error('Error fetching shell task executions:', err)
      return []
    }
  }

  /**
   * 执行 Shell 命令
   */
  const executeShellCommandAction = async (
    data: Omit<ShellTaskExecution, 'id' | 'created_at' | 'updated_at'>
  ) => {
    try {
      const response = await executeShellCommand(data)
      return response
    } catch (err) {
      error.value = err instanceof Error ? err.message : '执行命令失败'
      console.error('Error executing shell command:', err)
      return null
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
    shellTaskExecutions,
    currentShellTask,
    selectedTaskIds,
    pagination,
    error,
    errorCode,

    // Loading states
    tasksLoading,
    serversLoading,
    commandsLoading,
    createLoading,
    updateLoading,
    deleteLoading,

    // Computed
    selectedTasks,
    tasksByServer,
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
    fetchShellTaskExecutions,
    executeShellCommand: executeShellCommandAction,
    initializeData,
    clearSelection,
    toggleTaskSelection,
    selectAllTasks,
    clearError
  }
})
