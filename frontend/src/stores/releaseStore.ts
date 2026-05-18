/**
 * releaseStore.ts - 发布流程和历史状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ReleaseResponse, ReleaseEvent } from '@/types/api'
import type { ReleaseRecord } from '@/types/models'
import { releaseAPI } from '@/api/release'
import { useAppStore } from './appStore'
import { getCurrentUser } from '@/utils/auth'
import { getErrorMessage, isBusinessError } from '@/utils/error-handler'
import type { BusinessError } from '@/api/request'

export const useReleaseStore = defineStore('release', () => {
  // ============ State ============
  const currentRelease = ref<ReleaseResponse | null>(null)
  const releaseHistory = ref<ReleaseResponse[]>([])
  const releaseEvents = ref<ReleaseEvent[]>([])

  const isCreatingRelease = ref(false)
  const isPolling = ref(false)
  const isLoadingHistory = ref(false)
  const isRollingBack = ref(false)

  const error = ref<string | null>(null)
  const errorCode = ref<number | null>(null)
  const pollingInterval = ref<NodeJS.Timeout | null>(null)

  // ============ Getters ============
  const progressPercentage = computed(() => {
    if (!releaseEvents.value || releaseEvents.value.length === 0) {
      return 0
    }

    // 根据事件类型计算进度
    const eventTypes = releaseEvents.value.map(e => e.type)
    const hasStarted = eventTypes.some(type => type === 'started')
    const hasValidating = eventTypes.some(type => type === 'validating')
    const hasDeploying = eventTypes.some(type => type === 'deploying')
    const hasSuccess = eventTypes.some(type => type === 'success')
    const hasFailed = eventTypes.some(type => type === 'failed')

    if (hasFailed) return 0
    if (hasSuccess) return 100
    if (hasDeploying) return 75
    if (hasValidating) return 50
    if (hasStarted) return 25
    return 10
  })

  /**
   * 格式化发布记录（关联应用、环境、集群名称）
   */
  const formatReleaseRecord = (release: ReleaseResponse): ReleaseRecord => {
    const appStore = useAppStore()
    const app = appStore.getApplicationById(release.app_id)
    const env = appStore.getEnvironmentById(release.env_id)
    const cluster = appStore.getClusterById(release.cluster_id)

    return {
      ...release,
      app_name: app?.name || `App #${release.app_id}`,
      env_name: env?.name || `Env #${release.env_id}`,
      cluster_name: cluster?.name || `Cluster #${release.cluster_id}`
    }
  }

  const currentReleaseFormatted = computed(() => {
    if (!currentRelease.value) return null
    return formatReleaseRecord(currentRelease.value)
  })

  const releaseHistoryFormatted = computed(() => {
    return releaseHistory.value.map(r => formatReleaseRecord(r))
  })

  // ============ Actions ============

  /**
   * 创建发布
   */
  const createRelease = async (data: {
    app_id: number
    env_id: number
    cluster_id: number
    image: string
  }) => {
    isCreatingRelease.value = true
    error.value = null
    errorCode.value = null
    try {
      const response = await releaseAPI.createRelease({
        ...data,
        user: getCurrentUser() // 从认证系统获取当前用户
      })
      currentRelease.value = response
      return response
    } catch (err) {
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
      throw err
    } finally {
      isCreatingRelease.value = false
    }
  }

  /**
   * 获取发布状态
   */
  const fetchReleaseStatus = async (releaseId: number) => {
    try {
      const response = await releaseAPI.getRelease(releaseId)
      currentRelease.value = response
      // 同时获取事件
      await fetchReleaseEvents(releaseId)
      return response
    } catch (err) {
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
      throw err
    }
  }

  /**
   * 获取发布事件
   */
  const fetchReleaseEvents = async (releaseId: number) => {
    try {
      const events = await releaseAPI.getReleaseEvents(releaseId)
      releaseEvents.value = events
      return events
    } catch (err) {
      // 事件获取失败不中断流程
      console.error('Failed to fetch release events:', err)
    }
  }

  /**
   * 获取发布历史
   */
  const fetchReleaseHistory = async (limit = 20, offset = 0) => {
    isLoadingHistory.value = true
    error.value = null
    errorCode.value = null
    try {
      const response = await releaseAPI.listReleases(limit, offset)
      releaseHistory.value = response.data
      return response
    } catch (err) {
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
      throw err
    } finally {
      isLoadingHistory.value = false
    }
  }

  /**
   * 启动轮询（实时追踪发布进度）
   */
  const startPolling = async (releaseId: number, interval = 2000) => {
    // 防止重复轮询
    if (isPolling.value) {
      return
    }

    isPolling.value = true

    const poll = async () => {
      try {
        const release = await fetchReleaseStatus(releaseId)
        // 如果发布完成（成功或失败），停止轮询
        if (['success', 'failed', 'rolled_back'].includes(release.status)) {
          stopPolling()
        }
      } catch (err) {
        console.error('Polling error:', err)
        // 轮询出错不中断，继续尝试
      }
    }

    // 立即执行一次
    await poll()

    // 然后按间隔轮询
    if (isPolling.value) {
      pollingInterval.value = setInterval(poll, interval)
    }
  }

  /**
   * 停止轮询
   */
  const stopPolling = () => {
    if (pollingInterval.value) {
      clearInterval(pollingInterval.value)
      pollingInterval.value = null
    }
    isPolling.value = false
  }

  /**
   * 回滚发布
   */
  const rollback = async (releaseId: number) => {
    isRollingBack.value = true
    error.value = null
    errorCode.value = null
    try {
      const response = await releaseAPI.rollbackRelease(releaseId)
      currentRelease.value = response
      releaseEvents.value = [] // 清空事件，开始追踪回滚进度
      // 启动轮询追踪回滚进度
      await startPolling(response.id)
      return response
    } catch (err) {
      const message = getErrorMessage(err)
      error.value = message
      if (isBusinessError(err)) {
        errorCode.value = (err as BusinessError).code
      }
      throw err
    } finally {
      isRollingBack.value = false
    }
  }

  /**
   * 清空当前发布信息
   */
  const clearCurrent = () => {
    currentRelease.value = null
    releaseEvents.value = []
    stopPolling()
  }

  return {
    // State
    currentRelease,
    releaseHistory,
    releaseEvents,
    isCreatingRelease,
    isPolling,
    isLoadingHistory,
    isRollingBack,
    error,
    errorCode,

    // Getters
    progressPercentage,
    currentReleaseFormatted,
    releaseHistoryFormatted,

    // Actions
    createRelease,
    fetchReleaseStatus,
    fetchReleaseEvents,
    fetchReleaseHistory,
    startPolling,
    stopPolling,
    rollback,
    clearCurrent
  }
})
