/**
 * releaseStore.ts - 发布流程和历史状态管理
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ReleaseResponse, ReleaseEvent, ReleaseStatus } from '@/types/api'
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
  const eventSource = ref<EventSource | null>(null)

  // ============ Getters ============
  const progressPercentage = computed(() => {
    if (!releaseEvents.value || releaseEvents.value.length === 0) {
      return 0
    }

    // 根据事件类型计算进度
    const eventTypes = releaseEvents.value.map(e => e.type)
    
    // Phase 2: 增强的进度计算逻辑
    if (eventTypes.includes('deployment_success')) return 100
    if (eventTypes.includes('deployment_failed') || eventTypes.includes('pod_error_detected') || eventTypes.includes('pod_timeout')) return 0
    if (eventTypes.includes('pod_ready')) return 85
    if (eventTypes.includes('pod_running')) return 60
    if (eventTypes.includes('pod_created')) return 40
    if (eventTypes.includes('deployment_started')) return 20
    if (eventTypes.includes('started')) return 25
    if (eventTypes.includes('validating')) return 50
    if (eventTypes.includes('deploying')) return 75
    if (eventTypes.includes('success')) return 100
    if (eventTypes.includes('failed')) return 0
    
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
      
      // Phase 2: 自动启动SSE流以实时跟踪部署进度
      // 这样用户就能看到Pod创建、错误检测等事件
      console.log('Starting SSE stream for new release', response.id)
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
      // 响应拦截器已经提取了 data 字段，response 直接是 ReleaseResponse[] 数组
      releaseHistory.value = Array.isArray(response) ? response : response.data || []
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
   * 启动SSE流（实时追踪发布进度）
   * 使用 Server-Sent Events 替代定期轮询，减少网络请求并提供实时反馈
   */
  const startPolling = async (releaseId: number) => {
    // 防止重复连接
    if (eventSource.value) {
      console.warn('SSE already connected')
      return
    }

    isPolling.value = true
    console.log('Starting SSE stream for release', releaseId)

    try {
      // 建立SSE连接
      eventSource.value = new EventSource(`/api/v1/releases/${releaseId}/stream`)

      // 处理事件流中的事件
      eventSource.value.onmessage = (event) => {
        try {
          const releaseEvent: ReleaseEvent = JSON.parse(event.data)
          console.log('Received release event:', releaseEvent.type, releaseEvent.message)
          
          // 将事件添加到列表
          if (!releaseEvents.value.some(e => e.id === releaseEvent.id)) {
            releaseEvents.value.push(releaseEvent)
          }

          // 更新当前发布状态
          if (currentRelease.value) {
            // Phase 2: 扩展的事件类型映射
            const eventToStatusMap: { [key: string]: ReleaseStatus } = {
              'success': 'success',
              'deployment_success': 'success',
              'failed': 'failed',
              'deployment_failed': 'failed',
              'rolled_back': 'rolled_back',
              'completed': 'success',
              // Phase 2新增事件 - 保持进行中状态
              'pod_created': 'in_progress',
              'pod_running': 'in_progress',
              'pod_ready': 'in_progress',
              'pod_error_detected': 'failed',
              'pod_timeout': 'failed',
              'deployment_started': 'in_progress',
            }

            if (eventToStatusMap[releaseEvent.type]) {
              currentRelease.value.status = eventToStatusMap[releaseEvent.type]
            }

            // 如果发布完成，关闭连接
            if (['success', 'failed', 'rolled_back'].includes(releaseEvent.type)) {
              console.log('Release completed:', releaseEvent.type)
              stopPolling()
            }
          }
        } catch (err) {
          console.error('Failed to parse SSE event:', err)
        }
      }

      // 处理错误和连接关闭
      eventSource.value.onerror = (err) => {
        console.error('SSE connection error:', err)
        stopPolling()
        // 如果连接中断且发布未完成，显示错误
        if (currentRelease.value && !['success', 'failed', 'rolled_back'].includes(currentRelease.value.status)) {
          error.value = 'Lost connection to release stream'
        }
      }

      // 首先加载现有事件
      try {
        const existingEvents = await fetchReleaseEvents(releaseId)
        if (existingEvents) {
          releaseEvents.value = existingEvents
        }
      } catch (err) {
        console.warn('Failed to fetch existing events:', err)
      }
    } catch (err) {
      console.error('Failed to start SSE stream:', err)
      error.value = 'Failed to start release stream'
      stopPolling()
    }
  }

  /**
   * 停止SSE连接
   */
  const stopPolling = () => {
    if (eventSource.value) {
      eventSource.value.close()
      eventSource.value = null
      console.log('SSE stream closed')
    }
    
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
      // Phase 3: 后端现在返回新创建的回滚部署记录
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
