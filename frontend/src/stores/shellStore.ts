import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ShellServer, ShellCommand, ShellCommandExecution } from '@/types/api'
import { getErrorMessage, isBusinessError } from '@/utils/error-handler'
import type { BusinessError } from '@/api/request'
import {
  listShellServers,
  listShellCommands,
  listShellCommandExecutions,
  executeShellCommand
} from '@/api/shell'

export const useShellStore = defineStore('shell', () => {
  // ============ State ============
  const shellServers = ref<ShellServer[]>([])
  const shellCommands = ref<ShellCommand[]>([])
  const shellCommandExecutions = ref<ShellCommandExecution[]>([])

  // Loading states
  const serversLoading = ref(false)
  const commandsLoading = ref(false)

  // Error state
  const error = ref<string | null>(null)
  const errorCode = ref<number | null>(null)

  // ============ Computed ============
  
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
  const fetchShellCommandExecutions = async (
    page: number = 1,
    pageSize: number = 10,
    commandID?: number
  ) => {
    try {
      const response = await listShellCommandExecutions(page, pageSize, commandID)
      // 后端返回 {data: [...], page, pageSize, total, totalPages}
      shellCommandExecutions.value = Array.isArray(response.data) ? response.data : []
      return shellCommandExecutions.value
    } catch (err) {
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
      console.error('Error fetching shell command executions:', err)
      return []
    }
  }

  /**
   * 执行 Shell 命令
   */
  const executeShellCommandAction = async (
    data: Omit<ShellCommandExecution, 'id' | 'created_at' | 'updated_at'>
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
      fetchShellServers(),
      fetchShellCommands()
    ])
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
    shellServers,
    shellCommands,
    shellCommandExecutions,
    error,
    errorCode,

    // Loading states
    serversLoading,
    commandsLoading,

    // Computed
    getCommandName,
    getServerName,

    // Actions
    fetchShellServers,
    fetchShellCommands,
    fetchShellCommandExecutions,
    executeShellCommand: executeShellCommandAction,
    initializeData,
    clearError
  }
})
