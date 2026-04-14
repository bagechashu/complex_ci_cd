/**
 * format.ts - 格式化工具函数
 */

import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

/**
 * 格式化日期时间
 */
export const formatDateTime = (
  timestamp: string | null | undefined,
  format = 'YYYY-MM-DD HH:mm:ss'
): string => {
  if (!timestamp) return '-'
  return dayjs(timestamp).format(format)
}

/**
 * 格式化为相对时间（如"5分钟前"）
 */
export const formatRelativeTime = (timestamp: string | null | undefined): string => {
  if (!timestamp) return '-'
  return dayjs(timestamp).fromNow()
}

/**
 * 计算时间差（毫秒）
 */
export const calculateDuration = (
  startTime: string | null | undefined,
  endTime: string | null | undefined
): number | null => {
  if (!startTime || !endTime) return null
  const start = dayjs(startTime)
  const end = dayjs(endTime)
  return end.diff(start, 'ms')
}

/**
 * 格式化时长
 */
export const formatDuration = (ms: number | null | undefined): string => {
  if (!ms) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${Math.round(ms / 1000)}s`
  if (ms < 3600000) return `${Math.round(ms / 60000)}m`
  return `${Math.round(ms / 3600000)}h`
}

/**
 * 获取状态的中文描述
 */
export const getStatusLabel = (status: string): string => {
  const statusMap: Record<string, string> = {
    pending: '待发布',
    validating: '验证中',
    deploying: '部署中',
    success: '成功',
    failed: '失败',
    rolled_back: '已回滚'
  }
  return statusMap[status] || status
}

/**
 * 获取事件类型的中文描述
 */
export const getEventTypeLabel = (type: string): string => {
  const typeMap: Record<string, string> = {
    started: '发布开始',
    validating: '配置验证',
    deploying: '部署开始',
    pod_updated: 'Pod 更新',
    pod_ready: 'Pod 就绪',
    success: '部署成功',
    failed: '部署失败',
    rolled_back: '回滚完成',
    error: '错误'
  }
  return typeMap[type] || type
}

/**
 * 截断字符串
 */
export const truncateString = (
  str: string | null | undefined,
  length: number = 50
): string => {
  if (!str) return '-'
  if (str.length <= length) return str
  return str.substring(0, length) + '...'
}

/**
 * 复制文本到剪贴板
 */
export const copyToClipboard = async (text: string): Promise<boolean> => {
  try {
    await navigator.clipboard.writeText(text)
    return true
  } catch (err) {
    console.error('Failed to copy:', err)
    return false
  }
}
